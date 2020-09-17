package meta

import (
	"strings"
)

// LowercaseString type is used to avoid invalid mixing of normal strings
// and ones that are guaranteed to be lowercase.
type LowercaseString string

type ClassesMap struct {
	H map[LowercaseString]ClassInfo
}

func NewClassesMap() ClassesMap {
	return ClassesMap{H: make(map[LowercaseString]ClassInfo)}
}

func (m ClassesMap) Len() int           { return len(m.H) }
func (m ClassesMap) Delete(name string) { delete(m.H, toLower(name)) }

func (m ClassesMap) Get(name string) (ClassInfo, bool) {
	res, ok := m.H[toLower(name)]
	return res, ok
}

func (m ClassesMap) Set(name string, class ClassInfo) {
	m.H[toLower(name)] = class
}

type FunctionsMap struct {
	H map[LowercaseString]FuncInfo
}

func NewFunctionsMap() FunctionsMap {
	return FunctionsMap{H: make(map[LowercaseString]FuncInfo)}
}

func (m FunctionsMap) Len() int           { return len(m.H) }
func (m FunctionsMap) Delete(name string) { delete(m.H, toLower(name)) }

func (m FunctionsMap) Get(name string) (FuncInfo, bool) {
	res, ok := m.H[toLower(name)]
	return res, ok
}

func (m FunctionsMap) Set(name string, fn FuncInfo) {
	m.H[toLower(name)] = fn
}

// toLower is like strings.ToLower, but specialized for ascii-only.
// It also returns LowercaseString type.
func toLower(s string) LowercaseString {
	hasUpper := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
			break
		}
	}

	if !hasUpper {
		return LowercaseString(s)
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b.WriteByte(c)
	}
	return LowercaseString(b.String())
}
