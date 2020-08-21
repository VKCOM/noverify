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
		return x.ToBool(), true
	case meta.Integer:
		return x.ToInt() != 0, true
	case meta.Float:
		return x.ToFloat() != 0, true
	case meta.String:
		v := x.ToString()
		return v != "" && v != "0", true
	}
	return false, false
}

// ToString converts x constant to string constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToString(x meta.ConstantValue) (string, bool) {
	switch x.Type {
	case meta.Bool:
		if x.ToBool() {
			return "1", true
		}
		return "", true
	case meta.Integer:
		return strconv.FormatInt(x.ToInt(), 10), true
	case meta.String:
		return x.ToString(), true
	}
	return "", false
}
