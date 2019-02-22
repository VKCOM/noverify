package phpdoc

import "strings"

type CommentPart struct {
	Name   string   // e.g. "param" for "* @param something bla-bla-bla"
	Params []string // {"something", "bla-bla-bla"} in example above
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
	for _, ln := range lines {
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

		fields := strings.Fields(ln)
		if len(fields) == 0 {
			continue
		}

		res = append(res, CommentPart{Name: strings.TrimPrefix(fields[0], "@"), Params: fields[1:]})
	}

	return res
}
