package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile_Valid(t *testing.T) {
	validYAML := `
service:
  name: CatalogService
  package: catalog.v1
methods:
  - name: ListItems
    http:
      method: GET
      path: /v1/items
    request:
      fields:
        - name: page_size
          type: int32
          location: query
    response:
      fields:
        - name: items
          type: string[]
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(filePath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	spec, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if spec.Service.Name != "CatalogService" {
		t.Errorf("expected service name 'CatalogService', got %q", spec.Service.Name)
	}
	if spec.Service.Package != "catalog.v1" {
		t.Errorf("expected package 'catalog.v1', got %q", spec.Service.Package)
	}
	if len(spec.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(spec.Methods))
	}

	m := spec.Methods[0]
	if m.Name != "ListItems" {
		t.Errorf("expected method name 'ListItems', got %q", m.Name)
	}
	if m.HTTP.Method != "GET" || m.HTTP.Path != "/v1/items" {
		t.Errorf("expected GET /v1/items, got %s %s", m.HTTP.Method, m.HTTP.Path)
	}
	if len(m.Request.Fields) != 1 || m.Request.Fields[0].Name != "page_size" {
		t.Errorf("unexpected request fields: %+v", m.Request.Fields)
	}
	if len(m.Response.Fields) != 1 || m.Response.Fields[0].Name != "items" {
		t.Errorf("unexpected response fields: %+v", m.Response.Fields)
	}
}

func TestParseFile_Errors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError string
	}{
		{
			name: "malformed syntax",
			content: `
				service:
					 name: BadYAML
					 indentation_is_broken:
					`,
			expectError: "yaml",
		},
		{
			name: "unknown field rejected by KnownFields",
			content: `
service:
  name: CatalogService
  package: catalog.v1
  unexpected_field: true
methods: []
`,
			expectError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			_, err := ParseFile(filePath)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectError)
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.expectError)) {
				t.Errorf("expected error to contain %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func TestParseFile_NonExistent(t *testing.T) {
	_, err := ParseFile("does_not_exist.yaml")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}
