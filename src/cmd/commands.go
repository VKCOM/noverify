package cmd

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
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

	res += fmt.Sprintln("Usage:")
	res += fmt.Sprintln("  $ noverify [command] [options] /project/root")
	res += fmt.Sprintln()

	res += fmt.Sprintln("Commands:")
	res += c.printAlign(commands)

	return res
}

func (c *Commands) printAlign(commands []*SubCommand) string {
	buf := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(buf, 50, 0, 1, ' ', 0)

	for _, cmd := range commands {
		fmt.Fprintf(w, "  %s\t%s\n", cmd.Name, cmd.Description)
		fmt.Fprintf(w, "    Recipes:\n")
		for _, ex := range cmd.Examples {
			command := fmt.Sprintf("      $ noverify %s %s", cmd.Name, ex.Line)
			fmt.Fprintf(w, "%s\t%s\t\n", command, ex.Description)
		}
	}

	w.Flush()

	return buf.String()
}
