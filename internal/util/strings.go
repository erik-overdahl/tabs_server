package util

import (
	"strings"
)

func SnakeToCamel(s string) string {
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

func CamelToSnake(s string) string {
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

func Exportable(s string) string {
	if s[0] == '$' {
		s = s[1:]
	}
	if 96 < s[0] {
		return string(s[0]-32) + s[1:]
	}
	return s
}

// insert a newline every n characters in the string
func Linewrap(s string, lineLength int) string {
	if len(s) < lineLength {
		return s
	}
	pieces := strings.Split(s, " ")
	lines := []string{pieces[0]}
	for _, p := range pieces[1:] {
		currLine := lines[len(lines) - 1]
		// if adding p to the line would go past the line len,
		// start a new line
		if lineLength < len(currLine) + len(p) {
			lines = append(lines, p)
			continue
		}
		lines[len(lines) - 1] += " " + p
	}
	return strings.Join(lines, string('\n'))
}
