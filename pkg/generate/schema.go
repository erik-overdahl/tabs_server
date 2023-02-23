package generate

import (
	"fmt"
)

type SchemaItem interface {
	Type() string
	Base() *SchemaProperty
}

// the base type
type SchemaProperty struct {
	Id                 string
	Name               string
	Ref                string
	Extend             string
	Description        string
	Optional           bool
	Unsupported        bool
	Deprecated         bool
	Permissions        []string
	AllowedContexts    []string
	OnError            string
	MinManifestVersion int
	MaxManifestVersion int
}

func (_ SchemaProperty) Type() string {
	return "property"
}

func (this *SchemaProperty) Base() *SchemaProperty {
	return this
}

// if there is a "choices" property
type SchemaChoicesProperty struct {
	*SchemaProperty
	Choices []SchemaItem
	Default any
}

func (_ SchemaChoicesProperty) Type() string {
	return "choices"
}

type SchemaAnyProperty struct {
	*SchemaProperty
}

func (_ SchemaAnyProperty) Type() string {
	return "any"
}

type SchemaRefProperty struct {
	*SchemaProperty
}

func (_ SchemaRefProperty) Type() string {
	return "ref"
}

type SchemaNullProperty struct {
	*SchemaProperty
}

func (_ SchemaNullProperty) Type() string {
	return "null"
}

type SchemaValueProperty struct {
	*SchemaProperty
	Value any
}

func (_ SchemaValueProperty) Type() string {
	return "value"
}

type SchemaBoolProperty struct {
	*SchemaProperty
	Default bool
}

func (_ SchemaBoolProperty) Type() string {
	return "bool"
}

type SchemaIntProperty struct {
	*SchemaProperty
	Minimum int
	Maximum int
	Default int
}

func (_ SchemaIntProperty) Type() string {
	return "integer"
}

type SchemaFloatProperty struct {
	*SchemaProperty
	Minimum float64
	Maximum float64
	Default float64
}

func (_ SchemaFloatProperty) Type() string {
	return "float"
}

type SchemaArrayProperty struct {
	*SchemaProperty
	Items   SchemaItem
	Default any
}

func (_ SchemaArrayProperty) Type() string {
	return "array"
}

type SchemaEnumValue struct {
	Name        string
	Description string
}

type SchemaStringProperty struct {
	*SchemaProperty
	Enum      []SchemaEnumValue
	MinLength int
	MaxLength int
	Pattern   string
	Format    string
	Default   string
}

func (_ SchemaStringProperty) Type() string {
	return "string"
}

type SchemaObjectProperty struct {
	*SchemaProperty
	Properties           []SchemaItem
	AdditionalProperties SchemaItem
	PatternProperties    []SchemaItem
	Import               string
	IsInstanceOf         string
	Functions            []*SchemaFunctionProperty
	Events               []*SchemaFunctionProperty
	Default              any
}

func (_ SchemaObjectProperty) Type() string {
	return "object"
}

type SchemaFunctionProperty struct {
	*SchemaProperty
	Async                           bool
	RequireUserInput                bool
	Parameters                      []SchemaItem // uh
	ExtraParameters                 []SchemaItem
	Returns                         SchemaItem
	Filters                         []SchemaItem
	AllowAmbiguousOptionalArguments bool
	AllowCrossOriginArguments       bool
}

func (_ SchemaFunctionProperty) Type() string {
	return "function"
}

// a namespace will map to a file
type SchemaNamespace struct {
	*SchemaProperty
	Properties      []SchemaItem
	Types           []SchemaItem
	Functions       []*SchemaFunctionProperty
	Events          []*SchemaFunctionProperty
	DefaultContexts []string
	NoCompile       bool // what is the purpose of this
	Import          string
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

var actions map[string]func(SchemaItem, any) error

func init() {
	actions = map[string]func(SchemaItem, any) error{
		"id": func(s SchemaItem, val any) error {
			return set(&(s.Base().Id), val)
		},
		"name": func(s SchemaItem, val any) error {
			return set(&(s.Base().Name), val)
		},
		"$ref": func(s SchemaItem, val any) error {
			return set(&(s.Base().Ref), val)
		},
		"$extend": func(s SchemaItem, val any) error {
			return set(&(s.Base().Extend), val)
		},
		"description": func(s SchemaItem, val any) error {
			return set(&(s.Base().Description), val)
		},
		"min_manifest_version": func(s SchemaItem, val any) error {
			return set(&(s.Base().MinManifestVersion), val)
		},
		"max_manifest_version": func(s SchemaItem, val any) error {
			return set(&(s.Base().MaxManifestVersion), val)
		},
		"optional": func(s SchemaItem, val any) error {
			base := s.Base()
			switch v := val.(type) {
			case bool:
				base.Optional = v
			case string:
				base.Optional = v == "true" || v == "omit-if-key-missing"
			default:
				return ErrUnexpectedType{true, val}
			}
			return nil
		},
		"unsupported": func(s SchemaItem, val any) error {
			base := s.Base()
			switch v := val.(type) {
			case bool:
				base.Unsupported = v
			case string:
				base.Unsupported = val == "true"
			default:
				return ErrUnexpectedType{true, val}
			}
			return nil
		},
		"deprecated": func(s SchemaItem, val any) error {
			s.Base().Deprecated = true // I don't care about the explanations
			return nil
		},
		"permissions": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]string, error) {
				return parseList(lst, func(s string) (string, error) { return s, nil })
			})
			if err != nil {
				return err
			}
			s.Base().Permissions = v
			return nil
		},
		"allowedContexts": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]string, error) {
				return parseList(lst, func(s string) (string, error) { return s, nil })
			})
			if err != nil {
				return err
			}
			s.Base().AllowedContexts = v
			return nil
		},
		"choices": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]SchemaItem, error) {
				return parseList(lst, parseObject)
			})
			if err != nil {
				return err
			}
			s.(*SchemaChoicesProperty).Choices = v
			return nil
		},
		"default": func(s SchemaItem, val any) error {
			switch item := s.(type) {
			case *SchemaBoolProperty:
				return set(&item.Default, val)
			case *SchemaIntProperty:
				return set(&item.Default, val)
			case *SchemaFloatProperty:
				return set(&item.Default, val)
			case *SchemaStringProperty:
				return set(&item.Default, val)
			}
			var v any
			switch value := val.(type) {
			case *ObjNode:
				v = map[string]any{}
			case *ListNode:
				v = []any{}
			default:
				v = value
			}
			switch item := s.(type) {
			case *SchemaChoicesProperty:
				item.Default = v
			case *SchemaArrayProperty:
				item.Default = v
			case *SchemaObjectProperty:
				item.Default = v
			}
			return nil
		},
		"value": func(s SchemaItem, val any) error {
			s.(*SchemaValueProperty).Value = val
			return nil
		},
		"minimum": func(s SchemaItem, val any) error {
			switch item := s.(type) {
			case *SchemaFloatProperty:
				return set(&(item.Minimum), val)
			case *SchemaIntProperty:
				return set(&(item.Minimum), val)
			default:
				return fmt.Errorf("unexpected type: %T", item)
			}
		},
		"maximum": func(s SchemaItem, val any) error {
			switch item := s.(type) {
			case *SchemaFloatProperty:
				return set(&(item.Maximum), val)
			case *SchemaIntProperty:
				return set(&(item.Maximum), val)
			default:
				return fmt.Errorf("unexpected type: %T", item)
			}
		},
		"items": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, parseObject)
			if err != nil {
				return err
			}
			s.(*SchemaArrayProperty).Items = v
			return nil
		},
		"enum": func(s SchemaItem, val any) error {
			item, ok := s.(*SchemaStringProperty)
			if !ok {
				return nil // just skip it
			}
			v, err := castAndCall(val, func(lst *ListNode) ([]SchemaEnumValue, error) {
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
			})
			if err != nil {
				return err
			}
			item.Enum = v
			return nil
		},
		"pattern": func(s SchemaItem, val any) error {
			return set(&(s.(*SchemaStringProperty).Pattern), val)
		},
		"format": func(s SchemaItem, val any) error {
			return set(&(s.(*SchemaStringProperty).Format), val)
		},

		// objects
		"properties": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ObjNode) ([]SchemaItem, error) {
				return mapf(lst.Items, parseProperty)
			})
			if err != nil {
				return err
			}
			switch item := s.(type) {
			case *SchemaObjectProperty:
				item.Properties = v
			case *SchemaNamespace:
				item.Properties = v
			default:
				return fmt.Errorf("unexpected type: %T", item)
			}
			return nil
		},
		"additionalProperties": func(s SchemaItem, val any) error {
			switch value := val.(type) {
			case bool:
				s.(*SchemaObjectProperty).AdditionalProperties = &SchemaAnyProperty{}
			case *ObjNode:
				v, err := parseObject(value)
				if err != nil {
					return err
				}
				s.(*SchemaObjectProperty).AdditionalProperties = v
			}
			return nil
		},
		"patternProperties": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ObjNode) ([]SchemaItem, error) {
				return mapf(lst.Items, parseProperty)
			})
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}
			s.(*SchemaObjectProperty).PatternProperties = v
			return nil
		},
		"$import": func(s SchemaItem, val any) error {
			switch item := s.(type) {
			case *SchemaObjectProperty:
				return set(&(item.Import), val)
			case *SchemaNamespace:
				return set(&(item.Import), val)
			default:
				return fmt.Errorf("unexpected type: %T", item)
			}
		},
		"isInstanceOf": func(s SchemaItem, val any) error {
			return set(&(s.(*SchemaObjectProperty).IsInstanceOf), val)
		},
		"functions": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]*SchemaFunctionProperty, error) {
				return parseList(lst, parseFunction)
			})
			if err != nil {
				return err
			}
			switch item := s.(type) {
			case *SchemaObjectProperty:
				item.Functions = v
			case *SchemaNamespace:
				item.Functions = v
			default:
				return fmt.Errorf("unexpected type: %T", item)
			}
			return nil
		},
		"events": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]*SchemaFunctionProperty, error) {
				return parseList(lst, parseFunction)
			})
			if err != nil {
				return err
			}
			switch item := s.(type) {
			case *SchemaObjectProperty:
				item.Events = v
			case *SchemaNamespace:
				item.Events = v
			default:
				return fmt.Errorf("unexpected type: %T", item)
			}
			return nil
		},

		// functions
		"async": func(s SchemaItem, val any) error {
			s.(*SchemaFunctionProperty).Async = true
			return nil
		},
		"requireUserInput": func(s SchemaItem, val any) error {
			return set(&(s.(*SchemaFunctionProperty).RequireUserInput), val)
		},
		"parameters": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]SchemaItem, error) {
				return parseList(lst, parseObject)
			})
			if err != nil {
				return err
			}
			s.(*SchemaFunctionProperty).Parameters = v
			return nil
		},
		"extraParameters": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]SchemaItem, error) {
				return parseList(lst, parseObject)
			})
			if err != nil {
				return err
			}
			s.(*SchemaFunctionProperty).ExtraParameters = v
			return nil
		},
		"returns": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, parseObject)
			if err != nil {
				return err
			}
			s.(*SchemaFunctionProperty).Returns = v
			return nil
		},
		"filters": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]SchemaItem, error) {
				return parseList(lst, parseObject)
			})
			if err != nil {
				return err
			}
			s.(*SchemaFunctionProperty).Filters = v
			return nil
		},
		"allowAmbiguousOptionalArguments": func(s SchemaItem, val any) error {
			return set(&(s.(*SchemaFunctionProperty).AllowAmbiguousOptionalArguments), val)
		},
		"allowCrossOriginArguments": func(s SchemaItem, val any) error {
			return set(&(s.(*SchemaFunctionProperty).AllowCrossOriginArguments), val)
		},

		// namespaces
		"namespace": func(s SchemaItem, val any) error {
			return set(&(s.(*SchemaNamespace).Name), val)
		},
		"types": func(s SchemaItem, val any) error {
			v, err := castAndCall(val, func(lst *ListNode) ([]SchemaItem, error) {
				return parseList(lst, parseObject)
			})
			if err != nil {
				return err
			}
			s.(*SchemaNamespace).Types = v
			return nil
		},
	}
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

func determineType(json *ObjNode) (SchemaItem, error) {
	var item SchemaItem
	base := &SchemaProperty{}
	for _, kv := range json.Items {
		switch kv.Key {
		case "namespace":
			return &SchemaNamespace{SchemaProperty: base}, nil
		case "choices":
			return &SchemaChoicesProperty{SchemaProperty: base}, nil
		case "$ref":
			item = &SchemaRefProperty{SchemaProperty: base}
		case "value":
			return &SchemaValueProperty{SchemaProperty: base}, nil
		case "properties":
			return &SchemaObjectProperty{SchemaProperty: base}, nil
		case "type":
			typeName, ok := kv.Value.(string)
			if !ok {
				break // maybe this should be an error
			}
			switch typeName {
			case "value":
				return &SchemaValueProperty{SchemaProperty: base}, nil
			case "any":
				return &SchemaAnyProperty{SchemaProperty: base}, nil
			case "integer":
				return &SchemaIntProperty{SchemaProperty: base}, nil
			case "number":
				return &SchemaFloatProperty{SchemaProperty: base}, nil
			case "boolean":
				return &SchemaBoolProperty{SchemaProperty: base}, nil
			case "null":
				return &SchemaNullProperty{SchemaProperty: base}, nil
			case "string":
				return &SchemaStringProperty{SchemaProperty: base}, nil
			case "array":
				return &SchemaArrayProperty{SchemaProperty: base}, nil
			case "object":
				return &SchemaObjectProperty{SchemaProperty: base}, nil
			case "function":
				return &SchemaFunctionProperty{SchemaProperty: base}, nil
			}
		}
	}
	if item == nil {
		return base, nil
	}
	return item, nil
}


func parseObject(json *ObjNode) (SchemaItem, error) {
	item, err := determineType(json)
	if err != nil {
		return nil, err
	}
	for _, kv := range json.Items {
		// fmt.Printf("%s %T\n", kv.Key, kv.Value)
		f, exists := actions[kv.Key]
		if !exists {
			// if kv.Key != "type" {
			// 	fmt.Printf("nothing to do for key %s\n", kv.Key)
			// }
			continue
		} else if err := f(item, kv.Value); err != nil {
			return nil, ErrReadingKey{kv.Key, err}
		}
	}
	return item, nil
}

func parseProperty(json *KeyValueNode) (SchemaItem, error) {
	value, err := castAndCall(json.Value, parseObject)
	if err != nil {
		return nil, err
	}
	value.Base().Name = json.Key
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

func set[T any](field *T, val any) error {
	coerced, ok := val.(T)
	if !ok {
		return ErrUnexpectedType{coerced, val}
	}
	*field = coerced
	return nil
}

func parseList[ItemType any, To any](lst *ListNode, f func(ItemType) (To, error)) ([]To, error) {
	return mapf(lst.Items, func(item any) (To, error) {
		return castAndCall(item, f)
	})
}

func mapf[T any, Y any](lst []T, f func (T) (Y, error)) ([]Y, error) {
	result := make([]Y, len(lst))
	for i := range lst {
		item, err := f(lst[i])
		if err != nil {
			return nil, fmt.Errorf("map error: item %d: %w", i, err)
		}
		result[i] = item
	}
	return result, nil
}

// hahaha this is garbage
func castAndCall[From any, To any](param any, f func(From) (To, error)) (To, error) {
	if arg, ok := param.(From); !ok {
		var t From
		var zero To
		return zero, ErrUnexpectedType{t, param}
	} else {
		return f(arg)
	}
}

func identity[T any](v T) (T, error) {
	return v, nil
}
