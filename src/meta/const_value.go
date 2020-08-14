package meta

import (
	"fmt"
	"strconv"
)

type ConstantValueType uint8

// TODO: add bool const values?

//go:generate stringer -type=ConstantValueType
const (
	Undefined ConstantValueType = iota
	Integer
	Float
	String
)

var UnknownValue = ConstantValue{Type: Undefined}

type ConstantValue struct {
	Type  ConstantValueType
	Value interface{}
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
