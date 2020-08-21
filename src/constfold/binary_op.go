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
			return meta.IntValue(v)
		}
	case meta.Float:
		if y.Type == meta.Float {
			v := x.Value.(float64) + y.Value.(float64)
			return meta.FloatValue(v)
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
			return meta.IntValue(v)
		}
	case meta.Float:
		if y.Type == meta.Float {
			v := x.Value.(float64) * y.Value.(float64)
			return meta.FloatValue(v)
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
