package phpgrep

import (
	"bytes"
	"errors"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func findNamed(capture []CapturedNode, name string) (node.Node, bool) {
	for _, c := range capture {
		if c.Name == name {
			return c.Node, true
		}
	}
	return nil, false
}

func getNodePos(n node.Node) *position.Position {
	pos := n.GetPosition()
	if pos == nil {
		// FIXME: investigate how and why we're getting nil position for some nodes.
		// See #24.
		return nil
	}
	if pos.EndPos < 0 || pos.StartPos < 0 {
		// FIXME: investigate why we sometimes get out-of-range pos ranges.
		// We also get negative EndPos for some nodes, which is awkward.
		// See #24.
		return nil
	}
	return pos
}

func unquoted(s string) string {
	return s[1 : len(s)-1]
}

func nodeIsVar(n node.Node) bool {
	switch n.(type) {
	case *node.SimpleVar, *node.Var:
		return true
	default:
		return false
	}
}

func nodeIsExpr(n node.Node) bool {
	switch n.(type) {
	case *assign.Assign,
		*assign.BitwiseAnd,
		*assign.BitwiseOr,
		*assign.BitwiseXor,
		*assign.Concat,
		*assign.Div,
		*assign.Minus,
		*assign.Mod,
		*assign.Mul,
		*assign.Plus,
		*assign.Pow,
		*assign.Reference,
		*assign.ShiftLeft,
		*assign.ShiftRight,
		*node.Var,
		*node.SimpleVar,
		*binary.BitwiseAnd,
		*binary.BitwiseOr,
		*binary.BitwiseXor,
		*binary.BooleanAnd,
		*binary.BooleanOr,
		*binary.Coalesce,
		*binary.Concat,
		*binary.Div,
		*binary.Equal,
		*binary.Greater,
		*binary.GreaterOrEqual,
		*binary.Identical,
		*binary.LogicalAnd,
		*binary.LogicalOr,
		*binary.LogicalXor,
		*binary.Minus,
		*binary.Mod,
		*binary.Mul,
		*binary.NotEqual,
		*binary.NotIdentical,
		*binary.Plus,
		*binary.Pow,
		*binary.ShiftLeft,
		*binary.ShiftRight,
		*binary.Smaller,
		*binary.SmallerOrEqual,
		*binary.Spaceship,
		*expr.Array,
		*expr.ArrayDimFetch,
		*expr.ArrayItem,
		*expr.BitwiseNot,
		*expr.BooleanNot,
		*expr.ClassConstFetch,
		*expr.Clone,
		*expr.Closure,
		*expr.ClosureUse,
		*expr.ConstFetch,
		*expr.Empty,
		*expr.ErrorSuppress,
		*expr.Eval,
		*expr.Exit,
		*expr.FunctionCall,
		*expr.Include,
		*expr.IncludeOnce,
		*expr.InstanceOf,
		*expr.Isset,
		*expr.MethodCall,
		*expr.New,
		*expr.PostDec,
		*expr.PreInc,
		*expr.Print,
		*expr.PropertyFetch,
		*expr.Reference,
		*expr.Require,
		*expr.RequireOnce,
		*expr.ShellExec,
		*expr.StaticCall,
		*expr.StaticPropertyFetch,
		*expr.Ternary,
		*expr.UnaryMinus,
		*expr.UnaryPlus,
		*expr.Yield,
		*expr.YieldFrom,
		*scalar.Dnumber,
		*scalar.Encapsed,
		*scalar.EncapsedStringPart,
		*scalar.Heredoc,
		*scalar.Lnumber,
		*scalar.MagicConstant,
		*scalar.String,
		*stmt.Expression:
		return true

	default:
		return false
	}
}

func matchMetaVar(n node.Node, s string) bool {
	switch n := n.(type) {
	case *expr.ArrayItem:
		return n.Key == nil && matchMetaVar(n.Val, s)
	case *stmt.Expression:
		return matchMetaVar(n.Expr, s)
	case *node.Argument:
		return matchMetaVar(n.Expr, s)

	case *node.Var:
		nm, ok := n.Expr.(*scalar.String)
		return ok && unquoted(nm.Value) == s

	default:
		return false
	}
}

func parsePHP7(code []byte) (node.Node, []byte, error) {
	if bytes.HasPrefix(code, []byte("<?")) || bytes.HasPrefix(code, []byte("<?php")) {
		n, err := parsePHP7root(code)
		return n, code, err
	}
	return parsePHP7expr(code)
}

func parsePHP7expr(code []byte) (node.Node, []byte, error) {
	code = append([]byte("<?php "), code...)
	code = append(code, ';')
	root, err := parsePHP7root(code)
	if err != nil {
		return nil, code, err
	}
	stmts := root.(*node.Root).Stmts
	if len(stmts) == 0 {
		return &stmt.Nop{}, code, nil
	}
	return root.(*node.Root).Stmts[0], code, nil
}

func parsePHP7root(code []byte) (node.Node, error) {
	p := php7.NewParser(code)
	p.Parse()
	if len(p.GetErrors()) != 0 {
		return nil, errors.New(p.GetErrors()[0].String())
	}
	return p.GetRootNode(), nil
}
