package cmd

import (
	"fmt"
)

// Build* are initialized during the build via -ldflags
var (
	BuildVersion string
	BuildTime    string
	BuildOSUname string
	BuildCommit  string
)

func printVersion() {
	fmt.Print("NoVerify, ")
	if BuildCommit == "" {
		fmt.Printf("version %s: built without additional version info (try use `make install`)\n", BuildVersion)
	} else {
		fmt.Printf("version %s: built on: %s OS: %s Commit: %s\n", BuildVersion, BuildTime, BuildOSUname, BuildCommit)
	}
}
