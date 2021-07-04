package cmd

import (
	"flag"
)

type Argument struct {
	Name        string
	Description string
}

type Example struct {
	Line        string
	Description string
}

type Command struct {
	Name        string
	Description string
	Action      func(*AppContext) (int, error)

	Examples []Example

	Arguments     []*Argument
	RegisterFlags func(*AppContext) (fs *flag.FlagSet, groups *FlagsGroups)
	flagSet       *flag.FlagSet

	// Pure flag defines the command, which itself is responsible
	// for registering flags and arguments.
	// For such a command, help subcommand is not automatically generated.
	Pure bool

	Commands []*Command
	commands map[string]*Command
}

func (c *Command) prepareCommands() {
	if c.commands == nil {
		c.commands = map[string]*Command{}
	}

	for _, command := range c.Commands {
		command.prepareCommands()
		c.commands[command.Name] = command
	}
}
