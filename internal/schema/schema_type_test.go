package schema

import (
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/erik-overdahl/tabs_server/internal/util"
)

type typeTest struct {
	Name string
	Input Item
	Expected *jen.Statement
}

func (c typeTest) doTest(t *testing.T) {
	actual := typeTestString(c.Input.Type())
	expected := typeTestString(c.Expected)
	if actual != expected {
		t.Errorf("Expected: %s\n Got: %s", expected, actual)
	}
}

func typeTestString(typ *jen.Statement) string {
	f := jen.NewFile("foo")
	f.Var().Id("foo").Add(typ)
	s := strings.Builder{}
	f.Render(&s)
	return s.String()
}

func TestBasicTypes(t *testing.T) {
	cases := []typeTest{
		{
			Name: "String",
			Input: String{},
			Expected: jen.String(),
		},
		{
			Name: "Number",
			Input: Number{},
			Expected: jen.Float64(),
		},
		{
			Name: "Integer",
			Input: Int{},
			Expected: jen.Int(),
		},
		{
			Name: "Boolean",
			Input: Bool{},
			Expected: jen.Bool(),
		},
		{
			Name: "Null",
			Input: Null{},
			Expected: jen.Nil(),
		},
		{
			Name: "Any",
			Input: Any{},
			Expected: jen.Any(),
		},
		{
			Name: "Value with int",
			Input: Value{Value: 10},
			Expected: jen.Int(),
		},
		{
			Name: "Value with float",
			Input: Value{Value: 10.5},
			Expected: jen.Float64(),
		},
		{
			Name: "Value with string",
			Input: Value{Value: "foo"},
			Expected: jen.String(),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, c.doTest)
	}
}

func TestRefType(t *testing.T) {
	cases := []typeTest{
 		{
 			Name: "Ref type should be reference",
 			Input: &Ref{Property: Property{Name: "foo", Ref: "Bar"}},
 			Expected: jen.Id("Bar"),
 		},
		{
 			Name: "Ref to type in other package",
 			Input: &Ref{Property: Property{Name: "foo", Ref: "other.Bar"}},
 			Expected: jen.Qual("other","Bar"),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, c.doTest)
	}
}

// I guess this is really testing the Property.Type() but whatever
func TestEnumType(t *testing.T) {
	cases := []typeTest{
		{
			Name: "Enum with id",
			Input: &Enum{Property: Property{Id: "WindowType"}},
			Expected: jen.Id("WindowType"),
		},
		{
			// e.g. as the property of a struct
			Name: "Enum with name",
			Input: &Enum{Property: Property{Name: "windowType"}},
			Expected: jen.Id("WindowType"),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, c.doTest)
	}
}

func TestObjectType(t *testing.T) {
	cases := []typeTest{
		{
			Name: "Object with id",
			Input: &Object{Property: Property{Id: "Details"}},
			Expected: jen.Id("Details"),
		},
		{
			// e.g. a struct embedded in a larger struct
			Name: "Object with name",
			Input: &Object{Property: Property{Name: "details"}},
			Expected: jen.Id("Details"),
		},
		{
			Name: "Object with only AdditionalProperties is a map",
			Input: &Object{Property: Property{
				Name: "details"},
				AdditionalProperties: &String{},
			},
			Expected: jen.Map(jen.String()).String(),
		},
		{
			Name: "Object with only PatternProperties is a map",
			Input: &Object{Property: Property{
				Name: "details"},
				PatternProperties: []Item{
					&String{},
				},
			},
			Expected: jen.Map(jen.String()).String(),
		},
		{
			Name: "Object with multiple PatternProperties is a map of Any",
			Input: &Object{Property: Property{
				Name: "details"},
				PatternProperties: []Item{
					&String{},
					&Int{},
				},
			},
			Expected: jen.Map(jen.String()).Any(),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, c.doTest)
	}
}

func TestArrayType(t *testing.T) {
	cases := []typeTest{
		{
			Name: "Array of basic type",
			Input: &Array{Items: &String{}},
			Expected: jen.Index().String(),
		},
	}
	for _, c := range cases {
		t.Run(c.Name, c.doTest)
	}
}

func TestChoose(t *testing.T) {
	cases := []struct{
		Name string
		Input []Item
		Expected Item
	}{
		{
			Name: "String preferred to int",
			Input: []Item{
				&String{},
				&Int{},
			},
			Expected: &String{},
		},
		{
			Name: "Array of strings preferred over string",
			Input: []Item{
				&String{},
				&Array{
					Items: &String{},
				},
			},
			Expected: &Array{
				Items: &String{},
			},
		},
		{
			Name: "Array of $ref preferred over $ref",
			Input:[]Item{
				&Ref{Property: Property{Ref: "Tab"}},
				&Array{Items: &Ref{Property: Property{Ref: "Tab"}}},
			},
			Expected: &Array{Items: &Ref{Property: Property{Ref: "Tab"}}},
		},
		{
			Name: "Enum preferred to boolean",
			Input: []Item{
				&Enum{Enum: []EnumValue{{Name: "Screen"}, {Name: "Window"}}},
				&Bool{},
			},
			Expected: &Enum{Enum: []EnumValue{{Name: "Screen"}, {Name: "Window"}}},
		},
		{
			// not 100% on this
			Name: "Ref preferred to object",
			Input: []Item{
				&Ref{Property: Property{Ref: "ImageDataType"}},
				&Object{
					PatternProperties: []Item{
						&Ref{Property: Property{
							Name: "^[1-9]\\d*$",
							Ref: "ImageDataType",
						}},
					},
				},
			},
			Expected: &Ref{Property: Property{Ref: "ImageDataType"}},
		},
		{
			Name: "Ref preferred to enum",
			Input: []Item{
				&Enum{Enum: []EnumValue{{Name: ""}}},
				&Ref{Property: Property{Ref: "manifest.ExtensionUrl"}},
			},
			Expected: &Ref{Property: Property{Ref: "manifest.ExtensionUrl"}},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T){
			choices := Choices{Choices: c.Input}
			if !util.ValueEqual(choices.Choose(), c.Expected) {
				t.Errorf("Expected %#v\nGot %#v", c.Expected, choices.Choose())
			}
		})
	}
}

