package meta

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Preallocated and shared immutable type maps.
var (
	MixedType = NewTypesMap("mixed").Immutable()
	VoidType  = NewTypesMap("void").Immutable()
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

	// WGlobal means global variable.
	// E.g. global $Something;
	// Params: [Global variable name <string>]
	WGlobal

	// WConstant means constant
	// e.g. type of MINUTE constant
	// Params: [Constant name <string>]
	WConstant

	// WBaseMethodParam<0-N> is a way to inherit base type method type of nth parameter.
	// e.g. type of $x param of foo method from one of the implemented interfaces.
	// Params: [Index <uint8>] [Class name <string>] [Method name <string>]
	WBaseMethodParam

	// WMax must always be last to indicate which byte is the maximum value of a type byte
	WMax
)

// TypesMap holds a set of types and can be made immutable to prevent unexpected changes.
type TypesMap struct {
	immutable bool
	m         map[string]struct{}
}

// NewEmptyTypesMap creates new type map that has no types in it
func NewEmptyTypesMap(cap int) TypesMap {
	return TypesMap{m: make(map[string]struct{}, cap)}
}

// NewTypesMap returns new TypesMap that is initialized with the provided types (separated by "|" symbol)
func NewTypesMap(str string) TypesMap {
	m := make(map[string]struct{}, strings.Count(str, "|")+1)
	for _, s := range strings.Split(str, "|") {
		for strings.HasSuffix(s, "[]") {
			s = WrapArrayOf(strings.TrimSuffix(s, "[]"))
		}
		m[s] = struct{}{}
	}
	return TypesMap{m: m}
}

// MergeTypeMaps creates a new types map from union of specified type maps
func MergeTypeMaps(maps ...TypesMap) TypesMap {
	totalLen := 0
	for _, m := range maps {
		totalLen += m.Len()
	}

	t := NewEmptyTypesMap(totalLen)
	for _, m := range maps {
		t = t.Append(m)
	}

	return t
}

// NewTypesMapFromMap creates TypesMap from provided map[string]struct{}
func NewTypesMapFromMap(m map[string]struct{}) TypesMap {
	return TypesMap{m: m}
}

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

func unwrap1(s string) (one string) {
	return s[stringLenBytes+1:] // do not care about length, there is only 1 param
}

func unwrap2(s string) (one, two string) {
	var l int
	var b [stringLenBytes]byte
	var rawBuf [stringLenBytes / 2]byte

	pos := 1
	copy(b[:], s[pos:pos+stringLenBytes])
	hex.Decode(rawBuf[:], b[:])
	l = int(binary.LittleEndian.Uint16(rawBuf[:]))
	pos += stringLenBytes
	one = s[pos : pos+l]
	pos += l
	two = s[pos+stringLenBytes:] // do not care about length of last param

	return one, two
}

func unwrap3(s string) (b1 uint8, one, two string) {
	var l int
	var b [stringLenBytes]byte
	var rawBuf [stringLenBytes / 2]byte

	pos := 1
	copy(b[:], s[pos:pos+uint8fieldBytes])
	hex.Decode(rawBuf[:], b[:uint8fieldBytes])
	b1 = rawBuf[0]
	pos += uint8fieldBytes
	copy(b[:], s[pos:pos+stringLenBytes])
	hex.Decode(rawBuf[:], b[:])
	l = int(binary.LittleEndian.Uint16(rawBuf[:]))
	pos += stringLenBytes
	one = s[pos : pos+l]
	pos += l
	two = s[pos+stringLenBytes:] // do not care about length of last param

	return b1, one, two
}

func WrapBaseMethodParam(paramIndex int, className, methodName string) string {
	return wrap(WBaseMethodParam, []uint8{uint8(paramIndex)}, className, methodName)
}

func UnwrapBaseMethodParam(s string) (paramIndex uint8, className, methodName string) {
	return unwrap3(s)
}

func WrapStaticMethodCall(className, methodName string) string {
	return wrap(WStaticMethodCall, nil, className, methodName)
}

func UnwrapStaticMethodCall(s string) (className, methodName string) {
	return unwrap2(s)
}

func WrapInstanceMethodCall(typ, methodName string) string {
	return wrap(WInstanceMethodCall, nil, typ, methodName)
}

func UnwrapInstanceMethodCall(s string) (typ, methodName string) {
	return unwrap2(s)
}

func WrapStaticPropertyFetch(className, propName string) string {
	if !strings.HasPrefix(propName, "$") {
		propName = "$" + propName
	}
	return wrap(WStaticPropertyFetch, nil, className, propName)
}

func UnwrapStaticPropertyFetch(s string) (className, propName string) {
	return unwrap2(s)
}

func WrapInstancePropertyFetch(typ, propName string) string {
	return wrap(WInstancePropertyFetch, nil, typ, propName)
}

func UnwrapInstancePropertyFetch(s string) (typ, propName string) {
	return unwrap2(s)
}

func WrapFunctionCall(funcName string) string {
	return wrap(WFunctionCall, nil, funcName)
}

func UnwrapFunctionCall(s string) (funcName string) {
	return unwrap1(s)
}

func WrapArrayOf(typ string) string {
	return wrap(WArrayOf, nil, typ)
}

func UnwrapArrayOf(s string) (typ string) {
	return unwrap1(s)
}

func WrapElemOf(typ string) string {
	// ElemOf(ArrayOf(typ)) == typ
	if len(typ) >= 1+stringLenBytes && typ[0] == WArrayOf {
		return typ[1+stringLenBytes:]
	}

	return wrap(WElemOf, nil, typ)
}

func UnwrapElemOf(s string) (typ string) {
	return unwrap1(s)
}

func WrapGlobal(varName string) string {
	return wrap(WGlobal, nil, varName)
}

func UnwrapGlobal(s string) (varName string) {
	return unwrap1(s)
}

func WrapConstant(constName string) string {
	return wrap(WConstant, nil, constName)
}

func UnwrapConstant(s string) (constName string) {
	return unwrap1(s)
}

// Immutable returns immutable copy of TypesMap
func (m TypesMap) Immutable() TypesMap {
	return TypesMap{
		immutable: true,
		m:         m.m,
	}
}

// IsEmpty checks if map has no types at all
func (m TypesMap) IsEmpty() bool {
	return len(m.m) == 0
}

// Len returns number of different types in map
func (m TypesMap) Len() int {
	return len(m.m)
}

// IsInt checks if map contains only int type
func (m TypesMap) IsInt() bool {
	return m.Is("int") || m.Is("integer")
}

// IsString checks if map contains only string type
func (m TypesMap) IsString() bool {
	return m.Is("string")
}

// Is reports whether m contains exactly one specified type.
func (m TypesMap) Is(typ string) bool {
	if m.Len() != 1 {
		return false
	}

	_, ok := m.m[typ]
	return ok
}

// AppendString adds provided types to current map and returns new one (immutable maps are always copied)
func (m TypesMap) AppendString(str string) TypesMap {
	if !m.immutable {
		if m.m == nil {
			m.m = make(map[string]struct{}, strings.Count(str, "|")+1)
		}

		for _, s := range strings.Split(str, "|") {
			m.m[s] = struct{}{}
		}

		return m
	}

	mm := make(map[string]struct{}, m.Len()+strings.Count(str, "|")+1)
	for k, v := range m.m {
		mm[k] = v
	}

	for _, s := range strings.Split(str, "|") {
		mm[s] = struct{}{}
	}

	return TypesMap{m: mm}
}

func (m TypesMap) clone() TypesMap {
	if m.Len() == 0 || m.immutable {
		return m
	}

	mm := make(map[string]struct{}, m.Len())
	for typ := range m.m {
		mm[typ] = struct{}{}
	}
	return TypesMap{m: mm}
}

// Append adds provided types to current map and returns new one (immutable maps are always copied)
func (m TypesMap) Append(n TypesMap) TypesMap {
	if m.Len() == 0 {
		return n
	}
	if n.Len() == 0 {
		return m
	}

	if !m.immutable {
		if m.m == nil {
			if n.m == nil {
				return m
			}
			m.m = make(map[string]struct{}, n.Len())
		}

		for k, v := range n.m {
			m.m[k] = v
		}
		return m
	}

	mm := make(map[string]struct{}, m.Len()+n.Len())
	for k, v := range m.m {
		mm[k] = v
	}
	for k, v := range n.m {
		mm[k] = v
	}

	return TypesMap{m: mm}
}

func formatType(s string) (res string) {
	if len(s) == 0 || s[0] >= WMax {
		return s
	}

	defer func() {
		if r := recover(); r != nil {
			res = fmt.Sprintf("panic!(orig='%s', hex='%x')", s, s)
		}
	}()

	switch s[0] {
	case WGlobal:
		return "global_$" + formatType(UnwrapGlobal(s))
	case WConstant:
		return "constant(" + UnwrapConstant(s) + ")"
	case WArrayOf:
		return formatType(UnwrapArrayOf(s)) + "[]"
	case WElemOf:
		return "elem(" + formatType(UnwrapElemOf(s)) + ")"
	case WFunctionCall:
		return UnwrapFunctionCall(s) + "()"
	case WInstanceMethodCall:
		expr, methodName := UnwrapInstanceMethodCall(s)
		return "(" + formatType(expr) + ")->" + methodName + "()"
	case WInstancePropertyFetch:
		expr, propertyName := UnwrapInstancePropertyFetch(s)
		return "(" + formatType(expr) + ")->" + propertyName
	case WBaseMethodParam:
		index, className, methodName := unwrap3(s)
		return fmt.Sprintf("param(%s)::%s[%d]", className, methodName, index)
	case WStaticMethodCall:
		className, methodName := UnwrapStaticMethodCall(s)
		return className + "::" + methodName + "()"
	case WStaticPropertyFetch:
		className, propertyName := UnwrapStaticPropertyFetch(s)
		return className + "::" + propertyName
	}

	return "unknown(" + s + ")"
}

// String returns string representation of a map
func (m TypesMap) String() string {
	var types []string
	for k := range m.m {
		types = append(types, formatType(k))
	}
	sort.Strings(types)
	return strings.Join(types, "|")
}

// GobEncode is a custom gob marshaller
func (m TypesMap) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(m.immutable)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(m.m)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode is custom gob unmarshaller
func (m TypesMap) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.immutable)
	if err != nil {
		return err
	}
	return decoder.Decode(&m.m)
}

// Iterate applies cb to all contained types
func (m TypesMap) Iterate(cb func(typ string)) {
	if m.Len() == 0 {
		return
	}

	// We need to sort types so that we always iterate classes using the same order.
	keys := make([]string, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		cb(k)
	}
}
