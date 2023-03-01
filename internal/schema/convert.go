package schema

import (
	"fmt"
	_ "log"
	"reflect"

	"github.com/erik-overdahl/tabs_server/internal/util"
	ojson "github.com/erik-overdahl/tabs_server/internal/json"
)

type SchemaItem interface {
	Type() string
	Base() SchemaProperty
}

type ErrReadingKey struct {
	Key string
	error
}

func (e ErrReadingKey) Error() string {
	return fmt.Errorf("error reading key '%s': %w", e.Key, e.error).Error()
}

func Convert(json ojson.JSON) ([]SchemaItem, error) {
	switch j := json.(type) {
	case *ojson.ListNode:
		return parseList(j, parseObject)
	case *ojson.ObjNode:
		n, err := parseObject(j)
		if err != nil {
			return nil, err
		}
		return []SchemaItem{n}, nil
	case *ojson.KeyValueNode:
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
		spaces[info.Name] = util.Merge(ns, info)
		namespaces = util.Remove(i, namespaces)
	}
	return namespaces
}

func determineType(json *ojson.ObjNode) (SchemaItem, error) {
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

func parseObject(json *ojson.ObjNode) (SchemaItem, error) {
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

func setField(item SchemaItem, kv *ojson.KeyValueNode) error {
	itemValue := reflect.ValueOf(item).Elem()
	fieldName := util.Exportable(util.SnakeToCamel(kv.Key))
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
		case *ojson.ObjNode:
			v, err = parseObject(value)
		}
	case "properties", "patternProperties":
		v, err = util.CastAndCall(kv.Value, func(lst *ojson.ObjNode) ([]SchemaItem, error) {
			return util.Mapf(lst.Items, parseProperty)
		})
	default:
		switch field.Interface().(type) {
		case string, bool, int, float64:
			v = kv.Value
		case nil:
			_field, _ := itemValue.Type().FieldByName(fieldName)
			switch _field.Type.Name() {
			case "SchemaItem":
				v, err = util.CastAndCall(kv.Value, parseObject)
			default:
				v = kv.Value
			}
		case []string:
			v, err = util.CastAndCall(kv.Value, wrap(util.Identity[string]))
		case []SchemaItem:
			v, err = util.CastAndCall(kv.Value, wrap(parseObject))
		case []*SchemaFunctionProperty:
			v, err = util.CastAndCall(kv.Value, wrap(parseFunction))
		case []SchemaEnumValue:
			v, err = util.CastAndCall(kv.Value, parseEnum)
		}
	}
	field.Set(reflect.ValueOf(v))
	return err
}

func parseProperty(json *ojson.KeyValueNode) (SchemaItem, error) {
	value, err := util.CastAndCall(json.Value, parseObject)
	if err != nil {
		return nil, err
	}
	kv := &ojson.KeyValueNode{Key: "Name", Value: json.Key}
	setField(value, kv)
	return value, nil
}

func parseFunction(json *ojson.ObjNode) (*SchemaFunctionProperty, error) {
	if item, err := parseObject(json); err != nil {
		return nil, err
	} else if _func, ok := item.(*SchemaFunctionProperty); !ok {
		return nil, fmt.Errorf("failed to parse function: got %T", item)
	} else {
		return _func, nil
	}
}

func parseEnum(lst *ojson.ListNode) ([]SchemaEnumValue, error) {
	return parseList(lst, func(item any) (SchemaEnumValue, error) {
		enum := SchemaEnumValue{}
		switch item := item.(type) {
		case string:
			enum.Name = item
		case *ojson.ObjNode:
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

func wrap[T any, Y any](f func(T) (Y, error)) func(*ojson.ListNode) ([]Y, error) {
	return func(lst *ojson.ListNode) ([]Y, error) { return parseList(lst, f) }
}

func parseList[ItemType any, To any](lst *ojson.ListNode, f func(ItemType) (To, error)) ([]To, error) {
	return util.Mapf(lst.Items, func(item any) (To, error) {
		return util.CastAndCall(item, f)
	})
}

