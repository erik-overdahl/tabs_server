package json

import (
	"testing"
)

func TestTokenParser(t *testing.T) {
	cases := []struct{
		name string
		input string
		expected JSON
	}{
		{
			"Empty list should be list node",
			"[]",
			&List{Items: []any{}},
		},
		{
			"Empty object should be obj node",
			"{}",
			&Object{Items: []*KeyValue{}},
		},
		{
			"Integer value",
			`{"foo": 1}`,
			&Object{Items: []*KeyValue{
				{"foo", 1},
			}},
		},
		{
			"Boolean values",
			`{"foo": true, "bar": false}`,
			&Object{Items: []*KeyValue{
				{"foo", true},
				{"bar", false},
			}},
		},
		{
			"Float value",
			`{"foo": 1.0}`,
			&Object{Items: []*KeyValue{
				{"foo", 1.0},
			}},
		},
		{
			"Null value",
			`{"foo": null}`,
			&Object{Items: []*KeyValue{
				{"foo", nil},
			}},
		},
		{
			"String value",
			`{"foo": "null"}`,
			&Object{Items: []*KeyValue{
				{"foo", "null"},
			}},
		},
		{
			"List of objects",
			`[{"one": [1,2,3]}, {"two": 1.0, "three": null}]`,
			&List{Items: []any{
				&Object{Items: []*KeyValue{
					{"one", &List{Items: []any{1, 2, 3}}}},
				},
				&Object{Items: []*KeyValue{
					{"two", 1.0},
					{"three", nil}},
				},
			}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(test *testing.T) {
			// yes I know this is bad but I don't want to type out all
			// those damn token lists
			tokens, _ := TokenizeJson([]byte(c.input))
			parser := MakeTokenParser()
			actual, err := parser.Parse(tokens)
			if err != nil {
				t.Log(err)
				t.Fail()
			} else if !compareJSON(actual, c.expected) {
				t.Logf("expected %#v, got %#v", c.expected, actual)
				t.Fail()
			}
		})
	}
}

func compareJSON(n, m any) bool {
	switch n := n.(type) {
	case *List:
		m, ok := m.(*List)
		if !ok || len(n.Items) != len(m.Items) {
			return false
		}
		for i := range n.Items {
			if !compareJSON(n.Items[i], m.Items[i]) {
				return false
			}
		}
	case *Object:
		m, ok := m.(*Object)
		if !ok || len(n.Items) != len(m.Items) {
			return false
		}
		for i := range n.Items {
			if !compareJSON(n.Items[i], m.Items[i]) {
				return false
			}
		}
	case *KeyValue:
		m, ok := m.(*KeyValue)
		return ok && n.Key == m.Key && compareJSON(n.Value, m.Value)
	default:
		return n == m
	}
	return true
}
