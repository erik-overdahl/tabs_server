/*
 * - remove `deprecated` and `unsupported` elements
 * - remove InstanceTypes
 */

package generate

import "strings"

func Clean(node JSON) JSON {
	switch node := node.(type) {
	case *ListNode:
		return cleanList(node)
	case *ObjNode:
		return cleanObject(node)
	case *KeyValueNode:
		return cleanProperty(node)
	default:
		return cleanValue(node)
	}
}

func cleanList(list *ListNode) *ListNode {
	i := 0
	for i < len(list.Items) {
		item := list.Items[i]
		switch item := item.(type) {
		case *ObjNode:
			if cleanObject(item) == nil {
				list.Items = remove(i, list.Items)
				continue
			}
		case *ListNode:
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

func cleanObject(obj *ObjNode) *ObjNode {
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
		}
		if cleanProperty(prop) == nil {
			obj.Items = remove(i, obj.Items)
			continue
		}
		i++
	}
	if len(obj.Items) == 0 {
		return nil
	}
	return obj
}

func cleanProperty(prop *KeyValueNode) *KeyValueNode {
	switch value := prop.Value.(type) {
	case *ListNode:
		if cleanList(value) == nil {
			return nil
		}
	case *ObjNode:
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
