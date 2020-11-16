package cmd

import (
	"fmt"
	"io"
	"sort"
)

type Commands struct {
	Commands map[string]*SubCommand
}

func NewCommands() *Commands {
	return &Commands{
		Commands: map[string]*SubCommand{},
	}
}

func (c *Commands) RegisterCommand(command *SubCommand) {
	if command == nil {
		return
	}

	name := command.Name
	c.Commands[name] = command
}

func (c *Commands) OverrideCommand(command *SubCommand) {
	c.RegisterCommand(command)
}

func (c *Commands) GetCommand(name string) (*SubCommand, bool) {
	command, ok := c.Commands[name]
	return command, ok
}

func (c *Commands) OverrideCommands(commands *Commands) {
	if commands == nil {
		return
	}
	if commands.Commands == nil {
		return
	}

	for _, command := range commands.Commands {
		c.OverrideCommand(command)
	}
}

func (c *Commands) HelpPage() string {
	var res string

	commands := make([]*SubCommand, 0, len(c.Commands))
	for _, command := range c.Commands {
		commands = append(commands, command)
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	res += fmt.Sprintf("Supported sub-commands:\n")
	for _, cmd := range commands {
		res += cmd.String()
	}

	return res
}

func (c *Commands) PrintHelpPage() {
	fmt.Print(c.HelpPage())
}

func (c *Commands) WriteHelpPage(w io.Writer) {
	fmt.Fprint(w, c.HelpPage())
}
