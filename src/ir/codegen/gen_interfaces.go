package main

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
)

type getInterfaces struct {
	ctx *context
}

type interfaceRule struct {
	name       string
	ret        string
	typeString string
	endWith    string
}

var interfaces = []interfaceRule{
	{
		name: "Name",
		ret:  "*Identifier",

		typeString: "*ir.Identifier",
		endWith:    "Name",
	},
	{
		name: "GetAttributes",
		ret:  "[]*AttributeGroup",

		typeString: "[]*ir.AttributeGroup",
		endWith:    "AttrGroups",
	},
	{
		name: "DocComment",
		ret:  "phpdoc.Comment",

		typeString: "github.com/VKCOM/noverify/src/phpdoc.Comment",
		endWith:    "Doc",
	},
	{
		name: "ParamList",
		ret:  "[]Node",

		typeString: "[]ir.Node",
		endWith:    "Params",
	},
	{
		name: "ArgList",
		ret:  "[]Node",

		typeString: "[]ir.Node",
		endWith:    "Args",
	},
	{
		name: "ModifierList",
		ret:  "[]*Identifier",

		typeString: "[]*ir.Identifier",
		endWith:    "Modifiers",
	},
	{
		name: "TypeHint",
		ret:  "Node",

		typeString: "ir.Node",
		endWith:    "Type",
	},
}

func (g *getInterfaces) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer

	buf.WriteString(`
import (
	"github.com/VKCOM/noverify/src/phpdoc"
)
`)

	for _, typ := range ctx.irPkg.types {
		for _, iface := range interfaces {
			g.writeImplementInterface(&buf, iface, typ)
		}
		buf.WriteString("\n")
	}

	return ctx.WriteGoFile(codegenFile{
		filename: "interfaces_implements.go",
		pkgPath:  "ir",
		deps:     []string{},
		contents: buf.Bytes(),
	})
}

func (g *getInterfaces) writeImplementInterface(w *bytes.Buffer, iface interfaceRule, typ *typeData) {

	for i := 0; i < typ.info.NumFields(); i++ {
		field := typ.info.Field(i)

		if field.Embedded() {
			g.writeImplementInterface(w, iface, &typeData{
				name: typ.name,
				info: field.Type().Underlying().(*types.Struct),
			})
			return
		}

		println(field.Type().String())
		if strings.HasSuffix(field.Name(), iface.endWith) && field.Type().String() == iface.typeString {
			fmt.Fprintf(w, "func (n *%s) %s() %s {\n", typ.name, iface.name, iface.ret)
			fmt.Fprintf(w, "  return n.%s\n", field.Name())
			w.WriteString("}\n\n")
			return
		}
	}
}
