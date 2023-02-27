package generate

import (
	"fmt"
	"reflect"
)

type Stack[T any] struct {
	items []T
}

func MakeStack[T any]() *Stack[T] {
	return &Stack[T]{items: []T{}}
}

func (s *Stack[T]) Push(val T) {
	s.items = append(s.items, val)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

func (s Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s Stack[T]) Len() int {
	return len(s.items)
}

func remove[T any](i int, list []T) []T {
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

func snakeToCamel(s string) string {
	res := []byte{}
	upcase := false
	for i := range s {
		c := s[i]
		if c == '_' {
			upcase = true
			continue
		} else if upcase && 96 < c {
			c -= 32
			upcase = false
		}
		res = append(res, c)
	}
	return string(res)
}

func camelToSnake(s string) string {
	out := []byte{}
	for i := range s {
		c := s[i]
		if 64 < c && c < 91 && 0 < i && 96 < s[i-1] {
			out = append(out, '_', c+32)
		} else {
			out = append(out, c)
		}
	}
	return string(out)
}

func exportable(s string) string {
	if s[0] == '$' {
		s = s[1:]
	}
	if 96 < s[0] {
		return string(s[0]-32) + s[1:]
	}
	return s
}

func mapf[T any, Y any](lst []T, f func(T) (Y, error)) ([]Y, error) {
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

func merge[T any](target, source T) T {
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
				t.Set(reflect.ValueOf(merge(t.Interface(), s.Interface())))
			}
		}
	}
	return target
}

