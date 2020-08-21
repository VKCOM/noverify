package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

// ToBool converts x constant to boolean constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToBool(x meta.ConstantValue) (meta.ConstantValue, bool) {
	switch x.Type {
	case meta.Bool:
		return meta.ConstantValue{Type: meta.Bool, Value: x.Value.(bool)}, true
	case meta.Integer:
		return meta.ConstantValue{Type: meta.Bool, Value: x.Value.(int64) != 0}, true
	case meta.Float:
		return meta.ConstantValue{Type: meta.Bool, Value: x.Value.(float64) != 0}, true
	case meta.String:
		return meta.ConstantValue{Type: meta.Bool, Value: x.Value.(string) != "" && x.Value.(string) != "0"}, true
	}
	return meta.UnknownValue, false
}
