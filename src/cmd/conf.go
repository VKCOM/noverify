package cmd

import (
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/rules"
)

// InitEnvironment passes the state that may be required to finish
// custom linters initialization.
type InitEnvironment struct {
	RuleSets []*rules.Set

	MetaInfo *meta.Info
}

// MainConfig describes optional main function config.
// All zero field values have some defined behavior.
type MainConfig struct {
	// AfterFlagParse is called right after flag.Parse() returned.
	// Can be used to examine flags that were bound prior to the Main() call.
	//
	// If nil, behaves as a no-op function.
	AfterFlagParse func(InitEnvironment)

	// BeforeReport acts as both an on-report action and a filter.
	//
	// If false is returned, the given report will not be reported.
	BeforeReport func(*linter.Report) bool

	LinterVersion string

	LinterConfig *linter.Config
	linter       *linter.Linter
	rulesSets    []*rules.Set

	// RegisterCheckers is used to register additional checkers.
	RegisterCheckers func() []linter.CheckerInfo

	// ModifyApp is a callback function into which a standard
	// application is passed to modify a command, name or description.
	ModifyApp func(app *App)

	// If true, then the messages after reports is not displayed.
	DisableAfterReportsLog bool
}
