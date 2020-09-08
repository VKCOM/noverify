package phpdoc

type ExprShape uint8

type ExprKind uint8

type Type struct {
	Source string
	Expr   TypeExpr
}

func (typ Type) Clone() Type {
	return Type{Source: typ.Source, Expr: typ.Expr.Clone()}
}

func (typ Type) String() string { return typ.Source }

func (typ Type) IsEmpty() bool { return typ.Expr.Value == "" }

type TypeExpr struct {
	Kind  ExprKind
	Shape ExprShape
	Begin uint16
	End   uint16
	Value string
	Args  []TypeExpr
}

func (e TypeExpr) Clone() TypeExpr {
	cloned := e
	cloned.Args = make([]TypeExpr, len(e.Args))
	for i, a := range e.Args {
		cloned.Args[i] = a.Clone()
	}
	return cloned
}

//go:generate stringer -type=ExprKind -trimprefix=Expr
const (
	// ExprInvalid represents "failed to parse" type expression.
	ExprInvalid ExprKind = iota

	// ExprUnknown is a garbage-prefixed type expression.
	// Examples: `-int` `@@\Foo[]`
	// Args[0] - a valid expression that follows invalid prefix
	ExprUnknown

	// ExprName is a type that is identified by its name.
	// Examples: `int` `\Foo\Bar` `$this`
	ExprName

	// ExprKeyword is a special name-like type node.
	// Examples: `*` `...`
	ExprSpecialName

	// ExprInt is a digit-only type expression.
	// Examples: `0` `10`
	ExprInt

	// ExprKeyVal is `key:val` type.
	// Examples: `name: string` `id:int`
	// Args[0] - key expression (left)
	// Args[1] - val expression (right)
	ExprKeyVal

	// ExprMemberType is access to member.
	// Examples: `\Foo::SOME_CONSTANT` `\Foo::$a`
	// Args[0] - class type expression (left)
	// Args[1] - member name expression (right)
	ExprMemberType

	// ExprArray is `elem[]` or `[]elem` type.
	// Examples: `int[]` `(int|float)[]` `int[`
	// Args[0] - array element type
	// ShapeArrayPrefix: `[]T`
	//
	// Note: may miss second `]`.
	ExprArray

	// ExprParen is `(expr)` type.
	// Examples: `(int)` `(\Foo\Bar[])` `(int`
	// Args[0] - enclosed type
	//
	// Note: may miss closing `)`.
	ExprParen

	// ExprNullable is `?expr` type.
	// Examples: `?int` `?\Foo`
	// Args[0] - nullable element type
	ExprNullable

	// ExprOptional is `expr?` type.
	// Examples: `k?: int`
	// Args[0] - optional element type
	ExprOptional

	// ExprNot is `!expr` type.
	// Examples: `!int` `!(int|float)`
	// Args[0] - negated element type
	//
	// Note: only valid for phpgrep type filters.
	ExprNot

	// ExprUnion is `x|y` type.
	// Examples: `T1|T2` `int|float[]|false`
	// Args - type variants
	ExprUnion

	// ExprInter is `x&y` type.
	// Examples: `T1&T2` `I1&I2&I3`
	ExprInter

	// ExprGeneric is a parametrized `expr<T,...>` type.
	// Examples: `\Collection<T>` `Either<int[], false>` `Bad<int`
	// Args[0] - parametrized type
	// Args[1:] - type parameters
	// ShapeGenericParen: `T(X,Y)`
	// ShapeGenericBrace: `T{X,Y}`
	//
	// Note: may miss closing `>`.
	ExprGeneric

	// ExprTypedCallable is a parametrized `callable(A,...):B` type.
	// Examples: `callable():void` `callable(int, int) : float`
	// Args[0] - return type
	// Args[1:] - argument types
	ExprTypedCallable
)

const (
	ShapeDefault ExprShape = iota
	ShapeArrayPrefix
	ShapeGenericParen
	ShapeGenericBrace
)

var prefixPrecedenceTab = [256]byte{
	'?': 5,
	'[': 5,
	'!': 5,
}

var infixPrecedenceTab = [256]byte{
	':': 1,
	'|': 2,
	';': 3, // is ::
	'&': 4,
	'[': 5,
	'<': 6,
	'(': 6,
	'{': 6,
	'?': 5,
}

type TypeParser struct {
	input       string
	pos         uint
	skipUnknown bool
	insideGroup bool

	exprPool  []TypeExpr
	allocated uint
}

func NewTypeParser() *TypeParser {
	return &TypeParser{
		exprPool: make([]TypeExpr, 32),
	}
}

func (p *TypeParser) Parse(s string) Type {
	p.reset(s)
	p.skipWhitespace()
	typ := Type{Source: s, Expr: *p.parseExpr(0)}
	p.setValues(&typ.Expr)
	return typ
}

func (p *TypeParser) reset(input string) {
	p.input = input
	p.pos = 0
	p.allocated = 0
	p.skipUnknown = false
}

func (p *TypeParser) exprValue(e *TypeExpr) string {
	return p.input[e.Begin:e.End]
}

func (p *TypeParser) setValues(e *TypeExpr) {
	for i := range e.Args {
		p.setValues(&e.Args[i])
	}
	e.Value = p.exprValue(e)
}

func (p *TypeParser) parseExprInsideGroup() *TypeExpr {
	insideGroup := p.insideGroup
	p.insideGroup = true
	expr := p.parseExpr(0)
	p.insideGroup = insideGroup
	return expr
}

func (p *TypeParser) parseExpr(precedence byte) *TypeExpr {
	if p.insideGroup {
		p.skipWhitespace()
	}

	var left *TypeExpr
	begin := uint16(p.pos)
	ch := p.nextByte()

	switch {
	case ch == '$' || ch == '\\' || p.isFirstIdentChar(ch):
		for p.isNameChar(p.peek()) {
			p.pos++
		}
		left = p.newExpr(ExprName, begin, uint16(p.pos))
	case p.isDigit(ch):
		for p.isDigit(p.peek()) {
			p.pos++
		}
		left = p.newExpr(ExprInt, begin, uint16(p.pos))
	case ch == '[':
		if p.peek() == ']' {
			p.pos++
		}
		elem := p.parseExpr(prefixPrecedenceTab['['])
		left = p.newExprShape(ExprArray, ShapeArrayPrefix, begin, uint16(p.pos), elem)
	case ch == '(':
		if p.peek() == ')' {
			p.pos++
			expr := p.newExpr(ExprInvalid, begin+1, begin+1)
			left = p.newExpr(ExprParen, begin, uint16(p.pos), expr)
			break
		}
		expr := p.parseExprInsideGroup()
		if p.peek() == ')' {
			p.pos++
		}
		left = p.newExpr(ExprParen, begin, uint16(p.pos), expr)
	case ch == '?':
		elem := p.parseExpr(prefixPrecedenceTab['?'])
		left = p.newExpr(ExprNullable, begin, uint16(p.pos), elem)
	case ch == '!':
		elem := p.parseExpr(prefixPrecedenceTab['!'])
		left = p.newExpr(ExprNot, begin, uint16(p.pos), elem)
	case ch == '*':
		left = p.newExpr(ExprSpecialName, begin, uint16(p.pos))
	case ch == '.' && p.peekAt(p.pos+0) == '.' && p.peekAt(p.pos+1) == '.':
		p.pos += uint(len(".."))
		left = p.newExpr(ExprSpecialName, begin, uint16(p.pos))
	case ch == ' ':
		left = p.newExpr(ExprInvalid, begin, uint16(p.pos))
	default:
		// Try to handle invalid expressions somehow and continue
		// the parsing of valid expressions.
		if p.skipUnknown {
			return nil
		}
		p.skipUnknown = true
		for p.peek() != 0 {
			// Stop if we found infix or postfix token and emit invalid expr.
			// Stop if we found something that looks like a terminating token.
			ch := p.peek()
			if infixPrecedenceTab[ch] != 0 || ch == ')' || ch == '>' || ch == ']' || ch == ' ' {
				left = p.newExpr(ExprInvalid, begin, uint16(p.pos))
				break
			}
			pos := p.pos
			// Stop if we found a valid expression.
			x := p.parseExpr(0)
			if x != nil {
				left = p.newExpr(ExprUnknown, begin, uint16(p.pos), x)
				break
			}
			// Try again from the next byte pos.
			p.pos = pos + 1
		}
		p.skipUnknown = false
		// Found nothing, emit invalid expr.
		if left == nil {
			left = p.newExpr(ExprInvalid, begin, uint16(p.pos))
		}
	}

	if p.insideGroup {
		p.skipWhitespace()
	}

	calcPrecedence := func() byte {
		prc := infixPrecedenceTab[p.peek()]
		if p.peek() == ':' {
			ch := p.peekAt(p.pos + 1)
			if ch == ':' {
				prc = 3
			}
		}
		return prc
	}

	for precedence < calcPrecedence() {
		ch := p.nextByte()
		switch ch {
		case '?':
			left = p.newExpr(ExprOptional, begin, uint16(p.pos), left)
		case ':':
			isMemberType := p.peek() == ':'
			if isMemberType {
				_ = p.nextByte()
				right := p.parseExpr(infixPrecedenceTab[';'])
				left = p.newExpr(ExprMemberType, begin, uint16(p.pos), left, right)
			} else {
				right := p.parseExpr(infixPrecedenceTab[':'])
				left = p.newExpr(ExprKeyVal, begin, uint16(p.pos), left, right)
			}
		case '[':
			if p.peek() == ']' {
				p.pos++
			}
			left = p.newExpr(ExprArray, begin, uint16(p.pos), left)
		case '|':
			var right *TypeExpr
			switch p.peek() {
			case 0, ')':
				right = p.newExpr(ExprInvalid, uint16(p.pos), uint16(p.pos))
			default:
				right = p.parseExpr(infixPrecedenceTab['|'])
			}
			if left.Kind == ExprUnion {
				left.Args = append(left.Args, *right)
				left.End = right.End
			} else {
				left = p.newExpr(ExprUnion, begin, right.End, left, right)
			}
		case '&':
			var right *TypeExpr
			switch p.peek() {
			case 0, ')':
				right = p.newExpr(ExprInvalid, uint16(p.pos), uint16(p.pos))
			default:
				right = p.parseExpr(infixPrecedenceTab['&'])
			}
			if left.Kind == ExprInter {
				left.Args = append(left.Args, *right)
				left.End = right.End
			} else {
				left = p.newExpr(ExprInter, begin, right.End, left, right)
			}
		case '<', '(', '{':
			endCh := byte('>')
			shape := ShapeDefault
			switch ch {
			case '(':
				endCh = ')'
				shape = ShapeGenericParen
			case '{':
				endCh = '}'
				shape = ShapeGenericBrace
			}
			left = p.newExprShape(ExprGeneric, shape, begin, left.End, left)
			for {
				p.skipWhitespace()
				ch := p.peek()
				if ch == 0 {
					break
				}
				if ch == endCh {
					p.pos++
					break
				}
				x := p.parseExprInsideGroup()
				left.Args = append(left.Args, *x)
				p.skipWhitespace()
				if p.peek() == ',' {
					p.pos++
				}
			}
			// For `callable(...)` case we want to see whether we can peek ':'.
			// If we can, parse it as a typed callable.
			if shape == ShapeGenericParen && p.exprValue(&left.Args[0]) == "callable" {
				pos := p.pos
				p.skipWhitespace()
				if p.peek() == ':' {
					p.pos++
					returnType := p.parseExprInsideGroup()
					left.Args[0] = *returnType
					left.Kind = ExprTypedCallable
					left.Shape = ShapeDefault
				} else {
					p.pos = pos // Unread whitespace
				}
			}
			left.End = uint16(p.pos)
		}
	}

	return left
}

func (p *TypeParser) newExprShape(kind ExprKind, shape ExprShape, begin, end uint16, args ...*TypeExpr) *TypeExpr {
	e := p.newExpr(kind, begin, end, args...)
	e.Shape = shape
	return e
}

func (p *TypeParser) newExpr(kind ExprKind, begin, end uint16, args ...*TypeExpr) *TypeExpr {
	e := p.allocExpr()
	*e = TypeExpr{
		Kind:  kind,
		Begin: begin,
		End:   end,
		Args:  e.Args[:0],
	}
	for _, arg := range args {
		e.Args = append(e.Args, *arg)
	}
	return e
}

func (p *TypeParser) allocExpr() *TypeExpr {
	i := p.allocated
	if i < uint(len(p.exprPool)) {
		p.allocated++
		return &p.exprPool[i]
	}
	return &TypeExpr{}
}

func (p *TypeParser) isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func (p *TypeParser) isNameChar(ch byte) bool {
	// [\\a-zA-Z_\x7f-\xff0-9]
	return ch == '\\' || p.isFirstIdentChar(ch) || p.isDigit(ch)
}

func (p *TypeParser) isFirstIdentChar(ch byte) bool {
	// [a-zA-Z_\x7f-\xff]
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_' ||
		(ch >= 0x7f && ch <= 0xff)
}

func (p *TypeParser) nextByte() byte {
	if p.pos < uint(len(p.input)) {
		i := p.pos
		p.pos++
		return p.input[i]
	}
	return 0
}

func (p *TypeParser) peekAt(pos uint) byte {
	if pos < uint(len(p.input)) {
		return p.input[pos]
	}
	return 0
}

func (p *TypeParser) peek() byte {
	return p.peekAt(p.pos)
}

func (p *TypeParser) skipWhitespace() {
	for p.peek() == ' ' {
		p.pos++
	}
}
