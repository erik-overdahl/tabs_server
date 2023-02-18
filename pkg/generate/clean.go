/*
 * - remove `deprecated` and `unsupported` elements
 * - remove InstanceTypes
 */

package generate

import "strings"

func Clean(node Node) Node {
	switch node := node.(type) {
	case *listNode:
		return cleanList(node)
	case *objNode:
		return cleanObject(node)
	case *propNode:
		return cleanProperty(node)
	default:
		return cleanValue(node)
	}
}

func cleanList(list *listNode) *listNode {
	i := 0
	for i < len(list.Items) {
		item := list.Items[i]
		switch item := item.(type) {
		case *objNode:
			if cleanObject(item) == nil {
				list.Items = remove(i, list.Items)
				continue
			}
		case *listNode:
			if cleanList(item) == nil {
				list.Items = remove(i, list.Items)
				continue
			}
		default:
			val := cleanValue(item)
			if val == nil {
				list.Items = remove(i, list.Items)
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

func cleanObject(obj *objNode) *objNode {
	i := 0
	for i < len(obj.Properties) {
		prop := obj.Properties[i]
		switch prop.Name {
		case "unsupported", "deprecated":
			return nil
		case "id":
			val, isStr := prop.Value.(string)
			if isStr && strings.HasSuffix(val, "InstanceType") {
				return nil
			}
		case "additionalProperties":
		}
		if cleanProperty(prop) == nil {
			obj.Properties = remove(i, obj.Properties)
			continue
		}
		i++
	}
	if len(obj.Properties) == 0 {
		return nil
	}
	return obj
}

func cleanProperty(prop *propNode) *propNode {
	switch value := prop.Value.(type) {
	case *listNode:
		if cleanList(value) == nil {
			return nil
		}
	case *objNode:
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
