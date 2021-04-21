package types

import (
	"bytes"
	"encoding/gob"
	"sort"
	"strings"
)

// Preallocated and shared immutable type maps.
var (
	MixedType = NewMap("mixed").Immutable()
	VoidType  = NewMap("void").Immutable()
	NullType  = NewMap("null").Immutable()

	PreciseIntType    = NewPreciseMap("int").Immutable()
	PreciseFloatType  = NewPreciseMap("float").Immutable()
	PreciseBoolType   = NewPreciseMap("bool").Immutable()
	PreciseStringType = NewPreciseMap("string").Immutable()
)

type Type struct {
	Elem string
	Dims int
}

type mapFlags uint8

const (
	mapImmutable mapFlags = 1 << iota
	mapPrecise
)

// Map holds a set of types and can be made immutable to prevent unexpected changes.
type Map struct {
	flags mapFlags
	m     map[string]struct{}
}

// IsPrecise reports whether the type set represented by the map is precise
// enough to perform typecheck-like analysis.
//
// Type precision determined by a type information source.
// For example, Int literal has a precise type of `int`, while having
// a phpdoc that promises some variable to have type `T` is not precise enough.
//
// Adding an imprecise type to a types map makes the entire type map imprecise.
//
// Important invariant: a precise map contains no lazy types.
func (m Map) IsPrecise() bool { return m.flags&mapPrecise != 0 }

func (m *Map) MarkAsImprecise() {
	m.flags &^= mapPrecise
}

func (m Map) isImmutable() bool { return m.flags&mapImmutable != 0 }

// IsResolved reports whether all types inside types map are resolved.
//
// Users should not depend on the "false" result meaning.
// If "true" is returned, Map is guaranteed to be free of lazy types.
func (m Map) IsResolved() bool {
	// TODO: could do a `s[0] >= types.WMax` check over map keys
	// to check if it contains any lazy types.
	// It can be safe to start with maps of size 1.
	//
	// Looping over maps of arbitrary size can take more CPU time
	// than we would like to spend.
	// Note that most maps have a size that is less than 4, but
	// some of them can have 100+ elements.
	return m.IsPrecise()
}

// Map returns a new types map with the results of calling fn for every type contained inside m.
// The result type map is never marked as precise.
func (m Map) Map(fn func(string) string) Map {
	mapped := make(map[string]struct{}, len(m.m))
	for typ := range m.m {
		mapped[fn(typ)] = struct{}{}
	}
	return NewMapFromMap(mapped)
}

// Filter returns a new types map with the types of m for which fn returns true.
// The result type map is never marked as precise.
func (m Map) Filter(fn func(string) bool) Map {
	filtered := make(map[string]struct{}, len(m.m))
	for typ := range m.m {
		if fn(typ) {
			filtered[typ] = struct{}{}
		}
	}
	return NewMapFromMap(filtered)
}

// NewEmptyMap creates new type map that has no types in it
func NewEmptyMap(cap int) Map {
	return Map{m: make(map[string]struct{}, cap)}
}

func NewMapFromTypes(types []Type) Map {
	m := make(map[string]struct{}, len(types))
	for _, typ := range types {
		s := typ.Elem
		for i := 0; i < typ.Dims; i++ {
			s = WrapArrayOf(s)
		}
		m[s] = struct{}{}
	}
	return Map{m: m}
}

// NewMap returns new Map that is initialized with the provided types (separated by "|" symbol)
func NewMap(str string) Map {
	m := make(map[string]struct{}, strings.Count(str, "|")+1)
	for _, s := range strings.Split(str, "|") {
		if IsArray(s) {
			s = WrapArrayOf(strings.TrimSuffix(s, "[]"))
		}
		m[s] = struct{}{}
	}
	return Map{m: m}
}

func NewPreciseMap(str string) Map {
	m := NewMap(str)
	m.flags |= mapPrecise
	return m
}

// MergeMaps creates a new types map from union of specified type maps
func MergeMaps(maps ...Map) Map {
	totalLen := 0
	for _, m := range maps {
		totalLen += m.Len()
	}

	t := NewEmptyMap(totalLen)
	for _, m := range maps {
		t = t.Append(m)
	}

	return t
}

// NewMapFromMap creates Map from provided map[string]struct{}
func NewMapFromMap(m map[string]struct{}) Map {
	return Map{m: m}
}

// Immutable returns immutable view of Map
func (m Map) Immutable() Map {
	return Map{
		flags: m.flags | mapImmutable,
		m:     m.m,
	}
}

// Empty checks if map has no types at all
func (m Map) Empty() bool {
	return len(m.m) == 0
}

// Equals check if two typesmaps are the same
func (m Map) Equals(m2 Map) bool {
	if len(m.m) != len(m2.m) {
		return false
	}
	for k := range m.m {
		_, ok := m2.m[k]
		if !ok {
			return false
		}
	}
	return true
}

// Len returns number of different types in map
func (m Map) Len() int {
	return len(m.m)
}

// IsLazyArray checks if map contains only array of any type
func (m Map) IsLazyArray() bool {
	if len(m.m) != 1 {
		return false
	}

	for typ := range m.m {
		if len(typ) > 0 && typ[0] == WArrayOf {
			return true
		}
	}
	return false
}

// IsLazyArrayOf checks if map contains only array of given type
func (m Map) IsLazyArrayOf(typ string) bool {
	if len(m.m) != 1 {
		return false
	}

	_, ok := m.m[WrapArrayOf(typ)]
	return ok
}

// IsArray checks if map contains only array of any type
func (m Map) IsArray() bool {
	if len(m.m) != 1 {
		return false
	}

	for typ := range m.m {
		if IsArray(typ) {
			return true
		}
	}

	return false
}

// Is reports whether m contains exactly one specified type.
//
// Warning: typ must be a proper *lazy* or *solved* type.
func (m Map) Is(typ string) bool {
	if m.Len() != 1 {
		return false
	}

	_, ok := m.m[typ]
	return ok
}

func (m Map) Clone() Map {
	if m.Len() == 0 || m.isImmutable() {
		return m
	}

	mm := make(map[string]struct{}, m.Len())
	for typ := range m.m {
		mm[typ] = struct{}{}
	}
	return Map{m: mm, flags: m.flags}
}

// Append adds provided types to current map and returns new one (immutable maps are always copied)
func (m Map) Append(n Map) Map {
	if m.Len() == 0 {
		return n
	}
	if n.Len() == 0 {
		return m
	}

	if !m.isImmutable() {
		if m.m == nil {
			if n.m == nil {
				return m
			}
			m.m = make(map[string]struct{}, n.Len())
		}

		m.MarkAsImprecise()
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

	// Previously, returned map was always mutable, so we ignore mapImmutable flag.
	// If both maps are precise, we preserve that property.
	var flags mapFlags
	if m.IsPrecise() && n.IsPrecise() {
		flags |= mapPrecise
	}

	return Map{m: mm, flags: flags}
}

// String returns string representation of a map
func (m Map) String() string {
	if len(m.m) == 1 {
		for k := range m.m {
			return k
		}
	}

	types := make([]string, 0, len(m.m))
	for k := range m.m {
		types = append(types, FormatType(k))
	}
	sort.Strings(types)
	return strings.Join(types, "|")
}

// GobEncode is a custom gob marshaller
func (m Map) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(m.flags)
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
func (m *Map) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.flags)
	if err != nil {
		return err
	}
	return decoder.Decode(&m.m)
}

func (m Map) Contains(typ string) bool {
	if m.Len() == 0 {
		return false
	}
	for typ2 := range m.m {
		if typ2 == typ {
			return true
		}
	}
	return false
}

// Find applies a predicate function to every contained type.
// If callback returns true for any of them, this is a result of Find call.
// False is returned if none of the contained types made pred function return true.
func (m Map) Find(pred func(typ string) bool) bool {
	if m.Len() == 0 {
		return false
	}

	keys := make([]string, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, typ := range keys {
		if pred(typ) {
			return true
		}
	}

	return false
}

// Iterate applies cb to all contained types
func (m Map) Iterate(cb func(typ string)) {
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

// LazyArrayElemType returns type of array element. T[] -> T, T[][] -> T[].
// For *Lazy* type.
func (m Map) LazyArrayElemType() Map {
	if m.Len() == 0 {
		return MixedType
	}

	mm := make(map[string]struct{}, m.Len())
	for typ := range m.m {
		mm[UnwrapArrayOf(typ)] = struct{}{}
	}
	return Map{m: mm, flags: m.flags}
}
