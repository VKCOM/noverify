package main

import (
	"bytes"
	"fmt"
	"go/types"
)

type genWalk struct {
	ctx *context
}

func (g *genWalk) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	for _, typ := range ctx.irPkg.types {
		g.writeWalk(&buf, typ)
		buf.WriteString("\n")
	}

	return ctx.WriteGoFile(codegenFile{
		filename: "walk.go",
		pkgPath:  "ir",
		deps:     []string{},
		contents: buf.Bytes(),
	})
}

func (g *genWalk) writeWalk(w *bytes.Buffer, typ *typeData) {
	fmt.Fprintf(w, "func (n *%s) Walk(v Visitor) {\n", typ.name)
	w.WriteString("  if !v.EnterNode(n) { return }\n")
	g.writeFieldsWalk(w, typ.info)
	w.WriteString("  v.LeaveNode(n)\n")
	w.WriteString("}\n")
}

func (g *genWalk) writeFieldsWalk(w *bytes.Buffer, typ *types.Struct) {
	for i := 0; i < typ.NumFields(); i++ {
		g.writeFieldWalk(w, typ.Field(i))
	}
}

func (g *genWalk) writeFieldWalk(w *bytes.Buffer, field *types.Var) {
	// Embedded structs are handles as a normal members of the struct.
	if field.Embedded() {
		g.writeFieldsWalk(w, field.Type().Underlying().(*types.Struct))
		return
	}

	// Slices of nodes are walked via for loops.
	slice, ok := field.Type().Underlying().(*types.Slice)
	if ok {
		if !types.Implements(slice.Elem(), g.ctx.nodeIface) {
			return
		}
		fmt.Fprintf(w, "  for _, nn := range n.%s { nn.Walk(v) }\n", field.Name())
		return
	}

	// Skip fields that don't need to be traversed.
	if !types.Implements(field.Type(), g.ctx.nodeIface) {
		return
	}
	fmt.Fprintf(w, "  if n.%[1]s != nil { n.%[1]s.Walk(v) }\n", field.Name())
}
