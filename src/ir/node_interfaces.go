package ir

import (
	"github.com/VKCOM/noverify/src/phpdoc"
)

type NameOwner interface {
	Node
	Name() *Identifier
}

type DocOwner interface {
	Node
	DocComment() phpdoc.Comment
}

type AttributeOwner interface {
	Node
	GetAttributes() []*AttributeGroup
}

type ParamListOwner interface {
	Node
	ParamList() []Node
}

type ModifierListOwner interface {
	Node
	ModifierList() []*Identifier
}

type TypeHintOwner interface {
	Node
	TypeHint() Node
}

type Binary interface {
	Node
	Lhs() Node
	Rhs() Node
}
