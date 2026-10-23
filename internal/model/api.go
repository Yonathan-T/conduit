package model

type API struct {
	Package  string
	Services []Service
	Types    []Message
}

type Service struct {
	Name    string
	Methods []Method
}

type Method struct {
	Name       string
	HTTPMethod string
	Path       string
	Request    Message
	Response   Message
}

type Message struct {
	Name   string
	Fields []Field
}

type FieldLocation string

const (
	LocationPath  FieldLocation = "path"
	LocationQuery FieldLocation = "query"
	LocationBody  FieldLocation = "body"
)

type Field struct {
	Name     string
	Type     string
	Location FieldLocation
	Tag      int //for protobof field numbers
}
