package types

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/utils"
)

func NormalizedTypeHintTypes(normalizer Normalizer, typeNode ir.Node) Map {
	typeList := TypeHintTypes(typeNode)
	return NewMapWithNormalization(normalizer, typeList)
}

// TypeHintTypes converts type hint node to meta types.
//
// No normalization is performed.
func TypeHintTypes(typeNode ir.Node) []Type {
	n := typeNode

	var results []Type
	if nullable, ok := typeNode.(*ir.Nullable); ok {
		n = nullable.Expr
		results = make([]Type, 0, 2)
		results = append(results, Type{Elem: "null"})
	} else {
		results = make([]Type, 0, 1)
	}

	// There is a trick here.
	// Unlike with phpdoc types, having `integer` here
	// means that we need to force it to be interpreted as
	// `\integer`, not as `int`. This is why we prepend `\`.
	typ := Type{Elem: utils.NameNodeToString(n)}
	if IsAlias(typ.Elem) {
		typ.Elem = `\` + typ.Elem
	}

	results = append(results, typ)

	return results
}
