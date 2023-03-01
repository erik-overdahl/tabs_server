package json

import (
	"errors"
	"reflect"
	"testing"
)

func TestTokenizeValid(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []token
	}{
		{
			"Single line comment",
			"// This is a comment\n",
			[]token{jsonComment{pos: 0, Value: "// This is a comment\n"}},
		},
		{
			"Multiline comment",
			"/* \n * This is a multiline comment\n */\n",
			[]token{jsonComment{pos: 0, Value: "/* \n * This is a multiline comment\n */\n"}},
		},
		{
			"Parse int",
			"1234567890",
			[]token{jsonNumber{pos: 0, value: 1234567890}},
		},
		{
			"Parse negative int",
			"-1234567890",
			[]token{jsonNumber{pos: 0, value: -1234567890}},
		},
		{
			"Parse float",
			"123.456",
			[]token{jsonNumber{pos: 0, value: 123.456}},
		},
		{
			"Parse negative float",
			"-123.456",
			[]token{jsonNumber{pos: 0, value: -123.456}},
		},
		{
			"Parse true",
			"true",
			[]token{jsonBool{pos: 0, value: true}},
		},
		{
			"Parse false",
			"false",
			[]token{jsonBool{pos: 0, value: false}},
		},
		{
			"Parse null",
			"null",
			[]token{jsonNull{pos: 0}},
		},
		{
			"Empty object",
			"{}",
			[]token{jsonObjOpen{pos: 0}, jsonObjClose{pos: 1}},
		},
		{
			"Object with single key (string)",
			`{"foo":"bar"}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonString{pos:7, value: "bar"},
				jsonObjClose{pos: 12},
			},
		},
		{
			"Object with single key (int)",
			`{"foo":10}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonNumber{pos:7, value: 10},
				jsonObjClose{pos: 9},
			},
		},
		{
			"Object with single key (negative int)",
			`{"foo":-10}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonNumber{pos:7, value: -10},
				jsonObjClose{pos: 10},
			},
		},
		{
			"Object with single key (float)",
			`{"foo":1.0}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonNumber{pos:7, value: 1.0},
				jsonObjClose{pos: 10},
			},
		},
		{
			"Object with single key (negative float)",
			`{"foo":-1.0}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonNumber{pos:7, value: -1.0},
				jsonObjClose{pos: 11},
			},
		},
		{
			"Object with single key (negative int)",
			`{"foo":-10}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonNumber{pos:7, value: -10},
				jsonObjClose{pos: 10},
			},
		},
		{
			"Object with single key (negative int)",
			`{"foo":-10}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonNumber{pos:7, value: -10},
				jsonObjClose{pos: 10},
			},
		},
		{
			"Object with single key (list)",
			`{"foo":[1,2]}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonArrOpen{pos:7},
				jsonNumber{pos:8, value: 1},
				jsonComma{pos:9},
				jsonNumber{pos:10,value:2},
				jsonArrClose{pos:11},
				jsonObjClose{pos:12},
			},
		},
		{
			"Object with multiple keys",
			`{"foo":1, "bar":2}`,
			[]token{
				jsonObjOpen{pos:0},
				jsonString{pos:1, value: "foo"},
				jsonColon{pos:6},
				jsonNumber{pos:7, value: 1},
				jsonComma{pos:8},
				jsonString{pos:10, value: "bar"},
				jsonColon{pos:15},
				jsonNumber{pos:16, value: 2},
				jsonObjClose{pos: 17},
			},
		},
		{
			"Empty list",
			"[]",
			[]token{jsonArrOpen{pos: 0}, jsonArrClose{pos: 1}},
		},
		{
			"List of values",
			"[1,2,3]",
			[]token{
				jsonArrOpen{pos:0},
				jsonNumber{pos:1, value:1},
				jsonComma{pos:2},
				jsonNumber{pos:3, value:2},
				jsonComma{pos:4},
				jsonNumber{pos:5, value:3},
				jsonArrClose{pos:6},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(test *testing.T) {
			defer func(name string) {
				if r := recover(); r != nil {
					t.Logf("FAIL: Panic occurred during '%s':\n%v", name, r)
					t.Fail()
				}
			}(c.name)

			actual, err := TokenizeJson([]byte(c.input))
			if err != nil {
				t.Logf("FAIL: %s:\ngot error %#v", c.name, err)
				t.Fail()
			} else if !reflect.DeepEqual(actual, c.expected) {
				t.Logf("FAIL: %s:\nexpected %#v\ngot %#v", c.name, c.expected, actual)
				t.Fail()
			}
		})
	}
}

func TestTokenizeInvalid(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected error
	}{
		{
			"Multiline comment should error if not followed by newline or EOD",
			"/* \n * This \n * is \n * a \n * multiline \n * comment */ with some stuff after it",
			errTokenize{pos: 53, info:"multiline comment must be followed by newline or EOF: found ' '"},
		},
		{
			"Extra closing braces should error",
			"{}}",
			errTokenize{pos: 2, info: "unexpected char '}'"},
		},
		{
			"Opening brace must be followed by a string",
			"{[]}",
			errTokenize{pos: 1, info: "unexpected char reading key-value pair: expected '\"', got '['"},
		},
		{
			"Object keys must be strings",
			`{"foo": "bar", 10: "baz"}`,
			errTokenize{pos:14, info:"unexpected char reading key-value pair: expected '\"', got '1'"},
		},
		{
			"List should not contain key value pairs",
			`["foo":"bar"]`,
			errTokenize{pos: 6, info:"unexpected char reading array: ':'"},
		},
		{
			"Malformed number",
			`123abc`,
			errTokenize{pos: 3, info: "unexpected char while parsing number: 'a'"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(test *testing.T) {
			// defer func(name string) {
			// 	if r := recover(); r != nil {
			// 		t.Logf("FAIL: Panic occurred during '%s':\n%v", name, r)
			// 		t.Fail()
			// 	}
			// }(c.name)

			_, err := TokenizeJson([]byte(c.input))
			if !errors.Is(err, c.expected) {
				t.Logf("FAIL: %s:\nexpected %#v\ngot %#v", c.name, c.expected, err)
				t.Fail()
			}
		})
	}
}
