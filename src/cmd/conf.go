package cmd

// MainHooks provides integration points into cmd.Main for custom linters.
var MainHooks struct {
	// AfterFlagParse is called right after flag.Parse is invoked inside cmd.Main.
	AfterFlagParse func()
}
