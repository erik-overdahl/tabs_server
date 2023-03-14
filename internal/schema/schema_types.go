package schema

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/erik-overdahl/tabs_server/internal/util"
)

// should this include the pointer if `optional`?
func (p Property) Type() *jen.Statement {
	switch {
	case p.Id != "":
		return jen.Id(p.Id)
	case p.Name != "":
		return jen.Id(util.Exportable(util.SnakeToCamel(p.Name)))
	case p.Ref != "":
		if strings.Contains(p.Ref, ".") {
			pieces := strings.Split(p.Ref, ".")
			return jen.Qual(pieces[0], pieces[1])
		}
		return jen.Id(p.Ref)
	// TODO
	// case p.Extend != "":
	// case p.Import != "":
	default:
		return jen.Any()
	}
}

func (_ Any) Type() *jen.Statement {
	return jen.Any()
}

func (_ Null) Type() *jen.Statement {
	return jen.Nil()
}

func (_ Bool) Type() *jen.Statement {
	return jen.Bool()
}

func (_ Int) Type() *jen.Statement {
	return jen.Int()
}

func (_ Number) Type() *jen.Statement {
	return jen.Float64()
}

func (_ String) Type() *jen.Statement {
	return jen.String()
}

func (this Value) Type() *jen.Statement {
	switch this.Value.(type) {
	case int:
		return jen.Int()
	case float64:
		return jen.Float64()
	case string:
		return jen.String()
	}
	return jen.Any() // should this fall back to Property.Type()?
}

func (this Ref) Type() *jen.Statement {
	if strings.Contains(this.Ref, ".") {
		pieces := strings.Split(this.Ref, ".")
		return jen.Qual(pieces[0], pieces[1])
	}
	return jen.Id(this.Ref)
}

func (this Array) Type() *jen.Statement {
	return jen.Index().Add(this.Items.Type())
}

func (this Object) Type() *jen.Statement {
	// ugh
	if 0 < len(this.Properties) {
		return this.Property.Type()
	} else if this.AdditionalProperties != nil {
		return jen.Map(jen.String()).Add(this.AdditionalProperties.Type())
	} else if 1 == len(this.PatternProperties) {
		return jen.Map(jen.String()).Add(this.PatternProperties[0].Type())
	} else if 1 < len(this.PatternProperties) {
		return jen.Map(jen.String()).Any()
	}
	return this.Property.Type()
}

func (this Choices) Type() *jen.Statement {
	return this.Choose().Type()
}
// TODO
func (this Choices) Choose() Item {
	priority := func(t Item) int {
		switch t.(type) {
		case *Array:
			return 9
		case *Ref:
			return 8
		case *Enum:
			return 7
		case *Object:
			return 6
		case *String:
			return 5
		case *Bool:
			return 4
		case *Number:
			return 3
		case *Int:
			return 2
		case *Null:
			return 1
		}
		return 0
	}
	chosen := this.Choices[0]
	for _, option := range this.Choices {
		preference := priority(option) - priority(chosen)
		if 0 < preference {
			chosen = option
			continue
		} else if preference < 0 {
			continue
		}
		switch c := option.(type) {
		case *Array:
			o := option.(*Array)
			if priority(c.Items) < priority(o.Items) {
				chosen = option
			}
		// merge enums?
		// case *Enum:
		// 	o := option.(*Enum)
		// 	i
		}
	}
	return chosen
}
