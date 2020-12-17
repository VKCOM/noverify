package meta

import (
	"strings"
)

type Type string

func NewType(typ string) Type {
	return Type(typ)
}

func NewTypeFromPhpDocType(phpDocType PhpDocType) Type {
	typ := Type(phpDocType.Elem)
	for i := 0; i < phpDocType.Dims; i++ {
		typ = WrapArrayOf(typ)
	}
	return typ
}

func (t Type) String() string {
	return string(t)
}

func (t Type) IsEmpty() bool {
	return t == ""
}

func (t Type) ElementType() Type {
	if !t.IsEmpty() && t[0] == WArrayOf {
		return t.UnwrapArrayOf()
	}

	return NewType(strings.TrimSuffix(string(t), "[]"))
}

func (t Type) Is(str string) bool {
	return t.String() == str
}

func (t Type) IsLazy() bool {
	if t.IsEmpty() {
		return false
	}

	return t[0] < WMax
}

func (t Type) IsMixed() bool {
	return t == "mixed"
}

func (t Type) IsArray() bool {
	if t.IsEmpty() {
		return false
	}

	return strings.HasSuffix(string(t), "[]") || t[0] == WArrayOf
}

func (t Type) IsShape() bool {
	return strings.HasPrefix(string(t), `\shape$`)
}

func (t Type) IsClass() bool {
	return strings.HasPrefix(string(t), `\`) && !t.IsShape() && !t.IsArray()
}
