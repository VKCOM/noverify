package solver

import (
	"log"
	"reflect"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/expr/cast"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
)

func binaryMathOpType(sc *meta.Scope, cs *meta.ClassParseState, left, right node.Node, custom []CustomType) *meta.TypesMap {
	if ExprTypeLocalCustom(sc, cs, left, custom).IsInt() && ExprTypeLocalCustom(sc, cs, right, custom).IsInt() {
		return meta.NewTypesMap("int")
	}
	return meta.NewTypesMap("double")
}

// ExprType returns type of expression. Depending on whether or not is it "full mode",
// it will also recursively resolve all nested types
func ExprType(sc *meta.Scope, cs *meta.ClassParseState, n node.Node) *meta.TypesMap {
	return ExprTypeCustom(sc, cs, n, nil)
}

// ExprTypeCustom is ExprType that allows to specify custom types overrides
func ExprTypeCustom(sc *meta.Scope, cs *meta.ClassParseState, n node.Node, custom []CustomType) *meta.TypesMap {
	m := ExprTypeLocalCustom(sc, cs, n, custom)

	if !meta.IsIndexingComplete() {
		return m
	}

	newMap := make(map[string]struct{}, m.Len())
	visitedMap := make(map[string]struct{})

	m.Iterate(func(k string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic during parsing '%s'", meta.NewTypesMap(k))
				log.Printf("Scope: %s", sc)
				panic(r)
			}
		}()

		for kk := range ResolveType(k, visitedMap) {
			newMap[kk] = struct{}{}
		}
	})

	return meta.NewTypesMapFromMap(newMap)
}

func internalFuncType(nm string, sc *meta.Scope, cs *meta.ClassParseState, c *expr.FunctionCall, custom []CustomType) (typ *meta.TypesMap, ok bool) {
	fn, ok := meta.GetInternalFunctionInfo(nm)
	if !ok || fn.Typ.IsEmpty() {
		return nil, false
	}

	override, ok := meta.GetInternalFunctionOverrideInfo(nm)
	if !ok || len(c.Arguments) <= override.ArgNum {
		return fn.Typ, true
	}

	arg, ok := c.Arguments[override.ArgNum].(*node.Argument)
	if !ok {
		return fn.Typ, true
	}

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
	return nil, false
}

func arrayType(items []node.Node) *meta.TypesMap {
	if len(items) > 0 {
		if isConstantStringArray(items) {
			return meta.NewTypesMap("string[]")
		} else if isConstantIntArray(items) {
			return meta.NewTypesMap("int[]")
		} else if isConstantFloatArray(items) {
			return meta.NewTypesMap("double[]")
		}
	}

	return meta.NewTypesMap("array")
}

func isConstantStringArray(items []node.Node) bool {
	for _, n := range items {
		item, ok := n.(*expr.ArrayItem)
		if !ok {
			return false
		}

		if _, ok := item.Val.(*scalar.String); !ok {
			return false
		}
	}

	return true
}

func isConstantIntArray(items []node.Node) bool {
	for _, n := range items {
		item, ok := n.(*expr.ArrayItem)
		if !ok {
			return false
		}

		if _, ok := item.Val.(*scalar.Lnumber); !ok {
			return false
		}
	}

	return true
}

func isConstantFloatArray(items []node.Node) bool {
	for _, n := range items {
		item, ok := n.(*expr.ArrayItem)
		if !ok {
			return false
		}

		if _, ok := item.Val.(*scalar.Dnumber); !ok {
			return false
		}
	}

	return true
}

// CustomType specifies a mapping between some AST structure and concrete type (e.g. for <expr> instanceof <something>)
type CustomType struct {
	Node node.Node
	Typ  *meta.TypesMap
}

// ExprTypeLocal is basic expression type that does not resolve cross-file function calls and such
func ExprTypeLocal(sc *meta.Scope, cs *meta.ClassParseState, n node.Node) *meta.TypesMap {
	return ExprTypeLocalCustom(sc, cs, n, nil)
}

// ExprTypeLocalCustom is ExprTypeLocal that allows to specify custom types
func ExprTypeLocalCustom(sc *meta.Scope, cs *meta.ClassParseState, n node.Node, custom []CustomType) *meta.TypesMap {
	if n == nil || sc == nil {
		return &meta.TypesMap{}
	}

	for _, c := range custom {
		if reflect.DeepEqual(c.Node, n) {
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
			return &meta.TypesMap{}
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
			return &meta.TypesMap{}
		}

		nm, ok := GetClassName(cs, n.Class)
		if !ok {
			return &meta.TypesMap{}
		}

		return meta.NewTypesMap(meta.WrapStaticMethodCall(nm, id.Value))
	case *expr.StaticPropertyFetch:
		v, ok := n.Property.(*expr.Variable)
		if !ok {
			return &meta.TypesMap{}
		}

		id, ok := v.VarName.(*node.Identifier)
		if !ok {
			return &meta.TypesMap{}
		}

		nm, ok := GetClassName(cs, n.Class)
		if !ok {
			return &meta.TypesMap{}
		}

		return meta.NewTypesMap(meta.WrapStaticPropertyFetch(nm, "$"+id.Value))
	case *expr.Variable:
		id, ok := n.VarName.(*node.Identifier)
		if ok {
			typ, _ := sc.GetVarNameType(id.Value)
			return typ
		}
	case *expr.MethodCall:
		// Support only $obj->callSomething().
		// Do not support $obj->$method()
		id, ok := n.Method.(*node.Identifier)
		if !ok {
			return &meta.TypesMap{}
		}

		m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
		if m.IsEmpty() {
			return &meta.TypesMap{}
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
			return &meta.TypesMap{}
		}

		m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
		if m.IsEmpty() {
			return &meta.TypesMap{}
		}

		res := make(map[string]struct{}, m.Len())

		m.Iterate(func(className string) {
			res[meta.WrapInstancePropertyFetch(className, id.Value)] = struct{}{}
		})

		return meta.NewTypesMapFromMap(res)
	case *expr.ArrayDimFetch:
		m := ExprTypeLocalCustom(sc, cs, n.Variable, custom)
		if m.IsEmpty() {
			return &meta.TypesMap{}
		}

		res := make(map[string]struct{}, m.Len())

		m.Iterate(func(className string) {
			res[meta.WrapElemOf(className)] = struct{}{}
		})

		return meta.NewTypesMapFromMap(res)
	case *binary.Concat:
		return meta.NewTypesMap("string")
	case *expr.Array:
		return arrayType(n.Items)
	case *expr.ShortArray:
		return arrayType(n.Items)
	case *expr.BooleanNot, *binary.BooleanAnd, *binary.BooleanOr,
		*binary.Equal, *binary.NotEqual, *binary.Identical, *binary.NotIdentical,
		*binary.Greater, *binary.GreaterOrEqual,
		*binary.Smaller, *binary.SmallerOrEqual,
		*expr.Empty, *expr.Isset:
		return meta.NewTypesMap("bool")
	case *binary.Mul:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Div:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Plus:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Minus:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *binary.Mod:
		return binaryMathOpType(sc, cs, n.Left, n.Right, custom)
	case *cast.Array:
		return meta.NewTypesMap("array")
	case *cast.Bool:
		return meta.NewTypesMap("bool")
	case *cast.Double:
		return meta.NewTypesMap("double")
	case *cast.Int:
		return meta.NewTypesMap("int")
	case *cast.String:
		return meta.NewTypesMap("string")
	case *expr.ConstFetch:
		nm, ok := n.Constant.(*name.Name)
		if !ok {
			return &meta.TypesMap{}
		}

		// TODO: handle namespaces
		p := nm.Parts
		if len(p) == 1 {
			constName := p[0].(*name.NamePart).Value

			if constName == "false" || constName == "true" {
				return meta.NewTypesMap("bool")
			}

			if constName == "null" {
				return meta.NewTypesMap("null")
			}

			return meta.NewTypesMap(meta.WrapConstant(constName))
		}
	case *scalar.String:
		return meta.NewTypesMap("string")
	case *scalar.Encapsed:
		return meta.NewTypesMap("string")
	case *scalar.Lnumber:
		return meta.NewTypesMap("int")
	case *scalar.Heredoc:
		return meta.NewTypesMap("string")
	case *scalar.Dnumber:
		return meta.NewTypesMap("double")
	case *expr.Ternary:
		t := ExprTypeLocalCustom(sc, cs, n.IfTrue, custom)
		f := ExprTypeLocalCustom(sc, cs, n.IfFalse, custom)
		return meta.NewEmptyTypesMap(t.Len() + f.Len()).Append(t).Append(f)
	case *expr.New:
		nm, ok := GetClassName(cs, n.Class)
		if ok {
			return meta.NewTypesMap(nm)
		}
		return &meta.TypesMap{}
	case *assign.Assign:
		return ExprTypeLocalCustom(sc, cs, n.Expression, custom)
	case *expr.Closure:
		return meta.NewTypesMap(`\Closure`)
	}

	return &meta.TypesMap{}
}
