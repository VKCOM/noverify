package rules

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/linter/lintapi"
)

func cloneRuleList(rules []Rule) []Rule {
	res := make([]Rule, len(rules))
	for i, rule := range rules {
		res[i] = rule
		res[i].Matcher = rule.Matcher.Clone()
	}
	return res
}

func formatRule(r *Rule) string {
	var buf strings.Builder

	buf.WriteString("/**\n")

	switch r.Level {
	case lintapi.LevelError:
		buf.WriteString(" * @error " + r.Message + "\n")
	case lintapi.LevelWarning:
		buf.WriteString(" * @warning " + r.Message + "\n")
	case lintapi.LevelInformation:
		buf.WriteString(" * @info " + r.Message + "\n")
	case lintapi.LevelMaybe:
		buf.WriteString(" * @maybe " + r.Message + "\n")
	}

	if r.Location != "" {
		buf.WriteString(" * @location $" + r.Location + "\n")
	}

	if r.scope != "" {
		buf.WriteString(" * @scope " + r.scope + "\n")
	}

	for i, filters := range r.Filters {
		for name, filter := range filters {
			if len(filter.Types) != 0 {
				fmt.Fprintf(&buf, " * @type %s $%s\n", strings.Join(filter.Types, "|"), name)
			}
		}
		if i != len(r.Filters)-1 {
			buf.WriteString(" * @or\n")
		}
	}

	buf.WriteString(" */")

	return buf.String()

}
