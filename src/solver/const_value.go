package solver

import (
	"strconv"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
)

var undefinedValue = meta.ConstantValue{Type: meta.Undefined}

func GetConstantValue(c ir.Node) (meta.ConstantValue, bool) {
	switch c := c.(type) {
	case *ir.Lnumber:
		value, err := strconv.ParseInt(c.Value, 10, 64)
		return meta.ConstantValue{Type: meta.Integer, Value: value}, err == nil

	case *ir.Dnumber:
		value, err := strconv.ParseFloat(c.Value, 64)
		return meta.ConstantValue{Type: meta.Float, Value: value}, err == nil

	case *ir.String:
		v := irutil.Unquote(c.Value)
		return meta.ConstantValue{Value: v, Type: meta.String}, true

	case *ir.UnaryMinusExpr:
		v, ok := GetConstantValue(c.Expr)
		if !ok {
			return undefinedValue, false
		}
		switch value := v.Value.(type) {
		case int64:
			return meta.ConstantValue{Type: meta.Integer, Value: -value}, true
		case float64:
			return meta.ConstantValue{Type: meta.Float, Value: -value}, true
		}
		return v, true

	default:
		return undefinedValue, false
	}
}
