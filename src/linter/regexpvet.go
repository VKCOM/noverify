package linter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/quasilyte/regex/syntax"
)

type regexpVet struct {
	parser *syntax.Parser

	issues    []string
	exprFlags []string
}

func (c *regexpVet) CheckRegexp(pat regexpPattern) ([]string, error) {
	c.issues = c.issues[:0]
	c.exprFlags = c.exprFlags[:0]

	re, err := c.parser.ParsePCRE(pat.value)
	if err != nil {
		if strings.Contains(err.Error(), "not supported") {
			// Can't analyze, can't report an error.
			return nil, nil
		}
		return nil, err
	}
	c.walk(re.Expr)

	return c.issues, nil
}

func (c *regexpVet) walk(e syntax.Expr) {
	switch e.Op {
	case syntax.OpAlt:
		c.checkAltAnchor(e)
		c.checkAltDups(e)
		for _, a := range e.Args {
			c.walk(a)
		}

	case syntax.OpCharClass, syntax.OpNegCharClass:
		if c.checkCharClassRanges(e) {
			c.checkCharClassDups(e)
		}

	case syntax.OpStar, syntax.OpPlus:
		c.checkNestedQuantifier(e)
		c.walk(e.Args[0])

	case syntax.OpFlagOnlyGroup:
		c.checkFlags(e, e.Args[0].Value)
		c.exprFlags = append(c.exprFlags, e.Args[0].Value)
	case syntax.OpGroupWithFlags:
		nflags := len(c.exprFlags)
		c.checkFlags(e, e.Args[1].Value)
		c.exprFlags = append(c.exprFlags, e.Args[1].Value)
		c.walk(e.Args[0])
		c.exprFlags = c.exprFlags[:nflags]
	case syntax.OpGroup, syntax.OpCapture, syntax.OpNamedCapture:
		nflags := len(c.exprFlags)
		c.walk(e.Args[0])
		c.exprFlags = c.exprFlags[:nflags]

	default:
		for _, a := range e.Args {
			c.walk(a)
		}
	}
}

func (c *regexpVet) checkFlags(e syntax.Expr, flags string) {
	for _, fset := range c.exprFlags {
		if i := strings.IndexAny(flags, fset); i != -1 {
			c.warn("redundant flag %c in %s", flags[i], e.Value)
		}
	}
}

func (c *regexpVet) checkNestedQuantifier(e syntax.Expr) {
	x := e.Args[0]
	switch x.Op {
	case syntax.OpGroup, syntax.OpCapture, syntax.OpGroupWithFlags:
		if len(e.Args) == 1 {
			x = x.Args[0]
		}
	}

	switch x.Op {
	case syntax.OpPlus, syntax.OpStar:
		c.warn("repeated greedy quantifier in %s", e.Value)
	}
}

func (c *regexpVet) checkAltDups(alt syntax.Expr) {
	// Seek duplicated alternation expressions.

	set := make(map[string]struct{}, len(alt.Args))
	for _, a := range alt.Args {
		if _, ok := set[a.Value]; ok {
			c.warn("'%s' is duplicated in %s", a.Value, alt.Value)
		}
		set[a.Value] = struct{}{}
	}
}

func (c *regexpVet) checkAltAnchor(alt syntax.Expr) {
	// Seek suspicious anchors.

	// Case 1: an alternation of literals where 1st expr begins with ^ anchor.
	first := alt.Args[0]
	if first.Op == syntax.OpConcat && len(first.Args) > 0 && first.Args[0].Op == syntax.OpCaret {
		matched := true
		for _, a := range alt.Args[1:] {
			if a.Op != syntax.OpLiteral && a.Op != syntax.OpChar {
				matched = false
				break
			}
		}
		if matched {
			c.warn("^ applied only to '%s' in %s", first.Value[len(`^`):], alt.Value)
		}
	}

	// Case 2: an alternation of literals where last expr ends with $ anchor.
	last := alt.Args[len(alt.Args)-1]
	if last.Op == syntax.OpConcat && len(last.Args) > 0 && last.LastArg().Op == syntax.OpDollar {
		matched := true
		for _, a := range alt.Args[:len(alt.Args)-1] {
			if a.Op != syntax.OpLiteral && a.Op != syntax.OpChar {
				matched = false
				break
			}
		}
		if matched {
			c.warn("$ applied only to '%s' in %s", last.Value[:len(last.Value)-len(`$`)], alt.Value)
		}
	}
}

func (c *regexpVet) checkCharClassRanges(cc syntax.Expr) bool {
	// Seek for suspicious ranges like `!-_`.
	//
	// We permit numerical ranges (0-9, hex and octal literals)
	// and simple ascii letter ranges.

	for _, e := range cc.Args {
		if e.Op != syntax.OpCharRange {
			continue
		}
		if e.Args[0].Op == syntax.OpEscapeOctal || e.Args[0].Op == syntax.OpEscapeHex {
			continue
		}
		ch := c.charClassBoundRune(e.Args[0])
		if ch == 0 {
			return false
		}
		isLetter := unicode.IsLetter(ch)
		okay := isLetter || (ch >= '0' && ch <= '9')
		if !okay {
			c.warnSloppyCharRange(e.Value, cc.Value)
		}
	}

	return true
}

func (c *regexpVet) checkCharClassDups(cc syntax.Expr) {
	// Seek for excessive elements inside a character class.
	// Report them as intersections.

	if len(cc.Args) == 1 {
		return // Can't had duplicates.
	}

	type charRange struct {
		low    rune
		high   rune
		source string
	}
	ranges := make([]charRange, 0, 8)
	addRange := func(source string, low, high rune) {
		ranges = append(ranges, charRange{source: source, low: low, high: high})
	}
	addRange1 := func(source string, ch rune) {
		addRange(source, ch, ch)
	}

	// 1. Collect ranges, O(n).
	for _, e := range cc.Args {
		switch e.Op {
		case syntax.OpEscapeOctal:
			addRange1(e.Value, c.octalToRune(e))
		case syntax.OpEscapeHex:
			addRange1(e.Value, c.hexToRune(e))
		case syntax.OpChar:
			addRange1(e.Value, c.stringToRune(e.Value))
		case syntax.OpCharRange:
			addRange(e.Value, c.charClassBoundRune(e.Args[0]), c.charClassBoundRune(e.Args[1]))
		case syntax.OpEscapeMeta:
			addRange1(e.Value, rune(e.Value[1]))
		case syntax.OpEscapeChar:
			ch := c.stringToRune(e.Value[len(`\`):])
			if unicode.IsPunct(ch) {
				addRange1(e.Value, ch)
				break
			}
			switch e.Value {
			case `\|`, `\<`, `\>`, `\+`, `\=`: // How to cover all symbols?
				addRange1(e.Value, c.stringToRune(e.Value[len(`\`):]))
			case `\t`:
				addRange1(e.Value, '\t')
			case `\n`:
				addRange1(e.Value, '\n')
			case `\r`:
				addRange1(e.Value, '\r')
			case `\v`:
				addRange1(e.Value, '\v')
			case `\d`:
				addRange(e.Value, '0', '9')
			case `\D`:
				addRange(e.Value, 0, '0'-1)
				addRange(e.Value, '9'+1, utf8.MaxRune)
			case `\s`:
				addRange(e.Value, '\t', '\n') // 9-10
				addRange(e.Value, '\f', '\r') // 12-13
				addRange1(e.Value, ' ')       // 32
			case `\S`:
				addRange(e.Value, 0, '\t'-1)
				addRange(e.Value, '\n'+1, '\f'-1)
				addRange(e.Value, '\r'+1, ' '-1)
				addRange(e.Value, ' '+1, utf8.MaxRune)
			case `\w`:
				addRange(e.Value, '0', '9') // 48-57
				addRange(e.Value, 'A', 'Z') // 65-90
				addRange1(e.Value, '_')     // 95
				addRange(e.Value, 'a', 'z') // 97-122
			case `\W`:
				addRange(e.Value, 0, '0'-1)
				addRange(e.Value, '9'+1, 'A'-1)
				addRange(e.Value, 'Z'+1, '_'-1)
				addRange(e.Value, '_'+1, 'a'-1)
				addRange(e.Value, 'z'+1, utf8.MaxRune)
			default:
				// Give up: unknown escape sequence.
				return
			}
		default:
			// Give up: unexpected operation inside char class.
			return
		}
	}

	// 2. Sort ranges, O(nlogn).
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].low < ranges[j].low
	})

	// 3. Search for duplicates, O(n).
	for i := 0; i < len(ranges)-1; i++ {
		x := ranges[i+0]
		y := ranges[i+1]
		if x.high >= y.low {
			c.warnCharClassDup(x.source, y.source, cc.Value)
			break
		}
	}
}

func (c *regexpVet) charClassBoundRune(e syntax.Expr) rune {
	switch e.Op {
	case syntax.OpChar:
		return c.stringToRune(e.Value)
	case syntax.OpEscapeHex:
		return c.hexToRune(e)
	case syntax.OpEscapeOctal:
		return c.octalToRune(e)
	default:
		return 0
	}
}

func (c *regexpVet) octalToRune(e syntax.Expr) rune {
	v, _ := strconv.ParseInt(e.Value[len(`\`):], 8, 32)
	return rune(v)
}

func (c *regexpVet) hexToRune(e syntax.Expr) rune {
	var s string
	switch e.Form {
	case syntax.FormEscapeHexFull:
		s = e.Value[len(`\x{`) : len(e.Value)-len(`}`)]
	default:
		s = e.Value[len(`\x`):]
	}
	v, _ := strconv.ParseInt(s, 16, 32)
	return rune(v)
}

func (c *regexpVet) stringToRune(s string) rune {
	ch, _ := utf8.DecodeRuneInString(s)
	return ch
}

func (c *regexpVet) warn(format string, args ...interface{}) {
	c.issues = append(c.issues, fmt.Sprintf(format, args...))
}

func (c *regexpVet) warnSloppyCharRange(rng, charClass string) {
	c.warn("suspicious char range '%s' in %s", rng, charClass)
}

func (c *regexpVet) warnCharClassDup(x, y, charClass string) {
	if x == y {
		c.warn("'%s' is duplicated in %s", x, charClass)
	} else {
		c.warn("'%s' intersects with '%s' in %s", x, y, charClass)
	}
}
