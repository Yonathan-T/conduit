package preprocess

import (
	"errors"
	"strings"

	"conduit/internal/model"
	"conduit/internal/parser"
)

func Normalize(spec *parser.Spec) (*model.API, error) {
	if spec == nil {
		return nil, errors.New("spec is nil")
	}

	var methods []model.Method

	for _, m := range spec.Methods {
		methodName := toPascalCase(m.Name)

		methods = append(methods, model.Method{
			Name:       methodName,
			HTTPMethod: strings.ToUpper(strings.TrimSpace(m.HTTP.Method)),
			Path:       strings.TrimSpace(m.HTTP.Path),
			Request:    normalizeMessages(methodName+"Request", m.Request.Fields),
			Response:   normalizeMessages(methodName+"Response", m.Response.Fields),
		})
	}

	var types []model.Message
	for _, t := range spec.Types {
		typeName := toPascalCase(t.Name)
		types = append(types, normalizeMessages(typeName, t.Fields))
	}

	api := &model.API{
		Package: strings.TrimSpace(spec.Service.Package),
		Services: []model.Service{
			{
				Name:    toPascalCase(spec.Service.Name),
				Methods: methods,
			},
		},
		Types: types,
	}
	return api, nil
}

func normalizeMessages(name string, rawFields []parser.Field) model.Message {
	var fields []model.Field
	for i, f := range rawFields {
		fields = append(fields, model.Field{
			Name:     strings.TrimSpace(f.Name),
			Type:     strings.TrimSpace(f.Type),
			Location: model.FieldLocation(strings.ToLower(strings.TrimSpace(f.Location))),
			Tag:      i + 1,
		})
	}
	return model.Message{
		Name:   name,
		Fields: fields,
	}
}

func toPascalCase(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return ""
	}
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			b.WriteString(strings.ToUpper(part[:1]))
			b.WriteString(part[1:])
		}
	}
	return b.String()
}
