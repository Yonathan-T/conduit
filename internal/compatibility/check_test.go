package compatibility

import (
	"strings"
	"testing"

	"conduit/internal/model"
)

func baseAPI() *model.API {
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
								{Name: "query", Type: "string", Tag: 1},
								{Name: "limit", Type: "int32", Tag: 2},
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

func TestCheck_AddNewField_Allowed(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := baseAPI()

	// Add an optional new field with a new tag number
	newAPI.Services[0].Methods[0].Request.Fields = append(
		newAPI.Services[0].Methods[0].Request.Fields,
		model.Field{Name: "page_token", Type: "string", Tag: 3},
	)

	report := Check(oldAPI, newAPI)
	if report.HasBreaking() {
		t.Fatalf("expected no breaking changes when adding a field, got: %s", report.String())
	}
	if len(report.Changes) != 0 {
		t.Fatalf("expected 0 changes for compatible addition, got: %d", len(report.Changes))
	}
}

func TestCheck_RenamedField_Warning(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := baseAPI()

	// Rename "query" to "search_query" while keeping Tag 1 and Type "string"
	newAPI.Services[0].Methods[0].Request.Fields[0] = model.Field{
		Name: "search_query",
		Type: "string",
		Tag:  1,
	}

	report := Check(oldAPI, newAPI)
	if report.HasBreaking() {
		t.Fatalf("renaming field with same tag/type should be WARNING, not BREAKING: %s", report.String())
	}
	if !report.HasWarnings() {
		t.Fatalf("expected WARNING when renaming a field, got none: %s", report.String())
	}
	if !strings.Contains(report.String(), "was renamed to") {
		t.Errorf("expected report to mention rename, got: %s", report.String())
	}
}

func TestCheck_RenumberField_Breaking(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := baseAPI()

	// Change "query" tag from 1 to 5
	newAPI.Services[0].Methods[0].Request.Fields[0].Tag = 5

	report := Check(oldAPI, newAPI)
	if !report.HasBreaking() {
		t.Fatalf("expected BREAKING change when renumbering field tag, got none: %s", report.String())
	}
	if !strings.Contains(report.String(), "changed tag number 1 -> 5") {
		t.Errorf("unexpected report output: %s", report.String())
	}
}

func TestCheck_ChangeFieldType_Breaking(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := baseAPI()

	// Change "query" type from "string" to "int32"
	newAPI.Services[0].Methods[0].Request.Fields[0].Type = "int32"

	report := Check(oldAPI, newAPI)
	if !report.HasBreaking() {
		t.Fatalf("expected BREAKING change when changing field type, got none: %s", report.String())
	}
	if !strings.Contains(report.String(), "changed type \"string\" -> \"int32\"") {
		t.Errorf("unexpected report output: %s", report.String())
	}
}

func TestCheck_RemoveField_Breaking(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := baseAPI()

	// Remove "limit" field
	newAPI.Services[0].Methods[0].Request.Fields = newAPI.Services[0].Methods[0].Request.Fields[:1]

	report := Check(oldAPI, newAPI)
	if !report.HasBreaking() {
		t.Fatalf("expected BREAKING change when deleting a field, got none: %s", report.String())
	}
	if !strings.Contains(report.String(), "field \"limit\" (tag 2) was removed") {
		t.Errorf("unexpected report output: %s", report.String())
	}
}

func TestCheck_TagReusedDifferentType_Breaking(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := baseAPI()

	// Delete "query" (tag 1, string) and reuse Tag 1 for "user_id" (tag 1, int64)
	newAPI.Services[0].Methods[0].Request.Fields[0] = model.Field{
		Name: "user_id",
		Type: "int64",
		Tag:  1,
	}

	report := Check(oldAPI, newAPI)
	if !report.HasBreaking() {
		t.Fatalf("expected BREAKING change when tag is reused for different type, got none: %s", report.String())
	}
	if !strings.Contains(report.String(), "deleted tag 1") {
		t.Errorf("unexpected report output: %s", report.String())
	}
}

func TestCheck_RemoveRPCMethod_Breaking(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := baseAPI()

	// Remove SearchProducts RPC
	newAPI.Services[0].Methods = nil

	report := Check(oldAPI, newAPI)
	if !report.HasBreaking() {
		t.Fatalf("expected BREAKING change when removing an RPC, got none: %s", report.String())
	}
	if !strings.Contains(report.String(), "RPC method was removed") {
		t.Errorf("unexpected report output: %s", report.String())
	}
}

func TestCheck_RemoveService_Breaking(t *testing.T) {
	oldAPI := baseAPI()
	newAPI := &model.API{Package: "catalog.v1"}

	report := Check(oldAPI, newAPI)
	if !report.HasBreaking() {
		t.Fatalf("expected BREAKING change when removing a service, got none: %s", report.String())
	}
	if !strings.Contains(report.String(), "service was removed") {
		t.Errorf("unexpected report output: %s", report.String())
	}
}
