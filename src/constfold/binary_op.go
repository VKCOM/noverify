package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

// Plus performs arithmetic "+".
func Plus(x, y meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		if y.Type == meta.Integer {
			return meta.IntValue(x.ToInt() + y.ToInt())
		}
	case meta.Float:
		if y.Type == meta.Float {
			return meta.FloatValue(x.ToFloat() + y.ToFloat())
		}
	}
	return meta.UnknownValue
}

// Minus performs arithmetic "-".
func Minus(x, y meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		if y.Type == meta.Integer {
			return meta.IntValue(x.ToInt() - y.ToInt())
		}
	case meta.Float:
		if y.Type == meta.Float {
			return meta.FloatValue(x.ToFloat() - y.ToFloat())
		}
	}
	return meta.UnknownValue
}

// Mul performs arithmetic "*".
func Mul(x, y meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		if y.Type == meta.Integer {
			return meta.IntValue(x.ToInt() * y.ToInt())
		}
	case meta.Float:
		if y.Type == meta.Float {
			return meta.FloatValue(x.ToFloat() * y.ToFloat())
		}
	}
	return meta.UnknownValue
}

// Concat performs string "." operation.
func Concat(x, y meta.ConstantValue) meta.ConstantValue {
	v1, ok1 := ToString(x)
	v2, ok2 := ToString(y)
	if ok1 && ok2 {
		return meta.StringValue(v1 + v2)
	}
	return meta.UnknownValue
}

// Or performs logical "||".
// Also works for "or" operator.
func Or(x, y meta.ConstantValue) meta.ConstantValue {
	v1, ok1 := ToBool(x)
	v2, ok2 := ToBool(y)
	switch {
	case ok1 && v1:
		return meta.TrueValue
	case ok2 && v2:
		return meta.TrueValue
	case ok1 && ok2:
		return meta.BoolValue(v1 || v2)
	default:
		return meta.UnknownValue
	}
}

// And performs logical "&&".
// Also works for "and" operator.
func And(x, y meta.ConstantValue) meta.ConstantValue {
	v1, ok1 := ToBool(x)
	v2, ok2 := ToBool(y)
	switch {
	case ok1 && v1:
		return meta.FalseValue
	case ok2 && v2:
		return meta.FalseValue
	case ok1 && ok2:
		return meta.BoolValue(v1 && v2)
	default:
		return meta.UnknownValue
	}
}

// BitOr performs bitwise "|".
func BitOr(x, y meta.ConstantValue) meta.ConstantValue {
	v1, ok1 := ToInt(x)
	v2, ok2 := ToInt(y)
	if ok1 && ok2 {
		return meta.IntValue(v1 | v2)
	}
	return meta.UnknownValue
}

// BitAnd performs bitwise "&".
func BitAnd(x, y meta.ConstantValue) meta.ConstantValue {
	v1, ok1 := ToInt(x)
	v2, ok2 := ToInt(y)
	if ok1 && ok2 {
		return meta.IntValue(v1 & v2)
	}
	return meta.UnknownValue
}
