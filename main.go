package main

import (
	"log"

	"github.com/VKCOM/noverify/src/cmd"
)

// Build* are initialized during the build via -ldflags
var (
	BuildTime    string
	BuildOSUname string
	BuildCommit  string
)

func printVersion() {
	if BuildCommit == "" {
		log.Printf("built without version info (try using 'make install'?)")
	} else {
		log.Printf("built on: %s OS: %s Commit: %s\n", BuildTime, BuildOSUname, BuildCommit)
	}
}

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	// You can register your own rules here, see src/linter/custom.go

	printVersion()
	cmd.Main(&cmd.MainConfig{
		LinterVersion: BuildCommit,
	})
}
