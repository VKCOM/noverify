package main

import (
	"bytes"
	"fmt"
)

type genIterate struct {
	ctx *context
}

func (g *genIterate) Run() error {
	var buf bytes.Buffer
	ctx := g.ctx

	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "func (n *%s) IterateTokens(cb func (*token.Token) bool) {\n", typ.name)
		g.writeIterate(&buf, ctx.irPkg, typ)
		fmt.Fprintf(&buf, "}\n\n")
	}

	return ctx.WriteGoFile(codegenFile{
		filename: "iterate.go",
		pkgPath:  "ir",
		deps: []string{
			"github.com/VKCOM/php-parser/pkg/token",
		},
		contents: buf.Bytes(),
	})
}

func (g *genIterate) writeIterate(w *bytes.Buffer, pkg *packageData, typ *typeData) {
	for i := 0; i < typ.info.NumFields(); i++ {
		field := typ.info.Field(i)
		if field.Name() == "ParentNode" {
			continue
		}
		switch typeString := field.Type().String(); typeString {
		case "*github.com/VKCOM/php-parser/pkg/token.Token":
			fmt.Fprintf(w, "    if !traverseToken(n.%s, cb) {\n", field.Name())
			fmt.Fprintf(w, "        return\n")
			fmt.Fprintf(w, "    }\n")
		case "[]*github.com/VKCOM/php-parser/pkg/token.Token":
			fmt.Fprintf(w, "    for _, tk := range n.%s {\n", field.Name())
			fmt.Fprintf(w, "        if !traverseToken(tk, cb) {")
			fmt.Fprintf(w, "            return\n")
			fmt.Fprintf(w, "        }\n")
			fmt.Fprintf(w, "    }\n")
		}
	}
}
