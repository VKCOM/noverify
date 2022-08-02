package main

import (
	"bytes"
	"fmt"
	"strings"
)

type genEqual struct {
	ctx *context
}

func (g *genEqual) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	buf.WriteString("func NodeEqual(x, y ir.Node) bool {\n")
	buf.WriteString("  if x == nil || y == nil { return x == y }\n")
	buf.WriteString("  switch x := x.(type) {\n")
	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "  case *ir.%s:\n", typ.name)
		g.writeCompare(&buf, ctx.irPkg, typ)
		buf.WriteString("    return true\n")
	}
	buf.WriteString("  default:\n")
	buf.WriteString("    panic(fmt.Sprintf(`unhandled type %T`, x))\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n")

	return ctx.WriteGoFile(codegenFile{
		filename: "equal.go",
		pkgPath:  "ir/irutil",
		deps: []string{
			"fmt",
			"github.com/VKCOM/noverify/src/ir",
		},
		contents: buf.Bytes(),
	})
}

func (g *genEqual) writeCompare(w *bytes.Buffer, pkg *packageData, typ *typeData) {
	fmt.Fprintf(w, "    y, ok := y.(*ir.%s)\n", typ.name)
	w.WriteString("    if !ok || x == nil || y == nil { return x == y }\n")
	for i := 0; i < typ.info.NumFields(); i++ {
		field := typ.info.Field(i)
		if field.Name() == "ParentNode" {
			continue
		}

		switch typeString := field.Type().String(); typeString {
		case "string", "bool":
			fmt.Fprintf(w, "    if x.%[1]s != y.%[1]s { return false }\n", field.Name())
		case "[]ir.Node":
			fmt.Fprintf(w, "    if !NodeSliceEqual(x.%[1]s, y.%[1]s) { return false }\n", field.Name())
		case "github.com/VKCOM/noverify/src/phpdoc.Comment":
			fmt.Fprintf(w, "    if x.Doc.Raw != y.Doc.Raw { return false }\n")
		case "*github.com/VKCOM/php-parser/pkg/token.Token":
			// Do nothing.
		case "[]*github.com/VKCOM/php-parser/pkg/token.Token":
			// Do nothing.
		case "*github.com/VKCOM/php-parser/pkg/position.Position":
			// Do nothing.
		case "[]github.com/VKCOM/noverify/src/phpdoc.CommentPart":
			// Do nothing.
		case "ir.String":
			// Do nothing.
		case "ir.Class":
			fmt.Fprintf(w, "    if !classEqual(x.%[1]s, y.%[1]s) { return false }\n", field.Name())
		default:
			if !strings.HasPrefix(typeString, "[]") {
				fmt.Fprintf(w, "    if !NodeEqual(x.%[1]s, y.%[1]s) { return false }\n", field.Name())
				continue
			}
			fmt.Fprintf(w, "    if len(x.%[1]s) != len(y.%[1]s) { return false }\n", field.Name())
			fmt.Fprintf(w, "    for i := range x.%s {\n", field.Name())
			fmt.Fprintf(w, "      if !NodeEqual(x.%[1]s[i], y.%[1]s[i]) { return false }\n", field.Name())
			fmt.Fprintf(w, "    }\n")
		}
	}
}
