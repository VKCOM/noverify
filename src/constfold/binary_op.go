package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

// Plus performs arithmetic "+".
func Plus(x, y meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		if y.Type == meta.Integer {
			v := x.Value.(int64) + y.Value.(int64)
			return meta.ConstantValue{Type: meta.Integer, Value: v}
		}
	case meta.Float:
		if y.Type == meta.Float {
			v := x.Value.(float64) + y.Value.(float64)
			return meta.ConstantValue{Type: meta.Float, Value: v}
		}
	}
	return meta.UnknownValue
}

// Mul performs arithmetic "*".
func Mul(x, y meta.ConstantValue) meta.ConstantValue {
	switch x.Type {
	case meta.Integer:
		if y.Type == meta.Integer {
			v := x.Value.(int64) * y.Value.(int64)
			return meta.ConstantValue{Type: meta.Integer, Value: v}
		}
	case meta.Float:
		if y.Type == meta.Float {
			v := x.Value.(float64) * y.Value.(float64)
			return meta.ConstantValue{Type: meta.Float, Value: v}
		}
	}
	return meta.UnknownValue
}
