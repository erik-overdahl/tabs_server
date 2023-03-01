package util

import "reflect"

func ValueEqual(x, y any) bool {
	xVal := reflect.ValueOf(x)
	yVal := reflect.ValueOf(y)
	if !xVal.IsValid() || !yVal.IsValid() {
		return !xVal.IsValid() && !yVal.IsValid()
	}
	typeOfX := xVal.Type()
	if typeOfX != yVal.Type() {
		return false
	}
	switch typeOfX.Kind() {
	case reflect.Slice, reflect.Array:
		if xVal.Len() != yVal.Len() {
			return false
		}
		for i := 0; i < xVal.Len(); i++ {
			if !ValueEqual(xVal.Index(i).Interface(), yVal.Index(i).Interface()) {
				return false
			}
		}
	case reflect.Pointer, reflect.Interface:
		return ValueEqual(xVal.Elem().Interface(), yVal.Elem().Interface())
	case reflect.Struct:
		for i := 0; i < typeOfX.NumField(); i++ {
			if !typeOfX.Field(i).IsExported() {
				continue
			}
			if !ValueEqual(xVal.Field(i).Interface(), yVal.Field(i).Interface()) {
				return false
			}
		}
	default:
		return xVal.Interface() == yVal.Interface()
	}
	return true
}
