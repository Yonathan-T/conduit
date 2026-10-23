package preprocess

import (
	"strings"
	"testing"

	"conduit/internal/model"
	"conduit/internal/parser"
)

func TestPreprocess_Valid(t *testing.T) {
	spec := &parser.Spec{
		Service: parser.Service{
			Name:    "product_service",
			Package: "catalog.v1",
		},
		Methods: []parser.Method{
			{
				Name: "search_products",
				HTTP: parser.HTTPDefinition{
					Method: "get",
					Path:   "/v1/products/{category}",
				},
				Request: parser.Message{
					Fields: []parser.Field{
						{Name: "category", Type: "string", Location: "path"},
						{Name: "limit", Type: "int32", Location: "query"},
					},
				},
				Response: parser.Message{
					Fields: []parser.Field{
						{Name: "items", Type: "Product[]"},
					},
				},
			},
			{
				Name: "delete_product",
				HTTP: parser.HTTPDefinition{
					Method: "delete",
					Path:   "/v1/products/{id}",
				},
				Request: parser.Message{
					Fields: []parser.Field{
						{Name: "id", Type: "string", Location: "path"},
					},
				},
				Response: parser.Message{
					Fields: []parser.Field{
						{Name: "success", Type: "bool"},
					},
				},
			},
			{
				Name: "update_product",
				HTTP: parser.HTTPDefinition{
					Method: "put",
					Path:   "/v1/products/{id}",
				},
				Request: parser.Message{
					Fields: []parser.Field{
						{Name: "id", Type: "string", Location: "path"},
					},
				},
				Response: parser.Message{
					Fields: []parser.Field{
						{Name: "success", Type: "bool"},
					},
				},
			},
		},
	}

	api, err := Run(spec)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(api.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(api.Services))
	}
	svc := api.Services[0]
	if svc.Name != "ProductService" {
		t.Errorf("expected PascalCase 'ProductService', got %q", svc.Name)
	}

	if len(svc.Methods) != 3 {
		t.Fatalf("expected 3 methods, got %d", len(svc.Methods))
	}
	m := svc.Methods[0]
	if m.Name != "SearchProducts" {
		t.Errorf("expected PascalCase 'SearchProducts', got %q", m.Name)
	}
	if m.HTTPMethod != "GET" {
		t.Errorf("expected uppercase 'GET', got %q", m.HTTPMethod)
	}
	if svc.Methods[1].HTTPMethod != "DELETE" {
		t.Errorf("expected uppercase 'DELETE', got %q", svc.Methods[1].HTTPMethod)
	}
	if svc.Methods[2].HTTPMethod != "PUT" {
		t.Errorf("expected uppercase 'PUT', got %q", svc.Methods[2].HTTPMethod)
	}

	if m.Request.Name != "SearchProductsRequest" {
		t.Errorf("expected 'SearchProductsRequest', got %q", m.Request.Name)
	}
	if len(m.Request.Fields) != 2 {
		t.Fatalf("expected 2 request fields, got %d", len(m.Request.Fields))
	}
	if m.Request.Fields[0].Tag != 1 || m.Request.Fields[0].Location != model.LocationPath {
		t.Errorf("unexpected field 0: %+v", m.Request.Fields[0])
	}
	if m.Request.Fields[1].Tag != 2 || m.Request.Fields[1].Location != model.LocationQuery {
		t.Errorf("unexpected field 1: %+v", m.Request.Fields[1])
	}

	if m.Response.Name != "SearchProductsResponse" {
		t.Errorf("expected 'SearchProductsResponse', got %q", m.Response.Name)
	}
	if len(m.Response.Fields) != 1 || m.Response.Fields[0].Tag != 1 {
		t.Errorf("unexpected response fields: %+v", m.Response.Fields)
	}
}

func TestPreprocess_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		spec        *parser.Spec
		expectError string
	}{
		{
			name: "unsupported HTTP method",
			spec: &parser.Spec{
				Service: parser.Service{Name: "Svc"},
				Methods: []parser.Method{
					{
						Name: "TraceUser",
						HTTP: parser.HTTPDefinition{Method: "TRACE", Path: "/v1/users"},
					},
				},
			},
			expectError: "unsupported HTTP method",
		},
		{
			name: "duplicate method names",
			spec: &parser.Spec{
				Service: parser.Service{Name: "Svc"},
				Methods: []parser.Method{
					{Name: "GetUser", HTTP: parser.HTTPDefinition{Method: "GET", Path: "/v1/users/1"}},
					{Name: "GetUser", HTTP: parser.HTTPDefinition{Method: "GET", Path: "/v1/users/2"}},
				},
			},
			expectError: "duplicate method name",
		},
		{
			name: "missing path placeholder field",
			spec: &parser.Spec{
				Service: parser.Service{Name: "Svc"},
				Methods: []parser.Method{
					{
						Name: "GetUser",
						HTTP: parser.HTTPDefinition{Method: "GET", Path: "/v1/users/{user_id}"},
						Request: parser.Message{
							Fields: []parser.Field{
								{Name: "other_field", Type: "string", Location: "query"},
							},
						},
					},
				},
			},
			expectError: "route placeholder \"user_id\" has no matching path field",
		},
		{
			name: "multi-dimensional array rejected",
			spec: &parser.Spec{
				Service: parser.Service{Name: "Svc"},
				Methods: []parser.Method{
					{
						Name: "GetMatrix",
						HTTP: parser.HTTPDefinition{Method: "GET", Path: "/v1/matrix"},
						Response: parser.Message{
							Fields: []parser.Field{
								{Name: "grid", Type: "string[][]"},
							},
						},
					},
				},
			},
			expectError: "multi-dimensional arrays are not supported",
		},
		{
			name: "duplicate field in request",
			spec: &parser.Spec{
				Service: parser.Service{Name: "Svc"},
				Methods: []parser.Method{
					{
						Name: "Search",
						HTTP: parser.HTTPDefinition{Method: "GET", Path: "/v1/search"},
						Request: parser.Message{
							Fields: []parser.Field{
								{Name: "query", Type: "string", Location: "query"},
								{Name: "query", Type: "string", Location: "query"},
							},
						},
					},
				},
			},
			expectError: "duplicate field \"query\" in request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(tt.spec)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectError)
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.expectError)) {
				t.Errorf("expected error to contain %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}
