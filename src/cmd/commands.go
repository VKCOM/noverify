package cmd

import (
	"fmt"
	"io"
	"sort"

	"github.com/gosuri/uitable"
	"github.com/i582/cfmt"
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

func (c *Commands) PrintHelpPage() {
	fmt.Print(c.HelpPage())
}

func (c *Commands) WriteHelpPage(w io.Writer) {
	fmt.Fprint(w, c.HelpPage())
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

	res += cfmt.Sprintln("{{Usage:}}::yellow")
	res += cfmt.Sprintln("  {{$}}::gray noverify {{[command]}}::green {{[options]}}::gray /project/root")
	res += fmt.Sprintln()

	res += cfmt.Sprintln("{{Commands:}}::yellow")
	res += c.printAlign(commands)

	return res
}

func (c *Commands) printAlign(commands []*SubCommand) string {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = 65

	for _, cmd := range commands {
		var left string
		left += cfmt.Sprintf("  {{%s}}::green\n", cmd.Name)
		left += fmt.Sprintln("    Recipes:")
		for _, ex := range cmd.Examples {
			left += cfmt.Sprintf("      $ noverify %s %s\n", cmd.Name, ex.Line)
		}

		var right string
		right += fmt.Sprintln(cmd.Description)
		right += fmt.Sprintln()
		for _, ex := range cmd.Examples {
			right += fmt.Sprintln(ex.Description)
		}

		table.AddRow(left, right)
		table.Rows[len(table.Rows)-1].Cells[0].Width = 145
	}

	return table.String()
}
