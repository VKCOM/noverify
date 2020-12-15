package solver

import (
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
)

// CustomType specifies a mapping between some AST structure
// and concrete type (e.g. for <expr> instanceof <something>).
type CustomType struct {
	Node ir.Node
	Typ  meta.TypesMap
}

// ExprType returns type of expression. Depending on whether or not is it "full mode",
// it will also recursively resolve all nested types.
func ExprType(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node) meta.TypesMap {
	return ExprTypeCustom(sc, cs, n, nil)
}

// ExprTypeCustom is ExprType that allows to specify custom types overrides.
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

// ExprTypeLocal is basic expression type that does not resolve cross-file function calls and such.
func ExprTypeLocal(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node) meta.TypesMap {
	return ExprTypeLocalCustom(sc, cs, n, nil)
}

// ExprTypeLocalCustom is ExprTypeLocal that allows to specify custom types.
func ExprTypeLocalCustom(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node, custom []CustomType) meta.TypesMap {
	res := exprTypeLocalCustom(sc, cs, n, custom)
	if res.Len() == 0 {
		return meta.MixedType
	}
	return res
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
		return functionCallType(n, sc, cs, custom)
	case *ir.StaticCallExpr:
		return staticFunctionCallType(n, cs)
	case *ir.StaticPropertyFetchExpr:
		return staticPropertyFetchType(n, cs)
	case *ir.SimpleVar:
		return simpleVarType(n, sc)
	case *ir.MethodCallExpr:
		return methodCallType(n, sc, cs, custom)
	case *ir.PropertyFetchExpr:
		return propertyFetchType(n, sc, cs, custom)
	case *ir.ArrayDimFetchExpr:
		return arrayDimFetchType(n, sc, cs, custom)
	case *ir.BitwiseNotExpr:
		return unaryBitwiseOpType(sc, cs, n.Expr, custom)
	case *ir.BitwiseAndExpr:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.BitwiseOrExpr:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.BitwiseXorExpr:
		return bitwiseOpType(sc, cs, n.Left, n.Right, custom)
	case *ir.ConcatExpr:
		return meta.PreciseStringType
	case *ir.ArrayExpr:
		return arrayType(sc, cs, n.Items)
	case *ir.ArrayItemExpr:
		return ExprTypeLocalCustom(sc, cs, n.Val, custom)
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
		return typeCastType(n)
	case *ir.ShiftLeftExpr, *ir.ShiftRightExpr:
		return meta.PreciseIntType
	case *ir.ClassConstFetchExpr:
		return classConstFetchType(n, cs)
	case *ir.ConstFetchExpr:
		return constFetchType(n)
	case *ir.String, *ir.Encapsed, *ir.Heredoc:
		return meta.PreciseStringType
	case *ir.Lnumber:
		return meta.PreciseIntType
	case *ir.Dnumber:
		return meta.PreciseFloatType
	case *ir.TernaryExpr:
		return ternaryExprType(n, sc, cs, custom)
	case *ir.CoalesceExpr:
		return coalesceExprType(n, sc, cs, custom)
	case *ir.NewExpr:
		return newExprType(n, cs)
	case *ir.ParenExpr:
		return ExprTypeLocalCustom(sc, cs, n.Expr, custom)
	case *ir.Assign:
		return ExprTypeLocalCustom(sc, cs, n.Expression, custom)
	case *ir.AssignConcat:
		return meta.PreciseStringType
	case *ir.AssignShiftLeft, *ir.AssignShiftRight:
		return meta.PreciseIntType
	case *ir.CloneExpr:
		return ExprTypeLocalCustom(sc, cs, n.Expr, custom)
	case *ir.ClosureExpr:
		return meta.NewTypesMap(`\Closure`)
	case *ir.MagicConstant:
		return magicConstantType(n)
	}

	return meta.TypesMap{}
}

// unaryBitwiseOpType is used for unary bitwise operations.
func unaryBitwiseOpType(sc *meta.Scope, cs *meta.ClassParseState, x ir.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, x, custom).Is("string") {
		return meta.NewTypesMap("string")
	}
	return meta.NewTypesMap("int")
}

// bitwiseOpType is used for binary bitwise operations.
func bitwiseOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right ir.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, left, custom).Is("string") && ExprTypeLocalCustom(sc, cs, right, custom).Is("string") {
		return meta.NewTypesMap("string")
	}
	return meta.NewTypesMap("int")
}

// unaryMathOpType is used for unary arithmetic operations.
func unaryMathOpType(sc *meta.Scope, cs *meta.ClassParseState, x ir.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, x, custom).Is("int") {
		return meta.NewTypesMap("int")
	}
	return meta.NewTypesMap("float")
}

// binaryMathOpType is used for binary arithmetic operations.
func binaryMathOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right ir.Node, custom []CustomType) meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, left, custom).Is("int") && ExprTypeLocalCustom(sc, cs, right, custom).Is("int") {
		return meta.NewTypesMap("int")
	}
	return meta.NewTypesMap("float")
}

// binaryPlusOpType is a special case as "plus" is also used for array union operation.
func binaryPlusOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right ir.Node, custom []CustomType) meta.TypesMap {
	// TODO: PHP will raise fatal error if one operand is array and other is not, so we may check it too
	leftType := ExprTypeLocalCustom(sc, cs, left, custom)
	rightType := ExprTypeLocalCustom(sc, cs, right, custom)
	if leftType.IsArray() && rightType.IsArray() {
		return meta.MergeTypeMaps(leftType, rightType)
	}
	return binaryMathOpType(sc, cs, left, right, custom)
}

func classNameToString(cs *meta.ClassParseState, n ir.Node) (string, bool) {
	var name string

	switch n := n.(type) {
	case *ir.String:
		name = n.Value
	case *ir.ClassConstFetchExpr:
		if !strings.EqualFold(n.ConstantName.Value, "class") {
			return "", false
		}

		switch class := n.Class.(type) {
		case *ir.Name:
			name = class.Value
		case *ir.Identifier:
			name = class.Value
		case *ir.SimpleVar:
			name = "$" + class.Name
		default:
			return "", false
		}
	default:
		return "", false
	}

	className, ok := GetClassName(cs, &ir.Name{Value: name})
	if !ok {
		return "", false
	}

	return className, true
}

func internalFuncType(nm string, sc *meta.Scope, cs *meta.ClassParseState, c *ir.FunctionCallExpr, custom []CustomType) (typ meta.TypesMap, ok bool) {
	fn, ok := meta.GetInternalFunctionInfo(nm)
	if !ok || fn.Typ.IsEmpty() {
		return meta.TypesMap{}, false
	}

	override, ok := meta.GetInternalFunctionOverrideInfo(nm)
	if !ok || len(c.Args) <= override.ArgNum {
		return fn.Typ, true
	}

	arg := c.Arg(override.ArgNum)
	typ = ExprTypeLocalCustom(sc, cs, arg.Expr, custom)

	switch override.OverrideType {
	case meta.OverrideArgType:
		return typ, true

	case meta.OverrideElementType:
		newTyp := meta.NewEmptyTypesMap(typ.Len())
		typ.Iterate(func(t string) {
			newTyp = newTyp.AppendString(meta.WrapElemOf(t))
		})
		return newTyp, true

	case meta.OverrideClassType:
		// due to the fact that it is impossible for us to use constfold
		// here, we have to process only a part of the possible options,
		// although the most popular ones.
		className, ok := classNameToString(cs, arg.Expr)
		if !ok {
			return meta.NewTypesMap("mixed"), true
		}
		return meta.NewTypesMap(className + "|null"), true
	}

	log.Printf("Internal error: unexpected override type %d for function %s", override.OverrideType, nm)
	return meta.TypesMap{}, false
}

func arrayType(sc *meta.Scope, cs *meta.ClassParseState, items []*ir.ArrayItemExpr) meta.TypesMap {
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

	firstElementType := ExprTypeLocal(sc, cs, items[0])
	if items[0].Unpack {
		firstElementType = firstElementType.ArrayElemLazyType()
	}

	for _, item := range items[1:] {
		itemType := ExprTypeLocal(sc, cs, item)
		if item.Unpack {
			itemType = itemType.ArrayElemLazyType()
		}

		if !firstElementType.Equals(itemType) {
			return meta.NewTypesMap("mixed[]")
		}
	}

	wrapped := meta.NewEmptyTypesMap(firstElementType.Len())
	firstElementType.Iterate(func(t string) {
		wrapped.AppendString(meta.WrapArrayOf(t))
	})

	return wrapped
}

func newExprType(n *ir.NewExpr, cs *meta.ClassParseState) meta.TypesMap {
	if meta.NameNodeToString(n.Class) == "static" {
		return meta.NewTypesMap("static")
	}
	nm, ok := GetClassName(cs, n.Class)
	if ok {
		return meta.NewPreciseTypesMap(nm)
	}
	return meta.TypesMap{}
}

func ternaryExprType(n *ir.TernaryExpr, sc *meta.Scope, cs *meta.ClassParseState, custom []CustomType) meta.TypesMap {
	t := ExprTypeLocalCustom(sc, cs, n.IfTrue, custom)
	f := ExprTypeLocalCustom(sc, cs, n.IfFalse, custom)
	return meta.NewEmptyTypesMap(t.Len() + f.Len()).Append(t).Append(f)
}

func coalesceExprType(n *ir.CoalesceExpr, sc *meta.Scope, cs *meta.ClassParseState, custom []CustomType) meta.TypesMap {
	l := ExprTypeLocalCustom(sc, cs, n.Left, custom)
	r := ExprTypeLocalCustom(sc, cs, n.Right, custom)
	return meta.NewEmptyTypesMap(l.Len() + r.Len()).Append(l).Append(r)
}

func constFetchType(n *ir.ConstFetchExpr) meta.TypesMap {
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
	return meta.TypesMap{}
}

func classConstFetchType(n *ir.ClassConstFetchExpr, cs *meta.ClassParseState) meta.TypesMap {
	if n.ConstantName.Value == "class" {
		return meta.PreciseStringType
	}
	className, ok := GetClassName(cs, n.Class)
	if !ok {
		return meta.TypesMap{}
	}
	return meta.NewTypesMap(meta.WrapClassConstFetch(className, n.ConstantName.Value))
}

func typeCastType(n *ir.TypeCastExpr) meta.TypesMap {
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
	return meta.TypesMap{}
}

func arrayDimFetchType(n *ir.ArrayDimFetchExpr, sc *meta.Scope, cs *meta.ClassParseState, custom []CustomType) meta.TypesMap {
	m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
	if m.IsEmpty() {
		return meta.TypesMap{}
	}

	res := make(map[string]struct{}, m.Len())

	m.Iterate(func(className string) {
		switch dim := n.Dim.(type) {
		case *ir.String:
			res[meta.WrapElemOfKey(className, dim.Value)] = struct{}{}
		case *ir.Lnumber:
			res[meta.WrapElemOfKey(className, dim.Value)] = struct{}{}
		default:
			res[meta.WrapElemOf(className)] = struct{}{}
		}
	})

	return meta.NewTypesMapFromMap(res)
}

func propertyFetchType(n *ir.PropertyFetchExpr, sc *meta.Scope, cs *meta.ClassParseState, custom []CustomType) meta.TypesMap {
	// Support only $obj->some_prop.
	// Do not support $obj->$some_prop.
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
}

func methodCallType(n *ir.MethodCallExpr, sc *meta.Scope, cs *meta.ClassParseState, custom []CustomType) meta.TypesMap {
	// Support only $obj->callSomething().
	// Do not support $obj->$method().
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
}

func simpleVarType(n *ir.SimpleVar, sc *meta.Scope) meta.TypesMap {
	typ, _ := sc.GetVarNameType(n.Name)
	return typ
}

func staticPropertyFetchType(n *ir.StaticPropertyFetchExpr, cs *meta.ClassParseState) meta.TypesMap {
	v, ok := n.Property.(*ir.SimpleVar)
	if !ok {
		return meta.TypesMap{}
	}

	nm, ok := GetClassName(cs, n.Class)
	if !ok {
		return meta.TypesMap{}
	}

	return meta.NewTypesMap(meta.WrapStaticPropertyFetch(nm, "$"+v.Name))
}

func staticFunctionCallType(n *ir.StaticCallExpr, cs *meta.ClassParseState) meta.TypesMap {
	id, ok := n.Call.(*ir.Identifier)
	if !ok {
		return meta.TypesMap{}
	}

	nm, ok := GetClassName(cs, n.Class)
	if !ok {
		return meta.TypesMap{}
	}

	return meta.NewTypesMap(meta.WrapStaticMethodCall(nm, id.Value))
}

func functionCallType(n *ir.FunctionCallExpr, sc *meta.Scope, cs *meta.ClassParseState, custom []CustomType) meta.TypesMap {
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
}

func magicConstantType(n *ir.MagicConstant) meta.TypesMap {
	if n.Value == "__LINE__" {
		return meta.PreciseIntType
	}
	return meta.PreciseStringType
}
