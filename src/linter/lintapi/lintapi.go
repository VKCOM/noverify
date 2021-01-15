package lintapi

// This file exists to detach some parts from linter package
// and avoid cyclic package dependencies.
//
// For instance, rules package needs warning levels to create
// rule objects. Linter uses rules, so it can't simply import linter package.
//
// TODO: might want to replace this package with "reports" and move
// linter.Report type in here, as well as some related utilities.

const (
	LevelError    = 1
	LevelWarning  = 2
	LevelInfo     = 3
	LevelNotice   = 4 // do not treat this warning as a reason to reject if we get this kind of warning
	LevelSecurity = 5
)
