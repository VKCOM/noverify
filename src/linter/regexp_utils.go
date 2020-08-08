package linter

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
)

type regexpPattern struct {
	value  string
	quotes rune
}

func newRegexpPattern(lit *ir.String) (regexpPattern, bool) {
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

	var result regexpPattern

	switch lit.Value[0] {
	case '\'':
		// Single quotes only interpret \' and \\ sequences.
		pat := unquote(lit.Value)
		pat = strings.ReplaceAll(pat, `\\`, `\`)
		pat = strings.ReplaceAll(pat, `\'`, `'`)
		result.value = pat
		result.quotes = '\''
	case '"':
		// TODO: interpret the double-quoted literals?
		// Note that it would also be harder to reconstruct
		// the literal syntax for the simplified string.
		result.quotes = '"'
		return result, false
	default:
		return result, false
	}
	return result, true
}
