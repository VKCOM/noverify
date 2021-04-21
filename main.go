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
	log.SetFlags(log.Flags() | log.Ltime)

	// You can register your own rules here, see src/linter/custom.go

	printVersion()
	cmd.Main(&cmd.MainConfig{
		LinterVersion: BuildCommit,

		// example of modify
		// TODO: remove before PR merge
		ModifyApp: func(app *cmd.App) {
			app.Name = "phplinter"

			app.Commands = append(app.Commands, &cmd.Command{
				Name:        "version",
				Description: "print phplinter version and exit",
				Action: func(ctx *cmd.AppContext) (int, error) {
					printVersion()
					return 0, nil
				},
			})
		},
	})
}
