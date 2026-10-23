package convert

import (
	"encoding/json"
	"errors"

	"conduit/internal/model"
)

type RouteMetadata struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

func GenerateREST(api *model.API) (string, error) {
	routes := make([]RouteMetadata, 0)
	if api == nil {
		return "", errors.New("api is nil")
	}

	for _, svc := range api.Services {
		for _, m := range svc.Methods {
			routes = append(routes, RouteMetadata{
				Name:   m.Name,
				Method: m.HTTPMethod,
				Path:   m.Path,
			})
		}
	}

	data, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
