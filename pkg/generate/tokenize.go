package generate

import (
	"fmt"
	"strconv"
)

type token interface {
	Pos() int
}

type jsonObjOpen struct {
	pos   int
	Depth int
}

func (token jsonObjOpen) Pos() int {
	return token.pos
}

type jsonObjClose struct {
	pos   int
	Depth int
}

func (token jsonObjClose) Pos() int {
	return token.pos
}

type jsonArrOpen struct {
	pos   int
	Depth int
}

func (token jsonArrOpen) Pos() int {
	return token.pos
}

type jsonArrClose struct {
	pos   int
	Depth int
}

func (token jsonArrClose) Pos() int {
	return token.pos
}

type jsonColon struct {
	pos int
}

func (token jsonColon) Pos() int {
	return token.pos
}

type jsonComma struct {
	pos int
}

func (token jsonComma) Pos() int {
	return token.pos
}

type jsonComment struct {
	pos  int
	Value string
}

func (token jsonComment) Pos() int {
	return token.pos
}

type jsonValue interface {
	token
	Value() any
}

type jsonBool struct {
	pos int
	value bool
}

func (token jsonBool) Pos() int {
	return token.pos
}

func (token jsonBool) Value() any {
	return token.value
}

type jsonNull struct {
	pos int
}

func (token jsonNull) Pos() int {
	return token.pos
}

func (token jsonNull) Value() any {
	return nil
}

type jsonString struct {
	pos  int
	value string
}

func (token jsonString) Pos() int {
	return token.pos
}

func (token jsonString) Value() any {
	return token.value
}

type jsonNumber struct {
	pos  int
	value any // i hate this language
}

func (token jsonNumber) Pos() int {
	return token.pos
}

func (token jsonNumber) Value() any {
	return token.value
}

func TokenizeJson(data []byte) (tokens []token, err error) {
	// fmt.Printf("Tokenizing %d bytes\n", len(data))
	defer func() {
		if r := recover(); r != nil {
			tokens = nil
			err = fmt.Errorf("%s", r)
		}
	}()
	stack := Stack[byte]{}
	i := 0
	for i < len(data) {
		switch data[i] {
		case '/': // deal with comments
			// fmt.Printf("Reading comment: %d: %q...\n", i, data[i:i+5])
			start := i
			i++
			switch data[i] {
			case '*': // skip until */
				for i < len(data)-1 && !(data[i-1] == '*' && data[i] == '/') {
					i++
				}
			case '/': // skip until newline
				for i < len(data) && data[i] != '\n' {
					i++
				}
			}
			// fmt.Printf("COMMENT: %d: %d: %s\n", start, i, data[start : i+1])
			tokens = append(tokens, jsonComment{pos: i, Value: string(data[start : i+1])})

		case '{':
			stack.Push(data[i])
			tokens = append(tokens, jsonObjOpen{pos: i, Depth: stack.Len()})
			// fmt.Printf("%*s: %d\n", stack.Len(), "{", i)
			i++

		case '}':
			if val, exists := stack.Peek(); !exists {
				return nil, fmt.Errorf("unbalanced json: %q at pos %d with no corresponding '{'", data[i], i)
			} else if val != '{' {
				return nil, fmt.Errorf("Unexpected char %q at pos %d: should match %q", data[i], i, val)
			}
			// fmt.Printf("%*s: %d\n", stack.Len(), "}", i)
			tokens = append(tokens, jsonObjClose{pos: i, Depth: stack.Len()})
			stack.Pop()
			i++

		case '[':
			stack.Push(data[i])
			tokens = append(tokens, jsonArrOpen{pos: i, Depth: stack.Len()})
			// fmt.Printf("%*s: %d\n", stack.Len(), "[", i)
			i++

		case ']':
			if val, exists := stack.Peek(); !exists {
				return nil, fmt.Errorf("unbalanced json: %q at pos %d with no corresponding '['", data[i], i)
			} else if val != '[' {
				return nil, fmt.Errorf("Unexpected char %q at pos %d: should match %q", data[i], i, val)
			}
			// fmt.Printf("%*s: %d\n", stack.Len(), "]", i)
			stack.Pop()
			tokens = append(tokens, jsonArrClose{pos: i, Depth: stack.Len()})
			i++

		case ':':
			// fmt.Printf("COLON: %d: :\n", i)
			tokens = append(tokens, jsonColon{pos: i})
			i++

		case ',':
			// fmt.Printf("COMMA: %d: ,\n", i)
			tokens = append(tokens, jsonComma{pos: i})
			i++

		case '"':
			i++
			start := i
			for i < len(data) {
				if data[i] == '"' && data[i-1] != '\\' {
					break
				}
				i++
			}
			// fmt.Printf("STRING: %d: %d: %s\n", start, i, data[start : i])
			tokens = append(tokens, jsonString{pos: i, value: string(data[start:i])})
			i++

		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := i
			i++
			done := false
			for i < len(data) && !done {
				switch data[i] {
				case ',', '}', ']', ' ', '\r', '\n', '\t':
					done = true
				case '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					i++
				default:
					return nil, fmt.Errorf("unexpected char while parsing number: %q at pos %d", data[i], i)
				}
			}
			value, err := convertNum(string(data[start:i]))
			if err != nil {
				return nil, err
			}
			// fmt.Printf("NUMBER: %d: %d: %s\n", start, i, data[start : i])
			tokens = append(tokens, jsonNumber{pos: i, value: value})

		case 'f': // should we allow upper and lower case for booleans?
			for j, c := range []byte("false") {
				if data[i+j] != c {
					return nil, fmt.Errorf("unexpected char while parsing: expected %q, got %q at pos %d", c, data[i+j], i+j)
				}
			}
			tokens = append(tokens, jsonBool{pos: i, value: false})
			i += len("false")

		case 't':
			for j, c := range []byte("true") {
				if data[i+j] != c {
					return nil, fmt.Errorf("unexpected char while parsing: expected %q, got %q at pos %d", c, data[i+j], i+j)
				}
			}
			tokens = append(tokens, jsonBool{pos: i, value: true})
			i += len("true")

		case 'n':
			for j, c := range []byte("null") {
				if data[i+j] != c {
					return nil, fmt.Errorf("unexpected char while parsing: expected %q, got %q at pos %d", c, data[i+j], i+j)
				}
			}
			tokens = append(tokens, jsonNull{pos: i})
			i += len("null")
		case ' ', '\r', '\n', '\t': // skip whitespace
			i++
		default:
			return nil, fmt.Errorf("unexpected char %q at pos %d", data[i], i)
		}
	}
	if stack.Len() > 0 {
		return nil, fmt.Errorf("unbalanced json: %s were not closed", string(stack.items))
	}
	return tokens, nil
}

func convertNum(tokenValue string) (any, error) {
	var convertFunc func (string) (any, error) = func (s string) (any, error) { return strconv.Atoi(s); }
	for _, c := range tokenValue {
		if c == '.' {
			convertFunc = func (s string) (any, error) { return strconv.ParseFloat(s, 64); }
			break
		}
	}
	return convertFunc(tokenValue)
}
