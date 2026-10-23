package parser

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Spec struct {
	Service Service  `yaml:"service"`
	Types   []Type   `yaml:"types,omitempty"`
	Methods []Method `yaml:"methods"`
}

type Type struct {
	Name   string  `yaml:"name"`
	Fields []Field `yaml:"fields"`
}

type Service struct {
	Name    string `yaml:"name"`
	Package string `yaml:"package"`
}

type Method struct {
	Name     string         `yaml:"name"`
	HTTP     HTTPDefinition `yaml:"http"`
	Request  Message        `yaml:"request"`
	Response Message        `yaml:"response"`
}

type HTTPDefinition struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type Message struct {
	Fields []Field `yaml:"fields"`
}

type Field struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Location string `yaml:"location"`
}

func ParseFile(path string) (*Spec, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := yaml.NewDecoder(file)
	dec.KnownFields(true)

	var spec Spec
	if err := dec.Decode(&spec); err != nil {
		return nil, err
	}
	return &spec, nil
}
