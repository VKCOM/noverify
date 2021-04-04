package constfold

import (
	"github.com/VKCOM/noverify/src/meta"
)

// Plus performs arithmetic "+".
func Plus(x, y meta.ConstValue) meta.ConstValue {
	switch x.Type {
	case meta.Integer:
		switch y.Type {
		case meta.Integer:
			return meta.NewIntConst(x.GetInt() + y.GetInt())
		case meta.Float:
			return meta.NewFloatConst(float64(x.GetInt()) + y.GetFloat())
		}
	case meta.Float:
		switch y.Type {
		case meta.Integer:
			return meta.NewFloatConst(x.GetFloat() + float64(y.GetInt()))
		case meta.Float:
			return meta.NewFloatConst(x.GetFloat() + y.GetFloat())
		}
	}
	return meta.UnknownValue
}

// Minus performs arithmetic "-".
func Minus(x, y meta.ConstValue) meta.ConstValue {
	switch x.Type {
	case meta.Integer:
		switch y.Type {
		case meta.Integer:
			return meta.NewIntConst(x.GetInt() - y.GetInt())
		case meta.Float:
			return meta.NewFloatConst(float64(x.GetInt()) - y.GetFloat())
		}
	case meta.Float:
		switch y.Type {
		case meta.Integer:
			return meta.NewFloatConst(x.GetFloat() - float64(y.GetInt()))
		case meta.Float:
			return meta.NewFloatConst(x.GetFloat() - y.GetFloat())
		}
	}
	return meta.UnknownValue
}

// Mul performs arithmetic "*".
func Mul(x, y meta.ConstValue) meta.ConstValue {
	switch x.Type {
	case meta.Integer:
		switch y.Type {
		case meta.Integer:
			return meta.NewIntConst(x.GetInt() * y.GetInt())
		case meta.Float:
			return meta.NewFloatConst(float64(x.GetInt()) * y.GetFloat())
		}
	case meta.Float:
		switch y.Type {
		case meta.Integer:
			return meta.NewFloatConst(x.GetFloat() * float64(y.GetInt()))
		case meta.Float:
			return meta.NewFloatConst(x.GetFloat() * y.GetFloat())
		}
	}
	return meta.UnknownValue
}

// Concat performs string "." operation.
func Concat(x, y meta.ConstValue) meta.ConstValue {
	v1, ok1 := x.ToString()
	v2, ok2 := y.ToString()
	if ok1 && ok2 {
		return meta.NewStringConst(v1 + v2)
	}
	return meta.UnknownValue
}

// Or performs logical "||".
// Also works for "or" operator.
func Or(x, y meta.ConstValue) meta.ConstValue {
	v1, ok1 := x.ToBool()
	v2, ok2 := y.ToBool()
	if ok1 && ok2 {
		return meta.NewBoolConst(v1 || v2)
	}
	return meta.UnknownValue
}

// And performs logical "&&".
// Also works for "and" operator.
func And(x, y meta.ConstValue) meta.ConstValue {
	v1, ok1 := x.ToBool()
	v2, ok2 := y.ToBool()
	if ok1 && ok2 {
		return meta.NewBoolConst(v1 && v2)
	}
	return meta.UnknownValue
}

// BitOr performs bitwise "|".
func BitOr(x, y meta.ConstValue) meta.ConstValue {
	v1, ok1 := x.ToInt()
	v2, ok2 := y.ToInt()
	if ok1 && ok2 {
		return meta.NewIntConst(v1 | v2)
	}
	return meta.UnknownValue
}

// BitAnd performs bitwise "&".
func BitAnd(x, y meta.ConstValue) meta.ConstValue {
	v1, ok1 := x.ToInt()
	v2, ok2 := y.ToInt()
	if ok1 && ok2 {
		return meta.NewIntConst(v1 & v2)
	}
	return meta.UnknownValue
}
