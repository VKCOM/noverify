package linter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type fmtDirective struct {
	begin uint
	end   uint

	// specifier is a verb-like character.
	// Examples: `d` `s`
	specifier byte

	explicitArgNum bool

	// An argNum of -1 is used when the directive does not refer to any argument.
	argNum int

	// flags is a string that combines all directive flags.
	// Examples: `+` `'.`
	flags string

	// precision controls the specifier formatting (the exact effect depends on the specifier).
	// A precision of -1 means that this directive had no explicit precision specifier.
	// Examples: `.4` `.`
	precision int

	// width says how many characters (minimum) this conversion should result in.
	// A width of -1 means that this directive had no explicit width specifier.
	width int
}

func (d fmtDirective) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("%%%c", d.specifier))

	if d.argNum != -1 {
		parts = append(parts, fmt.Sprintf("arg=%d", d.argNum))
		if !d.explicitArgNum {
			parts = append(parts, "(implicit)")
		}
	}

	if d.precision != -1 {
		parts = append(parts, fmt.Sprintf("p=%d", d.precision))
	}
	if d.width != -1 {
		parts = append(parts, fmt.Sprintf("w=%d", d.width))
	}

	if d.flags != "" {
		parts = append(parts, fmt.Sprintf("flags=%s", d.flags))
	}

	return strings.Join(parts, " ")
}

type fmtString struct {
	directives []fmtDirective
}

func parseFormatString(s string) (fmtString, error) {
	var p fmtStringParser
	return p.Parse(s)
}

type fmtStringParser struct {
	input      string
	pos        uint
	directives []fmtDirective

	argNum int
	d      fmtDirective
}

func (p *fmtStringParser) Parse(s string) (fmtString, error) {
	p.reset(s)

	err := p.parse()

	result := fmtString{
		directives: p.directives,
	}

	return result, err
}

func (p *fmtStringParser) reset(s string) {
	p.input = s
	p.pos = 0
	p.argNum = 1
}

func (p *fmtStringParser) peek(offset uint) byte {
	pos := p.pos + offset
	if pos < uint(len(p.input)) {
		return p.input[pos]
	}
	return 0
}

func (p *fmtStringParser) parse() error {
	for p.pos < uint(len(p.input)) {
		if p.peek(0) != '%' {
			p.pos++
			continue
		}
		if err := p.parseDirective(); err != nil {
			return err
		}
	}
	return nil
}

func (p *fmtStringParser) parseDirective() error {
	p.d = fmtDirective{
		begin:     p.pos,
		width:     -1,
		precision: -1,
	}
	p.pos += uint(len("%"))

	p.parseArgnum()
	if err := p.parseFlags(); err != nil {
		return err
	}
	p.parseWidth()
	p.parsePrecision()
	if err := p.parseSpecifier(); err != nil {
		return err
	}

	p.d.end = p.pos
	p.directives = append(p.directives, p.d)
	return nil
}

func (p *fmtStringParser) parseArgnum() {
	digits := p.scanDigits()
	if digits == "" {
		return
	}
	if p.peek(uint(len(digits))) != '$' {
		return
	}
	p.d.argNum = p.atoi(digits)
	p.d.explicitArgNum = true
	p.pos += uint(len(digits) + len("$"))
}

func (p *fmtStringParser) parseFlags() error {
	length := uint(0)
loop:
	for {
		switch ch := p.peek(length); ch {
		case '-', '+', ' ', '0':
			length++
		case '\'':
			if p.peek(length+1) == 0 {
				return errors.New("'-flag expects a char, found end of string")
			}
			length += 2
		default:
			break loop
		}
	}

	p.d.flags = p.input[p.pos : p.pos+length]
	p.pos += length
	return nil
}

func (p *fmtStringParser) parseWidth() {
	digits := p.scanDigits()
	if digits == "" {
		return
	}
	p.d.width = p.atoi(digits)
	p.pos += uint(len(digits))
}

func (p *fmtStringParser) parsePrecision() {
	if p.peek(0) != '.' {
		return
	}
	p.pos += uint(len("."))
	digits := p.scanDigits()
	if digits == "" {
		return // Not a parse error, default is implied
	}
	p.d.precision = p.atoi(digits)
	p.pos += uint(len(digits))
}

func (p *fmtStringParser) parseSpecifier() error {
	ch := p.peek(0)
	switch ch {
	case 'b', 'c', 'd', 'e', 'E', 'f', 'F', 'g', 'G', 'o', 's', 'u', 'x', 'X':
		if !p.d.explicitArgNum {
			p.d.argNum = p.argNum
			p.argNum++
		}
	case '%':
		// Does not increment the argnum.
		if !p.d.explicitArgNum {
			p.d.argNum = -1
		}
	default:
		return fmt.Errorf("unexpected format specifier %c", ch)
	}
	p.d.specifier = ch
	p.pos++
	return nil
}

func (p *fmtStringParser) scanDigits() string {
	length := uint(0)
	for p.isDigit(p.peek(length)) {
		length++
	}
	if length == 0 {
		return ""
	}
	return p.input[p.pos : p.pos+length]
}

func (p *fmtStringParser) atoi(digits string) int {
	v, err := strconv.Atoi(digits)
	if err != nil {
		// Since digits string contains only decimal chars
		// atoi call should never fail unless we get something
		// wrong during the input slicing.
		panic(fmt.Sprintf("unexpected digits='%s' parse failure", digits))
	}
	return v
}

func (p *fmtStringParser) isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
