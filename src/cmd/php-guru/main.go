package main

import (
	"log"
	"os"
	"sort"

	"github.com/VKCOM/noverify/src/cmd/php-guru/dupcode"
	"github.com/VKCOM/noverify/src/cmd/php-guru/guru"
)

var commands []*subCommand

func guruCommands() []*subCommand {
	commands := []*subCommand{
		{
			name:    "help",
			main:    cmdHelp,
			summary: "print documentation based on the subject",
			examples: []subCommandExample{
				{
					comment: "show supported sub-commands",
					line:    "",
				},
			},
		},

		{
			name:    "dupcode",
			main:    dupcode.Main,
			summary: "find the code duplication",
			examples: []subCommandExample{
				{
					comment: "show dupcode sub-command help",
					line:    "-help",
				},
				{
					comment: "print all code duplicates across the project",
					line:    "path/to/project",
				},
			},
		},
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].name < commands[j].name
	})

	return commands
}

func main() {
	log.SetFlags(0)

	commands = guruCommands()

	if len(os.Args) < 2 {
		log.Println("Please provide a sub-command argument")
		printSupportedCommands(commands)
		os.Exit(1)
	}

	subcmdName := os.Args[1]
	subcmd := findSubCommand(commands, subcmdName)
	if subcmd == nil {
		log.Printf("Sub-command  %s doesn't exist\n\n", subcmdName)
		printSupportedCommands(commands)
		os.Exit(1)
	}

	subIdx := 1 // [0] is program name
	// Erase sub-command argument (index=1) to make it invisible for
	// sub commands themselves.
	os.Args = append(os.Args[:subIdx], os.Args[subIdx+1:]...)

	ctx := guru.NewContext()
	status, err := subcmd.main(ctx)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(status)
}
