package linter

import (
	"fmt"
	"hash/fnv"
	"math"
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

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func (b *BlockWalker) getHashForExpressionNode(x node.Node) (int64, bool) {
	switch x.(type) {
	case *binary.BitwiseAnd:
		y := x.(*binary.BitwiseAnd)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft & hashRight, true

	case *binary.BitwiseOr:
		y := x.(*binary.BitwiseOr)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft | hashRight, true

	case *binary.BitwiseXor:
		y := x.(*binary.BitwiseXor)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft ^ hashRight, true

	case *binary.Plus:
		y := x.(*binary.Plus)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft + hashRight, true

	case *binary.Minus:
		y := x.(*binary.Minus)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft - hashRight, true

	case *binary.Mul:
		y := x.(*binary.Mul)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft * hashRight, true

	case *binary.Div:
		y := x.(*binary.Div)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft / hashRight, true

	case *binary.Mod:
		y := x.(*binary.Mod)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return hashLeft % hashRight, true

	case *expr.UnaryPlus:
		y := x.(*expr.UnaryPlus)
		return b.getHashForExpressionNode(y.Expr)

	case *expr.UnaryMinus:
		y := x.(*expr.UnaryMinus)
		hash, ok := b.getHashForExpressionNode(y.Expr)
		return hash * -1, ok

	case *binary.Concat:
		y := x.(*binary.Concat)
		hashLeft, ok := b.getHashForExpressionNode(y.Left)
		if !ok {
			return 0, false
		}

		hashRight, ok := b.getHashForExpressionNode(y.Right)
		if !ok {
			return 0, false
		}
		return int64(hash(fmt.Sprint(hashLeft) + fmt.Sprint(hashRight))), true

	case *expr.ArrayDimFetch:
		y, _ := x.(*expr.ArrayDimFetch)
		variableName := y.Variable.(*node.SimpleVar).Name
		indexHash, ok := b.getHashForExpressionNode(y.Dim)
		if !ok {
			return 0, false
		}
		indexHashString := fmt.Sprint(indexHash)
		return int64(hash("$" + variableName + "[" + indexHashString + "]")), ok

	case *expr.ClassConstFetch:
		y := x.(*expr.ClassConstFetch)
		className := meta.NameToString(y.Class.(*name.Name))
		constName := y.ConstantName.Value
		return int64(hash(className + "::" + constName)), true

	case *expr.ConstFetch:
		y := x.(*expr.ConstFetch)
		constName := meta.NameToString(y.Constant.(*name.Name))
		return int64(hash("ConstFetch" + constName)), true

	case *expr.FunctionCall:
		y := x.(*expr.FunctionCall)
		functionName, ok := solver.GetFuncName(b.r.ctx.st, y.Function)
		if !ok {
			return 0, false
		}
		return int64(hash(functionName)), true

	case *expr.MethodCall:
		y := x.(*expr.MethodCall)
		variableName := y.Variable.(*node.SimpleVar).Name
		functionName := y.Method.(*node.Identifier).Value
		return int64(hash("$" + variableName + "." + functionName)), true

	case *expr.PropertyFetch:
		y := x.(*expr.PropertyFetch)
		variableName := y.Variable.(*node.SimpleVar).Name
		propertyName := y.Property.(*node.Identifier).Value
		return int64(hash("$" + variableName + "->" + propertyName)), true

	case *expr.StaticCall:
		y := x.(*expr.StaticCall)
		className := meta.NameToString(y.Class.(*name.Name))
		functionName := y.Call.(*node.Identifier).Value
		return int64(hash(className + "::" + functionName)), true

	case *expr.StaticPropertyFetch:
		y := x.(*expr.StaticPropertyFetch)
		className := meta.NameToString(y.Class.(*name.Name))
		propertyName := y.Property.(*node.SimpleVar).Name
		return int64(hash(className + "::$" + propertyName)), true

	case *node.SimpleVar:
		y := x.(*node.SimpleVar)
		return int64(hash("$" + y.Name)), true

	case *scalar.Dnumber:
		y := x.(*scalar.Dnumber)
		val, err := strconv.ParseFloat(y.Value, 64)
		if err != nil {
			return 0, false
		}
		val = math.Floor(val)
		return int64(val), true

	case *scalar.Lnumber:
		y := x.(*scalar.Lnumber)
		value, err := strconv.ParseInt(y.Value, 0, 64)
		if err != nil {
			return 0, false
		}
		return value, true

	case *scalar.String:
		y := x.(*scalar.String)
		return int64(hash("'" + unquote(y.Value) + "'")), true

	default:
		return 0, false
	}
}
