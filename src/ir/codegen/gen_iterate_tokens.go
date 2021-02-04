package main

import (
	"bytes"
	"fmt"
)

type genIterate struct {
	ctx *context
}

func (g *genIterate) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "func (n *%s) IterateTokens(cb func (*Token) bool) {\n", typ.name)
		g.writeIterate(&buf, ctx.irPkg, typ)
		fmt.Fprintf(&buf, "}\n\n")
	}

	return ctx.WriteGoFile(codegenFile{
		filename: "iterate.go",
		pkgPath:  "ir",
		deps:     []string{},
		contents: buf.Bytes(),
	})
}

func (g *genIterate) writeIterate(w *bytes.Buffer, pkg *packageData, typ *typeData) {
	for i := 0; i < typ.info.NumFields(); i++ {
		field := typ.info.Field(i)
		switch typeString := field.Type().String(); typeString {
		case "*github.com/z7zmey/php-parser/pkg/token.Token": // replace later with *github.com/z7zmey/php-parser/pkg/token.Token
			fmt.Fprintf(w, "    if !cb(n.%[1]s) {\n", field.Name())
			fmt.Fprintf(w, "        return\n")
			fmt.Fprintf(w, "    }\n")
		case "[]*github.com/z7zmey/php-parser/pkg/token.Token": // replace later with []*github.com/z7zmey/php-parser/pkg/token.Token
			fmt.Fprintf(w, "    for _, tk := range n.%[1]s {\n", field.Name())
			fmt.Fprintf(w, "        if !cb(tk) {\n")
			fmt.Fprintf(w, "            return\n")
			fmt.Fprintf(w, "        }\n")
			fmt.Fprintf(w, "    }\n")
		}
	}
}
