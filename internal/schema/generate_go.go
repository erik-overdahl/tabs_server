package schema

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/erik-overdahl/tabs_server/internal/util"
)

type GoItem interface {
	Item
	// no idea why jen.Code interface does not produce the same result
	ToGo() []*jen.Statement
}

func (this Enum) ToGo() []*jen.Statement {
	pieces := []*jen.Statement{}
	if this.Description != "" {
		pieces = append(pieces, jen.Comment(util.Linewrap(this.Description, 70)))
	}

	var name string
	if this.Base().Name != "" {
		name = util.Exportable(util.SnakeToCamel(this.Base().Name))
	} else if this.Base().Id != "" {
		name = util.Exportable(this.Base().Id)
	}
	alias := jen.Type().Id(name).String()
	pieces = append(pieces, alias)

	vals := []jen.Code{}
	for _, e := range this.Enum {
		if e.Description != "" {
			vals = append(vals, jen.Comment(util.Linewrap(e.Description, 70)))
		}
		sanitized := strings.ToUpper(util.CamelToSnake(e.Name))
		sanitized = strings.ReplaceAll(sanitized, "-", "_")
		sanitized = strings.ReplaceAll(sanitized, ".", "_")
		vals = append(vals, jen.Id(name + "_" + sanitized).Id(name).Op("=").Lit(e.Name))
	}
	pieces = append(pieces, jen.Const().Defs(vals...))
	return pieces
}

func (this Object) ToGo() []*jen.Statement {
	props := []jen.Code{}
	for _, prop := range this.Properties {
		info := prop.Base()
		if info.Description != "" {
			props = append(props, jen.Comment(util.Linewrap(info.Description, 73)))
		}
		code := jen.Id(util.Exportable(info.Name))
		tag := info.Name
		if info.Optional {
			switch prop.(type) {
			case *Object, *Ref:
				code.Op("*")
			}
			tag += ",omitempty"
		}
		code.Add(prop.Type()).Tag(map[string]string{"json":tag})
		props = append(props, code)
	}

	pieces := []*jen.Statement{}
	if this.Description != "" {
		pieces = append(pieces, jen.Comment(util.Linewrap(this.Description, 80)))
	}
	def := jen.Type().Id(this.Id).Struct(props...)
	pieces = append(pieces, def)
	return pieces
}

func (this Function) ToGo() []*jen.Statement {
	pieces := []*jen.Statement{}
	if this.Description != "" {
		pieces = append(pieces, jen.Comment(util.Linewrap(this.Description, 80)))
	}

	paramItems := []Item{}
	for _, param := range this.Parameters {
		if param.Base().Name == "callback" || param.Base().Name == "responseCallback" {
			continue
		}
		paramItems = append(paramItems, param)
	}

	params := []jen.Code{}
	requestData := jen.Nil()
	props := []jen.Code{}
	values := jen.Dict{}
	for _, p := range paramItems {
		param := jen.Id(p.Base().Name).Add(p.Type())
		params = append(params, param)

		values[jen.Id(util.Exportable(p.Base().Name))] = jen.Id(p.Base().Name)

		tag := p.Base().Name
		if p.Base().Optional {
			tag += ",omitempty"
		}
		prop := jen.Id(util.Exportable(p.Base().Name)).Add(p.Type()).
			Tag(map[string]string{"json": tag})
		props = append(props, prop)
	}

	if 0 < len(paramItems) {
		requestData = jen.Struct(props...).Values(values)
	}

	returns := []jen.Code{jen.Err().Error()}
	errReturn := []jen.Code{jen.Err()}
	responseVar := "_"
	var unmarshal *jen.Statement

	if this.Returns != nil {
		returns = []jen.Code{jen.Id("result").Add(this.Returns.Type()), jen.Err().Error()}
		errReturn = []jen.Code{jen.Id("result"), jen.Err()}
		responseVar = "response"
		unmarshal = jen.Else().If(
			jen.Err().Op(":=").Qual("json", "Unmarshal").Call(
				jen.Id(responseVar).Dot("Data"),
				jen.Op("&").Id("result"),
			),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Return(errReturn...),
		)
	}

	def := jen.Func().Params(jen.Id("client").Op("*").Id("Client")).
		Id(util.Exportable(this.Name)).
		Params(params...).
		Params(returns...).
		Block(
			jen.If(
				jen.List(jen.Id(responseVar), jen.Err()).Op(":=").
					Id("client").Dot("gateway").Dot("Request").
					Call(jen.Lit(this.Name), requestData),
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Return(errReturn...),
			).Add(unmarshal),
			jen.Return(),
		)
	pieces = append(pieces, def)
	return pieces
}
