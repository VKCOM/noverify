package phpdoc

import "strings"

type CommentPart struct {
	Line       int      // Comment part location inside phpdoc comment
	Name       string   // e.g. "param" for "* @param something bla-bla-bla"
	Params     []string // {"something", "bla-bla-bla"} in example above
	ParamsText string   // "something bla-bla-bla" in example above
}

// ContainsParam reports whether comment part contains param of specified name.
func (part *CommentPart) ContainsParam(name string) bool {
	for _, p := range part.Params {
		if p == name {
			return true
		}
	}
	return false
}

// IsPHPDoc checks if the string is a doc comment
func IsPHPDoc(doc string) bool {
	return strings.HasPrefix(doc, "/**")
}

// Parse returns parsed doc comment with interesting parts (ones that start "* @")
func Parse(doc string) (res []CommentPart) {
	if !IsPHPDoc(doc) {
		return nil
	}

	lines := strings.Split(doc, "\n")
	for i, ln := range lines {
		ln = strings.TrimSpace(ln)
		if len(ln) == 0 {
			continue
		}

		ln = strings.TrimPrefix(ln, "/**")
		ln = strings.TrimPrefix(ln, "*")
		ln = strings.TrimSuffix(ln, "*/")
		ln = strings.TrimSpace(ln)

		if !strings.HasPrefix(ln, "@") {
			continue
		}

		var text string
		nameEndPos := strings.Index(ln, " ")
		if nameEndPos != -1 {
			text = strings.TrimSpace(ln[nameEndPos:])
		}

		fields := strings.Fields(ln)
		if len(fields) == 0 {
			continue
		}

		res = append(res, CommentPart{
			Line:       i + 1,
			Name:       strings.TrimPrefix(fields[0], "@"),
			Params:     fields[1:],
			ParamsText: text,
		})
	}

	return res
}
