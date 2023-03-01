package schema

type Item interface {
	Type() string
	Base() Property
}

// the base type
type Property struct {
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

func (_ Property) Type() string {
	return "property"
}

func (this Property) Base() Property {
	return this
}

// if there is a "choices" property
type Choices struct {
	Property
	Choices []Item `json:"choices,omitempty"`
	Default any    `json:"default,omitempty"`
}

func (_ Choices) Type() string {
	return "choices"
}

type Any struct {
	Property
}

func (_ Any) Type() string {
	return "any"
}

type Ref struct {
	Property
}

func (_ Ref) Type() string {
	return "ref"
}

type Null struct {
	Property
}

func (_ Null) Type() string {
	return "null"
}

type Value struct {
	Property
	Value any `json:"value,omitempty"`
}

func (_ Value) Type() string {
	return "value"
}

type Bool struct {
	Property
	Default bool `json:"default,omitempty"`
}

func (_ Bool) Type() string {
	return "bool"
}

type Int struct {
	Property
	Minimum int `json:"minimum,omitempty"`
	Maximum int `json:"maximum,omitempty"`
	Default int `json:"default,omitempty"`
}

func (_ Int) Type() string {
	return "integer"
}

type Number struct {
	Property
	Minimum float64 `json:"minimum,omitempty"`
	Maximum float64 `json:"maximum,omitempty"`
	Default float64 `json:"default,omitempty"`
}

func (_ Number) Type() string {
	return "float64"
}

type Array struct {
	Property
	Items   Item `json:"items,omitempty"`
	Default any  `json:"default,omitempty"`
}

func (_ Array) Type() string {
	return "array"
}

type EnumValue struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type String struct {
	Property
	Enum      []EnumValue `json:"enum,omitempty"`
	MinLength int         `json:"minLength,omitempty"`
	MaxLength int         `json:"maxLength,omitempty"`
	Pattern   string      `json:"pattern,omitempty"`
	Format    string      `json:"format,omitempty"`
	Default   string      `json:"default,omitempty"`
}

func (_ String) Type() string {
	return "string"
}

type Object struct {
	Property
	Properties           []Item      `json:"properties,omitempty"`
	AdditionalProperties Item        `json:"additionalProperties,omitempty"`
	PatternProperties    []Item      `json:"patternProperties,omitempty"`
	IsInstanceOf         string      `json:"isInstanceOf,omitempty"`
	Functions            []*Function `json:"functions,omitempty"`
	Events               []*Function `json:"events,omitempty"`
	Default              any         `json:"default,omitempty"`
}

func (_ Object) Type() string {
	return "object"
}

type Function struct {
	Property
	Async                           bool   `json:"async,omitempty"`
	RequireUserInput                bool   `json:"requireUserInput,omitempty"`
	Parameters                      []Item `json:"parameters,omitempty"`
	ExtraParameters                 []Item `json:"extraParameters,omitempty"`
	Returns                         Item   `json:"returns,omitempty"`
	Filters                         []Item `json:"filters,omitempty"`
	AllowAmbiguousOptionalArguments bool   `json:"allowAmbiguousOptionalArguments,omitempty"`
	AllowCrossOriginArguments       bool   `json:"allowCrossOriginArguments,omitempty"`
}

func (_ Function) Type() string {
	return "function"
}

// a namespace will map to a file
type Namespace struct {
	Property
	Properties      []Item      `json:"properties,omitempty"`
	Types           []Item      `json:"types,omitempty"`
	Functions       []*Function `json:"functions,omitempty"`
	Events          []*Function `json:"events,omitempty"`
	DefaultContexts []string    `json:"defaultContexts,omitempty"`
	NoCompile       bool        `json:"noCompile,omitempty"`
	Import          string      `json:"import,omitempty"`
}

func (_ Namespace) Type() string {
	return "namespace"
}
