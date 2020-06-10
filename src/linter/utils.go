package linter

import (
	"fmt"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/printer"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
)

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
	_, ok := solver.FindMethod(class, methodName)
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

func typesMapToTypeExpr(p *phpdoc.TypeParser, m meta.TypesMap) phpdoc.Type {
	typeString := m.String()
	return p.Parse(typeString)
}

// typesIsCompatible reports whether val type is compatible with dst type.
func typeIsCompatible(dst, val phpdoc.TypeExpr) bool {
	// TODO: allow implementations to be compatible with interfaces.
	// TODO: allow derived classes to be compatible with base classes.

	for val.Kind == phpdoc.ExprParen {
		val = val.Args[0]
	}

	switch dst.Kind {
	case phpdoc.ExprParen:
		return typeIsCompatible(dst.Args[0], val)

	case phpdoc.ExprName:
		switch dst.Value {
		case "object":
			// For object we accept any kind of object instance.
			// https://wiki.php.net/rfc/object-typehint
			return val.Kind == dst.Kind && (val.Value == "object" || strings.HasPrefix(val.Value, `\`))
		case "array":
			return val.Kind == phpdoc.ExprArray
		}
		return val.Kind == dst.Kind && dst.Value == val.Value

	case phpdoc.ExprNot:
		return !typeIsCompatible(dst.Args[0], val)

	case phpdoc.ExprNullable:
		return val.Kind == dst.Kind && typeIsCompatible(dst.Args[0], val.Args[0])

	case phpdoc.ExprArray:
		return val.Kind == dst.Kind && typeIsCompatible(dst.Args[0], val.Args[0])

	case phpdoc.ExprUnion:
		if val.Kind == dst.Kind {
			return typeIsCompatible(dst.Args[0], val.Args[0]) &&
				typeIsCompatible(dst.Args[1], val.Args[1])
		}
		return typeIsCompatible(dst.Args[0], val) || typeIsCompatible(dst.Args[1], val)

	case phpdoc.ExprInter:
		// TODO: make it work as intended. (See #310)
		return false

	default:
		return false
	}
}

type funcCallInfo struct {
	canAnalyze bool
	defined    bool
	fqName     string
	info       meta.FuncInfo
}

// TODO: bundle type solving params somehow.
// We usually need ClassParseState+Scope+[]CustomType.
func resolveFunctionCall(sc *meta.Scope, st *meta.ClassParseState, customTypes []solver.CustomType, call *expr.FunctionCall) funcCallInfo {
	var res funcCallInfo
	res.canAnalyze = true
	if !meta.IsIndexingComplete() {
		return res
	}

	fqName, ok := solver.GetFuncName(st, call.Function)
	if ok {
		res.fqName = fqName
		res.info, res.defined = meta.Info.GetFunction(fqName)
	} else {
		solver.ExprTypeCustom(sc, st, call.Function, customTypes).Iterate(func(typ string) {
			if res.defined {
				return
			}
			m, ok := solver.FindMethod(typ, `__invoke`)
			res.info = m.Info
			res.defined = ok
		})
		if !res.defined {
			res.canAnalyze = false
		}
	}

	return res
}

// isCapitalized reports whether s starts with an upper case letter.
func isCapitalized(s string) bool {
	ch, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(ch)
}

// findVarNode returns expression variable node root.
// If expression doesn't start from a variable, returns nil.
func findVarNode(n node.Node) node.Node {
	switch n := n.(type) {
	case *node.Var, *node.SimpleVar:
		return n
	case *expr.PropertyFetch:
		return findVarNode(n.Variable)
	case *expr.ArrayDimFetch:
		return findVarNode(n.Variable)
	default:
		return nil
	}
}

func classHasProp(className, propName string) bool {
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

func binaryOpString(n node.Node) string {
	switch n.(type) {
	case *binary.BitwiseAnd:
		return "&"
	case *binary.BitwiseOr:
		return "|"
	case *binary.BitwiseXor:
		return "^"
	case *binary.LogicalAnd:
		return "and"
	case *binary.BooleanAnd:
		return "&&"
	case *binary.LogicalOr:
		return "or"
	case *binary.BooleanOr:
		return "||"
	case *binary.LogicalXor:
		return "xor"
	case *binary.Plus:
		return "+"
	case *binary.Minus:
		return "-"
	case *binary.Mul:
		return "*"
	case *binary.Div:
		return "/"
	case *binary.Mod:
		return "%"
	case *binary.Pow:
		return "**"
	case *binary.Equal:
		return "=="
	case *binary.NotEqual:
		return "!="
	case *binary.Identical:
		return "==="
	case *binary.NotIdentical:
		return "!=="
	case *binary.Smaller:
		return "<"
	case *binary.SmallerOrEqual:
		return "<="
	case *binary.Greater:
		return ">"
	case *binary.GreaterOrEqual:
		return ">="
	case *binary.Spaceship:
		return "<=>"

	default:
		return ""
	}
}

func findFreeFloatingToken(n node.Node, pos freefloating.Position, s string) bool {
	ff := n.GetFreeFloating()
	if ff == nil {
		return false
	}
	for _, tok := range (*ff)[pos] {
		if tok.StringType != freefloating.TokenType {
			continue
		}
		if tok.Value == s {
			return true
		}
	}
	return false
}

// Returns Node string representation if it's legal for array key.
// Illegal keys have to be processed with individual warning.
func getArrayKeyRepresentation(n node.Node) (view string, ok bool) {
	b := strings.Builder{}
	p := printer.NewPrinter(&b)

	ok = isLegalKey(n)
	p.Print(n)

	return b.String(), ok
}

// Checks if Node type is legal for array key.
func isLegalKey(n node.Node) (ok bool){
	switch n.(type){
	case *binary.BitwiseAnd, *binary.BitwiseOr, *binary.BitwiseXor,
		*binary.BooleanAnd, *binary.BooleanOr, *binary.Coalesce,
		*binary.Concat, *binary.Div, *binary.Equal, *binary.Greater,
		*binary.GreaterOrEqual, *binary.Identical, *binary.LogicalAnd,
		*binary.LogicalOr, *binary.LogicalXor, *binary.Minus, *binary.Mod,
		*binary.Mul, *binary.NotEqual, *binary.NotIdentical, *binary.Plus,
		*binary.Pow, *binary.ShiftLeft, *binary.ShiftRight,
		*binary.SmallerOrEqual, *binary.Smaller, *binary.Spaceship,

		*expr.ArrayDimFetch, *expr.ArrayItem, *expr.BitwiseNot,
		*expr.BooleanNot, *expr.ClassConstFetch, *expr.ConstFetch,
		*expr.Empty, *expr.FunctionCall, *expr.InstanceOf, *expr.Isset,
		*expr.MethodCall, *expr.PostDec, *expr.PostInc, *expr.PreDec,
		*expr.PreInc, *expr.PropertyFetch, *expr.StaticCall,
		*expr.StaticPropertyFetch, *expr.Ternary, *expr.UnaryMinus, *expr.UnaryPlus,

		*node.Var, *node.SimpleVar, *node.Identifier,

		*name.Name, *name.NamePart, *name.FullyQualified, *name.Relative,

		*scalar.Lnumber, *scalar.Dnumber, *scalar.String, *scalar.Heredoc:

		return true
	}

	return false
}

// List taken from https://wiki.php.net/rfc/context_sensitive_lexer
var phpKeywords = map[string]bool{
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
