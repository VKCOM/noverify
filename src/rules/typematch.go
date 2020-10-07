package rules

import (
	"strings"

	"github.com/VKCOM/noverify/src/phpdoc"
)

// TypesIsCompatible reports whether val type is compatible with dst type.
func TypeIsCompatible(dst, val phpdoc.TypeExpr) bool {
	// TODO: allow implementations to be compatible with interfaces.
	// TODO: allow derived classes to be compatible with base classes.

	for val.Kind == phpdoc.ExprParen {
		val = val.Args[0]
	}

	if val.Value == "mixed" {
		// Mixed type usually mean the we failed to infer the expression type.
		// We conservatively assume that mixed could be anything,
		// including something that is assignable to dst.
		// We may want to add a strict version of the type filter
		// that chooses a different stratage here.
		return false
	}

	switch dst.Kind {
	case phpdoc.ExprParen:
		return TypeIsCompatible(dst.Args[0], val)

	case phpdoc.ExprName:
		switch dst.Value {
		case "object":
			// For object we accept any kind of object instance.
			// https://wiki.php.net/rfc/object-typehint
			return val.Kind == dst.Kind && (val.Value == "object" || strings.HasPrefix(val.Value, `\`))
		case "array":
			return val.Kind == phpdoc.ExprArray
		}
		return val.Kind == dst.Kind && dst.Value == val.Value

	case phpdoc.ExprNot:
		return !TypeIsCompatible(dst.Args[0], val)

	case phpdoc.ExprNullable:
		if val.Value == "null" {
			return true
		}
		if TypeIsCompatible(dst.Args[0], val) {
			return true
		}
		return val.Kind == dst.Kind && TypeIsCompatible(dst.Args[0], val.Args[0])

	case phpdoc.ExprArray:
		return val.Kind == dst.Kind && TypeIsCompatible(dst.Args[0], val.Args[0])

	case phpdoc.ExprUnion:
		// TODO: sort the union types and avoid O(n^2) in the worst case?
		if val.Kind == phpdoc.ExprUnion {
			// Two union types are compatible if all their variants are compatible.
			for _, variant := range val.Args {
				if !TypeIsCompatible(dst, variant) {
					return false
				}
			}
			return true
		}
		for _, variant := range dst.Args {
			if TypeIsCompatible(variant, val) {
				return true
			}
		}
		return false

	case phpdoc.ExprInter:
		// TODO: make it work as intended. (See #310)
		return false

	default:
		return false
	}
}
