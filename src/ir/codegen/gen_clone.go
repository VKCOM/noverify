package main

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"
)

type genClone struct {
	ctx *context
}

func (g *genClone) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	buf.WriteString("func NodeClone(x ir.Node) ir.Node {\n")
	buf.WriteString("  if x == nil { return nil }\n")
	buf.WriteString("  switch x := x.(type) {\n")
	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "  case *ir.%s:\n", typ.name)
		g.writeCloneCase(&buf, ctx.irPkg, typ)
	}
	buf.WriteString("  }\n")
	buf.WriteString("  panic(fmt.Sprintf(`unhandled type %T`, x))\n")
	buf.WriteString("}\n")

	return ctx.WriteGoFile(codegenFile{
		filename: "clone.go",
		pkgPath:  "ir/irutil",
		deps: []string{
			"fmt",
			"github.com/VKCOM/noverify/src/ir",
		},
		contents: buf.Bytes(),
	})
}

func (g *genClone) writeAssign(w *bytes.Buffer, pad, lhs, rhs string, typ types.Type) {
	if typ.String() == "ir.Node" {
		// Avoid n.(ir.Node) type asserts that are redundant.
		fmt.Fprintf(w, "%s%s = %s\n", pad, lhs, rhs)
	} else {
		fmt.Fprintf(w, "%s%s = %s.(%s)\n", pad, lhs, rhs, formatType(typ))
	}
}

func (g *genClone) writeCloneCase(w *bytes.Buffer, pkg *packageData, typ *typeData) {
	// This clones all value-type fields.
	w.WriteString("    clone := *x\n")

	for i := 0; i < typ.info.NumFields(); i++ {
		field := typ.info.Field(i)
		switch typeString := field.Type().String(); typeString {
		case "[]ir.Node":
			fmt.Fprintf(w, "    clone.%[1]s = NodeSliceClone(x.%[1]s)\n", field.Name())
		case "*github.com/z7zmey/php-parser/pkg/token.Token":

		case "[]*github.com/z7zmey/php-parser/pkg/token.Token":

		case "github.com/VKCOM/noverify/src/php/parser/freefloating.Collection":
			// Do nothing.
		case "*github.com/z7zmey/php-parser/pkg/position.Position":
			// Do nothing.
		case "[]github.com/VKCOM/noverify/src/phpdoc.CommentPart":
			// Do nothing.
		case "string", "bool":
			// Do nothing.
		case "ir.Class":
			fmt.Fprintf(w, "    clone.%[1]s = classClone(x.%[1]s)\n", field.Name())
		default:
			if !strings.HasPrefix(typeString, "[]") {
				fmt.Fprintf(w, "    if x.%s != nil {\n", field.Name())
				g.writeAssign(w, "      ", "clone."+field.Name(), "NodeClone(x."+field.Name()+")", field.Type())
				fmt.Fprintf(w, "    }\n")
				continue
			}
			elemType := field.Type().(*types.Slice).Elem()
			fmt.Fprintf(w, "    {\n")
			fmt.Fprintf(w, "      sliceClone := make(%s, len(x.%s))\n", formatType(field.Type()), field.Name())
			fmt.Fprintf(w, "      for i := range x.%s {\n", field.Name())
			g.writeAssign(w, "        ", "sliceClone[i]", "NodeClone(x."+field.Name()+"[i])", elemType)
			fmt.Fprintf(w, "      }\n")
			fmt.Fprintf(w, "      clone.%s = sliceClone\n", field.Name())
			fmt.Fprintf(w, "    }\n")
		}
	}
	w.WriteString("    return &clone\n")
}
