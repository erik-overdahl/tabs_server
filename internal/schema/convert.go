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
		result := make([]Item, len(j.Items))
		for i, v := range j.Items {
			o, ok := v.(*ojson.Object)
			if !ok {
				return nil, fmt.Errorf("Err converting list: %w", util.ErrUnexpectedType{o, v})
			}
			obj, err := parseObject(o, nil)
			if err != nil {
				return nil, err
			}
			result[i] = obj
		}
		return result, nil
	case *ojson.Object:
		n, err := parseObject(j, nil)
		if err != nil {
			return nil, err
		}
		return []Item{n}, nil
	case *ojson.KeyValue:
		n, err := parseProperty(j, nil)
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

/*
 * Intentionally named to make you uncomfortable
 */

func itemFactory(json *ojson.Object, parent Item) Item {
	var item Item
	for _, kv := range json.Items {
		switch kv.Key {
		case "namespace":
			return &Namespace{Property: Property{parent: parent}}
		case "choices":
			return &Choices{Property: Property{parent: parent}}
		case "$ref":
			item = &Ref{Property: Property{parent: parent}}
		case "value":
			return &Value{Property: Property{parent: parent}}
		case "properties":
			return &Object{Property: Property{parent: parent}}
		case "enum":
			return &Enum{Property: Property{parent: parent}}
		case "type":
			typeName, ok := kv.Value.(string)
			if !ok {
				break // maybe this should be an error
			}
			switch typeName {
			case "value":
				return &Value{Property: Property{parent: parent}}
			case "any":
				return &Any{Property: Property{parent: parent}}
			case "integer":
				return &Int{Property: Property{parent: parent}}
			case "number":
				return &Number{Property: Property{parent: parent}}
			case "boolean":
				return &Bool{Property: Property{parent: parent}}
			case "null":
				return &Null{Property: Property{parent: parent}}
			case "string":
				item = &String{Property: Property{parent: parent}}
			case "array":
				return &Array{Property: Property{parent: parent}}
			case "object":
				return &Object{Property: Property{parent: parent}}
			case "function":
				return &Function{Property: Property{parent: parent}}
			}
		}
	}
	if item == nil {
		return &Property{parent: parent}
	}
	return item
}

var zero reflect.Value

func parseObject(json *ojson.Object, parent Item) (Item, error) {
	item := itemFactory(json, parent)
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
	}
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
			v = &Any{Property: Property{parent: item}}
		case *ojson.Object:
			v, err = parseObject(value, item)
		}
	case "properties", "patternProperties":
		val, ok := kv.Value.(*ojson.Object)
		if !ok {
			return util.ErrUnexpectedType{val, kv.Value}
		}
		lst := make([]Item, len(val.Items))
		for i := range val.Items {
			lst[i], err = parseProperty(val.Items[i], item)
			if err != nil {
				return err
			}
		}
		v = lst
	default:
		switch field.Interface().(type) {
		case string, bool, int, float64:
			v = kv.Value
		case nil:
			_field, _ := itemValue.Type().FieldByName(fieldName)
			switch _field.Type.Name() {
			case "Item":
				val, ok := kv.Value.(*ojson.Object)
				if !ok {
					return util.ErrUnexpectedType{val, kv.Value}
				}
				v, err = parseObject(val, item)
			default:
				v = kv.Value
			}
		}
		if v != nil {
			break
		}
		val, ok := kv.Value.(*ojson.List)
		if !ok {
			return util.ErrUnexpectedType{val, kv.Value}
		}
		switch field.Interface().(type) {
		case []string:
			lst := make([]string, len(val.Items))
			for i := range val.Items {
				if s, isStr := val.Items[i].(string); !isStr {
					return util.ErrUnexpectedType{s, val.Items[i]}
				} else {
					lst[i] = s
				}
			}
			v = lst
		case []Item:
			v := make([]Item, len(val.Items))
			for i := range val.Items {
				o, ok := val.Items[i].(*ojson.Object)
				if !ok {
					return util.ErrUnexpectedType{o, val.Items[i]}
				}
				v[i], err = parseObject(o, item)
				if err != nil {
					return err
				}
			}
		case []*Function:
			lst := make([]*Function, len(val.Items))
			for i := range val.Items {
				o, ok := val.Items[i].(*ojson.Object)
				if !ok {
					return util.ErrUnexpectedType{o, val.Items[i]}
				}
				if out, err := parseObject(o, item); err != nil {
					return err
				} else if _func, ok := out.(*Function); !ok {
					return util.ErrUnexpectedType{_func, out}
				} else {
					lst[i] = _func
				}
			}
			v = lst
		case []EnumValue:
			
			lst := make([]EnumValue, len(val.Items))
			for i := range val.Items {
				e, err := parseEnumValue(val.Items[i])
				if err != nil {
					return err
				}
				lst[i] = e
			}
			v = lst
		}
	}
	field.Set(reflect.ValueOf(v))
	return err
}

func parseProperty(kv *ojson.KeyValue, obj Item) (Item, error) {
	val, ok := kv.Value.(*ojson.Object)
	if !ok {
		return nil, util.ErrUnexpectedType{val, kv.Value}
	}
	value, err := parseObject(val, obj)
	if err != nil {
		return nil, err
	}
	name := &ojson.KeyValue{Key: "Name", Value: kv.Key}
	setField(value, name)
	return value, nil
}

func parseEnumValue(item any) (EnumValue, error) {
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
		return enum, util.ErrUnexpectedType{&ojson.Object{}, item}
	}
	return enum, nil
}
