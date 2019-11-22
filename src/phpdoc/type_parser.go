package phpdoc

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// TODO: use this parser inside linter as well after it's polished?
//
// TODO: consider using the grammar from https://github.com/phpstan/phpdoc-parser
//       if we ever want to understand array literal types, generics, callable with params.

// TypeParser handles phpdoc type expressions parsing.
//
// See https://github.com/php-fig/fig-standards/blob/master/proposed/phpdoc.md#abnf
type TypeParser struct {
	s string
}

// ParseType parses a phpdoc type out of a sting s.
func (p *TypeParser) ParseType(s string) (result TypeExpr, err error) {
	defer func() {
		r := recover()
		if err2, ok := r.(error); ok {
			err = err2
			return
		}
		if r != nil {
			panic(r)
		}
	}()

	p.s = strings.TrimSpace(s)
	return p.parseType(), nil
}

func (p *TypeParser) parseType() TypeExpr {
	if len(p.s) == 0 {
		panic(errors.New("unexpected end of input, expected type expr"))
	}

	left := p.parseOperand()
	for p.tryConsume("[") {
		if p.tryConsume("]") {
			left = &ArrayType{Elem: left}
		} else {
			panic(errors.New("missing closing `]`"))
		}
	}
	switch {
	case p.tryConsume("&"):
		return &InterType{X: left, Y: p.parseType()}
	case p.tryConsume("|"):
		return &UnionType{X: left, Y: p.parseType()}
	}
	return left
}

func (p *TypeParser) parseOperand() TypeExpr {
	switch {
	case isClassNameChar(p.s[0], true):
		i := 1
		for i < len(p.s) && isClassNameChar(p.s[i], false) {
			i++
		}
		typ := &NamedType{Name: p.s[:i]}
		p.consume(i)
		return typ

	case p.tryConsume("("):
		typ := p.parseType()
		if p.tryConsume(")") {
			return typ
		}
		panic(errors.New("missing closing `)`"))

	case p.tryConsume("!"):
		return &NotType{Expr: p.parseOperand()}

	case p.tryConsume("?"):
		return &NullableType{Expr: p.parseOperand()}

	case p.tryConsume("$this"): // The only permitted $-prefixed type
		return &NamedType{Name: "$this"}

	default:
		if len(p.s) == 0 {
			panic(errors.New("unexpected end of input, expected type operand"))
		}
		panic(fmt.Errorf("unexpected `%c` in type operand", p.s[0]))
	}
}

func (p *TypeParser) skipSpace() {
	p.s = strings.TrimLeftFunc(p.s, unicode.IsSpace)
}

func (p *TypeParser) tryConsume(s string) bool {
	p.skipSpace()
	if strings.HasPrefix(p.s, s) {
		p.consume(len(s))
		return true
	}
	return false
}

func (p *TypeParser) consume(n int) {
	p.s = p.s[n:]
}

func isClassNameChar(ch byte, first bool) bool {
	// ^[a-zA-Z_\x80-\xff][a-zA-Z0-9_\x80-\xff]*$
	switch {
	case ch == '\\':
		return true
	case ch == '_':
		return true
	case ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z':
		return true
	case ch >= '0' && ch <= '9':
		return !first
	case ch >= 0x80 && ch <= 0xff:
		return true
	default:
		return false
	}
}
