package preprocess

import (
	"conduit/internal/model"
	"conduit/internal/parser"
)

func Run(spec *parser.Spec) (*model.API, error) {
	if err := Filter(spec); err != nil {
		return nil, err
	}

	api, err := Normalize(spec)
	if err != nil {
		return nil, err
	}

	if err := Validate(api); err != nil {
		return nil, err
	}

	return api, nil
}
