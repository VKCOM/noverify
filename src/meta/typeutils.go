package meta

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// Constants for lazy ("wrap") types:
	// Here "<string>" means 2 bytes of length followed by string contents.
	// "<uint8>" means 1 byte field.
	// <uint8> fields must go before <string> fields.
	//
	// Note: both string length and <uint8> are represented using hex encoding.
	// One of the reasons is to avoid `|` inside type strings that were wrapped.
	// See tests for more details.

	// WStaticMethodCall type is "Wrap Static Method Call":
	// E.g. Class::doSomething()
	// Params: [Class name <string>] [Method name <string>]
	WStaticMethodCall byte = iota

	// WInstanceMethodCall is a method call on some expression.
	// You need to specify expression type (might be lazy type, e.g. <WStaticMethodCall, SomeClass, instance> ).
	// E.g. $var->callSomething()
	// Params: [Expression type <string>] [Method <string>]
	WInstanceMethodCall

	// WStaticPropertyFetch is a property fetch for static property :).
	// E.g. Test::$something
	// Params: [Class name <string>] [Property name with $ <string>]
	WStaticPropertyFetch

	// WClassConstFetch is a const fetch from a class.
	// E.g. Test::CONSTANT
	// Params: [Class name <string>] [Constant name <string>]
	WClassConstFetch

	// WInstancePropertyFetch is a property fetch from some instance.
	// You need to provide expression type, see example for WInstanceMethodCall.
	// E.g. $var->something
	// Params: [Expression type <string>] [Property name <string>]
	WInstancePropertyFetch

	// WFunctionCall represents a function call.
	// Function name must contain namespace. It will be first searched in the defined namespace
	// and then it will fall back to root namespace.
	// E.g. callSomething()
	// Params: [Function name with full namespace <string>]
	WFunctionCall

	// WArrayOf means that expression is array of another expression
	// E.g. <WArrayOf, string> would be normally written as "string[]"
	//      <WArrayOf, <WFunctionCall, callSomething>>
	// Params: [Expression type <string>]
	WArrayOf

	// WElemOf is the opposite of WArrayOf: it means the type of an element of the expression
	// E.g. $arr[0] would be "string" if $arr type is "string[]"
	// Params: [Expression type <string>]
	WElemOf

	// WElemOfKey is extended for of WElemOf where we also save the key that
	// was used during the indexing.
	// Params: [Expression type <string>] [Key <string>]
	WElemOfKey

	// WGlobal means global variable.
	// E.g. global $Something;
	// Params: [Global variable name <string>]
	WGlobal

	// WConstant means constant
	// e.g. type of MINUTE constant
	// Params: [Constant name <string>]
	WConstant

	// WBaseMethodParam is a way to inherit base type method type of nth parameter.
	// e.g. type of $x param of foo method from one of the implemented interfaces.
	// Params: [Index <uint8>] [Class name <string>] [Method name <string>]
	WBaseMethodParam

	// WMax must always be last to indicate which byte is the maximum value of a type byte
	WMax
)

func slice(typ byte, byteFields []uint8, args ...string) []byte {
	bufLen := 1 // hold type info
	bufLen += len(byteFields) * 2
	for _, a := range args {
		bufLen += stringLenBytes // string len
		bufLen += len(a)
	}
	res := make([]byte, 1, bufLen)
	res[0] = typ
	return res
}

const stringLenBytes = 4
const uint8fieldBytes = 2

func wrap(typ byte, byteFields []uint8, args ...string) string {
	var rawBuf [stringLenBytes / 2]byte
	var b [stringLenBytes]byte

	buf := slice(typ, byteFields, args...)
	for _, field := range byteFields {
		rawBuf[0] = field
		hex.Encode(b[:], rawBuf[:1])
		buf = append(buf, b[:uint8fieldBytes]...)
	}
	for _, s := range args {
		binary.LittleEndian.PutUint16(rawBuf[:], uint16(len(s)))
		hex.Encode(b[:], rawBuf[:])
		buf = append(buf, b[:]...)
		buf = append(buf, s...)
	}
	return string(buf)
}

func (t Type) unwrap1() (one string) {
	return t.String()[stringLenBytes+1:] // do not care about length, there is only 1 param
}

func (t Type) unwrap2() (one, two string) {
	var l int
	var b [stringLenBytes]byte
	var rawBuf [stringLenBytes / 2]byte

	pos := 1
	copy(b[:], t.String()[pos:pos+stringLenBytes])
	hex.Decode(rawBuf[:], b[:])
	l = int(binary.LittleEndian.Uint16(rawBuf[:]))
	pos += stringLenBytes
	one = t.String()[pos : pos+l]
	pos += l
	two = t.String()[pos+stringLenBytes:] // do not care about length of last param

	return one, two
}

func (t Type) unwrap3() (b1 uint8, one, two string) {
	var l int
	var b [stringLenBytes]byte
	var rawBuf [stringLenBytes / 2]byte

	tp := t.String()
	pos := 1
	copy(b[:], tp[pos:pos+uint8fieldBytes])
	hex.Decode(rawBuf[:], b[:uint8fieldBytes])
	b1 = rawBuf[0]
	pos += uint8fieldBytes
	copy(b[:], tp[pos:pos+stringLenBytes])
	hex.Decode(rawBuf[:], b[:])
	l = int(binary.LittleEndian.Uint16(rawBuf[:]))
	pos += stringLenBytes
	one = tp[pos : pos+l]
	pos += l
	two = tp[pos+stringLenBytes:] // do not care about length of last param

	return b1, one, two
}

func WrapBaseMethodParam(paramIndex int, className, methodName string) Type {
	return NewType(wrap(WBaseMethodParam, []uint8{uint8(paramIndex)}, className, methodName))
}

func (t Type) UnwrapBaseMethodParam() (paramIndex uint8, className, methodName string) {
	return t.unwrap3()
}

func WrapStaticMethodCall(className, methodName string) Type {
	return NewType(wrap(WStaticMethodCall, nil, className, methodName))
}

func (t Type) UnwrapStaticMethodCall() (className, methodName string) {
	return t.unwrap2()
}

func WrapInstanceMethodCall(typ Type, methodName string) Type {
	return NewType(wrap(WInstanceMethodCall, nil, typ.String(), methodName))
}

func (t Type) UnwrapInstanceMethodCall() (typ Type, methodName string) {
	rawType, methodName := t.unwrap2()
	typ = NewType(rawType)
	return typ, methodName
}

func WrapClassConstFetch(className, constName string) Type {
	return NewType(wrap(WClassConstFetch, nil, className, constName))
}

func (t Type) UnwrapClassConstFetch() (className, constName string) {
	return t.unwrap2()
}

func WrapStaticPropertyFetch(className, propName string) Type {
	if !strings.HasPrefix(propName, "$") {
		propName = "$" + propName
	}
	return NewType(wrap(WStaticPropertyFetch, nil, className, propName))
}

func (t Type) UnwrapStaticPropertyFetch() (className, propName string) {
	return t.unwrap2()
}

func WrapInstancePropertyFetch(typ, propName string) Type {
	return NewType(wrap(WInstancePropertyFetch, nil, typ, propName))
}

func (t Type) UnwrapInstancePropertyFetch() (typ Type, propName string) {
	rawType, propName := t.unwrap2()
	typ = NewType(rawType)
	return typ, propName
}

func WrapFunctionCall(funcName string) Type {
	return NewType(wrap(WFunctionCall, nil, funcName))
}

func (t Type) UnwrapFunctionCall() (funcName string) {
	return t.unwrap1()
}

func WrapArrayOf(typ Type) Type {
	return NewType(wrap(WArrayOf, nil, typ.String()))
}

func (t Type) UnwrapArrayOf() (typ Type) {
	return NewType(t.unwrap1())
}

func WrapElemOf(typ Type) Type {
	// ElemOf(ArrayOf(typ)) == typ
	if len(typ) >= 1+stringLenBytes && typ[0] == WArrayOf {
		return typ[1+stringLenBytes:]
	}

	return NewType(wrap(WElemOf, nil, typ.String()))
}

func (t Type) UnwrapElemOf() (typ Type) {
	return NewType(t.unwrap1())
}

func WrapElemOfKey(typ Type, key string) Type {
	if len(typ) >= 1+stringLenBytes && typ[0] == WArrayOf {
		return typ[1+stringLenBytes:]
	}
	return NewType(wrap(WElemOfKey, nil, typ.String(), key))
}

func (t Type) UnwrapElemOfKey() (typ Type, key string) {
	rawType, key := t.unwrap2()
	typ = NewType(rawType)
	return typ, key
}

func WrapGlobal(varName string) Type {
	return NewType(wrap(WGlobal, nil, varName))
}

func (t Type) UnwrapGlobal() (varName string) {
	return t.unwrap1()
}

func WrapConstant(constName string) Type {
	return NewType(wrap(WConstant, nil, constName))
}

func (t Type) UnwrapConstant() (constName string) {
	return t.unwrap1()
}

func (t Type) Format() (res string) {
	if t.IsEmpty() || !t.IsLazy() {
		return t.String()
	}

	defer func() {
		if r := recover(); r != nil {
			res = fmt.Sprintf("panic!(orig='%s', hex='%x')", t, t)
		}
	}()

	switch t[0] {
	case WGlobal:
		return "global_$" + NewType(t.UnwrapGlobal()).Format()
	case WConstant:
		return "constant(" + t.UnwrapConstant() + ")"
	case WArrayOf:
		return t.UnwrapArrayOf().Format() + "[]"
	case WElemOf:
		return "elem(" + t.UnwrapElemOf().Format() + ")"
	case WElemOfKey:
		typ, key := t.UnwrapElemOfKey()
		return fmt.Sprintf("elem(%s)[%s]", typ.Format(), key)
	case WFunctionCall:
		return t.UnwrapFunctionCall() + "()"
	case WInstanceMethodCall:
		expr, methodName := t.UnwrapInstanceMethodCall()
		return "(" + expr.Format() + ")->" + methodName + "()"
	case WInstancePropertyFetch:
		expr, propertyName := t.UnwrapInstancePropertyFetch()
		return "(" + expr.Format() + ")->" + propertyName
	case WBaseMethodParam:
		index, className, methodName := t.unwrap3()
		return fmt.Sprintf("param(%s)::%s[%d]", className, methodName, index)
	case WStaticMethodCall:
		className, methodName := t.UnwrapStaticMethodCall()
		return className + "::" + methodName + "()"
	case WStaticPropertyFetch:
		className, propertyName := t.UnwrapStaticPropertyFetch()
		return className + "::" + propertyName
	case WClassConstFetch:
		className, constName := t.UnwrapClassConstFetch()
		return className + "::" + constName
	}

	return "unknown(" + t.String() + ")"
}
