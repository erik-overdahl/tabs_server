package schema

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/erik-overdahl/tabs_server/internal/util"
)

type GoItem interface {
	Item
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
