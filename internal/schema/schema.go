package schema

// the base type
type SchemaProperty struct {
	Id                 string   `json:"id,omitempty"`
	Name               string   `json:"name,omitempty"`
	Ref                string   `json:"ref,omitempty"`
	Extend             string   `json:"extend,omitempty"`
	Import             string   `json:"import,omitempty"`
	Description        string   `json:"description,omitempty"`
	Optional           bool     `json:"optional,omitempty"`
	Unsupported        bool     `json:"unsupported,omitempty"`
	Deprecated         bool     `json:"deprecated,omitempty"`
	Permissions        []string `json:"permissions,omitempty"`
	AllowedContexts    []string `json:"allowedContexts,omitempty"`
	OnError            string   `json:"onError,omitempty"`
	MinManifestVersion int      `json:"minManifestVersion,omitempty"`
	MaxManifestVersion int      `json:"maxManifestVersion,omitempty"`
}

func (_ SchemaProperty) Type() string {
	return "property"
}

func (this SchemaProperty) Base() SchemaProperty {
	return this
}

// if there is a "choices" property
type SchemaChoicesProperty struct {
	SchemaProperty
	Choices []SchemaItem `json:"choices,omitempty"`
	Default any          `json:"default,omitempty"`
}

func (_ SchemaChoicesProperty) Type() string {
	return "choices"
}

type SchemaAnyProperty struct {
	SchemaProperty
}

func (_ SchemaAnyProperty) Type() string {
	return "any"
}

type SchemaRefProperty struct {
	SchemaProperty
}

func (_ SchemaRefProperty) Type() string {
	return "ref"
}

type SchemaNullProperty struct {
	SchemaProperty
}

func (_ SchemaNullProperty) Type() string {
	return "null"
}

type SchemaValueProperty struct {
	SchemaProperty
	Value any `json:"value,omitempty"`
}

func (_ SchemaValueProperty) Type() string {
	return "value"
}

type SchemaBoolProperty struct {
	SchemaProperty
	Default bool `json:"default,omitempty"`
}

func (_ SchemaBoolProperty) Type() string {
	return "bool"
}

type SchemaIntProperty struct {
	SchemaProperty
	Minimum int `json:"minimum,omitempty"`
	Maximum int `json:"maximum,omitempty"`
	Default int `json:"default,omitempty"`
}

func (_ SchemaIntProperty) Type() string {
	return "integer"
}

type SchemaFloatProperty struct {
	SchemaProperty
	Minimum float64 `json:"minimum,omitempty"`
	Maximum float64 `json:"maximum,omitempty"`
	Default float64 `json:"default,omitempty"`
}

func (_ SchemaFloatProperty) Type() string {
	return "float64"
}

type SchemaArrayProperty struct {
	SchemaProperty
	Items   SchemaItem `json:"items,omitempty"`
	Default any        `json:"default,omitempty"`
}

func (_ SchemaArrayProperty) Type() string {
	return "array"
}

type SchemaEnumValue struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type SchemaStringProperty struct {
	SchemaProperty
	Enum      []SchemaEnumValue `json:"enum,omitempty"`
	MinLength int               `json:"minLength,omitempty"`
	MaxLength int               `json:"maxLength,omitempty"`
	Pattern   string            `json:"pattern,omitempty"`
	Format    string            `json:"format,omitempty"`
	Default   string            `json:"default,omitempty"`
}

func (_ SchemaStringProperty) Type() string {
	return "string"
}

type SchemaObjectProperty struct {
	SchemaProperty
	Properties           []SchemaItem              `json:"properties,omitempty"`
	AdditionalProperties SchemaItem                `json:"additionalProperties,omitempty"`
	PatternProperties    []SchemaItem              `json:"patternProperties,omitempty"`
	IsInstanceOf         string                    `json:"isInstanceOf,omitempty"`
	Functions            []*SchemaFunctionProperty `json:"functions,omitempty"`
	Events               []*SchemaFunctionProperty `json:"events,omitempty"`
	Default              any                       `json:"default,omitempty"`
}

func (_ SchemaObjectProperty) Type() string {
	return "object"
}

type SchemaFunctionProperty struct {
	SchemaProperty
	Async                           bool         `json:"async,omitempty"`
	RequireUserInput                bool         `json:"requireUserInput,omitempty"`
	Parameters                      []SchemaItem `json:"parameters,omitempty"`
	ExtraParameters                 []SchemaItem `json:"extraParameters,omitempty"`
	Returns                         SchemaItem   `json:"returns,omitempty"`
	Filters                         []SchemaItem `json:"filters,omitempty"`
	AllowAmbiguousOptionalArguments bool         `json:"allowAmbiguousOptionalArguments,omitempty"`
	AllowCrossOriginArguments       bool         `json:"allowCrossOriginArguments,omitempty"`
}

func (_ SchemaFunctionProperty) Type() string {
	return "function"
}

// a namespace will map to a file
type SchemaNamespace struct {
	SchemaProperty
	Properties      []SchemaItem              `json:"properties,omitempty"`
	Types           []SchemaItem              `json:"types,omitempty"`
	Functions       []*SchemaFunctionProperty `json:"functions,omitempty"`
	Events          []*SchemaFunctionProperty `json:"events,omitempty"`
	DefaultContexts []string                  `json:"defaultContexts,omitempty"`
	NoCompile       bool                      `json:"noCompile,omitempty"`
	Import          string                    `json:"import,omitempty"`
}

func (_ SchemaNamespace) Type() string {
	return "namespace"
}
