package main

import (
	"bytes"
	"fmt"
)

type genGetNodeKind struct {
	ctx *context
}

func (g *genGetNodeKind) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	buf.WriteString("type NodeKind int\n")
	buf.WriteString("const (\n")
	for i, typ := range ctx.irPkg.types {
		if i == 0 {
			fmt.Fprintf(&buf, "  Kind%s NodeKind = iota\n", typ.name)
		} else {
			fmt.Fprintf(&buf, "  Kind%s\n", typ.name)
		}
	}
	buf.WriteString("\n")
	buf.WriteString("  NumKinds")
	buf.WriteString(")\n")

	buf.WriteString("func GetNodeKind(x Node) NodeKind {\n")
	buf.WriteString("  switch x := x.(type) {\n")
	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "  case *%s:\n", typ.name)
		fmt.Fprintf(&buf, "    return Kind%s\n", typ.name)
	}
	buf.WriteString("  default:\n")
	buf.WriteString("    panic(fmt.Sprintf(`unhandled type %T`, x))\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n")

	return ctx.WriteGoFile(codegenFile{
		filename: "get_node_kind.go",
		pkgPath:  "ir",
		deps: []string{
			"fmt",
		},
		contents: buf.Bytes(),
	})
}
