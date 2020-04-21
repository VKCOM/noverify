package linter

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/meta"
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

	switch nm := call.Function.(type) {
	case *name.Name:
		nameStr := meta.NameToString(nm)
		firstPart := nm.Parts[0].(*name.NamePart).Value
		if alias, ok := st.FunctionUses[firstPart]; ok {
			if len(nm.Parts) == 1 {
				nameStr = alias
			} else {
				// handle situations like 'use NS\Foo; Foo\Bar::doSomething();'
				nameStr = alias + `\` + meta.NamePartsToString(nm.Parts[1:])
			}
			res.fqName = nameStr
			res.info, res.defined = meta.Info.GetFunction(res.fqName)
		} else {
			res.fqName = st.Namespace + `\` + nameStr
			res.info, res.defined = meta.Info.GetFunction(res.fqName)
			if !res.defined && st.Namespace != "" {
				res.fqName = `\` + nameStr
				res.info, res.defined = meta.Info.GetFunction(res.fqName)
			}
		}

	case *name.FullyQualified:
		res.fqName = meta.FullyQualifiedToString(nm)
		res.info, res.defined = meta.Info.GetFunction(res.fqName)
	default:
		res.defined = false

		solver.ExprTypeCustom(sc, st, nm, customTypes).Iterate(func(typ string) {
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

// normalizeType adds namespaces to a type defined by the PHPDoc type string as well as
// converts notations like "array<int,string>" to <meta.WARRAY2, "int", "string">
func normalizeType(st *meta.ClassParseState, typStr string) string {
	if typStr == "" {
		return ""
	}

	nullable := false
	classNames := strings.Split(typStr, `|`)
	for idx, className := range classNames {
		// ignore things like \tuple(*)
		if braceIdx := strings.IndexByte(className, '('); braceIdx >= 0 {
			className = className[0:braceIdx]
		}

		// 0 for "bool", 1 for "bool[]", 2 for "bool[][]" and so on
		arrayDim := 0
		for strings.HasSuffix(className, "[]") {
			arrayDim++
			className = strings.TrimSuffix(className, "[]")
		}

		if len(className) == 0 {
			continue
		}

		if className[0] == '?' && len(className) > 1 {
			nullable = true
			className = className[1:]
		}

		switch className {
		case "bool", "true", "false", "float", "string", "int", "array", "resource", "mixed", "null", "callable", "void", "object":
			// Can't assign className here because it also erases [] for arrays.
			if classNames[idx][0] == '?' {
				classNames[idx] = classNames[idx][1:]
			}
			continue
		case "$this":
			// Handle `$this` as `static` alias in phpdoc context.
			classNames[idx] = "static"
			continue
		case "static":
			// Don't resolve `static` phpdoc type annotation too early
			// to make it possible to handle late static binding.
			continue
		}

		if className[0] == '\\' {
			continue
		}

		if className[0] <= meta.WMax {
			linterError(st.CurrentFile, "Bad type: '%s'", className)
			classNames[idx] = ""
			continue
		}

		// special types, e.g. "array<k,v>"
		if strings.ContainsAny(className, "<>") {
			classNames[idx] = parseAngleBracketedType(st, className)
			continue
		}

		fullClassName, ok := solver.GetClassName(st, meta.StringToName(className))
		if !ok {
			classNames[idx] = ""
			continue
		}

		if arrayDim > 0 {
			fullClassName += strings.Repeat("[]", arrayDim)
		}

		classNames[idx] = fullClassName
	}

	if nullable {
		classNames = append(classNames, "null")
	}

	return strings.Join(classNames, "|")
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
