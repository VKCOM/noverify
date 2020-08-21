package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

func Not(x meta.ConstantValue) meta.ConstantValue {
	v, ok := ToBool(x)
	if !ok {
		return meta.UnknownValue
	}
	return meta.BoolValue(!v)
}

func Neg(x meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		return meta.IntValue(-x.Value.(int64))
	case meta.Float:
		return meta.FloatValue(-x.Value.(float64))
	}
	return meta.UnknownValue
}
