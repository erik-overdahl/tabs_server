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

func Identity[T any](v T) (T, error) {
	return v, nil
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

/*
 * Given a _target_ and _source_ that are instances of the same struct,
 * recursively:
 *  - copy set fields on _source_ into unset fields on _target_
 *  - append elements from slices in _sources_ to slices in _target_
 *
 * A contrived example
 * ```
 * type Foo struct {
 *     Name string
 *     Children []int
 * }
 * target := &Foo{Children: []{1, 2, 3}}
 * source := &Foo{Name: "source", Children: []{4, 5, 6}}
 * Merge(target, source)
 * ```
 * _target_ will now have Name = "source" and Children = []int{1,2,3,4,5,6}
 *
 * Be careful with this, I didn't make it very robust.
 */
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
