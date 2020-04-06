package linter

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/quasilyte/regex/syntax"
)

type regexpSimplifier struct {
	parser *syntax.Parser

	// re is a regexp being analyzed (changed between passes).
	re *syntax.RegexpPCRE
	// out is a tmp buffer where we build a simplified regexp pattern.
	out *strings.Builder
	// score is a number of applied simplifications
	score int
	// escapedDelims counts how many times we've seen escaped
	// pattern delimiter inside the current regexp.
	escapedDelims int
}

func (c *regexpSimplifier) simplifyRegexp(lit *scalar.String) string {
	// We need to interpret the string literal so we see it in a same
	// was as regexp engine (PCRE) would.
	//
	// The '/\\\\/' pattern matches a single '\' literally:
	// - First we interpret 2 \\ pairs as two '\' characters.
	// - Then regexp engine sees two '\' and interprets them as a '\' escape.
	//
	// The '/\d/' works differently, since single quotes don't interptet
	// that escape sequence, so what you see is what you get.
	// The unfortunate consequence is that '/\\d/' is also correct.
	//
	// We probably want to suggest using `\d` instead of `\\d`.
	// Right now it happens as a side effect of regexp re-write process.
	// If pattern score is 0, no \d->\\d will be suggested.
	var pat string
	switch lit.Value[0] {
	case '\'':
		// Single quotes only interpret \' and \\ sequences.
		pat = unquote(lit.Value)
		pat = strings.ReplaceAll(pat, `\\`, `\`)
		pat = strings.ReplaceAll(pat, `\'`, `'`)
	case '"':
		// TODO: interpret the double-quoted literals?
		// Note that it would also be harder to reconstruct
		// the literal syntax for the simplified string.
		return ""
	default:
		return "" // Unreachable?
	}

	// Only do up to 2 passes for now.
	// We can increase this threshold later if there
	// will be any simplifications that are enabled by it.
	const maxPasses = 2
	simplified := pat
	for pass := 0; pass < maxPasses; pass++ {
		candidate := c.simplify(pass, simplified)
		if candidate == "" {
			break
		}
		simplified = candidate
	}

	if simplified == pat {
		return ""
	}
	return simplified
}

func (c *regexpSimplifier) simplify(pass int, pat string) string {
	re, err := c.parser.ParsePCRE(pat)
	if err != nil {
		return ""
	}
	c.re = re
	c.score = 0
	c.escapedDelims = 0
	c.out.Reset()

	c.walk(c.re.Expr)

	// Since '/' delimiter is so ubiquitous, tolerate 1 '\/' without
	// suggesting an alternative delimiter.
	minDelims := 2
	if c.re.Delim[0] != '/' {
		minDelims = 1
	}
	if c.escapedDelims >= minDelims {
		// Try to find a delimiter that may allow us to escape less.
		// Good candidates are enumerated in 'delimiters' string in
		// the order of their priority (more common go first).
		const delimiters = "/~@#"
		var newDelim byte
		for _, d := range delimiters {
			if c.re.Delim[0] == byte(d) {
				continue // Already using this one
			}
			if strings.IndexByte(c.re.Pattern, byte(d)) >= 0 {
				continue // Delimiter char is used inside a pattern
			}
			newDelim = byte(d)
			break
		}
		if newDelim != 0 {
			// It can be tricky to keep all simplifications we've done during
			// this pass, this is why we try to re-parse pattern with
			// new delimiters and re-run  the current pass.
			// This is not very efficient, but robust.
			//
			// simplify() sets score back to 0, but we know that there are
			// at least 2 escaped delimiters that can now be unescaped.
			newPat := fmt.Sprintf("%c%s%c%s", newDelim, c.re.Pattern, newDelim, c.re.Modifiers)
			return c.simplify(pass, newPat)
		}
	}

	if c.score == 0 {
		return ""
	}

	simplified := c.out.String()
	return fmt.Sprintf("%c%s%c%s", c.re.Delim[0], simplified, c.re.Delim[1], c.re.Modifiers)
}

func (c *regexpSimplifier) walk(e syntax.Expr) {
	out := c.out

	switch e.Op {
	case syntax.OpConcat:
		c.walkConcat(e)

	case syntax.OpAlt:
		c.walkAlt(e)

	case syntax.OpCharRange:
		s := c.simplifyCharRange(e)
		if s != "" {
			out.WriteString(s)
			c.score++
		} else {
			out.WriteString(e.Value)
		}

	case syntax.OpGroupWithFlags:
		out.WriteString("(")
		out.WriteString(e.Args[1].Value)
		out.WriteString(":")
		c.walk(e.Args[0])
		out.WriteString(")")
	case syntax.OpGroup:
		c.walkGroup(e)
	case syntax.OpCapture:
		out.WriteString("(")
		c.walk(e.Args[0])
		out.WriteString(")")
	case syntax.OpNamedCapture:
		out.WriteString("(?P<")
		out.WriteString(e.Args[1].Value)
		out.WriteString(">")
		c.walk(e.Args[0])
		out.WriteString(")")

	case syntax.OpRepeat:
		// TODO(quasilyte): is it worth it to analyze repeat argument
		// more closely and handle `{n,n} -> {n}` cases?
		rep := e.Args[1].Value
		switch rep {
		case "{0,1}":
			c.walk(e.Args[0])
			out.WriteString("?")
			c.score++
		case "{1,}":
			c.walk(e.Args[0])
			out.WriteString("+")
			c.score++
		case "{0,}":
			c.walk(e.Args[0])
			out.WriteString("*")
			c.score++
		case "{0}":
			// Maybe {0} should be reported by another check, regexpLint?
			c.score++
		case "{1}":
			c.walk(e.Args[0])
			c.score++
		default:
			c.walk(e.Args[0])
			out.WriteString(rep)
		}

	case syntax.OpPosixClass:
		out.WriteString(e.Value)

	case syntax.OpNegCharClass:
		s := c.simplifyNegCharClass(e)
		if s != "" {
			c.out.WriteString(s)
			c.score++
		} else {
			out.WriteString("[^")
			for _, e := range e.Args {
				c.walk(e)
			}
			out.WriteString("]")
		}

	case syntax.OpCharClass:
		s := c.simplifyCharClass(e)
		if s != "" {
			c.out.WriteString(s)
			c.score++
		} else {
			out.WriteString("[")
			for _, e := range e.Args {
				c.walk(e)
			}
			out.WriteString("]")
		}

	case syntax.OpEscapeChar:
		isDelim := c.re.Delim[0] == e.Value[1] || c.re.Delim[1] == e.Value[1]
		if isDelim {
			c.escapedDelims++
		}
		switch e.Value {
		case `\&`, `\#`, `\!`, `\@`, `\%`, `\<`, `\>`, `\:`, `\;`, `\/`, `\,`, `\=`, `\.`:
			if !isDelim {
				c.score++
				out.WriteString(e.Value[len(`\`):])
				return
			}
		}
		out.WriteString(e.Value)

	case syntax.OpQuestion, syntax.OpNonGreedy:
		c.walk(e.Args[0])
		out.WriteString("?")
	case syntax.OpStar:
		c.walk(e.Args[0])
		out.WriteString("*")
	case syntax.OpPlus, syntax.OpPossessive:
		c.walk(e.Args[0])
		out.WriteString("+")

	default:
		out.WriteString(e.Value)
	}
}

func (c *regexpSimplifier) walkGroup(g syntax.Expr) {
	switch g.Args[0].Op {
	case syntax.OpChar, syntax.OpEscapeChar, syntax.OpEscapeMeta, syntax.OpCharClass:
		c.walk(g.Args[0])
		c.score++
		return
	}

	c.out.WriteString("(?:")
	c.walk(g.Args[0])
	c.out.WriteString(")")
}

func (c *regexpSimplifier) simplifyNegCharClass(e syntax.Expr) string {
	switch e.Value {
	case `[^0-9]`:
		return `\D`
	case `[^\s]`:
		return `\S`
	case `[^\S]`:
		return `\s`
	case `[^\w]`:
		return `\W`
	case `[^\W]`:
		return `\w`
	case `[^\d]`:
		return `\D`
	case `[^\D]`:
		return `\d`
	case `[^[:^space:]]`:
		return `\s`
	case `[^[:space:]]`:
		return `\S`
	case `[^[:^word:]]`:
		return `\w`
	case `[^[:word:]]`:
		return `\W`
	case `[^[:^digit:]]`:
		return `\d`
	case `[^[:digit:]]`:
		return `\D`
	}

	return ""
}

func (c *regexpSimplifier) simplifyCharClass(e syntax.Expr) string {
	switch e.Value {
	case `[0-9]`:
		return `\d`
	case `[[:word:]]`:
		return `\w`
	case `[[:^word:]]`:
		return `\W`
	case `[[:digit:]]`:
		return `\d`
	case `[[:^digit:]]`:
		return `\D`
	case `[[:space:]]`:
		return `\s`
	case `[[:^space:]]`:
		return `\S`
	}

	if len(e.Args) == 1 {
		switch e.Args[0].Op {
		case syntax.OpChar:
			switch v := e.Args[0].Value; v {
			case "|", "*", "+", "?", ".", "[", "^", "$", "(", ")":
				// Can't take outside of the char group without escaping.
			default:
				return v
			}
		case syntax.OpEscapeChar:
			return e.Args[0].Value
		}
	}

	return ""
}

func (c *regexpSimplifier) canMerge(x, y syntax.Expr) bool {
	if x.Op != y.Op {
		return false
	}
	switch x.Op {
	case syntax.OpChar, syntax.OpCharClass, syntax.OpEscapeMeta, syntax.OpEscapeChar, syntax.OpNegCharClass, syntax.OpGroup:
		return x.Value == y.Value
	default:
		return false
	}
}

func (c *regexpSimplifier) canCombine(x, y syntax.Expr) (threshold int, ok bool) {
	if x.Op != y.Op {
		return 0, false
	}

	switch x.Op {
	case syntax.OpDot:
		return 3, true

	case syntax.OpChar:
		if x.Value != y.Value {
			return 0, false
		}
		if x.Value == " " {
			return 1, true
		}
		return 4, true

	case syntax.OpEscapeMeta, syntax.OpEscapeChar:
		if x.Value == y.Value {
			return 2, true
		}

	case syntax.OpCharClass, syntax.OpNegCharClass, syntax.OpGroup:
		if x.Value == y.Value {
			return 1, true
		}
	}

	return 0, false
}

func (c *regexpSimplifier) walkAlt(alt syntax.Expr) {
	allChars := true
	for _, e := range alt.Args {
		if e.Op != syntax.OpChar {
			allChars = false
			break
		}
	}

	if allChars {
		c.score++
		c.out.WriteString("[")
		for _, e := range alt.Args {
			c.out.WriteString(e.Value)
		}
		c.out.WriteString("]")
	} else {
		for i, e := range alt.Args {
			c.walk(e)
			if i != len(alt.Args)-1 {
				c.out.WriteString("|")
			}
		}
	}
}

func (c *regexpSimplifier) walkConcat(concat syntax.Expr) {
	i := 0
	for i < len(concat.Args) {
		x := concat.Args[i]
		c.walk(x)
		i++

		if i >= len(concat.Args) {
			break
		}

		// Try merging `xy*` into `x+` where x=y.
		if concat.Args[i].Op == syntax.OpStar {
			if c.canMerge(x, concat.Args[i].Args[0]) {
				c.out.WriteString("+")
				c.score++
				i++
				continue
			}
		}

		// Try combining `xy` into `x{2}` where x=y.
		threshold, ok := c.canCombine(x, concat.Args[i])
		if !ok {
			continue
		}
		n := 1 // Can combine at least 1 pair.
		for j := i + 1; j < len(concat.Args); j++ {
			_, ok := c.canCombine(x, concat.Args[j])
			if !ok {
				break
			}
			n++
		}
		if n >= threshold {
			fmt.Fprintf(c.out, "{%d}", n+1)
			c.score++
			i += n
		}
	}
}

func (c *regexpSimplifier) simplifyCharRange(rng syntax.Expr) string {
	if rng.Args[0].Op != syntax.OpChar || rng.Args[1].Op != syntax.OpChar {
		return ""
	}

	lo := rng.Args[0].Value
	hi := rng.Args[1].Value
	if len(lo) == 1 && len(hi) == 1 {
		switch hi[0] - lo[0] {
		case 0:
			return lo
		case 1:
			return fmt.Sprintf("%s%s", lo, hi)
		case 2:
			return fmt.Sprintf("%s%s%s", lo, string(lo[0]+1), hi)
		}
	}

	return ""
}
