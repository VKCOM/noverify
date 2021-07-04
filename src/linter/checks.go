package linter

import (
	"regexp"
)

type Checks struct {
	All []CheckerInfo

	EnableAll bool
	Allowed   map[string]bool
	Excluded  map[string]bool
	Critical  map[string]bool

	ExcludeFileRegexp *regexp.Regexp
}

func NewChecks() *Checks {
	return &Checks{
		Allowed:  map[string]bool{},
		Excluded: map[string]bool{},
		Critical: map[string]bool{},
	}
}

func (c *Checks) IsEnabledCheck(checkName string) bool {
	if !c.EnableAll && !c.Allowed[checkName] {
		return false // Not enabled by --allow-checks.
	}

	if c.Excluded[checkName] {
		return false // Disabled by --exclude-checks.
	}

	return true
}

func (c *Checks) IsCriticalReport(r *Report) bool {
	if len(c.Critical) != 0 {
		return c.Critical[r.CheckName]
	}
	return r.IsCritical()
}

func (c *Checks) IsEnabledReport(r *Report) bool {
	if !c.IsEnabledCheck(r.CheckName) {
		return false
	}

	if c.ExcludeFileRegexp == nil {
		return true
	}

	// Disabled by a file comment.
	return !c.ExcludeFileRegexp.MatchString(r.Filename)
}
