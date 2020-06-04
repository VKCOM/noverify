package phpdoc

import (
	"strings"
)

type CommentPart interface {
	Line() int
	Name() string
}

type RawCommentPart struct {
	line       int      // Comment part location inside phpdoc comment
	name       string   // e.g. "param" for "* @param something bla-bla-bla"
	Params     []string // {"something", "bla-bla-bla"} in example above
	ParamsText string   // "something bla-bla-bla" in example above
}

func (c *RawCommentPart) Line() int    { return c.line }
func (c *RawCommentPart) Name() string { return c.name }

type TypeCommentPart struct {
	line int
	name string
	Type Type
	Rest string
}

func (c *TypeCommentPart) Line() int    { return c.line }
func (c *TypeCommentPart) Name() string { return c.name }

type TypeVarCommentPart struct {
	line       int
	name       string
	VarIsFirst bool
	Var        string
	Type       Type
	Rest       string
}

func (c *TypeVarCommentPart) Line() int    { return c.line }
func (c *TypeVarCommentPart) Name() string { return c.name }

// IsPHPDoc checks if the string is a doc comment
func IsPHPDoc(doc string) bool {
	// See #289.
	return strings.HasPrefix(doc, "/* @var ") ||
		strings.HasPrefix(doc, "/**")
}

// Parse returns parsed doc comment with interesting parts (ones that start "* @")
func Parse(parser *TypeParser, doc string) (res []CommentPart) {
	if !IsPHPDoc(doc) {
		return nil
	}

	var lines []string
	if strings.HasPrefix(doc, "/* @var ") && strings.Count(doc, "\n") == 0 {
		lines = []string{doc}
	} else {
		lines = strings.Split(doc, "\n")
	}

	for i, ln := range lines {
		ln = strings.TrimSpace(ln)
		if len(ln) == 0 {
			continue
		}

		// A combination of /* and * trimming works for both /** and /* comments.
		ln = strings.TrimPrefix(ln, "/*")
		ln = strings.TrimPrefix(ln, "*")
		ln = strings.TrimSuffix(ln, "*/")
		ln = strings.TrimSpace(ln)

		if !strings.HasPrefix(ln, "@") {
			continue
		}

		var text string
		var name string
		nameEndPos := strings.Index(ln, " ")
		if nameEndPos != -1 {
			text = strings.TrimSpace(ln[nameEndPos:])
			name = ln[:nameEndPos]
		} else {
			name = ln
		}
		name = strings.TrimPrefix(name, "@")

		line := i + 1
		var part CommentPart
		switch name {
		case "param", "var", "property":
			part = parseTypeVarComment(parser, line, name, text)
		case "return":
			part = parseTypeComment(parser, line, name, text)
		default:
			part = parseRawComment(line, name, text)
		}

		res = append(res, part)
	}

	return res
}

func parseRawComment(line int, name, text string) *RawCommentPart {
	fields := strings.Fields(text)
	return &RawCommentPart{
		line:       line,
		name:       name,
		Params:     fields,
		ParamsText: text,
	}
}

func parseTypeComment(parser *TypeParser, line int, name, text string) *TypeCommentPart {
	typ, rest := nextTypeField(parser, text)
	return &TypeCommentPart{
		line: line,
		name: name,
		Type: typ,
		Rest: rest,
	}
}

func parseTypeVarComment(parser *TypeParser, line int, name, text string) *TypeVarCommentPart {
	result := TypeVarCommentPart{line: line, name: name}

	result.VarIsFirst = strings.HasPrefix(text, "$")
	if result.VarIsFirst {
		variable, rest := nextField(text)
		result.Var = variable
		typ, rest := nextTypeField(parser, rest)
		result.Type = typ
		result.Rest = rest
	} else {
		typ, rest := nextTypeField(parser, text)
		result.Type = typ
		variable, rest := nextField(rest)
		result.Var = variable
		result.Rest = rest
	}

	return &result
}

func nextField(s string) (field, rest string) {
	delim := strings.IndexByte(s, ' ')
	if delim == -1 {
		return s, ""
	}
	return s[:delim], strings.TrimLeft(s[delim:], " ")
}

func nextTypeField(parser *TypeParser, s string) (field Type, rest string) {
	typ := parser.Parse(s).Clone()
	return typ, strings.TrimLeft(s[typ.Expr.End:], " ")
}
