package phpdoctypes

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/linter/autogen"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/types"
)

type RealPHPDocTypes struct {
	Types    []types.Type
	Shapes   types.ShapesMap
	Closures types.ClosureMap
	Warning  string
}

// ToRealType extracts types out of the PHPDoc type string.
//
// No normalization is performed, but some PHPDoc-specific types
// are simplified to be compatible with our type system.
func ToRealType(classFQNProvider func(string) (string, bool), kphp bool, typ phpdoc.Type) *RealPHPDocTypes {
	conv := TypeConverter{
		classFQNProvider:   classFQNProvider,
		shapeNameGenerator: autogen.GenerateShapeName,
		shapes:             make(types.ShapesMap),
		closures:           make(types.ClosureMap),
		kphp:               kphp,
	}

	result := conv.mapType(typ.Expr)
	if conv.nullable {
		alreadyHasNull := false

		for _, tp := range result {
			if tp.Elem == "null" {
				alreadyHasNull = true
				conv.warning = "Repeated nullable doesn't make sense"
				break
			}
		}

		if !alreadyHasNull {
			result = append(result, types.Type{Elem: "null"})
		}
	}

	return &RealPHPDocTypes{
		Types:    result,
		Shapes:   conv.shapes,
		Closures: conv.closures,
		Warning:  conv.warning,
	}
}

type TypeConverter struct {
	classFQNProvider   func(string) (string, bool)
	shapeNameGenerator func([]types.ShapeProp) string
	warning            string
	nullable           bool
	shapes             types.ShapesMap
	closures           types.ClosureMap
	kphp               bool
}

func (conv *TypeConverter) mapType(e phpdoc.TypeExpr) []types.Type {
	switch e.Kind {
	case phpdoc.ExprInvalid, phpdoc.ExprUnknown:
		if e.Value == "-" {
			conv.warn(`Expected a type, found '-'; if you want to express 'any' type, use 'mixed'`)
			return []types.Type{{Elem: "mixed"}}
		}

	case phpdoc.ExprParen:
		return conv.mapType(e.Args[0])

	case phpdoc.ExprName:
		if suggest, has := types.Alias(e.Value); has {
			conv.warn(fmt.Sprintf("Use %s type instead of %s", suggest, e.Value))
		}
		return []types.Type{{Elem: e.Value}}

	case phpdoc.ExprMemberType:
		return []types.Type{{Elem: "mixed"}}

	case phpdoc.ExprGeneric:
		typ := e.Args[0]
		params := e.Args[1:]

		isCallable := strings.Contains(typ.Value, "callable")
		if isCallable {
			conv.warn("Lost return type for callable(...), if the function returns nothing, specify void explicitly")
		}

		isArray := typ.Value == "array" ||
			typ.Value == "list" ||
			typ.Value == "iterable" ||
			strings.Contains(typ.Value, "-")

		if isArray {
			if e.Shape == phpdoc.ShapeGenericBrace {
				return conv.mapShapeType(params, true)
			}
			switch len(params) {
			case 1:
				return conv.mapArrayType(params[0])
			case 2:
				return conv.mapArrayType(params[1])
			}
		}
		if typ.Value == "shape" || typ.Value == `\shape` {
			return conv.mapShapeType(params, false)
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
			conv.warn("Array syntax is T[], not []T")
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
		// processed by the mapType function, then in the mapType function the ExprOptional
		// type can only be in one case, if it is an incorrect syntax of the optional type.
		conv.warn("Nullable syntax is ?T, not T?")

	case phpdoc.ExprLiteral:
		return []types.Type{{Elem: "string"}}

	case phpdoc.ExprTypedCallable:
		closureName := `\Closure$(`
		argsStart := 0
		var returnType phpdoc.TypeExpr
		if strings.IndexByte(e.Value, ':') != -1 {
			returnType = e.Args[0]
			argsStart = 1
		}

		for i := argsStart; i < len(e.Args); i++ {
			closureName += e.Args[i].Value
			if i != len(e.Args)-1 {
				closureName += ","
			}
		}

		closureName += ")"
		if returnType.Value != "" {
			closureName += ":" + strings.ReplaceAll(returnType.Value, "|", "/")
		} else {
			returnType = phpdoc.TypeExpr{
				Kind:  phpdoc.ExprName,
				Value: "void",
			}
		}

		var argsTypes [][]types.Type
		for _, arg := range e.Args[argsStart:] {
			argsTypes = append(argsTypes, conv.mapType(arg))
		}

		closure := types.ClosureInfo{
			Name:       closureName,
			ReturnType: conv.mapType(returnType),
			ParamTypes: argsTypes,
		}
		conv.closures[closure.Name] = closure

		return []types.Type{{Elem: closureName}}
	}

	return nil
}

func (conv *TypeConverter) mapArrayType(elem phpdoc.TypeExpr) []types.Type {
	typeList := conv.mapType(elem)
	if len(typeList) == 0 {
		return []types.Type{{Elem: "mixed", Dims: 1}}
	}
	for i := range typeList {
		typeList[i].Dims++
	}
	return typeList
}

func (conv *TypeConverter) mapShapeType(params []phpdoc.TypeExpr, allowedMixing bool) []types.Type {
	props := make([]types.ShapeProp, 0, len(params))
	for i, p := range params {
		if p.Value == "*" || p.Value == "..." {
			continue
		}
		if p.Kind != phpdoc.ExprKeyVal {
			if !allowedMixing {
				conv.warn(fmt.Sprintf("Shape param #%d: want key:type, found %s", i+1, p.Value))
				continue
			}

			typeList := conv.mapType(p)

			props = append(props, types.ShapeProp{
				Key:   fmt.Sprintf("%d", i),
				Types: typeList,
			})

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
			conv.warn(fmt.Sprintf("Invalid shape key: %s", key.Value))
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

			if conv.kphp && (typ.Elem == "any" || typ.Elem == "kmixed" || typ.Elem == "future" || typ.Elem == "future_queue") {
				continue
			}

			if conv.classFQNProvider == nil {
				continue
			}

			className, ok := conv.classFQNProvider(typ.Elem)
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
			conv.nullable = false
		}

		props = append(props, types.ShapeProp{
			Key:   key.Value,
			Types: typeList,
		})
	}

	shape := types.ShapeInfo{
		Name:  conv.shapeNameGenerator(props),
		Props: props,
	}
	conv.shapes[shape.Name] = shape

	return []types.Type{{Elem: shape.Name}}
}

func (conv *TypeConverter) mapTupleType(params []phpdoc.TypeExpr) []types.Type {
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

	return conv.mapShapeType(typeList, false)
}

func (conv *TypeConverter) warn(msg string) {
	if conv.warning == "" {
		conv.warning = msg
	}
}
