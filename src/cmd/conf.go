package cmd

import (
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
)

// InitEnvironment passes the state that may be required to finish
// custom linters initialization.
type InitEnvironment struct {
	RuleSets []*rules.Set
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

	// OverriddenCommands is a list of new commands and
	// commands that override existing commands.
	OverriddenCommands *Commands
}
