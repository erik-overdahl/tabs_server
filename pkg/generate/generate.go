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

func (pkg *Pkg) AddAlias(item SchemaItem) {
	f := pkg.TypeFile
	switch item := item.(type) {
	case *SchemaObjectProperty:
		if 0 < len(item.PatternProperties) {
			// TODO
			f.Type().Id(exportable(snakeToCamel(item.Name))).
				Map(String()).Any()
		}
	default:
		var name string
		if item.Base().Name != "" {
			name = exportable(snakeToCamel(item.Base().Name))
		} else if item.Base().Id != "" {
			name = exportable(item.Base().Id)
		}
		f.Type().Id(exportable(snakeToCamel(name))).Add(getPropertyType(item))
	}
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
				g.Comment(e.Description)
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
	var paramItems, returnItems []SchemaItem
	var callback SchemaItem
	for _, param := range item.Parameters {
		if param.Base().Name == "callback" || param.Base().Name == "responseCallback" {
			callback = param
			continue
		}
		paramItems = append(paramItems, param)
		// switch t := param.(type) {
		// case *SchemaArrayProperty:

		// }
	}

	if item.Returns != nil {
		returnItems = append(returnItems, item.Returns)
	} else if item.Async && callback != nil {
		c, ok := callback.(*SchemaFunctionProperty)
		if !ok {
			return fmt.Errorf("function '%s': callback: expected %T, got %T", item.Name, c, callback)
		}
		for _, param := range c.Parameters {
			// create a struct if necessary
			// if obj, ok := param.(*SchemaObjectProperty); ok {
			// 	continue
			// }
			returnItems = append(returnItems, param)
		}
	}

	if item.Description != "" {
		pkg.ClientFile.Comment(item.Description)
	}
	pkg.ClientFile.Func().Params(jen.Id("client").Op("*").Id("Client")).
		Id(exportable(item.Name)).
		ParamsFunc(func(g *jen.Group) {
			for _, param := range paramItems {
				g.Add(funcParamId(param).Add(getPropertyType(param)))
			}
		}).
		ParamsFunc(func(g *jen.Group) {
			if len(returnItems) > 0 {
				retItem := returnItems[0]
				if retItem.Base().Name != "" {
					g.Add(funcParamId(retItem)).Add(getPropertyType(retItem))
				} else {
					g.Id("result").Add(getPropertyType(retItem))
				}
			}
			g.Err().Error()
		}).
		BlockFunc(func(g *jen.Group) {
			var unmarshal jen.Code
			data := jen.Nil()
			varname := "_"
			// in practice, only ever 1 of these
			if len(paramItems) > 0 {
				structValues := jen.Dict{}
				data = jen.StructFunc(func(g *jen.Group) {
					for _, param := range paramItems {
						tag := param.Base().Name
						if param.Base().Optional {
							tag += ",omitempty"
						}
						structId := jen.Id(exportable(param.Base().Name))
						structValues[structId.Clone()] = funcParamId(param)
						g.Add(structId).Add(getPropertyType(param)).
							Tag(map[string]string{"json": tag})
					}
				}).Values(structValues)
			}
			if len(returnItems) > 0 {
				varname = "response"
				unmarshal = jen.Else().If(
					jen.Err().Op(":=").Qual("json", "Unmarshal").
						CallFunc(func(g *jen.Group) {
							g.Id(varname).Dot("Data")
							if returnItems[0].Base().Name != "" {
								g.Op("&").Add(funcParamId(returnItems[0]))
							} else {
								g.Op("&").Add(jen.Id("result"))
							}
						}),
					jen.Err().Op("!=").Nil(),
				).Block(jen.Return(jen.Nil(), jen.Err()))
			}
			g.If(
				jen.List(jen.Id(varname), jen.Err()).Op(":=").
					Id("client").Dot("gateway").Dot("Request").
					Call(jen.Lit(item.Name), data),
				jen.Err().Op("!=").Nil()).
				Block(jen.Return(jen.Nil(), jen.Err())).
				Add(unmarshal)
			g.Return()
		})
	return nil
}

func funcParamId(param SchemaItem) *jen.Statement {
	name := param.Base().Name
	if jen.IsReservedWord(name) {
		name = "_" + name
	}
	return jen.Id(name)
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

func (pkg *Pkg) AddEvent(item *SchemaFunctionProperty) {
	// f := pkg.TypeFile
	if 1 < len(item.Parameters) {
		// collect those parameters into an object
		_struct := &SchemaObjectProperty{SchemaProperty:
			&SchemaProperty{Name: exportable(item.Name) + "Event"},
			Properties: item.Parameters,
		}
		pkg.AddStruct(_struct, "")
	} else if 1 == len(item.Parameters) {
		if ret, ok := item.Parameters[0].(*SchemaObjectProperty); ok {
			pkg.AddStruct(ret, exportable(item.Name) + exportable(ret.Name))
		}
	}
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
