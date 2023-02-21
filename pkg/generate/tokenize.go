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
}

func (token jsonObjOpen) Pos() int {
	return token.pos
}

type jsonObjClose struct {
	pos   int
}

func (token jsonObjClose) Pos() int {
	return token.pos
}

type jsonArrOpen struct {
	pos   int
}

func (token jsonArrOpen) Pos() int {
	return token.pos
}

type jsonArrClose struct {
	pos   int
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

type errTokenize struct {
	pos int
	info string
}

func (e errTokenize) Error() string {
	return fmt.Sprintf("error reading token: %d: %s", e.pos, e.info)
}

type JsonTokenizer struct {
	data   []byte
	tokens []token
}

func TokenizeJson(data []byte) ([]token, error) {
	tokenizer := JsonTokenizer{
		data: data,
		tokens: []token{},
	}
	pos := 0
	for pos < len(data) {
		if n, err := tokenizer.readJson(pos); err != nil {
			return nil, err
		} else {
			pos += n
		}
	}
	return tokenizer.tokens, nil
}

func (this *JsonTokenizer) readJson(pos int) (int, error) {
	end := pos
	if n, err := this.readWhitespaceOrComment(pos); err != nil {
		return 0, err
	} else if len(this.data) <= n {
		return n, nil
	} else {
		end += n
	}

	var f func(int) (int, error)
	switch this.data[end] {
	case '{':
		f = this.readObject
	case '[':
		f = this.readArray
	case '"':
		f = this.readString
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		f = this.readNumber
	case 't':
		f = this.readTrue
	case 'f':
		f = this.readFalse
	case 'n':
		f = this.readNull
	default:
		return 0, errTokenize{pos: end, info: fmt.Sprintf("unexpected char %q", this.data[end])}
	}
	if n, err := f(end); err != nil {
		return 0, err
	} else {
		end += n
	}

	if n, err := this.readWhitespaceOrComment(end); err != nil {
		return 0, err
	} else {
		return end + n - pos, nil
	}
}

func (this *JsonTokenizer) readWhitespaceOrComment(pos int) (int, error) {
	end := pos
	for end < len(this.data) {
		switch this.data[end] {
		case ' ', '\n', '\t', '\r':
			end++
		case '/':
			if n, err := this.readComment(end); err != nil {
				return 0, err
			} else {
				end += n
			}
		default:
			return end - pos, nil
		}
	}
	return end + 1 - pos, nil
}

func (this *JsonTokenizer) readComment(pos int) (int, error) {
	end := pos + 1
	switch this.data[end] {
	case '/':
	case '*':
		return this.readMultilineComment(pos)
	default:
		return 0, errTokenize{pos: pos + 1, info: fmt.Sprintf("unexpected char reading comment: expected '/' or '*', got %q", this.data[end])}
	}
	for end < len(this.data) {
		if this.data[end] == '\n' {
			this.tokens = append(this.tokens, jsonComment{pos: pos, Value: string(this.data[pos:end+1])})
			break
		}
		end++
	}
	return end + 1 - pos, nil
}

func (this *JsonTokenizer) readMultilineComment(pos int) (int, error) {
	end := pos + 1
	complete := false
	for end < len(this.data) && !complete {
		if this.data[end-1] == '*' && this.data[end] == '/' {
			complete = true
		}
		end++
	}
	if !complete {
		return 0, errTokenize{pos: end, info: "unexpected end of data reading multiline comment"}
	} else if end < len(this.data) - 1 && this.data[end] != '\n' {
		return 0, errTokenize{pos: end, info: fmt.Sprintf("multiline comment must be followed by newline or EOF: found %q", this.data[end])}
	}
	this.tokens = append(this.tokens, jsonComment{pos: pos, Value: string(this.data[pos:end+1])})
	return end + 1 - pos, nil
}

// read the object starting at index `i` in `this.data`
// adds the resulting this.tokens to `this.tokens`
// returns index of first byte in `this.data` after end of object
func (this *JsonTokenizer) readObject(pos int) (int, error) {
	this.tokens = append(this.tokens, jsonObjOpen{pos:pos})
	end := pos + 1
	for end < len(this.data) && this.data[end] != '}' {
		if n, err := this.readKeyValuePair(end); err != nil {
			return 0, err
		} else {
			end += n
		}

		if n, err := this.readWhitespaceOrComment(end); err != nil {
			return 0, err
		} else {
			end += n
		}
		if this.data[end] == ',' {
			this.tokens = append(this.tokens, jsonComma{pos: end})
			end++
		} else if this.data[end] == '}' {
			break
		} else {
			return 0, errTokenize{pos: end, info: fmt.Sprintf("unexpected char reading object: %q", this.data[end])}
		}
	}
	if end == len(this.data) {
		return 0, errTokenize{pos: pos, info: "unexpected end of data"}
	}
	this.tokens = append(this.tokens, jsonObjClose{pos: end})
	return end + 1 - pos, nil
}

// reads the this.tokens jsonString, jsonColon, and any starting at index `i`
// in `this.data`; add the resulting this.tokens to `this.tokens` and returns index of
// first byte in `this.data` after end of value token
func (this *JsonTokenizer) readKeyValuePair(pos int) (int, error) {
	end := pos
	if n, err := this.readWhitespaceOrComment(end); err != nil {
		return 0, err
	} else {
		end += n
	}
	if this.data[end] != '"' {
		return 0, errTokenize{pos: pos, info: fmt.Sprintf("unexpected char reading key-value pair: expected '\"', got %q", this.data[end])}
	} else if n, err := this.readString(end); err != nil {
		return 0, err
	} else {
		end += n
	}

	if n,err := this.readWhitespaceOrComment(end); err != nil {
		return 0,err
	} else {
		end += n
	}

	if this.data[end] != ':' {
		return 0, errTokenize{pos: end, info: fmt.Sprintf("unexpected char reading key-value pair: expected ':', got %q", this.data[pos])}
	}
	this.tokens = append(this.tokens, jsonColon{pos: end})
	end++

	if n, err := this.readJson(end); err != nil {
		return 0, err
	} else {
		end += n
	}

	return end - pos, nil
}

// read the array starting at index `i` in `this.data`
// adds the resulting this.tokens to `this.tokens`
// returns index of first byte in `this.data` after end of object
func (this *JsonTokenizer) readArray(pos int) (int, error) {
	this.tokens = append(this.tokens, jsonArrOpen{pos: pos})
	end := pos + 1
	for end < len(this.data) && this.data[end] != ']' {
		if n, err := this.readJson(end); err != nil {
			return 0, err
		} else {
			end += n
		}
		if this.data[end] == ',' {
			this.tokens = append(this.tokens, jsonComma{pos: end})
			end++
		} else if this.data[end] == ']' {
			break
		} else {
			return 0, errTokenize{pos: end, info: fmt.Sprintf("unexpected char reading array: %q", this.data[end])}
		}
	}
	if end == len(this.data) {
		return 0, errTokenize{pos: pos, info: "unexpected end of data"}
	}
	this.tokens = append(this.tokens, jsonArrClose{pos: end})
	return end + 1 - pos, nil
}


func (this *JsonTokenizer) readString(pos int) (int, error) {
	end := pos + 1
	for end < len(this.data) {
		switch this.data[end] {
		case '"':
			if this.data[end - 1] != '\\' {
				this.tokens = append(this.tokens, jsonString{pos: pos, value: string(this.data[pos+1:end])})
				return end + 1 - pos, nil
			}
		default:
			end++
		}
	}
	return 0, errTokenize{pos: end, info: "unexpected end of data"}
}

func (this *JsonTokenizer) readNumber(pos int) (int, error) {
	reading := true
	end := pos
	if this.data[end] == '-' {
		end++
	}
	for end < len(this.data) && reading {
		switch this.data[end] {
		case ',', '}', ']', ' ', '\r', '\n', '\t':
			reading = false
		case '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			end++
		default:
			return 0, errTokenize{pos: end, info: fmt.Sprintf("unexpected char while parsing number: %q", this.data[end])}
		}
	}
	value, err := convertNum(string(this.data[pos:end]))
	if err != nil {
		return 0, err
	}
	this.tokens = append(this.tokens, jsonNumber{pos: pos, value: value})
	return end - pos, nil
}

func (this *JsonTokenizer) readTrue(pos int) (int, error) {
	if _, err := this.readValue(pos, "true"); err != nil {
		return 0, err
	}
	this.tokens = append(this.tokens, jsonBool{pos: pos, value: true})
	return 4, nil
}

func (this *JsonTokenizer) readFalse(pos int) (int, error) {
	if _, err := this.readValue(pos, "false"); err != nil {
		return 0, err
	}
	this.tokens = append(this.tokens, jsonBool{pos: pos, value: false})
	return 5, nil
}

func (this *JsonTokenizer) readNull(pos int) (int, error) {
	if _, err := this.readValue(pos, "null"); err != nil {
		return 0, err
	}
	this.tokens = append(this.tokens, jsonNull{pos: pos})
	return 4, nil
}

func (this *JsonTokenizer) readValue(pos int, value string) (int, error) {
	for j, c := range []byte(value) {
		if len(this.data) <= pos+j {
			return 0, errTokenize{pos: pos+j, info: "unexpected end of data"}
		} else if this.data[pos+j] != c {
			return 0, errTokenize{pos: pos+j, info: fmt.Sprintf("unexpected char reading %s: expected %q, got %q", value, c, this.data[pos+j])}
		}
	}
	return len(value), nil
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
