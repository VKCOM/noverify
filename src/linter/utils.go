package linter

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
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

// This function returns true if two subtrees are identical
// This function is very primitive, as it works with a small number of possible nodes
func compareSubTreeForArrayItemsKey(firstSubTreeNode node.Node, secondSubTreeNode node.Node) bool {

	// If the node types do not match, then immediately return a false
	if reflect.TypeOf(firstSubTreeNode) != reflect.TypeOf(secondSubTreeNode) {
		return false
	}

	switch reflect.TypeOf(firstSubTreeNode).String() {
	case "*binary.Plus": // 8 + 4
		firstNode := firstSubTreeNode.(*binary.Plus)
		secondNode := secondSubTreeNode.(*binary.Plus)

		// Compare the left and right subtrees
		return compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Left) &&
			compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Right) ||
			// And compare them the other way around.
			compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Right) &&
				compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Left)

	case "*binary.Minus": // 8 - 9
		firstNode := firstSubTreeNode.(*binary.Minus)
		secondNode := secondSubTreeNode.(*binary.Minus)

		// Compare the left and right subtrees
		return compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Left) &&
			compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Right)

	case "*binary.Mul": // 8 * 9
		firstNode := firstSubTreeNode.(*binary.Mul)
		secondNode := secondSubTreeNode.(*binary.Mul)

		// Compare the left and right subtrees
		return compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Left) &&
			compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Right) ||
			// And compare them the other way around.
			compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Right) &&
				compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Left)

	case "*binary.Div": // 8 / 7
		firstNode := firstSubTreeNode.(*binary.Div)
		secondNode := secondSubTreeNode.(*binary.Div)

		// Compare the left and right subtrees
		return compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Left) &&
			compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Right)

	case "*binary.Concat": // "some" . $a
		firstNode := firstSubTreeNode.(*binary.Concat)
		secondNode := secondSubTreeNode.(*binary.Concat)

		// Compare the left and right subtrees
		return compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Left) &&
			compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Right) ||
			// And compare them the other way around.
			compareSubTreeForArrayItemsKey(firstNode.Left, secondNode.Right) &&
				compareSubTreeForArrayItemsKey(firstNode.Right, secondNode.Left)

	case "*scalar.String": // Strings
		firstNode := firstSubTreeNode.(*scalar.String)
		secondNode := secondSubTreeNode.(*scalar.String)

		// Compare values
		return firstNode.Value == secondNode.Value

	case "*scalar.Lnumber": // Integer numbers
		firstNode := firstSubTreeNode.(*scalar.Lnumber)
		secondNode := secondSubTreeNode.(*scalar.Lnumber)

		// Compare values
		return firstNode.Value == secondNode.Value

	case "*scalar.Dnumber": // Real numbers
		firstNode := firstSubTreeNode.(*scalar.Dnumber)
		secondNode := secondSubTreeNode.(*scalar.Dnumber)

		// Note: PHP rounds down real numbers in keys, therefore,
		// keys 14.6 and 14.5 will be one key equal to 14.

		// Therefore, for starters, convert the string value to real
		floatValue1, _ := strconv.ParseFloat(firstNode.Value, 64)
		floatValue2, _ := strconv.ParseFloat(secondNode.Value, 64)

		// And compare the values that were rounded down.
		return math.Floor(floatValue1) == math.Floor(floatValue2)

	case "*expr.ConstFetch": // Constants
		firstNode := firstSubTreeNode.(*expr.ConstFetch)
		secondNode := secondSubTreeNode.(*expr.ConstFetch)

		firstConstantNode := firstNode.Constant.(*name.Name)
		secondConstantNode := secondNode.Constant.(*name.Name)

		return meta.NameToString(firstConstantNode) == meta.NameToString(secondConstantNode)

	case "*expr.ClassConstFetch": // Class constants
		firstNode := firstSubTreeNode.(*expr.ClassConstFetch)
		secondNode := secondSubTreeNode.(*expr.ClassConstFetch)

		return firstNode.ConstantName.Value == secondNode.ConstantName.Value

	case "*node.SimpleVar": // Variables
		firstNode := firstSubTreeNode.(*node.SimpleVar)
		secondNode := secondSubTreeNode.(*node.SimpleVar)

		return firstNode.Name == secondNode.Name

	case "*expr.ArrayDimFetch": // Access to array
		firstNode := firstSubTreeNode.(*expr.ArrayDimFetch)
		secondNode := secondSubTreeNode.(*expr.ArrayDimFetch)

		firstSimpleVarNode := firstNode.Variable.(*node.SimpleVar)
		secondSimpleVarNode := secondNode.Variable.(*node.SimpleVar)

		// If the variable names do not match, then immediately return false
		if firstSimpleVarNode.Name != secondSimpleVarNode.Name {
			return false
		}

		return compareSubTreeForArrayItemsKey(firstNode.Dim, secondNode.Dim)
	}

	return false
}
