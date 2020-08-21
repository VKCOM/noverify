package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

func Not(x meta.ConstantValue) meta.ConstantValue {
	v, ok := ToBool(x)
	if !ok {
		return meta.UnknownValue
	}
	return meta.ConstantBoolValue(!v)
}

func Neg(x meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		return meta.ConstantIntValue(-x.ToInt())
	case meta.Float:
		return meta.ConstantFloatValue(-x.ToFloat())
	}
	return meta.UnknownValue
}
