package solver

import (
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/astutil"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/cast"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
)

func bitwiseOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right node.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, left, custom).Is("string") && ExprTypeLocalCustom(sc, cs, right, custom).Is("string") {
		return meta.NewTypesMap("string")
	}
	return meta.NewTypesMap("int")
}

func unaryMathOpType(sc *meta.Scope, cs *meta.ClassParseState, x node.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, x, custom).Is("int") {
		return meta.NewTypesMap("int")
	}
	return meta.NewTypesMap("float")
}

// binaryMathOpType is used for binary arithmetic operations
func binaryMathOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right node.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, left, custom).Is("int") && ExprTypeLocalCustom(sc, cs, right, custom).Is("int") {
		return meta.NewTypesMap("int")
	}
	return meta.NewTypesMap("float")
}

// binaryPlusOpType is a special case as "plus" is also used for array union operation
func binaryPlusOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right node.Node, custom []CustomType) meta.TypesMap {
	// TODO: PHP will raise fatal error if one operand is array and other is not, so we may check it too
	leftType := ExprTypeLocalCustom(sc, cs, left, custom)
	rightType := ExprTypeLocalCustom(sc, cs, right, custom)
	if leftType.IsArray() && rightType.IsArray() {
		return meta.MergeTypeMaps(leftType, rightType)
	}
	return binaryMathOpType(sc, cs, left, right, custom)
}

// ExprType returns type of expression. Depending on whether or not is it "full mode",
// it will also recursively resolve all nested types
func ExprType(sc *meta.Scope, cs *meta.ClassParseState, n node.Node) meta.TypesMap {
	return ExprTypeCustom(sc, cs, n, nil)
}

// ExprTypeCustom is ExprType that allows to specify custom types overrides
func ExprTypeCustom(sc *meta.Scope, cs *meta.ClassParseState, n node.Node, custom []CustomType) meta.TypesMap {
	m := ExprTypeLocalCustom(sc, cs, n, custom)

	if !meta.IsIndexingComplete() {
		return m
	}
	if m.IsResolved() {
		return m
	}

	visitedMap := make(ResolverMap)
	resolvedTypes := ResolveTypes(cs.CurrentClass, m, visitedMap)
	return meta.NewTypesMapFromMap(resolvedTypes)
}

func internalFuncType(nm string, sc *meta.Scope, cs *meta.ClassParseState, c *expr.FunctionCall, custom []CustomType) (typ meta.TypesMap, ok bool) {
	fn, ok := meta.GetInternalFunctionInfo(nm)
	if !ok || fn.Typ.IsEmpty() {
		return meta.TypesMap{}, false
	}

	override, ok := meta.GetInternalFunctionOverrideInfo(nm)
	if !ok || len(c.ArgumentList.Arguments) <= override.ArgNum {
		return fn.Typ, true
	}

	arg := c.ArgumentList.Arguments[override.ArgNum].(*node.Argument)
	typ = ExprTypeLocalCustom(sc, cs, arg.Expr, custom)
	if override.OverrideType == meta.OverrideArgType {
		return typ, true
	} else if override.OverrideType == meta.OverrideElementType {
		newTyp := meta.NewEmptyTypesMap(typ.Len())
		typ.Iterate(func(t string) {
			newTyp = newTyp.AppendString(meta.WrapElemOf(t))
		})
		return newTyp, true
	}

	log.Printf("Internal error: unexpected override type %d for function %s", override.OverrideType, nm)
	return meta.TypesMap{}, false
}

func arrayType(items []*expr.ArrayItem) meta.TypesMap {
	if len(items) == 0 {
		// Used as a placeholder until more specific type is discovered.
		//
		// Should be resolved before used.
		//
		// We do this to simplify resolving `mixed[]|int[]` to `int[]`.
		// If we know that an array is empty, then it's valid array
		// for any mono-typed array, so we can just throw away "empty_array"
		// in that case. If element type is unknown, "empty_array" is
		// resolved into "mixed[]".
		return meta.NewTypesMap("empty_array")
	}

	if len(items) > 0 {
		switch {
		case isConstantStringArray(items):
			return meta.NewTypesMap("string[]")
		case isConstantIntArray(items):
			return meta.NewTypesMap("int[]")
		case isConstantFloatArray(items):
			return meta.NewTypesMap("float[]")
		}
	}

	return meta.NewTypesMap("mixed[]")
}

func isConstantStringArray(items []*expr.ArrayItem) bool {
	for _, item := range items {
		if _, ok := item.Val.(*scalar.String); !ok {
			return false
		}
	}

	return true
}

func isConstantIntArray(items []*expr.ArrayItem) bool {
	for _, item := range items {
		if _, ok := item.Val.(*scalar.Lnumber); !ok {
			return false
		}
	}

	return true
}

func isConstantFloatArray(items []*expr.ArrayItem) bool {
	for _, item := range items {
		if _, ok := item.Val.(*scalar.Dnumber); !ok {
			return false
		}
	}

	return true
}

// CustomType specifies a mapping between some AST structure and concrete type (e.g. for <expr> instanceof <something>)
type CustomType struct {
	Node node.Node
	Typ  meta.TypesMap
}

// ExprTypeLocal is basic expression type that does not resolve cross-file function calls and such
func ExprTypeLocal(sc *meta.Scope, cs *meta.ClassParseState, n node.Node) meta.TypesMap {
	return ExprTypeLocalCustom(sc, cs, n, nil)
}

func exprTypeLocalCustom(sc *meta.Scope, cs *meta.ClassParseState, n node.Node, custom []CustomType) meta.TypesMap {
	if n == nil || sc == nil {
		return meta.TypesMap{}
	}

	for _, c := range custom {
		if astutil.NodeEqual(c.Node, n) {
			return c.Typ
		}
	}

	switch n := n.(type) {
	case *expr.FunctionCall:
		nm, ok := n.Function.(*name.Name)
		if !ok {
			if nm, ok := n.Function.(*name.FullyQualified); ok {
				funcName := meta.FullyQualifiedToString(nm)
				if strings.Count(funcName, `\`) == 1 {
					typ, ok := internalFuncType(strings.TrimPrefix(funcName, `\`), sc, cs, n, custom)
					if ok {
						return typ
					}
				}
				return meta.NewTypesMap(meta.WrapFunctionCall(funcName))
			}
			return meta.TypesMap{}
		}

		funcName := meta.NameToString(nm)
		typ, ok := internalFuncType(`\`+funcName, sc, cs, n, custom)
		if ok {
			return typ
		}

		return meta.NewTypesMap(meta.WrapFunctionCall(cs.Namespace + `\` + funcName))
	case *expr.StaticCall:
		id, ok := n.Call.(*node.Identifier)
		if !ok {
			return meta.TypesMap{}
		}

		nm, ok := GetClassName(cs, n.Class)
		if !ok {
			return meta.TypesMap{}
		}

		return meta.NewTypesMap(meta.WrapStaticMethodCall(nm, id.Value))
	case *expr.StaticPropertyFetch:
		v, ok := n.Property.(*node.SimpleVar)
		if !ok {
			return meta.TypesMap{}
		}

		nm, ok := GetClassName(cs, n.Class)
		if !ok {
			return meta.TypesMap{}
		}

		return meta.NewTypesMap(meta.WrapStaticPropertyFetch(nm, "$"+v.Name))
	case *node.SimpleVar:
		typ, _ := sc.GetVarNameType(n.Name)
		return typ
	case *expr.MethodCall:
		// Support only $obj->callSomething().
		// Do not support $obj->$method()
		id, ok := n.Method.(*node.Identifier)
		if !ok {
			return meta.TypesMap{}
		}

		m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
		if m.IsEmpty() {
			return meta.TypesMap{}
		}

		res := make(map[string]struct{}, m.Len())

		m.Iterate(func(className string) {
			res[meta.WrapInstanceMethodCall(className, id.Value)] = struct{}{}
		})

		return meta.NewTypesMapFromMap(res)
	case *expr.PropertyFetch:
		// Support only $obj->some_prop.
		// Do not support $obj->$some_prop
		id, ok := n.Property.(*node.Identifier)
		if !ok {
			return meta.TypesMap{}
		}

		m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
		if m.IsEmpty() {
			return meta.TypesMap{}
		}

		res := make(map[string]struct{}, m.Len())

		m.Iterate(func(className string) {
			res[meta.WrapInstancePropertyFetch(className, id.Value)] = struct{}{}
		})

		return meta.NewTypesMapFromMap(res)
	case *expr.ArrayDimFetch:
		m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
		if m.IsEmpty() {
			return meta.TypesMap{}
		}

		res := make(map[string]struct{}, m.Len())

		m.Iterate(func(className string) {
			switch dim := n.Dim.(type) {
			case *scalar.String:
				key := dim.Value[len(`"`) : len(dim.Value)-len(`"`)]
				res[meta.WrapElemOfKey(className, key)] = struct{}{}
			case *scalar.Lnumber:
				res[meta.WrapElemOfKey(className, dim.Value)] = struct{}{}
			default:
				res[meta.WrapElemOf(className)] = struct{}{}
			}
		})

		return meta.NewTypesMapFromMap(res)
	case *expr.BitwiseNot:
		if ExprTypeLocalCustom(sc, cs, n.Expr, custom).Is("string") {
			return meta.NewTypesMap("string")
		}
		return meta.NewTypesMap("int")
	case *binary.BitwiseAnd:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.BitwiseOr:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.BitwiseXor:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Concat:
		return meta.PreciseStringType
	case *expr.Array:
		return arrayType(n.Items)
	case *expr.BooleanNot, *binary.BooleanAnd, *binary.BooleanOr,
		*binary.Equal, *binary.NotEqual, *binary.Identical, *binary.NotIdentical,
		*binary.Greater, *binary.GreaterOrEqual,
		*binary.Smaller, *binary.SmallerOrEqual,
		*expr.Empty, *expr.Isset:
		return meta.PreciseBoolType
	case *expr.UnaryMinus:
		return unaryMathOpType(sc, cs, n.Expr, custom)
	case *expr.UnaryPlus:
		return unaryMathOpType(sc, cs, n.Expr, custom)
	case *binary.Mul:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Div:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Plus:
		return binaryPlusOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Minus:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Mod:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *cast.Array:
		return meta.NewTypesMap("mixed[]")
	case *cast.Bool:
		return meta.PreciseBoolType
	case *cast.Double:
		return meta.PreciseFloatType
	case *cast.Int, *binary.ShiftLeft, *binary.ShiftRight:
		return meta.PreciseIntType
	case *cast.String:
		return meta.PreciseStringType
	case *expr.ClassConstFetch:
		className, ok := GetClassName(cs, n.Class)
		if !ok {
			return meta.TypesMap{}
		}
		return meta.NewTypesMap(meta.WrapClassConstFetch(className, n.ConstantName.Value))
	case *expr.ConstFetch:
		nm, ok := n.Constant.(*name.Name)
		if !ok {
			return meta.TypesMap{}
		}

		// TODO: handle namespaces
		p := nm.Parts
		if len(p) == 1 {
			constName := p[0].(*name.NamePart).Value

			if constName == "false" || constName == "true" {
				return meta.PreciseBoolType
			}

			if constName == "null" {
				return meta.NewTypesMap("null")
			}

			return meta.NewTypesMap(meta.WrapConstant(constName))
		}
	case *scalar.String, *scalar.Encapsed, *scalar.Heredoc:
		return meta.PreciseStringType
	case *scalar.Lnumber:
		return meta.PreciseIntType
	case *scalar.Dnumber:
		return meta.PreciseFloatType
	case *expr.Ternary:
		t := ExprTypeLocalCustom(sc, cs, n.IfTrue, custom)
		f := ExprTypeLocalCustom(sc, cs, n.IfFalse, custom)
		return meta.NewEmptyTypesMap(t.Len() + f.Len()).Append(t).Append(f)
	case *expr.New:
		if meta.NameNodeToString(n.Class) == "static" {
			return meta.NewTypesMap("static")
		}
		nm, ok := GetClassName(cs, n.Class)
		if ok {
			return meta.NewPreciseTypesMap(nm)
		}
		return meta.TypesMap{}
	case *expr.Paren:
		return ExprTypeLocalCustom(sc, cs, n.Expr, custom)
	case *assign.Assign:
		return ExprTypeLocalCustom(sc, cs, n.Expression, custom)
	case *expr.Clone:
		return ExprTypeLocalCustom(sc, cs, n.Expr, custom)
	case *expr.Closure:
		return meta.NewTypesMap(`\Closure`)
	}

	return meta.TypesMap{}
}

// ExprTypeLocalCustom is ExprTypeLocal that allows to specify custom types
func ExprTypeLocalCustom(sc *meta.Scope, cs *meta.ClassParseState, n node.Node, custom []CustomType) meta.TypesMap {
	res := exprTypeLocalCustom(sc, cs, n, custom)
	if res.Len() == 0 {
		return meta.MixedType
	}
	return res
}
