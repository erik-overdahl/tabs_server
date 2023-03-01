/*
 * - remove `deprecated` and `unsupported` elements
 * - remove InstanceTypes
 */

package schema

import (
	"strings"

	"github.com/erik-overdahl/tabs_server/internal/util"
	ojson "github.com/erik-overdahl/tabs_server/internal/json"
)

func Clean(node ojson.JSON) ojson.JSON {
	switch node := node.(type) {
	case *ojson.List:
		return cleanList(node)
	case *ojson.Object:
		return cleanObject(node)
	case *ojson.KeyValue:
		return cleanProperty(node)
	default:
		return cleanValue(node)
	}
}

func cleanList(list *ojson.List) *ojson.List {
	i := 0
	for i < len(list.Items) {
		item := list.Items[i]
		switch item := item.(type) {
		case *ojson.Object:
			if cleanObject(item) == nil {
				list.Items = util.Remove(i, list.Items)
				continue
			}
		case *ojson.List:
			if cleanList(item) == nil {
				list.Items = util.Remove(i, list.Items)
				continue
			}
		default:
			val := cleanValue(item)
			if val == nil {
				list.Items = util.Remove(i, list.Items)
				continue
			}
			list.Items[i] = val
		}
		i++
	}
	if len(list.Items) == 0 {
		return nil
	}
	return list
}

func cleanObject(obj *ojson.Object) *ojson.Object {
	i := 0
	for i < len(obj.Items) {
		prop := obj.Items[i]
		switch prop.Key {
		case "unsupported", "deprecated":
			return nil
		case "id":
			val, isStr := prop.Value.(string)
			if isStr && strings.HasSuffix(val, "InstanceType") {
				return nil
			}
		case "additionalProperties":
			addProp, ok := prop.Value.(*ojson.Object)
			if !ok {
				break
			}
			for _, p := range addProp.Items {
				if p.Key != "$ref" {
					continue
				}
				removed := false
				switch p.Value.(string) {
				case "UnrecognizedProperty",
					"ImageDataOrExtensionURL",
					"ThemeColor":
					util.Remove(i, obj.Items)
					removed = true
				}
				if removed {
					break
				}
			}
		case "$ref":
			switch prop.Value.(string) {
			case "UnrecognizedProperty":
				prop.Value = "any"
			case "PersistenBackgroundProperty":
				prop.Value = "boolean"
			}
		}
		if cleanProperty(prop) == nil {
			obj.Items = util.Remove(i, obj.Items)
			continue
		}
		i++
	}
	if len(obj.Items) == 0 {
		return nil
	}
	return obj
}

func cleanProperty(prop *ojson.KeyValue) *ojson.KeyValue {
	switch value := prop.Value.(type) {
	case *ojson.List:
		if cleanList(value) == nil {
			return nil
		}
	case *ojson.Object:
		if cleanObject(value) == nil {
			return nil
		}
	default:
		prop.Value = cleanValue(value)
		if prop.Value == nil {
			return nil
		}
	}
	return prop
}

func cleanValue(value any) any {
	return value
}
