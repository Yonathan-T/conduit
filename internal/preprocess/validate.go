package preprocess

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"conduit/internal/model"
)

var (
	primitiveTypes = map[string]bool{
		"string": true,
		"int32":  true,
		"int64":  true,
		"bool":   true,
		"float":  true,
		"double": true,
	}

	placeholderRegex = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)
)

func Validate(api *model.API) error {
	if api == nil {
		return errors.New("api is nil")
	}

	var errs []error

	seenTypes := make(map[string]bool)
	for _, t := range api.Types {
		if seenTypes[t.Name] {
			errs = append(errs, fmt.Errorf("duplicate type name %q", t.Name))
		}
		seenTypes[t.Name] = true

		seenTypeFields := make(map[string]bool)
		for _, f := range t.Fields {
			if seenTypeFields[f.Name] {
				errs = append(errs, fmt.Errorf("type %q: duplicate field %q", t.Name, f.Name))
			}
			seenTypeFields[f.Name] = true

			if strings.HasSuffix(f.Type, "[][]") {
				errs = append(errs, fmt.Errorf("type %q: field %q has nested array %q (multi-dimensional arrays are not supported in Protobuf)", t.Name, f.Name, f.Type))
			} else if !isValidType(f.Type) {
				errs = append(errs, fmt.Errorf("type %q: field %q has unsupported type %q", t.Name, f.Name, f.Type))
			}
		}
	}

	for _, s := range api.Services {
		seenMethods := make(map[string]bool)

		for _, m := range s.Methods {
			if seenMethods[m.Name] {
				errs = append(errs, fmt.Errorf("service %q: duplicate method name %q", s.Name, m.Name))
			}
			seenMethods[m.Name] = true

			placeholders := extractPlaceholders(m.Path)
			for _, placeholder := range placeholders {
				found := false
				for _, f := range m.Request.Fields {
					if f.Name == placeholder && f.Location == model.LocationPath {
						found = true
						break
					}
				}
				if !found {
					errs = append(errs, fmt.Errorf("method %q: route placeholder %q has no matching path field in request", m.Name, placeholder))
				}
			}

			seenReqFields := make(map[string]bool)
			for _, f := range m.Request.Fields {
				if seenReqFields[f.Name] {
					errs = append(errs, fmt.Errorf("method %q: duplicate field %q in request", m.Name, f.Name))
				}
				seenReqFields[f.Name] = true

				if strings.HasSuffix(f.Type, "[][]") {
					errs = append(errs, fmt.Errorf("method %q: field %q has nested array %q (multi-dimensional arrays are not supported in Protobuf)", m.Name, f.Name, f.Type))
				} else if !isValidType(f.Type) {
					errs = append(errs, fmt.Errorf("method %q: field %q has unsupported type %q", m.Name, f.Name, f.Type))
				}
			}

			seenRespFields := make(map[string]bool)
			for _, f := range m.Response.Fields {
				if seenRespFields[f.Name] {
					errs = append(errs, fmt.Errorf("method %q: duplicate field %q in response", m.Name, f.Name))
				}
				seenRespFields[f.Name] = true

				if strings.HasSuffix(f.Type, "[][]") {
					errs = append(errs, fmt.Errorf("method %q: field %q has nested array %q (multi-dimensional arrays are not supported in Protobuf)", m.Name, f.Name, f.Type))
				} else if !isValidType(f.Type) {
					errs = append(errs, fmt.Errorf("method %q: field %q has unsupported type %q", m.Name, f.Name, f.Type))
				}
			}
		}
	}

	return errors.Join(errs...)
}

func extractPlaceholders(path string) []string {
	var placeholders []string
	matches := placeholderRegex.FindAllStringSubmatch(path, -1)
	for _, m := range matches {
		placeholders = append(placeholders, m[1])
	}
	return placeholders
}

func isValidType(rawType string) bool {
	base := strings.TrimSuffix(rawType, "[]")
	if base == "" {
		return false
	}

	if primitiveTypes[base] {
		return true
	}

	return isCustomMessageName(base)
}

func isCustomMessageName(name string) bool {
	if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
		return false
	}

	for i := 1; i < len(name); i++ {
		if !isIdentChar(name[i]) {
			return false
		}
	}

	return true
}

func isIdentChar(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_':
		return true
	default:
		return false
	}
}
