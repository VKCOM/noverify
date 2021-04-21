package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

type AppContext struct {
	App *App

	MainConfig  *MainConfig
	ParsedFlags ParsedFlags
	ParsedArgs  []string
	FlagSet     *flag.FlagSet
}

func (ctx *AppContext) FormatFlags() (res string) {
	ctx.FlagSet.VisitAll(func(f *flag.Flag) {
		defaultVal := f.DefValue
		if f.DefValue != "" {
			defaultVal = fmt.Sprintf("(default: %s)", f.DefValue)
		}
		res += fmt.Sprintf("  -%s %s\n      %s\n", f.Name, defaultVal, f.Usage)
	})
	return res
}

func (ctx *AppContext) CountDefinedFlags() int {
	var res int
	ctx.FlagSet.VisitAll(func(*flag.Flag) { res++ })
	return res
}

type App struct {
	Name        string
	Description string

	Commands []*Command

	commands map[string]*Command
}

func (a *App) getCommand(name string) (*Command, bool) {
	cmd, found := a.commands[name]
	return cmd, found
}

func (a *App) prepareCommands() {
	if a.commands == nil {
		a.commands = map[string]*Command{}
	}

	for _, command := range a.Commands {
		a.commands[command.Name] = command

		if command.Name != "help" {
			a.addDefaultHelpCommand(command)
		}

		command.prepareCommands()
	}

	helpCommand, has := a.commands["help"]
	if has {
		if helpCommand.Action == nil {
			helpCommand.Action = func(ctx *AppContext) (int, error) {
				a.showHelp()
				return 0, nil
			}
		}
	}
}

func (a *App) addDefaultHelpCommand(command *Command) {
	a.commands[command.Name].Commands = append(a.commands[command.Name].Commands, &Command{
		Name:        "help",
		Description: "The command to show help for " + command.Name + " command",
		Action: func(ctx *AppContext) (int, error) {
			if command.RegisterFlags != nil {
				fs := command.RegisterFlags(ctx)
				err := fs.Parse(os.Args)
				if err != nil {
					return 2, err
				}
				command.flagSet = fs
				ctx.FlagSet = fs
			}

			withFlags := ctx.CountDefinedFlags() != 0
			withArgs := len(command.Arguments) != 0

			var res string

			options := ""
			if withFlags {
				options = " [options]"
			}
			args := ""
			if withArgs {
				args = " [args]"
			}

			res += fmt.Sprintln("Usage:")
			res += fmt.Sprintf("  $ %s %s%s%s - %s\n", a.Name, command.Name, options, args, command.Description)
			res += fmt.Sprintln()

			if withArgs {
				res += fmt.Sprintln("Arguments:")

				for _, arg := range command.Arguments {
					res += fmt.Sprintf("  %s - %s\n", arg.Name, arg.Description)
				}
			}

			res += fmt.Sprintln()

			if withFlags {
				res += fmt.Sprintln("Options:")

				res += ctx.FormatFlags()
			}

			fmt.Println(res)

			return 0, nil
		},
	})
}

func (a *App) showHelp() {
	var res string

	res += fmt.Sprintf("%s - %s\n", a.Name, a.Description)
	res += fmt.Sprintln()
	res += fmt.Sprintln("Usage:")
	res += fmt.Sprintf("  $ %s [command]\n", a.Name)
	res += fmt.Sprintln()

	res += fmt.Sprintln("Commands:")

	w := tabwriter.NewWriter(os.Stdout, 15, 0, 3, ' ', 0)
	printCommands(w, 1, a.commands)

	fmt.Print(res)

	w.Flush()

	res = fmt.Sprintln()
	res += fmt.Sprintln("Help:")

	res += fmt.Sprintln("  To get help for command, use the help subcommand:")
	res += fmt.Sprintf("    $ %s check help\n", a.Name)
	res += fmt.Sprintf("    $ %s checkers help\n", a.Name)

	fmt.Println(res)
}

func printCommands(w io.Writer, level int, commands map[string]*Command) {
	for _, command := range commands {
		if command.Name == "help" {
			continue
		}

		fmt.Fprintf(w, "%s%s\t%s\n", strings.Repeat("  ", level), command.Name, command.Description)

		if len(command.Commands) != 0 {
			printCommands(w, level+1, command.commands)
		}
	}
}

func (a *App) getCommandByArgs(args []string, commands map[string]*Command) (*Command, bool) {
	if len(args) == 0 {
		return nil, false
	}

	commandName := args[0]
	command, found := commands[commandName]
	if !found {
		return nil, false
	}

	os.Args = os.Args[1:]

	if len(command.commands) == 0 {
		return command, true
	}

	subCommand, found := a.getCommandByArgs(args[1:], command.commands)
	if found {
		return subCommand, true
	}

	return command, true
}

func (a *App) Run(cfg *MainConfig) (int, error) {
	if len(os.Args) == 1 {
		a.showHelp()
	}

	os.Args = os.Args[1:]

	a.prepareCommands()

	command, found := a.getCommandByArgs(os.Args, a.commands)
	if !found {
		a.showHelp()
		return 0, nil
	}

	ctx := &AppContext{
		App:         a,
		MainConfig:  cfg,
		ParsedFlags: ParsedFlags{},
	}

	var fs *flag.FlagSet

	if command.RegisterFlags != nil {
		fs = command.RegisterFlags(ctx)
	} else {
		fs = flag.NewFlagSet("empty", flag.ContinueOnError)
	}

	err := fs.Parse(os.Args)
	if err != nil {
		return 2, err
	}
	command.flagSet = fs
	ctx.ParsedArgs = fs.Args()
	ctx.FlagSet = fs

	return command.Action(ctx)
}
