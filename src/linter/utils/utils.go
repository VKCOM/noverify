package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

// WalkNode is a convenience wrapper for EnterNode-only traversals.
// It gives a way to traverse a node without defining a new kind of walker.
//
// enterNode function is called in place where EnterNode method would be called.
// If n is nil, no traversal is performed.
func WalkNode(n ir.Node, enterNode func(ir.Node) bool) {
	if n == nil {
		return
	}
	v := nodeVisitor{enterNode: enterNode}
	n.Walk(v)
}

type nodeVisitor struct {
	enterNode func(ir.Node) bool
}

func (v nodeVisitor) LeaveNode(n ir.Node) {}

func (v nodeVisitor) EnterNode(n ir.Node) bool {
	return v.enterNode(n)
}

// HaveMagicMethod checks for the presence of a magic method in the passed class.
func HaveMagicMethod(class string, methodName string) bool {
	_, ok := solver.FindMethod(class, methodName)
	return ok
}

// ClassHasProp checks for the property in the passed class.
func ClassHasProp(className, propName string) bool {
	var nameWithDollar string
	var nameWithoutDollar string
	if strings.HasPrefix(propName, "$") {
		nameWithDollar = propName
		nameWithoutDollar = strings.TrimPrefix(propName, "$")
	} else {
		nameWithDollar = "$" + propName
		nameWithoutDollar = propName
	}

	// Static props stored with leading "$".
	if _, ok := solver.FindProperty(className, nameWithDollar); ok {
		return true
	}
	_, ok := solver.FindProperty(className, nameWithoutDollar)
	return ok
}

// IsQuote reports whether r is quote.
func IsQuote(r rune) bool {
	return r == '"' || r == '\''
}

// IsCapitalized reports whether s starts with an upper case letter.
func IsCapitalized(s string) bool {
	ch, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(ch)
}

func VarToString(v ir.Node) string {
	switch t := v.(type) {
	case *ir.SimpleVar:
		return t.Name
	case *ir.Var:
		return "$" + VarToString(t.Expr)
	case *ir.FunctionCallExpr:
		// TODO: support function calls here :)
		return ""
	case *ir.String:
		// Things like ${"x"}
		return "${" + t.Value + "}"
	default:
		return ""
	}
}

type FuncCallInfo struct {
	CanAnalyze bool
	Defined    bool
	FqName     string
	Info       meta.FuncInfo
}

// ResolveFunctionCall
// TODO: bundle type solving params somehow.
// We usually need ClassParseState+Scope+[]CustomType.
func ResolveFunctionCall(sc *meta.Scope, st *meta.ClassParseState, customTypes []solver.CustomType, call *ir.FunctionCallExpr) FuncCallInfo {
	var res FuncCallInfo
	res.CanAnalyze = true
	if !meta.IsIndexingComplete() {
		return res
	}

	fqName, ok := solver.GetFuncName(st, call.Function)
	if ok {
		res.FqName = fqName
		res.Info, res.Defined = meta.Info.GetFunction(fqName)
	} else {
		solver.ExprTypeCustom(sc, st, call.Function, customTypes).Iterate(func(typ string) {
			if res.Defined {
				return
			}
			m, ok := solver.FindMethod(typ, `__invoke`)
			res.Info = m.Info
			res.Defined = ok
		})
		if !res.Defined {
			res.CanAnalyze = false
		}
	}

	return res
}

// FindVarNode returns expression variable node root.
// If expression doesn't start from a variable, returns nil.
func FindVarNode(n ir.Node) ir.Node {
	switch n := n.(type) {
	case *ir.Var, *ir.SimpleVar:
		return n
	case *ir.PropertyFetchExpr:
		return FindVarNode(n.Variable)
	case *ir.ArrayDimFetchExpr:
		return FindVarNode(n.Variable)
	default:
		return nil
	}
}

func BinaryOpString(n ir.Node) string {
	switch n.(type) {
	case *ir.BitwiseAndExpr:
		return "&"
	case *ir.BitwiseOrExpr:
		return "|"
	case *ir.BitwiseXorExpr:
		return "^"
	case *ir.LogicalAndExpr:
		return "and"
	case *ir.BooleanAndExpr:
		return "&&"
	case *ir.LogicalOrExpr:
		return "or"
	case *ir.BooleanOrExpr:
		return "||"
	case *ir.LogicalXorExpr:
		return "xor"
	case *ir.PlusExpr:
		return "+"
	case *ir.MinusExpr:
		return "-"
	case *ir.MulExpr:
		return "*"
	case *ir.DivExpr:
		return "/"
	case *ir.ModExpr:
		return "%"
	case *ir.PowExpr:
		return "**"
	case *ir.EqualExpr:
		return "=="
	case *ir.NotEqualExpr:
		return "!="
	case *ir.IdenticalExpr:
		return "==="
	case *ir.NotIdenticalExpr:
		return "!=="
	case *ir.SmallerExpr:
		return "<"
	case *ir.SmallerOrEqualExpr:
		return "<="
	case *ir.GreaterExpr:
		return ">"
	case *ir.GreaterOrEqualExpr:
		return ">="
	case *ir.SpaceshipExpr:
		return "<=>"
	default:
		return ""
	}
}

// PhpKeywords list taken from https://wiki.php.net/rfc/context_sensitive_lexer
var PhpKeywords = map[string]bool{
	"callable":     true,
	"class":        true,
	"trait":        true,
	"extends":      true,
	"implements":   true,
	"static":       true,
	"abstract":     true,
	"final":        true,
	"public":       true,
	"protected":    true,
	"private":      true,
	"const":        true,
	"enddeclare":   true,
	"endfor":       true,
	"endforeach":   true,
	"endif":        true,
	"endwhile":     true,
	"and":          true,
	"global":       true,
	"goto":         true,
	"instanceof":   true,
	"insteadof":    true,
	"interface":    true,
	"namespace":    true,
	"new":          true,
	"or":           true,
	"xor":          true,
	"try":          true,
	"use":          true,
	"var":          true,
	"exit":         true,
	"list":         true,
	"clone":        true,
	"include":      true,
	"include_once": true,
	"throw":        true,
	"array":        true,
	"print":        true,
	"echo":         true,
	"require":      true,
	"require_once": true,
	"return":       true,
	"else":         true,
	"elseif":       true,
	"default":      true,
	"break":        true,
	"continue":     true,
	"switch":       true,
	"yield":        true,
	"function":     true,
	"if":           true,
	"endswitch":    true,
	"finally":      true,
	"for":          true,
	"foreach":      true,
	"declare":      true,
	"case":         true,
	"do":           true,
	"while":        true,
	"as":           true,
	"catch":        true,
	"die":          true,
	"self":         true,
	"parent":       true,
}
