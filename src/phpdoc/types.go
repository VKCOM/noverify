package phpdoc

// TypeExpr is an arbitrary type expression.
type TypeExpr interface {
	typeExpr()

	// String returns a textual representation of a type.
	String() string
}

// Types that implement TypeExpr.
type (
	// NamedType is a type that is identified by its name.
	NamedType struct{ Name string }

	// NotType is `!expr` type.
	NotType struct{ Expr TypeExpr }

	// NullableType is `?expr` type.
	NullableType struct{ Expr TypeExpr }

	// ArrayType is `elem[]` type.
	ArrayType struct{ Elem TypeExpr }

	// UnionType is `x|y` type.
	// Union type requires a type to "implement" either X or Y.
	UnionType struct{ X, Y TypeExpr }

	// InterType is `x&y` type.
	// Intersection type requires a type to "implement" both X and Y.
	InterType struct{ X, Y TypeExpr }
)

func (*NamedType) typeExpr()    {}
func (*NotType) typeExpr()      {}
func (*NullableType) typeExpr() {}
func (*ArrayType) typeExpr()    {}
func (*UnionType) typeExpr()    {}
func (*InterType) typeExpr()    {}

func (typ *NamedType) String() string    { return typ.Name }
func (typ *NotType) String() string      { return "!" + typ.Expr.String() }
func (typ *NullableType) String() string { return "?" + typ.Expr.String() }
func (typ *ArrayType) String() string    { return typ.Elem.String() + "[]" }
func (typ *UnionType) String() string    { return "(" + typ.X.String() + "|" + typ.Y.String() + ")" }
func (typ *InterType) String() string    { return "(" + typ.X.String() + "&" + typ.Y.String() + ")" }
