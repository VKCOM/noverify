package cmd

// MainConfig describes optional main function config.
// All zero field values have some defined behavior.
type MainConfig struct {
	// AfterFlagParse is called right after flag.Parse() returned.
	// Can be used to examine flags that were bound prior to the Main() call.
	//
	// If nil, behaves as a no-op function.
	AfterFlagParse func()
}
