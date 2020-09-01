package main

import (
	"bytes"
	"fmt"
)

type genGetPosition struct {
	ctx *context
}

func (g *genGetPosition) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	buf.WriteString("func GetPosition(n Node) *position.Position {\n")
	buf.WriteString("  switch n := n.(type) {\n")
	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "  case *%s:\n", typ.name)
		buf.WriteString("    return n.Position\n")
	}
	buf.WriteString("  default:\n")
	buf.WriteString("    panic(fmt.Sprintf(`unhandled type %T`, n))\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n")

	return ctx.WriteGoFile(codegenFile{
		filename: "get_position.go",
		pkgPath:  "ir",
		deps: []string{
			"fmt",
			"github.com/VKCOM/noverify/src/php/parser/position",
		},
		contents: buf.Bytes(),
	})
}
