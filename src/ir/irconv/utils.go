package irconv

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
)

func fullyQualifiedToString(n *name.FullyQualified) string {
	s := make([]string, 1, len(n.Parts)+1)
	for _, v := range n.Parts {
		s = append(s, v.(*name.NamePart).Value)
	}
	return strings.Join(s, `\`)
}

// namePartsToString converts slice of *name.NamePart to string
func namePartsToString(parts []node.Node) string {
	s := make([]string, 0, len(parts))
	for _, v := range parts {
		s = append(s, v.(*name.NamePart).Value)
	}
	return strings.Join(s, `\`)
}

// interpretString returns s with all escape sequences replaced.
//
// It tries to follow PHP rules as close are possible, but also
// expects strings to be valid and parseable by the compliant PHP-parser.
//
// If, for whatever reason, a bad strign escape was encountered,
// second returned value will be false.
func interpretString(s string, quote byte) (string, bool) {
	switch quote {
	case '\'', '"':
		// OK
	default:
		return "", false
	}

	if !strings.Contains(s, `\`) {
		// Fast path: nothing to replace.
		return s, true
	}

	// To understand what's going on, consult the manual:
	// https://www.php.net/manual/en/language.types.string.php#language.types.string.syntax.double

	if quote == '"' {
		return interpretStringQ2(s)
	}
	return interpretStringQ1(s)
}

// interpretStringQ1 returns s interpreted value as a single-quoted PHP string.
func interpretStringQ1(s string) (string, bool) {
	var out strings.Builder
	out.Grow(len(s))

	i := 0
	for i < len(s) {
		ch := s[i]

		switch {
		case ch == '\\':
			if !hasOffset(s, i+1) {
				return "", false
			}
			switch s[i+1] {
			case '\'':
				out.WriteByte('\'')
				i += 2
			case '\\':
				out.WriteByte(s[i+1])
				i += 2
			default:
				out.WriteString(s[i : i+2])
				i += 2
			}

		case ch <= unicode.MaxASCII:
			out.WriteByte(ch)
			i++

		default:
			r, n := utf8.DecodeRuneInString(s[i:])
			out.WriteRune(r)
			i += n
		}
	}

	return out.String(), true
}

// interpretStringQ2 returns s interpreted value as a double-quoted PHP string.
func interpretStringQ2(s string) (string, bool) {
	var out strings.Builder
	out.Grow(len(s))

	i := 0
	for i < len(s) {
		ch := s[i]

		switch {
		case ch == '\\':
			if !hasOffset(s, i+1) {
				return "", false
			}
			switch s[i+1] {
			case 'u': // \u{[0-9A-Fa-f]+}
				if !hasOffset(s, i+2) || s[i+2] != '{' {
					out.WriteString(`\u`)
					i += 2
					break
				}
				end := strings.IndexByte(s[i+len(`\u`):], '}')
				if end == -1 {
					return "", false
				}
				codepoints := s[i+len(`\u{`) : i+len(`\u{`)+end-len(`}`)]
				goLiteral := `\U` + zeros[:8-len(codepoints)] + codepoints
				ch, _, _, err := strconv.UnquoteChar(goLiteral, '"')
				if err != nil {
					return "", false
				}
				out.WriteRune(ch)
				i += len(`\u{`) + len(codepoints) + len(`}`)
			case '\'':
				out.WriteString(`\'`)
				i += 2
			case '"':
				out.WriteByte('"')
				i += 2
			case '$':
				out.WriteByte('$')
				i += 2
			case 'n':
				out.WriteByte('\n')
				i += 2
			case 'r':
				out.WriteByte('\r')
				i += 2
			case 't':
				out.WriteByte('\t')
				i += 2
			case 'v':
				out.WriteByte('\v')
				i += 2
			case 'f':
				out.WriteByte('\f')
				i += 2
			case 'e':
				out.WriteByte(0x1B) // ESC
				i += 2
			case '\\':
				out.WriteByte(s[i+1])
				i += 2
			case '0', '1', '2', '3', '4', '5', '6', '7':
				digits := 1
				if hasOffset(s, i+2) && isOctalDigit(s[i+2]) {
					digits++
				}
				if hasOffset(s, i+3) && isOctalDigit(s[i+3]) {
					digits++
				}
				v, err := strconv.ParseInt(s[i+len(`\`):i+len(`\`)+digits], 8, 64)
				if err == nil {
					out.WriteByte(byte(v)) // Overflow is OK
				} else {
					out.WriteString(s[i : i+len(`\`)+digits])
				}
				i += len(`\`) + digits
			case 'x':
				digits := 0
				if hasOffset(s, i+2) && isHexDigit(s[i+2]) {
					digits++
				}
				if hasOffset(s, i+3) && isHexDigit(s[i+3]) {
					digits++
				}
				if digits == 0 {
					out.WriteString(`\x`)
					i += 2
					break
				}
				v, err := strconv.ParseInt(s[i+len(`\x`):i+len(`\x`)+digits], 16, 64)
				if err == nil && v <= 255 {
					out.WriteByte(byte(v))
				} else {
					out.WriteString(s[i : i+len(`\x`)+digits])
				}
				i += len(`\x`) + digits
			default:
				out.WriteString(s[i : i+2])
				i += 2
			}

		case ch <= unicode.MaxASCII:
			out.WriteByte(ch)
			i++

		default:
			r, n := utf8.DecodeRuneInString(s[i:])
			out.WriteRune(r)
			i += n
		}
	}

	return out.String(), true
}

var zeros = "00000000"

func isOctalDigit(ch byte) bool { return ch >= '0' && ch <= '7' }

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') ||
		(ch >= 'a' && ch <= 'f') ||
		(ch >= 'A' && ch <= 'F')
}

func hasOffset(s string, offset int) bool {
	return len(s) > offset
}
