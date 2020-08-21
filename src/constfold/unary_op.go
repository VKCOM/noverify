package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

func Not(x meta.ConstantValue) meta.ConstantValue {
	v, ok := ToBool(x)
	if !ok {
		return meta.UnknownValue
	}
	return meta.ConstantValue{Type: meta.Bool, Value: !v.Value.(bool)}
}

func Neg(x meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		return meta.ConstantValue{Type: meta.Integer, Value: -x.Value.(int64)}
	case meta.Float:
		return meta.ConstantValue{Type: meta.Float, Value: -x.Value.(float64)}
	}
	return meta.UnknownValue
}
