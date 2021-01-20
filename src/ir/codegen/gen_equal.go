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
			"bytes",
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
		switch typeString := field.Type().String(); typeString {
		case "string", "bool":
			fmt.Fprintf(w, "    if x.%[1]s != y.%[1]s { return false }\n", field.Name())
		case "[]ir.Node": // replace later with *github.com/z7zmey/php-parser/pkg/token.Token
			fmt.Fprintf(w, "    if !NodeSliceEqual(x.%[1]s, y.%[1]s) { return false }\n", field.Name())
		case "*ir.Token": // replace later with []*github.com/z7zmey/php-parser/pkg/token.Token
			fmt.Fprintf(w, "    if (x.%[1]s != nil || y.%[1]s != nil) && x.%[1]s == nil || y.%[1]s == nil { return false }\n", field.Name())
			fmt.Fprintf(w, "    if x.%[1]s != nil && y.%[1]s != nil && !bytes.Equal(x.%[1]s.Value, y.%[1]s.Value) { return false }\n", field.Name())
		case "[]*ir.Token":
			fmt.Fprintf(w, "    for i := range x.%s {\n", field.Name())
			fmt.Fprintf(w, "      if (x.%[1]s != nil || y.%[1]s != nil) && x.%[1]s[i] == nil || y.%[1]s[i] == nil { return false }\n", field.Name())
			fmt.Fprintf(w, "      if x.%[1]s != nil && y.%[1]s != nil && !bytes.Equal(x.%[1]s[i].Value, y.%[1]s[i].Value) { return false }\n", field.Name())
			fmt.Fprintf(w, "    }\n")
		case "github.com/VKCOM/noverify/src/php/parser/freefloating.Collection":
			// Do nothing.
		case "*github.com/VKCOM/noverify/src/php/parser/position.Position":
			// Do nothing.
		case "[]github.com/VKCOM/noverify/src/phpdoc.CommentPart":
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
