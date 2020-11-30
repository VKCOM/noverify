package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

// Not performs unary "!".
func Not(x meta.ConstValue) meta.ConstValue {
	v, ok := x.ToBool()
	if !ok {
		return meta.UnknownValue
	}
	return meta.NewBoolConst(!v)
}

// Neg performs unary "-".
func Neg(x meta.ConstValue) meta.ConstValue {
	switch x.Type {
	case meta.Integer:
		return meta.NewIntConst(-x.GetInt())
	case meta.Float:
		return meta.NewFloatConst(-x.GetFloat())
	}
	return meta.UnknownValue
}
