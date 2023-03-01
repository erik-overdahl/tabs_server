package util

import "testing"


type greeter interface {
	Greeting() string
}

type simpleStruct struct {
	Name string
}

func (s simpleStruct) Greeting() string {
	return s.Name
}

type ptrStruct struct {
	Child *simpleStruct
}

type sliceStruct struct {
	Children []simpleStruct
}

type ptrSliceStruct struct {
	Children []*simpleStruct
}

type interfaceStruct struct {
	Child greeter
}

type interfaceSliceStruct struct {
	Children []greeter
}

func TestValueEqual(t *testing.T) {
	cases := []struct{
		Name string
		X, Y any
		Expected bool
	}{
		{"Equal strings", "foo", "foo", true},
		{"Unequal strings", "foo", "bar", false},
		{
			"Equal structs",
			simpleStruct{Name: "foo"},
			simpleStruct{Name: "foo"},
			true,
		},
		{
			"Unequal structs",
			simpleStruct{Name: "foo"},
			simpleStruct{Name: "bar"},
			false,
		},
		{
			"Equal structs different pointers",
			ptrStruct{Child: &simpleStruct{Name: "foo"}},
			ptrStruct{Child: &simpleStruct{Name: "foo"}},
			true,
		},
		{
			"Unequal structs different pointers",
			ptrStruct{Child: &simpleStruct{Name: "foo"}},
			ptrStruct{Child: &simpleStruct{Name: "bar"}},
			false,
		},
		{
			"Equal string slice",
			[]string{"foo"},
			[]string{"foo"},
			true,
		},
		{
			"Unequal string slice",
			[]string{"foo"},
			[]string{"bar"},
			false,
		},
		{
			"Equal slice of structs",
			[]simpleStruct{{"foo"}},
			[]simpleStruct{{"foo"}},
			true,
		},
		{
			"Unequal slice of structs",
			[]simpleStruct{{"foo"}},
			[]simpleStruct{{"bar"}},
			false,
		},
		{
			"Equal slice as field on struct",
			sliceStruct{Children: []simpleStruct{{"foo"}}},
			sliceStruct{Children: []simpleStruct{{"foo"}}},
			true,
		},
		{
			"Unequal slice as field on struct",
			sliceStruct{Children: []simpleStruct{{"foo"}}},
			sliceStruct{Children: []simpleStruct{{"bar"}}},
			false,
		},
		{
			"Equal slice of pointers to structs",
			[]*simpleStruct{{"foo"}},
			[]*simpleStruct{{"foo"}},
			true,
		},
		{
			"Unequal slice of pointers to structs",
			[]*simpleStruct{{"foo"}},
			[]*simpleStruct{{"bar"}},
			false,
		},
		{
			"Equal slice of interfaces",
			[]greeter{&simpleStruct{"foo"}},
			[]greeter{&simpleStruct{"foo"}},
			true,
		},
		{
			"Unequal slice of interfaces",
			[]greeter{&simpleStruct{"foo"}},
			[]greeter{&simpleStruct{"bar"}},
			false,
		},
	}
	for _, c := range cases {
		if actual := ValueEqual(c.X, c.Y); actual != c.Expected {
			t.Errorf("%s: %#v, %#v: expected %v, got %v", c.Name, c.X, c.Y, c.Expected, actual)
		}
	}
}
