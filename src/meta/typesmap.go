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

// Equals check if two typesmaps are the same
func (m *TypesMap) Equals(m2 *TypesMap) bool {
	return m.String() == m2.String()
}

// Len returns number of different types in map
func (m TypesMap) Len() int {
	return len(m.m)
}

// IsInt checks if map contains only int type
func (m *TypesMap) IsInt() bool {
	return m.Is("int")
}

// IsString checks if map contains only string type
func (m *TypesMap) IsString() bool {
	return m.Is("string")
}

// IsArray checks if map contains only array of any type
func (m TypesMap) IsArray() bool {
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

// IsArrayOf checks if map contains only array of given type
func (m *TypesMap) IsArrayOf(typ string) bool {
	if m == nil {
		return false
	}

	if len(m.m) != 1 {
		return false
	}

	_, ok := m.m[WrapArrayOf(typ)]
	return ok
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
func (m *TypesMap) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.immutable)
	if err != nil {
		return err
	}
	return decoder.Decode(&m.m)
}

// Find applies a predicate function to every contained type.
// If callback returns true for any of them, this is a result of Find call.
// False is returned if none of the contained types made pred function return true.
func (m TypesMap) Find(pred func(typ string) bool) bool {
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
