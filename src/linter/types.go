package linter

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
)

// TODO: reflect source line in shape names.

type warningString string

type shapeTypeProp struct {
	key   string
	types []meta.Type
}

type shapeTypeInfo struct {
	name  string
	props []shapeTypeProp
}

// typesFromPHPDoc extracts types out of the PHPDoc type string.
//
// No normalization is performed, but some PHPDoc-specific types
// are simplified to be compatible with our type system.
func typesFromPHPDoc(ctx *rootContext, typ phpdoc.Type) ([]meta.Type, warningString) {
	conv := phpdocTypeConverter{ctx: ctx}
	types := conv.mapType(typ.Expr)
	if conv.nullable {
		types = append(types, meta.Type{Elem: "null"})
	}
	return types, conv.warning
}

type phpdocTypeConverter struct {
	ctx      *rootContext
	warning  warningString
	nullable bool
}

func (conv *phpdocTypeConverter) mapType(e phpdoc.TypeExpr) []meta.Type {
	switch e.Kind {
	case phpdoc.ExprInvalid, phpdoc.ExprUnknown:
		if e.Value == "-" {
			conv.warn(`expected a type, found '-'; if you want to express 'any' type, use 'mixed'`)
			return []meta.Type{{Elem: "mixed"}}
		}

	case phpdoc.ExprParen:
		return conv.mapType(e.Args[0])

	case phpdoc.ExprName:
		if suggest, ok := typeAliases[e.Value]; ok {
			conv.warn(fmt.Sprintf("use %s type instead of %s", suggest, e.Value))
		}
		return []meta.Type{{Elem: e.Value}}

	case phpdoc.ExprMemberType:
		return []meta.Type{{Elem: "mixed"}}

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
		types := make([]meta.Type, 0, len(e.Args))
		for _, a := range e.Args {
			types = append(types, conv.mapType(a)...)
		}
		return types

	case phpdoc.ExprOptional:
		// Due to the fact that the optional keys for shape are processed in the mapShapeType
		// function, while the optionality for the key is cleared, and the key itself is not
		// processed by the mapType function, then in the mapType function the phpdoc.ExprOptional
		// type can only be in one case, if it is an incorrect syntax of the optional type.
		conv.warn(e.Value + ": nullable syntax is ?T, not T?")
	}

	return nil
}

func (conv *phpdocTypeConverter) mapArrayType(elem phpdoc.TypeExpr) []meta.Type {
	types := conv.mapType(elem)
	if len(types) == 0 {
		return []meta.Type{{Elem: "mixed", Dims: 1}}
	}
	for i := range types {
		types[i].Dims++
	}
	return types
}

func (conv *phpdocTypeConverter) mapShapeType(params []phpdoc.TypeExpr) []meta.Type {
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
		types := conv.mapType(typeExpr)

		// We need to resolve the class names as well as static,
		// self and $this here for it to work properly.
		for i, typ := range types {
			if _, ok := typeAliases[typ.Elem]; ok {
				continue
			}

			if trivialTypes[typ.Elem] {
				continue
			}

			className, ok := solver.GetClassName(conv.ctx.st, &ir.Name{Value: typ.Elem})
			if !ok {
				continue
			}

			types[i].Elem = className
		}
		if conv.nullable {
			types = append(types, meta.Type{
				Elem: "null",
				Dims: 0,
			})
		}

		props = append(props, shapeTypeProp{
			key:   key.Value,
			types: types,
		})
	}

	shape := shapeTypeInfo{
		name:  conv.ctx.generateShapeName(),
		props: props,
	}
	conv.ctx.shapes = append(conv.ctx.shapes, shape)

	return []meta.Type{{Elem: shape.name}}
}

func (conv *phpdocTypeConverter) mapTupleType(params []phpdoc.TypeExpr) []meta.Type {
	types := make([]phpdoc.TypeExpr, 0, len(params))
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
		types = append(types, typeExpr)
	}

	return conv.mapShapeType(types)
}

func (conv *phpdocTypeConverter) warn(msg string) {
	if conv.warning == "" {
		conv.warning = warningString(msg)
	}
}

// typesFromNode converts type hint node to meta types.
//
// No normalization is performed.
func typesFromNode(typeNode ir.Node) []meta.Type {
	n := typeNode

	var results []meta.Type
	if nullable, ok := typeNode.(*ir.Nullable); ok {
		n = nullable.Expr
		results = make([]meta.Type, 0, 2)
		results = append(results, meta.Type{Elem: "null"})
	} else {
		results = make([]meta.Type, 0, 1)
	}

	// There is a trick here.
	// Unlike with phpdoc types, having `integer` here
	// means that we need to force it to be interpreted as
	// `\integer`, not as `int`. This is why we prepend `\`.
	typ := meta.Type{Elem: meta.NameNodeToString(n)}
	if _, isAlias := typeAliases[typ.Elem]; isAlias {
		typ.Elem = `\` + typ.Elem
	}

	results = append(results, typ)

	return results
}

type typeNormalizer struct {
	st   *meta.ClassParseState
	kphp bool
}

func (n typeNormalizer) NormalizeTypes(types []meta.Type) {
	for i := range types {
		n.normalizeType(&types[i])
	}
}

func (n typeNormalizer) string2name(s string) *ir.Name {
	// TODO: Can avoid extra work by holding 1 tmp name inside
	// typeNormalizer, since we never need more than one at the time.
	return &ir.Name{Value: s}
}

func (n typeNormalizer) normalizeType(typ *meta.Type) {
	if trivialTypes[typ.Elem] {
		return
	}

	if typename, ok := typeAliases[typ.Elem]; ok {
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
		fullClassName, ok := solver.GetClassName(n.st, n.string2name(typ.Elem))
		if !ok {
			panic(fmt.Sprintf("can't expand type name: '%s'", typ.Elem))
		}
		typ.Elem = fullClassName
	}
}

var trivialTypes = map[string]bool{
	"bool":     true,
	"callable": true,
	"float":    true,
	"int":      true,
	"mixed":    true,
	"object":   true,
	"resource": true,
	"string":   true,
	"void":     true,
	"iterable": true,
	"array":    true,

	"null":  true,
	"true":  true,
	"false": true,
}

var typeAliases = map[string]string{
	"integer": "int",
	"long":    "int",

	"boolean": "bool",

	"real":   "float",
	"double": "float",

	"callback": "callable",
}
