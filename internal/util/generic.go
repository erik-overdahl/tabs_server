package util

import (
	"fmt"
	"reflect"
)

func Remove[T any](i int, list []T) []T {
	if !(-1 < i || i < len(list)) {
		return list
	}
	switch i {
	case 0:
		return list[1:]
	case len(list) - 1:
		return list[:i]
	default:
		return append(list[:i], list[i+1:]...)
	}
}

func Mapf[T any, Y any](lst []T, f func(T) (Y, error)) ([]Y, error) {
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

type ErrUnexpectedType struct {
	Expected any
	Actual   any
}

func (e ErrUnexpectedType) Error() string {
	return fmt.Sprintf("type error: expected %T, got %T", e.Expected, e.Actual)
}

// hahaha this is garbage
func CastAndCall[From any, To any](param any, f func(From) (To, error)) (To, error) {
	if arg, ok := param.(From); !ok {
		var t From
		var zero To
		return zero, ErrUnexpectedType{t, param}
	} else {
		return f(arg)
	}
}

func Identity[T any](v T) (T, error) {
	return v, nil
}

func Merge[T any](target, source T) T {
	vTarget := reflect.ValueOf(target).Elem()
	vSource := reflect.ValueOf(source).Elem()
	if vTarget == vSource {
		return target
	}
	typ := vTarget.Type()
	for i := 0; i < typ.NumField(); i++ {
		t, s := vTarget.Field(i), vSource.Field(i)
		switch {
		case s.IsZero() || !t.CanSet():
			continue
		case t.IsZero():
			t.Set(s)
		default:
			switch typ.Field(i).Type.Kind() {
			case reflect.Slice:
				t.Set(reflect.AppendSlice(t, s))
			case reflect.Pointer, reflect.Interface:
				t.Set(reflect.ValueOf(Merge(t.Interface(), s.Interface())))
			}
		}
	}
	return target
}
