package generate

import (
	"fmt"
	_ "log"
	"reflect"
)

type SchemaItem interface {
	Type() string
	Base() SchemaProperty
}

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

type ErrUnexpectedType struct {
	Expected any
	Actual   any
}

func (e ErrUnexpectedType) Error() string {
	return fmt.Sprintf("type error: expected %T, got %T", e.Expected, e.Actual)
}

type ErrReadingKey struct {
	Key string
	error
}

func (e ErrReadingKey) Error() string {
	return fmt.Errorf("error reading key '%s': %w", e.Key, e.error).Error()
}

func Convert(json JSON) ([]SchemaItem, error) {
	switch j := json.(type) {
	case *ListNode:
		return parseList(j, parseObject)
	case *ObjNode:
		n, err := parseObject(j)
		if err != nil {
			return nil, err
		}
		return []SchemaItem{n}, nil
	case *KeyValueNode:
		n, err := parseProperty(j)
		if err != nil {
			return nil, err
		}
		return []SchemaItem{n}, nil
	}
	return nil, fmt.Errorf("cannot read object of type %T to a schema object", json)
}

func MergeNamespaces(namespaces []*SchemaNamespace) []*SchemaNamespace {
	spaces := map[string]*SchemaNamespace{}
	i := 0
	for i < len(namespaces) {
		info := namespaces[i]
		ns, exists := spaces[info.Name]
		if !exists {
			spaces[info.Name] = info
			i++
			continue
		}
		spaces[info.Name] = merge(ns, info)
		namespaces = remove(i, namespaces)
	}
	return namespaces
}

func determineType(json *ObjNode) (SchemaItem, error) {
	var item SchemaItem
	// base := SchemaProperty{}
	for _, kv := range json.Items {
		switch kv.Key {
		case "namespace":
			return &SchemaNamespace{}, nil
		case "choices":
			return &SchemaChoicesProperty{}, nil
		case "$ref":
			item = &SchemaRefProperty{}
		case "value":
			return &SchemaValueProperty{}, nil
		case "properties":
			return &SchemaObjectProperty{}, nil
		case "type":
			typeName, ok := kv.Value.(string)
			if !ok {
				break // maybe this should be an error
			}
			switch typeName {
			case "value":
				return &SchemaValueProperty{}, nil
			case "any":
				return &SchemaAnyProperty{}, nil
			case "integer":
				return &SchemaIntProperty{}, nil
			case "number":
				return &SchemaFloatProperty{}, nil
			case "boolean":
				return &SchemaBoolProperty{}, nil
			case "null":
				return &SchemaNullProperty{}, nil
			case "string":
				return &SchemaStringProperty{}, nil
			case "array":
				return &SchemaArrayProperty{}, nil
			case "object":
				return &SchemaObjectProperty{}, nil
			case "function":
				return &SchemaFunctionProperty{}, nil
			}
		}
	}
	if item == nil {
		return &SchemaProperty{}, nil
	}
	return item, nil
}

var zero reflect.Value

func parseObject(json *ObjNode) (SchemaItem, error) {
	item, err := determineType(json)
	if err != nil {
		return nil, err
	}
	for _, kv := range json.Items {
		if err := setField(item, kv); err != nil {
			return nil, fmt.Errorf("key '%s': %w", kv.Key, err)
		}
	}
	return item, nil
}

func setField(item SchemaItem, kv *KeyValueNode) error {
	itemValue := reflect.ValueOf(item).Elem()
	fieldName := exportable(snakeToCamel(kv.Key))
	if kv.Key == "namespace" {
		fieldName = "Name"
	} //else if kv.Key == "type" {
	// 	fieldName = "Type_"
	// }
	field := itemValue.FieldByName(fieldName)
	if field == zero {
		return nil
	}
	// specific actions
	var v any
	var err error
	switch kv.Key {
	case "optional", "unsupported":
		switch val := kv.Value.(type) {
		case bool:
			v = val
		case string:
			v = val == "true" || val == "omit-if-key-missing"
		}
	case "deprecated", "async":
		v = true
	case "additionalProperties":
		switch value := kv.Value.(type) {
		case bool:
			v = &SchemaAnyProperty{}
		case *ObjNode:
			v, err = parseObject(value)
		}
	case "properties", "patternProperties":
		v, err = castAndCall(kv.Value, func(lst *ObjNode) ([]SchemaItem, error) {
			return mapf(lst.Items, parseProperty)
		})
	default:
		switch field.Interface().(type) {
		case string, bool, int, float64:
			v = kv.Value
		case nil:
			_field, _ := itemValue.Type().FieldByName(fieldName)
			switch _field.Type.Name() {
			case "SchemaItem":
				v, err = castAndCall(kv.Value, parseObject)
			default:
				v = kv.Value
			}
		case []string:
			v, err = castAndCall(kv.Value, wrap(identity[string]))
		case []SchemaItem:
			v, err = castAndCall(kv.Value, wrap(parseObject))
		case []*SchemaFunctionProperty:
			v, err = castAndCall(kv.Value, wrap(parseFunction))
		case []SchemaEnumValue:
			v, err = castAndCall(kv.Value, parseEnum)
		}
	}
	field.Set(reflect.ValueOf(v))
	return err
}

func parseProperty(json *KeyValueNode) (SchemaItem, error) {
	value, err := castAndCall(json.Value, parseObject)
	if err != nil {
		return nil, err
	}
	kv := &KeyValueNode{Key: "Name", Value: json.Key}
	setField(value, kv)
	return value, nil
}

func parseFunction(json *ObjNode) (*SchemaFunctionProperty, error) {
	if item, err := parseObject(json); err != nil {
		return nil, err
	} else if _func, ok := item.(*SchemaFunctionProperty); !ok {
		return nil, fmt.Errorf("failed to parse function: got %T", item)
	} else {
		return _func, nil
	}
}

func parseEnum(lst *ListNode) ([]SchemaEnumValue, error) {
	return parseList(lst, func(item any) (SchemaEnumValue, error) {
		enum := SchemaEnumValue{}
		switch item := item.(type) {
		case string:
			enum.Name = item
		case *ObjNode:
			for _, kv := range item.Items {
				if kv.Key == "name" {
					enum.Name = kv.Value.(string)
				} else if kv.Key == "description" {
					enum.Description = kv.Value.(string)
				}
			}
		default:
			return enum, fmt.Errorf("unexpected type: %T", item)
		}
		return enum, nil
	})
}

func wrap[T any, Y any](f func(T) (Y, error)) func(*ListNode) ([]Y, error) {
	return func(lst *ListNode) ([]Y, error) { return parseList(lst, f) }
}

func parseList[ItemType any, To any](lst *ListNode, f func(ItemType) (To, error)) ([]To, error) {
	return mapf(lst.Items, func(item any) (To, error) {
		return castAndCall(item, f)
	})
}

