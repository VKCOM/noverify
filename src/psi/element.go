package psi

type Element interface {
	Walk(Visitor)
}

type Visitor interface {
	EnterNode(Element) bool
	LeaveNode(Element)
}

type Stub interface {
}

type StubbedElement interface {
	Element
	Stub() Stub
}
