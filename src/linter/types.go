package linter

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
)

// TODO: reflect source line in shape names.

type warningString string

type shapeTypeProp struct {
	key   string
	types []types.Type
}

type shapeTypeInfo struct {
	name  string
	props []shapeTypeProp
}

// typesFromPHPDoc extracts types out of the PHPDoc type string.
//
// No normalization is performed, but some PHPDoc-specific types
// are simplified to be compatible with our type system.
func typesFromPHPDoc(ctx *rootContext, typ phpdoc.Type) ([]types.Type, warningString) {
	conv := phpdocTypeConverter{ctx: ctx}
	result := conv.mapType(typ.Expr)
	if conv.nullable {
		result = append(result, types.Type{Elem: "null"})
	}
	return result, conv.warning
}

type phpdocTypeConverter struct {
	ctx      *rootContext
	warning  warningString
	nullable bool
}

func (conv *phpdocTypeConverter) mapType(e phpdoc.TypeExpr) []types.Type {
	switch e.Kind {
	case phpdoc.ExprInvalid, phpdoc.ExprUnknown:
		if e.Value == "-" {
			conv.warn(`expected a type, found '-'; if you want to express 'any' type, use 'mixed'`)
			return []types.Type{{Elem: "mixed"}}
		}

	case phpdoc.ExprParen:
		return conv.mapType(e.Args[0])

	case phpdoc.ExprName:
		if suggest, has := types.Alias(e.Value); has {
			conv.warn(fmt.Sprintf("use %s type instead of %s", suggest, e.Value))
		}
		return []types.Type{{Elem: e.Value}}

	case phpdoc.ExprMemberType:
		return []types.Type{{Elem: "mixed"}}

	case phpdoc.ExprGeneric:
		typ := e.Args[0]
		params := e.Args[1:]
		if typ.Value == "array" {
			if e.Shape == phpdoc.ShapeGenericBrace {
				return conv.mapShapeType(params)
			}
			switch len(params) {
			case 1:
				return conv.mapArrayType(params[0])
			case 2:
				return conv.mapArrayType(params[1])
			}
		}
		if typ.Value == "shape" || typ.Value == `\shape` {
			return conv.mapShapeType(params)
		}
		if typ.Value == "tuple" || typ.Value == `\tuple` {
			return conv.mapTupleType(params)
		}

		return conv.mapType(typ)

	case phpdoc.ExprNullable:
		conv.nullable = true
		return conv.mapType(e.Args[0])

	case phpdoc.ExprArray:
		if e.Shape == phpdoc.ShapeArrayPrefix {
			conv.warn(e.Value + ": array syntax is T[], not []T")
		}
		return conv.mapArrayType(e.Args[0])

	case phpdoc.ExprUnion:
		typeList := make([]types.Type, 0, len(e.Args))
		for _, a := range e.Args {
			typeList = append(typeList, conv.mapType(a)...)
		}
		return typeList

	case phpdoc.ExprOptional:
		// Due to the fact that the optional keys for shape are processed in the mapShapeType
		// function, while the optionality for the key is cleared, and the key itself is not
		// processed by the mapType function, then in the mapType function the phpdoc.ExprOptional
		// type can only be in one case, if it is an incorrect syntax of the optional type.
		conv.warn(e.Value + ": nullable syntax is ?T, not T?")
	}

	return nil
}

func (conv *phpdocTypeConverter) mapArrayType(elem phpdoc.TypeExpr) []types.Type {
	typeList := conv.mapType(elem)
	if len(typeList) == 0 {
		return []types.Type{{Elem: "mixed", Dims: 1}}
	}
	for i := range typeList {
		typeList[i].Dims++
	}
	return typeList
}

func (conv *phpdocTypeConverter) mapShapeType(params []phpdoc.TypeExpr) []types.Type {
	props := make([]shapeTypeProp, 0, len(params))
	for i, p := range params {
		if p.Value == "*" || p.Value == "..." {
			continue
		}
		if p.Kind != phpdoc.ExprKeyVal {
			conv.warn(fmt.Sprintf("shape param #%d: want key:type, found %s", i+1, p.Value))
			continue
		}
		key := p.Args[0]
		typeExpr := p.Args[1]
		if key.Kind == phpdoc.ExprOptional {
			key = key.Args[0]
		}
		switch key.Kind {
		case phpdoc.ExprName, phpdoc.ExprInt:
			// OK.
		default:
			conv.warn(fmt.Sprintf("invalid shape key: %s", key.Value))
			continue
		}
		typeList := conv.mapType(typeExpr)

		// We need to resolve the class names as well as static,
		// self and $this here for it to work properly.
		for i, typ := range typeList {
			if types.IsAlias(typ.Elem) {
				continue
			}

			if types.IsTrivial(typ.Elem) {
				continue
			}

			if typ.Elem == "array" {
				continue
			}

			className, ok := solver.GetClassName(conv.ctx.st, &ir.Name{Value: typ.Elem})
			if !ok {
				continue
			}

			typeList[i].Elem = className
		}
		if conv.nullable {
			typeList = append(typeList, types.Type{
				Elem: "null",
				Dims: 0,
			})
		}

		props = append(props, shapeTypeProp{
			key:   key.Value,
			types: typeList,
		})
	}

	shape := shapeTypeInfo{
		name:  conv.ctx.generateShapeName(),
		props: props,
	}
	conv.ctx.shapes = append(conv.ctx.shapes, shape)

	return []types.Type{{Elem: shape.name}}
}

func (conv *phpdocTypeConverter) mapTupleType(params []phpdoc.TypeExpr) []types.Type {
	typeList := make([]phpdoc.TypeExpr, 0, len(params))
	for i, p := range params {
		if p.Value == "*" || p.Value == "..." {
			continue
		}

		key := phpdoc.TypeExpr{
			Kind:  phpdoc.ExprInt,
			Value: fmt.Sprint(i),
		}
		typ := p
		args := []phpdoc.TypeExpr{key, typ}

		typeExpr := phpdoc.TypeExpr{
			Kind: phpdoc.ExprKeyVal,
			Args: args,
		}
		typeList = append(typeList, typeExpr)
	}

	return conv.mapShapeType(typeList)
}

func (conv *phpdocTypeConverter) warn(msg string) {
	if conv.warning == "" {
		conv.warning = warningString(msg)
	}
}

// typesFromNode converts type hint node to meta types.
//
// No normalization is performed.
func typesFromNode(typeNode ir.Node) []types.Type {
	n := typeNode

	var results []types.Type
	if nullable, ok := typeNode.(*ir.Nullable); ok {
		n = nullable.Expr
		results = make([]types.Type, 0, 2)
		results = append(results, types.Type{Elem: "null"})
	} else {
		results = make([]types.Type, 0, 1)
	}

	// There is a trick here.
	// Unlike with phpdoc types, having `integer` here
	// means that we need to force it to be interpreted as
	// `\integer`, not as `int`. This is why we prepend `\`.
	typ := types.Type{Elem: meta.NameNodeToString(n)}
	if types.IsAlias(typ.Elem) {
		typ.Elem = `\` + typ.Elem
	}

	results = append(results, typ)

	return results
}

type typeNormalizer struct {
	st   *meta.ClassParseState
	kphp bool
}

func (n typeNormalizer) NormalizeTypes(typeList []types.Type) {
	for i := range typeList {
		n.normalizeType(&typeList[i])
	}
}

func (n typeNormalizer) normalizeType(typ *types.Type) {
	if types.IsTrivial(typ.Elem) {
		return
	}

	if typename, has := types.Alias(typ.Elem); has {
		typ.Elem = typename
		return
	}

	if typ.Elem == "any" && n.kphp {
		// `any` is a special KPHP type that is more-or-less
		// identical to `mixed|object`. In PHP, `mixed` already covers
		// objects, so there is no need to add `object`.
		// See https://php.watch/versions/8.0/mixed-type
		typ.Elem = "mixed"
		return
	}

	switch typ.Elem {
	case "array":
		// Rewrite `array` to `mixed[]`.
		// If it's `array[]`, it'll become `mixed[][]`.
		typ.Dims++
		typ.Elem = "mixed"
	case "$this":
		// Handle `$this` as `static` alias in phpdoc context.
		typ.Elem = "static"
	case "static":
		// Don't replace `static` phpdoc type annotation too early
		// to make it possible to handle late static binding.
	default:
		if typ.Elem[0] == '\\' {
			return // Already FQN?
		}
		fullClassName, ok := solver.GetClassName(n.st, &ir.Name{Value: typ.Elem})
		if !ok {
			panic(fmt.Sprintf("can't expand type name: '%s'", typ.Elem))
		}
		typ.Elem = fullClassName
	}
}
