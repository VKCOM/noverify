package main

import (
	"bytes"
	"fmt"
)

type genParent struct {
	ctx *context
}

func (g *genParent) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	for _, typ := range ctx.irPkg.types {
		g.writeParent(&buf, typ)
		buf.WriteString("\n")
	}

	return ctx.WriteGoFile(codegenFile{
		filename: "get_parent.go",
		pkgPath:  "ir",
		deps:     []string{},
		contents: buf.Bytes(),
	})
}

func (g *genParent) writeParent(w *bytes.Buffer, typ *typeData) {
	fmt.Fprintf(w, "func (n *%s) Parent() Node {\n", typ.name)
	w.WriteString("   return n.ParentNode\n")
	w.WriteString("}\n")
}
