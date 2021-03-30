package main

import (
	"bytes"
	"fmt"
	"strings"
)

type genGetFirstToken struct {
	ctx *context
}

func (g *genGetFirstToken) Run() error {
	var buf bytes.Buffer
	ctx := g.ctx

	buf.WriteString("func GetFirstToken(n Node) *token.Token {\n")
	buf.WriteString("  switch n := n.(type) {\n")
	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "  case *%s:\n", typ.name)
		g.writeGet(&buf, ctx.irPkg, typ)
	}
	buf.WriteString("  default:\n")
	buf.WriteString("    panic(fmt.Sprintf(`unhandled type %T`, n))\n")
	buf.WriteString("  }\n")
	buf.WriteString("  return nil\n")
	buf.WriteString("}\n")

	return ctx.WriteGoFile(codegenFile{
		filename: "get_first_token.go",
		pkgPath:  "ir",
		deps: []string{
			"fmt",
			"github.com/z7zmey/php-parser/pkg/token",
		},
		contents: buf.Bytes(),
	})
}

func (g *genGetFirstToken) writeGet(w *bytes.Buffer, pkg *packageData, typ *typeData) {
	for i := 0; i < typ.info.NumFields(); i++ {
		field := typ.info.Field(i)
		switch typeString := field.Type().String(); typeString {
		case "*github.com/z7zmey/php-parser/pkg/token.Token":
			fmt.Fprintf(w, "    if n.%s != nil {\n", field.Name())
			fmt.Fprintf(w, "        return n.%s\n", field.Name())
			fmt.Fprintf(w, "    }\n")
		case "[]*github.com/z7zmey/php-parser/pkg/token.Token":
			fmt.Fprintf(w, "    if len(n.%s) != 0 {\n", field.Name())
			fmt.Fprintf(w, "        return n.%s[0]\n", field.Name())
			fmt.Fprintf(w, "    }\n")
		case "[]ir.Node":
			fmt.Fprintf(w, "    if len(n.%s) != 0 {\n", field.Name())
			fmt.Fprintf(w, "        if n.%s[0] != nil {\n", field.Name())
			fmt.Fprintf(w, "            return GetFirstToken(n.%s[0])\n", field.Name())
			fmt.Fprintf(w, "        }\n")
			fmt.Fprintf(w, "    }\n")
		case "ir.Node":
			fmt.Fprintf(w, "    if n.%s != nil {\n", field.Name())
			fmt.Fprintf(w, "        return GetFirstToken(n.%s)\n", field.Name())
			fmt.Fprintf(w, "    }\n")
		case "*github.com/z7zmey/php-parser/pkg/position.Position":
			// Do nothing.
		case "[]github.com/VKCOM/noverify/src/phpdoc.CommentPart":
			// Do nothing.
		case "string", "bool":
			// Do nothing.
		case "ir.Doc":
			// Do nothing.
		case "ir.Class":
			// Do nothing.
		default:
			if strings.HasPrefix(typeString, "[]") {
				fmt.Fprintf(w, "    if n.%s[0] != nil {\n", field.Name())
				fmt.Fprintf(w, "        return GetFirstToken(n.%s[0])\n", field.Name())
				fmt.Fprintf(w, "    }\n")
				continue
			}

			fmt.Fprintf(w, "    if n.%s != nil {\n", field.Name())
			fmt.Fprintf(w, "        return GetFirstToken(n.%s)\n", field.Name())
			fmt.Fprintf(w, "    }\n")
		}
	}
}
