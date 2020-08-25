package meta

import (
	"fmt"
	"math"
	"strconv"
)

type ConstantValueType uint8

//go:generate stringer -type=ConstantValueType
const (
	Undefined ConstantValueType = iota
	Integer
	Float
	String
	Bool
)

var (
	UnknownValue = ConstantValue{Type: Undefined}
	TrueValue    = ConstantValue{Type: Bool, Value: true}
	FalseValue   = ConstantValue{Type: Bool, Value: false}
)

type ConstantValue struct {
	Type  ConstantValueType
	Value interface{}
}

func ConstantIntValue(v int64) ConstantValue {
	return ConstantValue{Type: Integer, Value: v}
}

func ConstantFloatValue(v float64) ConstantValue {
	return ConstantValue{Type: Float, Value: v}
}

func ConstantStringValue(v string) ConstantValue {
	return ConstantValue{Type: String, Value: v}
}

func ConstantBoolValue(v bool) ConstantValue {
	return ConstantValue{Type: Bool, Value: v}
}

// GetInt returns the value stored in c.Value cast to int type.
func (c ConstantValue) GetInt() int64 {
	return c.Value.(int64)
}

// GetFloat returns the value stored in c.Value cast to float type.
func (c ConstantValue) GetFloat() float64 {
	return c.Value.(float64)
}

// GetString returns the value stored in c.Value cast to string type.
func (c ConstantValue) GetString() string {
	return c.Value.(string)
}

// GetBool returns the value stored in c.Value cast to bool type.
func (c ConstantValue) GetBool() bool {
	return c.Value.(bool)
}

// ToBool converts x constant to boolean constants following PHP conversion rules.
// Second bool result tells whether that conversion was successful.
func (c ConstantValue) ToBool() (bool, bool) {
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
func (c ConstantValue) ToInt() (int64, bool) {
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
func (c ConstantValue) ToString() (string, bool) {
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

func (c ConstantValue) GobEncode() ([]byte, error) {
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

func (c *ConstantValue) GobDecode(buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("data corrupted")
	}

	tp := ConstantValueType(buf[0])
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

func (c ConstantValue) String() string {
	if c.Type == Undefined {
		return "Undefined type"
	}

	return fmt.Sprintf("%s(%v)", c.Type, c.Value)
}

func (c ConstantValue) IsEqual(v ConstantValue) bool {
	if v.Type == Undefined || c.Type == Undefined {
		return false
	}

	return c.Value == v.Value
}

type ConstantInfo struct {
	Pos         ElementPosition
	Typ         TypesMap
	AccessLevel AccessLevel
	Value       ConstantValue
}
