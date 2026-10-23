package convert

import (
	"encoding/json"
	"strings"
	"testing"

	"conduit/internal/model"
)

func sampleAPI() *model.API {
	return &model.API{
		Package: "catalog.v1",
		Services: []model.Service{
			{
				Name: "ProductService",
				Methods: []model.Method{
					{
						Name:       "SearchProducts",
						HTTPMethod: "GET",
						Path:       "/v1/products",
						Request: model.Message{
							Name: "SearchProductsRequest",
							Fields: []model.Field{
								{Name: "query", Type: "string", Location: model.LocationQuery, Tag: 1},
								{Name: "limit", Type: "int32", Location: model.LocationQuery, Tag: 2},
							},
						},
						Response: model.Message{
							Name: "SearchProductsResponse",
							Fields: []model.Field{
								{Name: "products", Type: "Product[]", Tag: 1},
							},
						},
					},
				},
			},
		},
	}
}

func TestGenerateProto_Valid(t *testing.T) {
	api := sampleAPI()

	output, err := GenerateProto(api)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedSnippets := []string{
		`syntax = "proto3";`,
		`package catalog.v1;`,
		`service ProductService {`,
		`rpc SearchProducts (SearchProductsRequest) returns (SearchProductsResponse);`,
		`message SearchProductsRequest {`,
		`string query = 1;`,
		`int32 limit = 2;`,
		`message SearchProductsResponse {`,
		`repeated Product products = 1;`,
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(output, snippet) {
			t.Errorf("expected proto output to contain %q, but it was missing.\nOutput:\n%s", snippet, output)
		}
	}
}

func TestGenerateProto_Nil(t *testing.T) {
	_, err := GenerateProto(nil)
	if err == nil {
		t.Fatal("expected error for nil API, got nil")
	}
}

func TestGenerateProto_DeduplicateMessages(t *testing.T) {
	commonMsg := model.Message{
		Name: "Empty",
	}

	api := &model.API{
		Package: "test.v1",
		Services: []model.Service{
			{
				Name: "TestService",
				Methods: []model.Method{
					{
						Name:     "MethodA",
						Request:  commonMsg,
						Response: commonMsg,
					},
					{
						Name:     "MethodB",
						Request:  commonMsg,
						Response: commonMsg,
					},
				},
			},
		},
	}

	output, err := GenerateProto(api)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	count := strings.Count(output, "message Empty {")
	if count != 1 {
		t.Errorf("expected 'message Empty {' to appear exactly once, appeared %d times", count)
	}
}

func TestGenerateREST_Valid(t *testing.T) {
	api := sampleAPI()

	output, err := GenerateREST(api)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var routes []RouteMetadata
	if err := json.Unmarshal([]byte(output), &routes); err != nil {
		t.Fatalf("expected valid JSON output, got unmarshal error: %v\nJSON:\n%s", err, output)
	}

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	r := routes[0]
	if r.Name != "SearchProducts" {
		t.Errorf("expected route name 'SearchProducts', got %q", r.Name)
	}
	if r.Method != "GET" {
		t.Errorf("expected method 'GET', got %q", r.Method)
	}
	if r.Path != "/v1/products" {
		t.Errorf("expected path '/v1/products', got %q", r.Path)
	}
}

func TestGenerateREST_Nil(t *testing.T) {
	_, err := GenerateREST(nil)
	if err == nil {
		t.Fatal("expected error for nil API, got nil")
	}
}

func TestGenerateREST_EmptyServices(t *testing.T) {
	api := &model.API{
		Package:  "empty.v1",
		Services: []model.Service{},
	}

	output, err := GenerateREST(api)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if strings.TrimSpace(output) != "[]" {
		t.Errorf("expected '[]', got %q", output)
	}
}
