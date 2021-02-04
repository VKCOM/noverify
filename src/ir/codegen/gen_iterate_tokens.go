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

	fmt.Fprintf(&buf, `func handleToken(t *token.Token, cb func(*token.Token) bool) bool {
	if t == nil {
		return true
	}
	
	if !cb(t) {
		return false
	}

	needReturn := true
	for _, ff := range t.FreeFloating {
		needReturn = needReturn && handleToken(ff, cb)
	}

	return needReturn
}

`)

	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "func (n *%s) IterateTokens(cb func (*token.Token) bool) {\n", typ.name)
		g.writeIterate(&buf, ctx.irPkg, typ)
		fmt.Fprintf(&buf, "}\n\n")
	}

	return ctx.WriteGoFile(codegenFile{
		filename: "iterate.go",
		pkgPath:  "ir",
		deps: []string{
			"github.com/z7zmey/php-parser/pkg/token",
		},
		contents: buf.Bytes(),
	})
}

func (g *genIterate) writeIterate(w *bytes.Buffer, pkg *packageData, typ *typeData) {
	for i := 0; i < typ.info.NumFields(); i++ {
		field := typ.info.Field(i)
		switch typeString := field.Type().String(); typeString {
		case "*github.com/z7zmey/php-parser/pkg/token.Token":
			fmt.Fprintf(w, "    handleToken(n.%s, cb)\n", field.Name())
		case "[]*github.com/z7zmey/php-parser/pkg/token.Token":
			fmt.Fprintf(w, "    for _, tk := range n.%s {\n", field.Name())
			fmt.Fprintf(w, "        handleToken(tk, cb)")
			fmt.Fprintf(w, "    }\n")
		}
	}
}
