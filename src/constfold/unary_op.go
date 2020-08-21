package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

// Not performs unary "!".
func Not(x meta.ConstantValue) meta.ConstantValue {
	v, ok := x.ToBool()
	if !ok {
		return meta.UnknownValue
	}
	return meta.ConstantBoolValue(!v)
}

// Neg performs unary "-".
func Neg(x meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		return meta.ConstantIntValue(-x.GetInt())
	case meta.Float:
		return meta.ConstantFloatValue(-x.GetFloat())
	}
	return meta.UnknownValue
}
