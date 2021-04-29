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
	RegisterFlags func(*AppContext) *flag.FlagSet
	flagSet       *flag.FlagSet

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
