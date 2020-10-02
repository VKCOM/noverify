package meta

import (
	"fmt"
	"math"
	"strconv"
)

type ConstValueType uint8

//go:generate stringer -type=ConstValueType
const (
	Undefined ConstValueType = iota
	Integer
	Float
	String
	Bool
)

var (
	UnknownValue = ConstValue{Type: Undefined}
	TrueValue    = ConstValue{Type: Bool, Value: true}
	FalseValue   = ConstValue{Type: Bool, Value: false}
)

// ConstValue structure is used to store
// the value and type of a constant.
type ConstValue struct {
	Type  ConstValueType
	Value interface{}
}

// NewIntConst returns a new constant value with the
// preset int type and the passed value v.
func NewIntConst(v int64) ConstValue {
	return ConstValue{Type: Integer, Value: v}
}

// NewFloatConst returns a new constant value with the
// preset float type and the passed value v.
func NewFloatConst(v float64) ConstValue {
	return ConstValue{Type: Float, Value: v}
}

// NewStringConst returns a new constant value with the
// preset string type and the passed value v.
func NewStringConst(v string) ConstValue {
	return ConstValue{Type: String, Value: v}
}

// NewBoolConst returns a new constant value with the
// preset bool type and the passed value v.
func NewBoolConst(v bool) ConstValue {
	return ConstValue{Type: Bool, Value: v}
}

// GetInt returns the value stored in c.Value cast to int type.
//
// Used with care, it can panic if the type is not equal to the
// required one. Usually used in places where the type has already
// been clearly defined and the probability of panic is 0.
func (c ConstValue) GetInt() int64 {
	return c.Value.(int64)
}

// GetFloat returns the value stored in c.Value cast to float type.
//
// Used with care, it can panic if the type is not equal to the
// required one. Usually used in places where the type has already
// been clearly defined and the probability of panic is 0.
func (c ConstValue) GetFloat() float64 {
	return c.Value.(float64)
}

// GetString returns the value stored in c.Value cast to string type.
//
// Used with care, it can panic if the type is not equal to the
// required one. Usually used in places where the type has already
// been clearly defined and the probability of panic is 0.
func (c ConstValue) GetString() string {
	return c.Value.(string)
}

// GetBool returns the value stored in c.Value cast to bool type.
//
// Used with care, it can panic if the type is not equal to the
// required one. Usually used in places where the type has already
// been clearly defined and the probability of panic is 0.
func (c ConstValue) GetBool() bool {
	return c.Value.(bool)
}

// ToBool converts x constant to boolean constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func (c ConstValue) ToBool() (bool, bool) {
	switch c.Type {
	case Bool:
		return c.GetBool(), true
	case Integer:
		return c.GetInt() != 0, true
	case Float:
		eps := 1.11e-15
		return math.Abs(c.GetFloat()-0) < eps, true
	case String:
		v := c.GetString()
		return v != "" && v != "0", true
	}
	return false, false
}

// ToInt converts x constant to int constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func (c ConstValue) ToInt() (int64, bool) {
	switch c.Type {
	case Bool:
		if c.GetBool() {
			return 1, true
		}
		return 0, true
	case Integer:
		return c.GetInt(), true
	case Float:
		return int64(c.GetFloat()), true
	}
	return 0, false
}

// ToString converts x constant to string constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func (c ConstValue) ToString() (string, bool) {
	switch c.Type {
	case Bool:
		if c.GetBool() {
			return "1", true
		}
		return "", true
	case Integer:
		return strconv.FormatInt(c.GetInt(), 10), true
	case String:
		return c.GetString(), true
	}
	return "", false
}

// IsEqual checks for equality with the passed constant value.
//
// If any of the constants are undefined, false is returned.
func (c ConstValue) IsEqual(v ConstValue) bool {
	if v.Type == Undefined || c.Type == Undefined {
		return false
	}

	return c.Value == v.Value
}

func (c ConstValue) String() string {
	if c.Type == Undefined {
		return "Undefined type"
	}

	return fmt.Sprintf("%s(%v)", c.Type, c.Value)
}

func (c ConstValue) GobEncode() ([]byte, error) {
	switch c.Type {
	case Float:
		val, ok := c.Value.(float64)
		if !ok {
			return nil, fmt.Errorf("corrupted float")
		}
		str := fmt.Sprintf("%c%f", c.Type, val)
		return []byte(str), nil
	case Integer:
		val, ok := c.Value.(int64)
		if !ok {
			return nil, fmt.Errorf("corrupted integer")
		}
		str := fmt.Sprintf("%c%d", c.Type, val)
		return []byte(str), nil
	case String:
		val, ok := c.Value.(string)
		if !ok {
			return nil, fmt.Errorf("corrupted string")
		}
		str := fmt.Sprintf("%c%s", c.Type, val)
		return []byte(str), nil
	case Bool:
		val, ok := c.Value.(bool)
		if !ok {
			return nil, fmt.Errorf("corrupted bool")
		}
		x := "f"
		if val {
			x = "t"
		}
		str := fmt.Sprintf("%c%s", c.Type, x)
		return []byte(str), nil
	}

	return nil, fmt.Errorf("unhandeled type")
}

func (c *ConstValue) GobDecode(buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("data corrupted")
	}

	tp := ConstValueType(buf[0])
	buf = buf[1:]
	val := string(buf)

	switch tp {
	case Float:
		value, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid float")
		}
		c.Value = value
	case Integer:
		value, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer")
		}
		c.Value = value
	case String:
		c.Value = val
	case Bool:
		switch val {
		case "t":
			c.Value = true
		case "f":
			c.Value = false
		default:
			return fmt.Errorf("invalid bool: %q", val)
		}
	}

	c.Type = tp

	return nil
}
