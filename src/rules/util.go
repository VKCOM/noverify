package rules

import (
	"strings"

	"github.com/VKCOM/noverify/src/linter/lintapi"
)

func formatRule(r *Rule) string {
	var buf strings.Builder

	buf.WriteString("/**\n")

	switch r.Level {
	case lintapi.LevelError:
		buf.WriteString(" * @error " + r.Message + "\n")
	case lintapi.LevelWarning:
		buf.WriteString(" * @warning " + r.Message + "\n")
	case lintapi.LevelNotice:
		buf.WriteString(" * @maybe " + r.Message + "\n")
	}

	if len(r.Path) > 0 {
		for _, path := range r.Path {
			buf.WriteString(" * @path " + path + "\n")
		}
	}

	if r.PathExcludes != nil {
		for pathExclude := range r.PathExcludes {
			buf.WriteString(" * @path-exclude " + pathExclude + "\n")
		}
	}

	if r.Location != "" {
		buf.WriteString(" * @location $" + r.Location + "\n")
	}

	if r.scope != "" {
		buf.WriteString(" * @scope " + r.scope + "\n")
	}

	for i, filters := range r.Filters {
		for name, filter := range filters {
			if filter.Type != nil {
				buf.WriteString(" * @type ")
				buf.WriteString(filter.Type.String())
				buf.WriteString(" $" + name + "\n")
			}
		}
		if i != len(r.Filters)-1 {
			buf.WriteString(" * @or\n")
		}
	}

	buf.WriteString(" */")

	return buf.String()

}
