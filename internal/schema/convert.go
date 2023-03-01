package schema

import (
	"fmt"
	_ "log"
	"reflect"

	"github.com/erik-overdahl/tabs_server/internal/util"
	ojson "github.com/erik-overdahl/tabs_server/internal/json"
)

type ErrReadingKey struct {
	Key string
	error
}

func (e ErrReadingKey) Error() string {
	return fmt.Errorf("error reading key '%s': %w", e.Key, e.error).Error()
}

func Convert(json ojson.JSON) ([]Item, error) {
	switch j := json.(type) {
	case *ojson.List:
		return parseList(j, parseObject)
	case *ojson.Object:
		n, err := parseObject(j)
		if err != nil {
			return nil, err
		}
		return []Item{n}, nil
	case *ojson.KeyValue:
		n, err := parseProperty(j)
		if err != nil {
			return nil, err
		}
		return []Item{n}, nil
	}
	return nil, fmt.Errorf("cannot read object of type %T to a schema object", json)
}

func MergeNamespaces(namespaces []*Namespace) []*Namespace {
	spaces := map[string]*Namespace{}
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

func determineType(json *ojson.Object) (Item, error) {
	var item Item
	for _, kv := range json.Items {
		switch kv.Key {
		case "namespace":
			return &Namespace{}, nil
		case "choices":
			return &Choices{}, nil
		case "$ref":
			item = &Ref{}
		case "value":
			return &Value{}, nil
		case "properties":
			return &Object{}, nil
		case "enum":
			return &Enum{}, nil
		case "type":
			typeName, ok := kv.Value.(string)
			if !ok {
				break // maybe this should be an error
			}
			switch typeName {
			case "value":
				return &Value{}, nil
			case "any":
				return &Any{}, nil
			case "integer":
				return &Int{}, nil
			case "number":
				return &Number{}, nil
			case "boolean":
				return &Bool{}, nil
			case "null":
				return &Null{}, nil
			case "string":
				item = &String{}
			case "array":
				return &Array{}, nil
			case "object":
				return &Object{}, nil
			case "function":
				return &Function{}, nil
			}
		}
	}
	if item == nil {
		return &Property{}, nil
	}
	return item, nil
}

var zero reflect.Value

func parseObject(json *ojson.Object) (Item, error) {
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

func setField(item Item, kv *ojson.KeyValue) error {
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
			v = &Any{}
		case *ojson.Object:
			v, err = parseObject(value)
		}
	case "properties", "patternProperties":
		v, err = util.CastAndCall(kv.Value, func(lst *ojson.Object) ([]Item, error) {
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
		case []Item:
			v, err = util.CastAndCall(kv.Value, wrap(parseObject))
		case []*Function:
			v, err = util.CastAndCall(kv.Value, wrap(parseFunction))
		case []EnumValue:
			v, err = util.CastAndCall(kv.Value, parseEnum)
		}
	}
	field.Set(reflect.ValueOf(v))
	return err
}

func parseProperty(json *ojson.KeyValue) (Item, error) {
	value, err := util.CastAndCall(json.Value, parseObject)
	if err != nil {
		return nil, err
	}
	kv := &ojson.KeyValue{Key: "Name", Value: json.Key}
	setField(value, kv)
	return value, nil
}

func parseFunction(json *ojson.Object) (*Function, error) {
	if item, err := parseObject(json); err != nil {
		return nil, err
	} else if _func, ok := item.(*Function); !ok {
		return nil, fmt.Errorf("failed to parse function: got %T", item)
	} else {
		return _func, nil
	}
}

func parseEnum(lst *ojson.List) ([]EnumValue, error) {
	return parseList(lst, func(item any) (EnumValue, error) {
		enum := EnumValue{}
		switch item := item.(type) {
		case string:
			enum.Name = item
		case *ojson.Object:
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

func wrap[T any, Y any](f func(T) (Y, error)) func(*ojson.List) ([]Y, error) {
	return func(lst *ojson.List) ([]Y, error) { return parseList(lst, f) }
}

func parseList[ItemType any, To any](lst *ojson.List, f func(ItemType) (To, error)) ([]To, error) {
	return util.Mapf(lst.Items, func(item any) (To, error) {
		return util.CastAndCall(item, f)
	})
}

