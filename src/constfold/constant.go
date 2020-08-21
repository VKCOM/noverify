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
		return x.GetBool(), true
	case meta.Integer:
		return x.GetInt() != 0, true
	case meta.Float:
		return x.GetFloat() != 0, true
	case meta.String:
		v := x.GetString()
		return v != "" && v != "0", true
	}
	return false, false
}

// ToInt converts x constant to int constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToInt(x meta.ConstantValue) (int64, bool) {
	switch x.Type {
	case meta.Bool:
		if x.GetBool() {
			return 1, true
		}
		return 0, true
	case meta.Integer:
		return x.GetInt(), true
	case meta.Float:
		return int64(x.GetFloat()), true
	}
	return 0, false
}

// ToString converts x constant to string constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func ToString(x meta.ConstantValue) (string, bool) {
	switch x.Type {
	case meta.Bool:
		if x.GetBool() {
			return "1", true
		}
		return "", true
	case meta.Integer:
		return strconv.FormatInt(x.GetInt(), 10), true
	case meta.String:
		return x.GetString(), true
	}
	return "", false
}
