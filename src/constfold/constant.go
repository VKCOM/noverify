package constfold

import (
	"strconv"

	"github.com/VKCOM/noverify/src/meta"
)

// ToBool converts x constant to boolean constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToBool(x meta.ConstantValue) (bool, bool) {
	switch x.Type {
	case meta.Bool:
		return x.Value.(bool), true
	case meta.Integer:
		return x.Value.(int64) != 0, true
	case meta.Float:
		return x.Value.(float64) != 0, true
	case meta.String:
		return x.Value.(string) != "" && x.Value.(string) != "0", true
	}
	return false, false
}

// ToString converts x constant to string constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToString(x meta.ConstantValue) (string, bool) {
	switch x.Type {
	case meta.Bool:
		if x.Value.(bool) {
			return "1", true
		}
		return "", true
	case meta.Integer:
		return strconv.FormatInt(x.Value.(int64), 10), true
	case meta.String:
		return x.Value.(string), true
	}
	return "", false
}
