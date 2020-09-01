package main

import (
	"bytes"
	"fmt"
)

type genGetFreeFloating struct {
	ctx *context
}

func (g *genGetFreeFloating) Run() error {
	ctx := g.ctx

	var buf bytes.Buffer
	for _, typ := range ctx.irPkg.types {
		fmt.Fprintf(&buf, "func (n *%s) GetFreeFloating() *freefloating.Collection { return &n.FreeFloating }\n\n",
			typ.name)
	}

	return ctx.WriteGoFile(codegenFile{
		filename: "freefloating.go",
		pkgPath:  "ir",
		deps: []string{
			"github.com/VKCOM/noverify/src/php/parser/freefloating",
		},
		contents: buf.Bytes(),
	})
}
