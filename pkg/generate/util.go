package generate

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
		if 64 < c && c < 91 {
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
