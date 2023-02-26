package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	jen "github.com/dave/jennifer/jen"
)

/*
   Turn a Namespace into a Go package.

   The package name is the name of the Namespace.

   Namespace Properties become exported package-level `vars` and `const`
   elements

   If there are Functions, Events, or Types with methods, a `Client`
   struct is added to the package. The `Client` takes a `gateway.Client`
   in its constructor, which allows the usage of multiple APIs over the
   same connection.
*/

func RenderGo(code *jen.File) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := code.Render(b); err != nil {
		return nil, err
	}
	// return format.Source(b.Bytes())
	return b.Bytes(), nil
}

func BuildGo(item SchemaItem) (jen.Code, error) {
	return nil, nil
}

type Pkg struct {
	Name                 string
	TypeFile, ClientFile *jen.File
	// Comments, Vars, Consts, Types, Funcs *jen.Statement
}

func MakePkg(name string) *Pkg {
	if strings.Contains(name, ".") {
		path := strings.Split(name, ".")
		name = path[len(path)-1]
	}
	return &Pkg{
		Name:       name,
		TypeFile:   jen.NewFile(name),
		ClientFile: jen.NewFile(name),
	}
}

func (pkg *Pkg) AddNamespaceProperties(props []SchemaItem) error {
	f := pkg.TypeFile
	for _, prop := range props {
		name := exportable(prop.Base().Name)
		switch prop := prop.(type) {
		case *SchemaValueProperty:
			if prop.Description != "" {
				f.Comment(prop.Description)
			}
			f.Const().Id(name).Op("=").Lit(prop.Value)

		case *SchemaRefProperty:
			if prop.Description != "" {
				f.Comment(prop.Description)
			}
			ref := strings.Split(prop.Ref, ".")
			if len(ref) == 2 {
				f.Var().Id(name).Qual(ref[0], ref[1])
			} else {
				f.Var().Id(name).Id(prop.Ref)
			}
		// case *SchemaObjectProperty:
		// 	for _, p := range prop.Properties {

		// 	}
		case *SchemaStringProperty:
			alias := exportable(prop.Name)
			if len(prop.Enum) > 0 {
				pkg.AddEnum(prop)
			} else {
				f.Var().Id(alias).String()
			}
		}
	}
	return nil
}

func (pkg *Pkg) AddEnum(enum *SchemaStringProperty) error {
	f := pkg.TypeFile
	var name string
	if enum.Base().Name != "" {
		name = exportable(enum.Base().Name)
	} else if enum.Base().Id != "" {
		name = exportable(enum.Base().Id)
	}
	pkg.TypeFile.Type().Id(name).String()
	f.Const().DefsFunc(func(g *jen.Group) {
		for _, e := range enum.Enum {
			if e.Description != "" {
				f.Comment(e.Description)
			}
			sanitized := strings.ReplaceAll(e.Name, "-", "_")
			g.Id(name + "_" + strings.ToUpper(camelToSnake(sanitized))).
				Id(name).Op("=").Lit(e.Name)
		}
	})
	return nil
}

func (pkg *Pkg) AddClient() {
	f := pkg.ClientFile
	f.Type().Id("Client").Struct(
		jen.Id("gateway").Op("*").Qual("gateway", "Client"),
	)
	f.Func().Id("MakeClient").Params(
		jen.Id("gateway").Op("*").Qual("gateway", "Client"),
	).Op("*").Id("Client").Block(
		jen.Return(jen.Op("&").Id("Client").Values(
			jen.Dict{
				jen.Id("gateway"): jen.Id("gateway"),
			},
		)),
	)
}

func (pkg *Pkg) AddFunction(item *SchemaFunctionProperty) error {
	params := []jen.Code{}
	var callback SchemaItem
	for _, param := range item.Parameters {
		if param.Base().Name == "callback" || param.Base().Name == "responseCallback" {
			callback = param
			continue
		}
		p := jen.Id(param.Base().Name).Add(getPropertyType(param))
		params = append(params, p)
	}

	returns := jen.Error()
	if item.Returns != nil {
		returns = jen.Params(getPropertyType(item.Returns), jen.Error())
	} else if item.Async && callback != nil {
		c, ok := callback.(*SchemaFunctionProperty)
		if !ok {
			return fmt.Errorf("function '%s': callback: expected %T, got %T", item.Name, c, callback)
		}
		if len(c.Parameters) > 0 {
			returns = jen.ParamsFunc(func(g *jen.Group) {
				for _, param := range c.Parameters {
					// if obj, ok := param.(*SchemaObjectProperty); ok {

					// 	continue
					// }
					g.Add(jen.Id(param.Base().Name).
						Add(getPropertyType(param)))
				}
				g.Add(jen.Err().Error())
			})
		}
	}

	if item.Description != "" {
		pkg.ClientFile.Comment(item.Description)
	}
	pkg.ClientFile.Func().Params(jen.Id("client").Op("*").Id("Client")).
		Id(exportable(item.Name)).
		Params(params...).Add(returns).
		Block().Do(func(s *jen.Statement) {
		if len(*returns) > 1 {
			// for i := 0; i < len(*returns) - 2; i++ {
			// 	s.Var().
			// }
		}
	})
	return nil
}

func (pkg *Pkg) AddStruct(item *SchemaObjectProperty, name string) {
	f := pkg.TypeFile
	if item.Description != "" {
		f.Comment(item.Description)
	}

	switch {
	case name != "":
	case item.Name != "":
		name = item.Name
	case item.Id != "":
		name = item.Id
	}

	f.Type().Id(exportable(name)).StructFunc(func(g *jen.Group) {
		for _, raw := range item.Properties {
			info := raw.Base()
			if info.Description != "" {
				g.Comment(info.Description)
			}
			tag := info.Name
			if info.Optional {
				tag += ",omitempty"
			}
			g.Id(exportable(info.Name)).Add(getPropertyType(raw)).
				Tag(map[string]string{"json": tag})
		}
		if item.AdditionalProperties != nil {
			g.Id("AdditionalProperties").Map(jen.String()).Any().
				Tag(map[string]string{"json": "additionalProperties,omitempty"})
		}
	})
}

func buildEvent(item SchemaFunctionProperty) (jen.Code, error) {
	out := jen.Func()
	return out, nil
}

func getPropertyType(item SchemaItem) jen.Code {
	switch item := item.(type) {
	case *SchemaObjectProperty, *SchemaRefProperty, *SchemaProperty:
		p := item.Base()
		var typ jen.Code
		switch {
		case p.Id != "":
			typ = jen.Id(exportable(p.Id))
		case p.Ref != "":
			if strings.Contains(p.Ref, ".") {
				pieces := strings.Split(p.Ref, ".")
				typ = jen.Qual(pieces[0], pieces[1])
			} else {
				typ = jen.Id(exportable(p.Ref))
			}
		case p.Import != "":
			typ = jen.Id(exportable(p.Import))
		case p.Name != "":
			typ = jen.Id(exportable(p.Name))
		default:
			if o, ok := item.(*SchemaObjectProperty); ok && o.IsInstanceOf != "" {
				typ = jen.Id(o.IsInstanceOf)
			}
		}
		if item.Base().Optional {
			return jen.Op("*").Add(typ)
		}
		if typ == nil {
			b, _ := json.Marshal(item)
			fmt.Printf("Where is the type for %s???\n", string(b))

		}
		return typ
	case *SchemaStringProperty:
		return jen.String()
	case *SchemaBoolProperty:
		return jen.Bool()
	case *SchemaFloatProperty:
		return jen.Float64()
	case *SchemaIntProperty:
		return jen.Int()
	case *SchemaArrayProperty:
		return jen.Index().Add(getPropertyType(item.Items))
	case *SchemaChoicesProperty:
		return getPropertyType(chooseType(item.Choices))
	default:
		return jen.Any()
	}
}

func chooseType(choices []SchemaItem) SchemaItem {
	priority := map[string]int{
		"array":    8,
		"string":   7,
		"ref":      6,
		"object":   5,
		"boolean":  4,
		"float64":  3,
		"integer":  2,
		"null":     1,
		"function": 0,
	}
	chosen := choices[0]
	for _, option := range choices {
		preference := priority[option.Type()] - priority[chosen.Type()]
		if 0 < preference {
			chosen = option
			continue
		} else if preference < 0 {
			continue
		}
		switch c := option.(type) {
		// case *SchemaStringProperty:
		// 	o := option.(*SchemaStringProperty)
		case *SchemaArrayProperty:
			o := option.(*SchemaArrayProperty)
			if o.Items == chooseType([]SchemaItem{o.Items, c.Items}) {
				chosen = option
			}
		case *SchemaObjectProperty:

		}
	}
	return chosen
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
