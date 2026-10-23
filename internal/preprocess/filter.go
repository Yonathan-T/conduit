package preprocess

import (
	"conduit/internal/parser"
	"fmt"
	"strings"
)

var allowedMethods = map[string]struct{}{
	"GET":    {},
	"POST":   {},
	"PUT":    {},
	"PATCH":  {},
	"DELETE": {},
}

func Filter(spec *parser.Spec) error {
	for _, m := range spec.Methods {
		method := strings.ToUpper(strings.TrimSpace(m.HTTP.Method))

		if _, ok := allowedMethods[method]; !ok {
			return fmt.Errorf("method %q: invalid or unsupported HTTP method %q",
				m.Name, method)
		}
	}

	return nil
}
