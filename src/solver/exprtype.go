package solver

import (
	"log"
	"strconv"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
)

func bitwiseOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right ir.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, left, custom).Is("string") && ExprTypeLocalCustom(sc, cs, right, custom).Is("string") {
		return meta.NewTypesMap("string")
	}
	return meta.NewTypesMap("int")
}

func unaryMathOpType(sc *meta.Scope, cs *meta.ClassParseState, x ir.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, x, custom).Is("int") {
		return meta.NewTypesMap("int")
	}
	return meta.NewTypesMap("float")
}

// binaryMathOpType is used for binary arithmetic operations
func binaryMathOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right ir.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, left, custom).Is("int") && ExprTypeLocalCustom(sc, cs, right, custom).Is("int") {
		return meta.NewTypesMap("int")
	}
	return meta.NewTypesMap("float")
}

// binaryPlusOpType is a special case as "plus" is also used for array union operation
func binaryPlusOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right ir.Node, custom []CustomType) meta.TypesMap {
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
func ExprType(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node) meta.TypesMap {
	return ExprTypeCustom(sc, cs, n, nil)
}

// ExprTypeCustom is ExprType that allows to specify custom types overrides
func ExprTypeCustom(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node, custom []CustomType) meta.TypesMap {
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

func internalFuncType(nm string, sc *meta.Scope, cs *meta.ClassParseState, c *ir.FunctionCallExpr, custom []CustomType) (typ meta.TypesMap, ok bool) {
	fn, ok := meta.GetInternalFunctionInfo(nm)
	if !ok || fn.Typ.IsEmpty() {
		return meta.TypesMap{}, false
	}

	override, ok := meta.GetInternalFunctionOverrideInfo(nm)
	if !ok || len(c.ArgumentList.Arguments) <= override.ArgNum {
		return fn.Typ, true
	}

	arg := c.ArgumentList.Arguments[override.ArgNum].(*ir.Argument)
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

func arrayType(items []*ir.ArrayItemExpr) meta.TypesMap {
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

func isConstantStringArray(items []*ir.ArrayItemExpr) bool {
	for _, item := range items {
		if _, ok := item.Val.(*ir.String); !ok {
			return false
		}
	}

	return true
}

func isConstantIntArray(items []*ir.ArrayItemExpr) bool {
	for _, item := range items {
		if _, ok := item.Val.(*ir.Lnumber); !ok {
			return false
		}
	}

	return true
}

func isConstantFloatArray(items []*ir.ArrayItemExpr) bool {
	for _, item := range items {
		if _, ok := item.Val.(*ir.Dnumber); !ok {
			return false
		}
	}

	return true
}

// CustomType specifies a mapping between some AST structure and concrete type (e.g. for <expr> instanceof <something>)
type CustomType struct {
	Node ir.Node
	Typ  meta.TypesMap
}

// ExprTypeLocal is basic expression type that does not resolve cross-file function calls and such
func ExprTypeLocal(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node) meta.TypesMap {
	return ExprTypeLocalCustom(sc, cs, n, nil)
}

func exprTypeLocalCustom(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node, custom []CustomType) meta.TypesMap {
	if n == nil || sc == nil {
		return meta.TypesMap{}
	}

	for _, c := range custom {
		if irutil.NodeEqual(c.Node, n) {
			return c.Typ
		}
	}

	switch n := n.(type) {
	case *ir.FunctionCallExpr:
		nm, ok := n.Function.(*ir.Name)
		if !ok {
			return meta.TypesMap{}
		}
		if nm.IsFullyQualified() {
			if nm.NumParts() == 1 {
				typ, ok := internalFuncType(strings.TrimPrefix(nm.Value, `\`), sc, cs, n, custom)
				if ok {
					return typ
				}
			}
			return meta.NewTypesMap(meta.WrapFunctionCall(nm.Value))
		}
		typ, ok := internalFuncType(`\`+nm.Value, sc, cs, n, custom)
		if ok {
			return typ
		}
		return meta.NewTypesMap(meta.WrapFunctionCall(cs.Namespace + `\` + nm.Value))
	case *ir.StaticCallExpr:
		id, ok := n.Call.(*ir.Identifier)
		if !ok {
			return meta.TypesMap{}
		}

		nm, ok := GetClassName(cs, n.Class)
		if !ok {
			return meta.TypesMap{}
		}

		return meta.NewTypesMap(meta.WrapStaticMethodCall(nm, id.Value))
	case *ir.StaticPropertyFetchExpr:
		v, ok := n.Property.(*ir.SimpleVar)
		if !ok {
			return meta.TypesMap{}
		}

		nm, ok := GetClassName(cs, n.Class)
		if !ok {
			return meta.TypesMap{}
		}

		return meta.NewTypesMap(meta.WrapStaticPropertyFetch(nm, "$"+v.Name))
	case *ir.SimpleVar:
		typ, _ := sc.GetVarNameType(n.Name)
		return typ
	case *ir.MethodCallExpr:
		// Support only $obj->callSomething().
		// Do not support $obj->$method()
		id, ok := n.Method.(*ir.Identifier)
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
	case *ir.PropertyFetchExpr:
		// Support only $obj->some_prop.
		// Do not support $obj->$some_prop
		id, ok := n.Property.(*ir.Identifier)
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
	case *ir.ArrayDimFetchExpr:
		m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
		if m.IsEmpty() {
			return meta.TypesMap{}
		}

		res := make(map[string]struct{}, m.Len())

		m.Iterate(func(className string) {
			switch dim := n.Dim.(type) {
			case *ir.String:
				key := dim.Value[len(`"`) : len(dim.Value)-len(`"`)]
				res[meta.WrapElemOfKey(className, key)] = struct{}{}
			case *ir.Lnumber:
				res[meta.WrapElemOfKey(className, dim.Value)] = struct{}{}
			default:
				res[meta.WrapElemOf(className)] = struct{}{}
			}
		})

		return meta.NewTypesMapFromMap(res)
	case *ir.BitwiseNotExpr:
		if ExprTypeLocalCustom(sc, cs, n.Expr, custom).Is("string") {
			return meta.NewTypesMap("string")
		}
		return meta.NewTypesMap("int")
	case *ir.BitwiseAndExpr:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.BitwiseOrExpr:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.BitwiseXorExpr:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.ConcatExpr:
		return meta.PreciseStringType
	case *ir.ArrayExpr:
		return arrayType(n.Items)
	case *ir.BooleanNotExpr, *ir.BooleanAndExpr, *ir.BooleanOrExpr,
		*ir.EqualExpr, *ir.NotEqualExpr, *ir.IdenticalExpr, *ir.NotIdenticalExpr,
		*ir.GreaterExpr, *ir.GreaterOrEqualExpr,
		*ir.SmallerExpr, *ir.SmallerOrEqualExpr,
		*ir.EmptyExpr, *ir.IssetExpr:
		return meta.PreciseBoolType
	case *ir.UnaryMinusExpr:
		return unaryMathOpType(sc, cs, n.Expr, custom)
	case *ir.UnaryPlusExpr:
		return unaryMathOpType(sc, cs, n.Expr, custom)
	case *ir.MulExpr:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.DivExpr:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.PlusExpr:
		return binaryPlusOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.MinusExpr:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.ModExpr:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.PostIncExpr:
		return unaryMathOpType(sc, cs, n.Variable, custom)
	case *ir.PreIncExpr:
		return unaryMathOpType(sc, cs, n.Variable, custom)
	case *ir.PostDecExpr:
		return unaryMathOpType(sc, cs, n.Variable, custom)
	case *ir.PreDecExpr:
		return unaryMathOpType(sc, cs, n.Variable, custom)
	case *ir.TypeCastExpr:
		switch n.Type {
		case "array":
			return meta.NewTypesMap("mixed[]")
		case "int":
			return meta.PreciseIntType
		case "string":
			return meta.PreciseStringType
		case "float":
			return meta.PreciseFloatType
		case "bool":
			return meta.PreciseBoolType
		}
	case *ir.ShiftLeftExpr, *ir.ShiftRightExpr:
		return meta.PreciseIntType
	case *ir.ClassConstFetchExpr:
		className, ok := GetClassName(cs, n.Class)
		if !ok {
			return meta.TypesMap{}
		}
		return meta.NewTypesMap(meta.WrapClassConstFetch(className, n.ConstantName.Value))
	case *ir.ConstFetchExpr:
		// TODO: handle namespaces
		nm := n.Constant
		switch nm.Value {
		case "false", "true":
			return meta.PreciseBoolType
		case "null":
			return meta.NewTypesMap("null")
		default:
			if nm.NumParts() == 0 {
				return meta.NewTypesMap(meta.WrapConstant(nm.Value))
			}
		}
	case *ir.String, *ir.Encapsed, *ir.Heredoc:
		return meta.PreciseStringType
	case *ir.Lnumber:
		return meta.PreciseIntType
	case *ir.Dnumber:
		return meta.PreciseFloatType
	case *ir.TernaryExpr:
		t := ExprTypeLocalCustom(sc, cs, n.IfTrue, custom)
		f := ExprTypeLocalCustom(sc, cs, n.IfFalse, custom)
		return meta.NewEmptyTypesMap(t.Len() + f.Len()).Append(t).Append(f)
	case *ir.NewExpr:
		if meta.NameNodeToString(n.Class) == "static" {
			return meta.NewTypesMap("static")
		}
		nm, ok := GetClassName(cs, n.Class)
		if ok {
			return meta.NewPreciseTypesMap(nm)
		}
		return meta.TypesMap{}
	case *ir.ParenExpr:
		return ExprTypeLocalCustom(sc, cs, n.Expr, custom)
	case *ir.Assign:
		return ExprTypeLocalCustom(sc, cs, n.Expression, custom)
	case *ir.CloneExpr:
		return ExprTypeLocalCustom(sc, cs, n.Expr, custom)
	case *ir.ClosureExpr:
		return meta.NewTypesMap(`\Closure`)
	}

	return meta.TypesMap{}
}

// ExprTypeLocalCustom is ExprTypeLocal that allows to specify custom types
func ExprTypeLocalCustom(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node, custom []CustomType) meta.TypesMap {
	res := exprTypeLocalCustom(sc, cs, n, custom)
	if res.Len() == 0 {
		return meta.MixedType
	}
	return res
}

func GetConstantValue(c *ir.ConstantStmt) (meta.ConstantValue, bool) {
	switch c := c.Expr.(type) {
	case *ir.Lnumber:
		return getConstantValue(c, 1)
	case *ir.Dnumber:
		return getConstantValue(c, 1)
	case *ir.String:
		return getConstantValue(c, 1)
	case *ir.UnaryMinusExpr:
		return getConstantValue(c.Expr, -1)
	default:
		return meta.NewUndefinedConstantValue(), false
	}
}

func getConstantValue(n ir.Node, modifier int64) (meta.ConstantValue, bool) {
	switch c := n.(type) {
	case *ir.Lnumber:
		value, err := strconv.ParseInt(c.Value, 10, 64)
		return meta.NewConstantValueFromInt(value * modifier), err == nil
	case *ir.Dnumber:
		value, err := strconv.ParseFloat(c.Value, 64)
		return meta.NewConstantValueFromFloat(value * float64(modifier)), err == nil
	case *ir.String:
		return meta.NewConstantValueFromString(c.Value), true
	default:
		return meta.NewUndefinedConstantValue(), false
	}
}
