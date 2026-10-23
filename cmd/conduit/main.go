package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"conduit/internal/compatibility"
	"conduit/internal/convert"
	"conduit/internal/parser"
	"conduit/internal/preprocess"
)

func writeToFile(dir string, filename string, content string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	targetPath := filepath.Join(dir, filename)
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", targetPath, err)
	}

	return nil
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func main() {
	specPath := flag.String("spec", "", "Path to the canonical API spec (YAML)")
	outDir := flag.String("out", "./generated", "Output directory for generated files")
	baselinePath := flag.String("baseline", "", "Path to baseline API spec to check for breaking changes")
	flag.Parse()

	if *specPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --spec flag is required")
		flag.Usage()
		os.Exit(1)
	}

	parsed, err := parser.ParseFile(*specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading spec from %s: %v\n", *specPath, err)
		os.Exit(1)
	}

	ir, err := preprocess.Run(parsed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: preprocessing spec: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully compiled to IR: service %q with %d method(s)\n", ir.Services[0].Name, len(ir.Services[0].Methods))

	if *baselinePath != "" {
		baselineParsed, err := parser.ParseFile(*baselinePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading baseline: %v\n", err)
			os.Exit(1)
		}

		baselineIR, err := preprocess.Run(baselineParsed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error preprocessing baseline: %v\n", err)
			os.Exit(1)
		}

		report := compatibility.Check(baselineIR, ir)
		if len(report.Changes) > 0 {
			fmt.Println("--- Compatibility Report ---")
			fmt.Println(report.String())
			fmt.Println("----------------------------")
		}

		if report.HasBreaking() {
			fmt.Fprintln(os.Stderr, "Generation rejected: breaking changes detected against baseline.")
			os.Exit(2)
		}
	}

	protoContent, err := convert.GenerateProto(ir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating proto: %v\n", err)
		os.Exit(1)
	}

	restContent, err := convert.GenerateREST(ir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating REST: %v\n", err)
		os.Exit(1)
	}

	// Follow official Protobuf directory structure matching package name (e.g. catalog.v1 -> catalog/v1/)
	targetDir := *outDir
	if ir.Package != "" {
		parts := strings.Split(ir.Package, ".")
		targetDir = filepath.Join(append([]string{*outDir}, parts...)...)
	}

	// File name in lower_snake_case (e.g. ProductService -> product_service.proto)
	baseName := "service"
	if len(ir.Services) > 0 && ir.Services[0].Name != "" {
		baseName = toSnakeCase(ir.Services[0].Name)
	}

	protoFilename := fmt.Sprintf("%s.proto", baseName)
	if err := writeToFile(targetDir, protoFilename, protoContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing proto file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Protobuf Generated in: %s\n", filepath.Join(targetDir, protoFilename))

	restFilename := fmt.Sprintf("%s_routes.json", baseName)
	if err := writeToFile(targetDir, restFilename, restContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing REST routes: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("REST Generated in: %s\n", filepath.Join(targetDir, restFilename))
}
