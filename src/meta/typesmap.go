package meta

import (
	"bytes"
	"encoding/gob"
	"sort"
	"strings"
)

// Preallocated and shared immutable type maps.
var (
	MixedType = NewTypesMap("mixed").Immutable()
	VoidType  = NewTypesMap("void").Immutable()

	PreciseIntType    = NewPreciseTypesMap("int").Immutable()
	PreciseFloatType  = NewPreciseTypesMap("float").Immutable()
	PreciseBoolType   = NewPreciseTypesMap("bool").Immutable()
	PreciseStringType = NewPreciseTypesMap("string").Immutable()
)

type Type struct {
	Elem string
	Dims int
}

type MapFlags uint8

const (
	mapImmutable MapFlags = 1 << iota
	mapPrecise
)

// TypesMap holds a set of types and can be made immutable to prevent unexpected changes.
type TypesMap struct {
	Flags MapFlags
	M     map[string]struct{}
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
func (m TypesMap) IsPrecise() bool { return m.Flags&mapPrecise != 0 }

func (m *TypesMap) MarkAsImprecise() {
	m.Flags &^= mapPrecise
}

func (m TypesMap) isImmutable() bool { return m.Flags&mapImmutable != 0 }

// IsResolved reports whether all types inside types map are resolved.
//
// Users should not depend on the "false" result meaning.
// If "true" is returned, TypesMap is guaranteed to be free of lazy types.
func (m TypesMap) IsResolved() bool {
	// TODO: could do a `s[0] >= meta.WMax` check over map keys
	// to check if it contains any lazy types.
	// It can be safe to start with maps of size 1.
	//
	// Looping over maps of arbitrary size can take more CPU time
	// than we would like to spend.
	// Note that most maps have a size that is less than 4, but
	// some of them can have 100+ elements.
	return m.IsPrecise()
}

// NewEmptyTypesMap creates new type map that has no types in it
func NewEmptyTypesMap(cap int) TypesMap {
	return TypesMap{M: make(map[string]struct{}, cap)}
}

func NewTypesMapFromTypes(types []Type) TypesMap {
	m := make(map[string]struct{}, len(types))
	for _, typ := range types {
		s := typ.Elem
		for i := 0; i < typ.Dims; i++ {
			s = WrapArrayOf(s)
		}
		m[s] = struct{}{}
	}
	return TypesMap{M: m}
}

// NewTypesMap returns new TypesMap that is initialized with the provided types (separated by "|" symbol)
func NewTypesMap(str string) TypesMap {
	m := make(map[string]struct{}, strings.Count(str, "|")+1)
	for _, s := range strings.Split(str, "|") {
		if IsArrayType(s) {
			s = WrapArrayOf(strings.TrimSuffix(s, "[]"))
		}
		m[s] = struct{}{}
	}
	return TypesMap{M: m}
}

func NewPreciseTypesMap(str string) TypesMap {
	m := NewTypesMap(str)
	m.Flags |= mapPrecise
	return m
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
	return TypesMap{M: m}
}

// Immutable returns immutable view of TypesMap
func (m TypesMap) Immutable() TypesMap {
	return TypesMap{
		Flags: m.Flags | mapImmutable,
		M:     m.M,
	}
}

// IsEmpty checks if map has no types at all
func (m TypesMap) IsEmpty() bool {
	return len(m.M) == 0
}

// Equals check if two typesmaps are the same
func (m TypesMap) Equals(m2 TypesMap) bool {
	if len(m.M) != len(m2.M) {
		return false
	}
	for k := range m.M {
		_, ok := m2.M[k]
		if !ok {
			return false
		}
	}
	return true
}

// Len returns number of different types in map
func (m TypesMap) Len() int {
	return len(m.M)
}

// IsArray checks if map contains only array of any type
//
// Warning: use only for *lazy* types!
func (m TypesMap) IsArray() bool {
	if len(m.M) != 1 {
		return false
	}

	for typ := range m.M {
		if len(typ) > 0 && typ[0] == WArrayOf {
			return true
		}
	}
	return false
}

// IsArrayOf checks if map contains only array of given type
//
// Warning: use only for *lazy* types!
func (m TypesMap) IsArrayOf(typ string) bool {
	if len(m.M) != 1 {
		return false
	}

	_, ok := m.M[WrapArrayOf(typ)]
	return ok
}

// Is reports whether m contains exactly one specified type.
//
// Warning: typ must be a proper *lazy* or *solved* type.
func (m TypesMap) Is(typ string) bool {
	if m.Len() != 1 {
		return false
	}

	_, ok := m.M[typ]
	return ok
}

// AppendString adds provided types to current map and returns new one (immutable maps are always copied)
func (m TypesMap) AppendString(str string) TypesMap {
	if !m.isImmutable() {
		if m.M == nil {
			m.M = make(map[string]struct{}, strings.Count(str, "|")+1)
		}

		for _, s := range strings.Split(str, "|") {
			m.M[s] = struct{}{}
		}

		// Since we have no idea where str is coming from,
		// mark map as imprecise.
		m.MarkAsImprecise()

		return m
	}

	mm := make(map[string]struct{}, m.Len()+strings.Count(str, "|")+1)
	for k, v := range m.M {
		mm[k] = v
	}

	for _, s := range strings.Split(str, "|") {
		mm[s] = struct{}{}
	}

	// The returned map is mutable and imprecise.
	return TypesMap{M: mm}
}

func (m TypesMap) Clone() TypesMap {
	if m.Len() == 0 || m.isImmutable() {
		return m
	}

	mm := make(map[string]struct{}, m.Len())
	for typ := range m.M {
		mm[typ] = struct{}{}
	}
	return TypesMap{M: mm, Flags: m.Flags}
}

// Append adds provided types to current map and returns new one (immutable maps are always copied)
func (m TypesMap) Append(n TypesMap) TypesMap {
	if m.Len() == 0 {
		return n
	}
	if n.Len() == 0 {
		return m
	}

	if !m.isImmutable() {
		if m.M == nil {
			if n.M == nil {
				return m
			}
			m.M = make(map[string]struct{}, n.Len())
		}

		m.MarkAsImprecise()
		for k, v := range n.M {
			m.M[k] = v
		}
		return m
	}

	mm := make(map[string]struct{}, m.Len()+n.Len())
	for k, v := range m.M {
		mm[k] = v
	}
	for k, v := range n.M {
		mm[k] = v
	}

	// Previously, returned map was always mutable, so we ignore mapImmutable flag.
	// If both maps are precise, we preserve that property.
	var flags MapFlags
	if m.IsPrecise() && n.IsPrecise() {
		flags |= mapPrecise
	}

	return TypesMap{M: mm, Flags: flags}
}

// String returns string representation of a map
func (m TypesMap) String() string {
	if len(m.M) == 1 {
		for k := range m.M {
			return k
		}
	}

	types := make([]string, 0, len(m.M))
	for k := range m.M {
		types = append(types, formatType(k))
	}
	sort.Strings(types)
	return strings.Join(types, "|")
}

// GobEncode is a custom gob marshaller
func (m TypesMap) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(m.Flags)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(m.M)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode is custom gob unmarshaller
func (m *TypesMap) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.Flags)
	if err != nil {
		return err
	}
	return decoder.Decode(&m.M)
}

func (m TypesMap) Contains(typ string) bool {
	if m.Len() == 0 {
		return false
	}
	for typ2 := range m.M {
		if typ2 == typ {
			return true
		}
	}
	return false
}

// Find applies a predicate function to every contained type.
// If callback returns true for any of them, this is a result of Find call.
// False is returned if none of the contained types made pred function return true.
func (m TypesMap) Find(pred func(typ string) bool) bool {
	if m.Len() == 0 {
		return false
	}

	keys := make([]string, 0, len(m.M))
	for k := range m.M {
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
func (m TypesMap) Iterate(cb func(typ string)) {
	if m.Len() == 0 {
		return
	}

	// We need to sort types so that we always iterate classes using the same order.
	keys := make([]string, 0, len(m.M))
	for k := range m.M {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		cb(k)
	}
}

// Returns type of array element. T[] -> T, T[][] -> T[].
// For *Lazy* type.
func (m TypesMap) ArrayElemLazyType() TypesMap {
	if m.Len() == 0 {
		return MixedType
	}

	mm := make(map[string]struct{}, m.Len())
	for typ := range m.M {
		mm[UnwrapArrayOf(typ)] = struct{}{}
	}
	return TypesMap{M: mm, Flags: m.Flags}
}
