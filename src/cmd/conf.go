package cmd

import (
	"github.com/VKCOM/noverify/src/linter"
)

// MainConfig describes optional main function config.
// All zero field values have some defined behavior.
type MainConfig struct {
	// AfterFlagParse is called right after flag.Parse() returned.
	// Can be used to examine flags that were bound prior to the Main() call.
	//
	// If nil, behaves as a no-op function.
	AfterFlagParse func()

	// BeforeReport acts as both an on-report action and a filter.
	//
	// If false is returned, the given report will not be reported.
	BeforeReport func(*linter.Report) bool
}
