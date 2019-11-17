package linter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/printer"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
)

// FmtNode is used for debug purposes and returns string representation of a specified node.
func FmtNode(n node.Node) string {
	var b bytes.Buffer
	printer.NewPrettyPrinter(&b, " ").Print(n)
	return b.String()
}

// FlagsToString is designed for debugging flags.
func FlagsToString(f int) string {
	var res []string

	if (f & FlagReturn) == FlagReturn {
		res = append(res, "Return")
	}

	if (f & FlagDie) == FlagDie {
		res = append(res, "Die")
	}

	if (f & FlagThrow) == FlagThrow {
		res = append(res, "Throw")
	}

	if (f & FlagBreak) == FlagBreak {
		res = append(res, "Break")
	}

	return "Exit flags: [" + strings.Join(res, ", ") + "], digits: " + fmt.Sprintf("%d", f)
}

func haveMagicMethod(class string, methodName string) bool {
	_, _, ok := solver.FindMethod(class, methodName)
	return ok
}

func isQuote(r rune) bool {
	return r == '"' || r == '\''
}

// walkNode is a convenience wrapper for EnterNode-only traversals.
// It gives a way to traverse a node without defining a new kind of walker.
//
// enterNode function is called in place where EnterNode method would be called.
// If n is nil, no traversal is performed.
func walkNode(n node.Node, enterNode func(walker.Walkable) bool) {
	if n == nil {
		return
	}
	v := nodeVisitor{enterNode: enterNode}
	n.Walk(v)
}

type nodeVisitor struct {
	enterNode func(walker.Walkable) bool
}

func (v nodeVisitor) LeaveNode(w walker.Walkable) {}

func (v nodeVisitor) EnterNode(w walker.Walkable) bool {
	return v.enterNode(w)
}

func varToString(v node.Node) string {
	switch t := v.(type) {
	case *node.SimpleVar:
		return t.Name
	case *node.Var:
		return "$" + varToString(t.Expr)
	case *expr.FunctionCall:
		// TODO: support function calls here :)
		return ""
	case *scalar.String:
		// Things like ${"x"}
		return "${" + t.Value + "}"
	default:
		return ""
	}
}

func typeIsCompatible(actual meta.TypesMap, want phpdoc.TypeExpr) bool {
	// TODO: compare without converting a TypesMap into TypeExpr?
	// Or maybe store TypeExpr inside a TypesMap instead of strings?
	have := typesMapToTypeExpr(actual)
	return typeExprIsCompatible(want, have)
}

var (
	typeEmpty = &phpdoc.NamedType{}
	typeArray = &phpdoc.ArrayType{Elem: &phpdoc.NamedType{Name: "mixed"}}
)

func typesMapToTypeExpr(m meta.TypesMap) phpdoc.TypeExpr {
	// TODO: when ExprType stops returning
	// "empty_array" type, remove the extra check.
	if m.Is("empty_array") {
		return typeArray
	}

	var p phpdoc.TypeParser
	typeExpr, err := p.ParseType(m.String())
	if err != nil {
		return typeEmpty
	}
	return typeExpr
}

// typeExprIsCompatible reports whether val type is compatible with dst type.
func typeExprIsCompatible(dst, val phpdoc.TypeExpr) bool {
	// TODO: allow implementations to be compatible with interfaces.
	// TODO: allow derived classes to be compatible with base classes.

	switch x := dst.(type) {
	case *phpdoc.NamedType:
		switch x.Name {
		case "object":
			// For object we accept any kind of object instance.
			// https://wiki.php.net/rfc/object-typehint
			y, ok := val.(*phpdoc.NamedType)
			return ok && (y.Name == "object" || strings.HasPrefix(y.Name, `\`))
		case "array":
			_, ok := val.(*phpdoc.ArrayType)
			return ok
		}
		y, ok := val.(*phpdoc.NamedType)
		return ok && x.Name == y.Name

	case *phpdoc.NotType:
		return !typeExprIsCompatible(x.Expr, val)

	case *phpdoc.NullableType:
		y, ok := val.(*phpdoc.NullableType)
		return ok && typeExprIsCompatible(x.Expr, y.Expr)

	case *phpdoc.ArrayType:
		y, ok := val.(*phpdoc.ArrayType)
		return ok && typeExprIsCompatible(x.Elem, y.Elem)

	case *phpdoc.UnionType:
		if y, ok := val.(*phpdoc.UnionType); ok {
			return typeExprIsCompatible(x.X, y.X) && typeExprIsCompatible(x.Y, y.Y)
		}
		return typeExprIsCompatible(x.X, val) || typeExprIsCompatible(x.Y, val)

	case *phpdoc.InterType:
		// TODO: make it work as intended. (See #310)
		return false

	default:
		return false
	}
}
