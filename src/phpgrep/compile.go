package phpgrep

import (
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

type compiler struct {
	src []byte
}

func compile(opts *Compiler, pattern []byte) (*Matcher, error) {
	root, src, err := parsePHP7(pattern)
	if err != nil {
		return nil, err
	}

	if st, ok := root.(*stmt.Expression); ok {
		root = st.Expr
	}

	c := compiler{src: src}
	root.Walk(&c)

	m := &Matcher{m: matcher{root: root}}

	return m, nil
}

func (c *compiler) EnterNode(w walker.Walkable) bool {
	v, ok := w.(*node.Var)
	if !ok {
		return true
	}
	s, ok := v.Expr.(*scalar.String)
	if !ok {
		return true
	}
	value := unquoted(s.Value)

	var name string
	var class string

	colon := strings.Index(value, ":")
	if colon == -1 {
		// Anonymous matcher.
		name = "_"
		class = value
	} else {
		// Named matcher.
		name = value[:colon]
		class = value[colon+len(":"):]
	}

	switch class {
	case "var":
		v.Expr = anyVar{metaNode{name: name}}
	case "int":
		v.Expr = anyInt{metaNode{name: name}}
	case "float":
		v.Expr = anyFloat{metaNode{name: name}}
	case "str":
		v.Expr = anyStr{metaNode{name: name}}
	case "num":
		v.Expr = anyNum{metaNode{name: name}}
	case "expr":
		v.Expr = anyExpr{metaNode{name: name}}
	case "const":
		v.Expr = anyConst{metaNode{name: name}}
	case "func":
		v.Expr = anyFunc{metaNode{name: name}}
	}

	return true
}

func (c *compiler) LeaveNode(w walker.Walkable) {}
