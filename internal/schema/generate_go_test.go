package schema

import (
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
)

type genTest struct {
	Name string
	Input GoItem
	Expected []*jen.Statement
}

func (this genTest) doTest(t *testing.T) {
	actual, err := this.genTestString(this.Input.ToGo()...)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	expected, _ := this.genTestString(this.Expected...)
	if actual != expected {
		t.Errorf("Expected: %s\n Got: %s", expected, actual)
	}
}

func (this genTest) genTestString(stuff ...*jen.Statement) (string, error) {
	f := jen.NewFile("foo")
	for _, thing := range stuff {
		f.Add(thing)
	}
	s := strings.Builder{}
	if err := f.Render(&s); err != nil {
		return "", err
	}
	return s.String(), nil
}

func TestGenEnum(t *testing.T) {
	cases := []genTest{
		{
			Name: "Simple top level enum",
			Input: &Enum{Property: Property{
				Id: "WindowType",},
				Enum: []EnumValue{
					{Name: "normal"},
					{Name: "popup"},
				},
			},
			Expected: []*jen.Statement{
				jen.Type().Id("WindowType").String(),
				jen.Const().Defs(
					jen.Id("WindowType_NORMAL").Id("WindowType").Op("=").Lit("normal"),
					jen.Id("WindowType_POPUP").Id("WindowType").Op("=").Lit("popup"),
				),
			},
		},
		{
			Name: "Enum with illegal chars in values",
			Input: Enum{Property: Property{
				Id: "Version"},
				Enum: []EnumValue{
					{Name: "4-rc3"},
					{Name: "release1.2.3.24"},
				},
			},
			Expected: []*jen.Statement{
				jen.Type().Id("Version").String(),
				jen.Const().Defs(
					jen.Id("Version_4_RC3").Id("Version").Op("=").Lit("4-rc3"),
					jen.Id("Version_RELEASE1_2_3_24").Id("Version").Op("=").Lit("release1.2.3.24"),
				),
			},
		},
		{
			Name: "Enum with value descriptions",
			Input: &Enum{Property: Property{
				Description: "The type of window.",
				Id: "WindowType",},
				Enum: []EnumValue{
					{Name: "normal"},
					{Name: "popup"},
				},
			},
			Expected: []*jen.Statement{
				jen.Comment("The type of window."),
				jen.Type().Id("WindowType").String(),
				jen.Const().Defs(
					jen.Id("WindowType_NORMAL").Id("WindowType").Op("=").Lit("normal"),
					jen.Id("WindowType_POPUP").Id("WindowType").Op("=").Lit("popup"),
				),
			},
		},
	}
	{
		e := &Enum{Property: Property{
			Name: "status",},
			Enum: []EnumValue{
				{Name: "unkown"},
				{Name: "up"},
				{Name: "down"},
			},
		}
		obj := &Object{Property: Property{
			Id: "NetworkLinkInfo"},
			Properties: []Item{e},
		}
		e.parent = obj

		cases = append(cases,
		genTest{
			Name: "Embedded enum",
			Input: e,
			Expected: []*jen.Statement{
				jen.Type().Id("NetworkLinkInfoStatus").String(),
				jen.Const().Defs(
					jen.Id("NetworkLinkInfoStatus_UNKNOWN").Id("NetworkLinkInfoStatus").
						Op("=").Lit("unknown"),
					jen.Id("NetworkLinkInfoStatus_UP").Id("NetworkLinkInfoStatus").
						Op("=").Lit("up"),
					jen.Id("NetworkLinkInfoStatus_DOWN").Id("NetworkLinkInfoStatus").
						Op("=").Lit("down"),
				),
			},
		})
	}
	for _, c := range cases {
		t.Run(c.Name, c.doTest)
	}
}

func TestGenStruct(t *testing.T) {
	cases := []genTest{
		{
			Name: "Simple top level struct",
			Input: &Object{Property: Property{Id: "Foo"},
				Properties: []Item{
					&String{Property: Property{Name: "bar"}},
				},
			},
			Expected: []*jen.Statement{
				jen.Type().Id("Foo").Struct(
					jen.Id("Bar").String().Tag(map[string]string{"json":"bar"}),
				)},
		},
		{
			Name: "Struct with optional struct property",
			Input: &Object{Property: Property{
				Id: "Foo"},
				Properties: []Item{
					&Ref{Property: Property{
						Name:"bar",
						Ref: "Tab",
						Optional: true}},
				},
			},
			Expected: []*jen.Statement{
				jen.Type().Id("Foo").Struct(
					jen.Id("Bar").Op("*").Id("Tab").Tag(map[string]string{"json":"bar,omitempty"}),
				),
			},
		},
		{
			Name: "Struct with optional non struct property",
			Input: &Object{Property: Property{
				Id: "Foo"},
				Properties: []Item{
					&String{Property: Property{
						Name: "bar",
						Optional: true}},
				},
			},
			Expected: []*jen.Statement{
				jen.Type().Id("Foo").Struct(
					jen.Id("Bar").String().Tag(map[string]string{"json":"bar,omitempty"}),
				),
			},
		},
		{
			Name: "Struct with comments",
			Input: &Object{Property: Property{
				Id: "Foo",
				Description: "This is a foo"},
			},
			Expected: []*jen.Statement{
				jen.Comment("This is a foo"),
				jen.Type().Id("Foo").Struct(),
			},
		},
		{
			Name: "Struct with commented properties",
			Input: &Object{Property: Property{
				Id: "Foo"},
				Properties: []Item{
					&String{Property: Property{
						Name: "bar",
						Description: "The type of bar"},
					},
				},
			},
			Expected: []*jen.Statement{
				jen.Type().Id("Foo").Struct(
					jen.Comment("The type of bar"),
					jen.Id("Bar").String().Tag(map[string]string{"json":"bar"}),
				),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, c.doTest)
	}
}
