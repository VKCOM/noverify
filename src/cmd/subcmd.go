package cmd

import (
	"fmt"
	"strings"
)

type subCommand struct {
	name     string
	main     func(*MainConfig) (int, error)
	summary  string
	examples []subCommandExample
	cfg      *MainConfig
}

type subCommandExample struct {
	comment string
	line    string
}

func findSubCommand(list []*subCommand, name string) *subCommand {
	for _, cmd := range list {
		if cmd.name == name {
			return cmd
		}
	}
	return nil
}

func looksLikeCommandName(s string) bool {
	return !strings.HasPrefix(s, "-") &&
		!strings.Contains(s, ".") &&
		!strings.Contains(s, "/")
}

func printSupportedCommands(list []*subCommand) {
	fmt.Printf("Supported sub-commands:\n")
	for _, cmd := range list {
		fmt.Printf("\n\tnoverify %s\n", cmd.name)
		fmt.Printf("\tDescription: %s.\n", cmd.summary)
		for _, ex := range cmd.examples {
			fmt.Printf("\t%s:\n", ex.comment)
			fmt.Printf("\t\t$ noverify %s %s\n", cmd.name, ex.line)
		}
	}
}
