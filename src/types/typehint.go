package types

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/utils"
)

func NormalizedTypeHintTypes(normalizer Normalizer, typeNode ir.Node) Map {
	if typeNode == nil {
		return Map{}
	}

	typeList := TypeHintTypes(typeNode)
	return NewMapWithNormalization(normalizer, typeList)
}

func TypeHintHasMoreAccurateType(typeHintType, phpDocType Map) bool {
	// If is not array typehint.
	if !typeHintType.IsLazyArrayOf("mixed") {
		return true
	}

	// If it has more accurate type.
	if !phpDocType.Empty() {
		return true
	}

	return false
}

// TypeHintTypes converts type hint node to meta types.
//
// No normalization is performed.
func TypeHintTypes(typeNode ir.Node) []Type {
	var results []Type

	switch n := typeNode.(type) {
	case *ir.Union:
		results = make([]Type, 0, len(n.Types))

		for _, unionTyp := range n.Types {
			results = append(results, handleSingleType(unionTyp))
		}

		return results

	case *ir.Nullable:
		return []Type{
			{Elem: "null"},
			handleSingleType(n.Expr),
		}

	default:
		return []Type{handleSingleType(n)}
	}
}

func handleSingleType(n ir.Node) Type {
	// There is a trick here.
	// Unlike with phpdoc types, having `integer` here
	// means that we need to force it to be interpreted as
	// `\integer`, not as `int`. This is why we prepend `\`.
	typ := Type{Elem: utils.NameNodeToString(n)}
	if IsAlias(typ.Elem) {
		typ.Elem = `\` + typ.Elem
	}
	return typ
}
