package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

func Not(x meta.ConstantValue) meta.ConstantValue {
	v, ok := x.ToBool()
	if !ok {
		return meta.UnknownValue
	}
	return meta.ConstantBoolValue(!v)
}

func Neg(x meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		return meta.ConstantIntValue(-x.GetInt())
	case meta.Float:
		return meta.ConstantFloatValue(-x.GetFloat())
	}
	return meta.UnknownValue
}
