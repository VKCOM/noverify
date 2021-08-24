package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

type AppContext struct {
	App *App

	MainConfig  *MainConfig
	ParsedFlags ParsedFlags
	ParsedArgs  []string
	FlagSet     *flag.FlagSet
	flagGroups  *FlagsGroups

	CustomFlags interface{}
}

func (ctx *AppContext) FormatFlags() (res string) {
	if ctx.FlagSet == nil {
		return ""
	}

	first := true

	for _, group := range ctx.flagGroups.Groups {
		name := group
		flags, ok := ctx.flagGroups.Map[group]
		if !ok {
			continue
		}

		if !first {
			res += "\n"
		}
		first = false

		res += fmt.Sprintf(" %s:\n", name)
		for _, flagName := range flags {
			f := ctx.FlagSet.Lookup(flagName)
			if f == nil {
				continue
			}

			defaultVal := f.DefValue
			if f.DefValue != "" {
				defaultVal = fmt.Sprintf("(default: %s)", f.DefValue)
			}
			res += fmt.Sprintf("   --%s %s\n      %s\n", f.Name, defaultVal, f.Usage)
		}
	}

	return res
}

func (ctx *AppContext) CountDefinedFlags() int {
	if ctx.FlagSet == nil {
		return 0
	}
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

func (a *App) GetRawCommandByName(name string) *Command {
	for _, command := range a.Commands {
		if command.Name == name {
			return command
		}
	}

	return nil
}

func (a *App) addHelpCommands(commands []*Command) {
	for _, command := range commands {
		if command.Name != "help" {
			a.addDefaultHelpCommand(command)
		}

		if len(command.Commands) > 0 {
			a.addHelpCommands(command.Commands)
		}
	}
}

func (a *App) prepareCommands() {
	if a.commands == nil {
		a.commands = map[string]*Command{}
	}

	a.addHelpCommands(a.Commands)

	for _, command := range a.Commands {
		a.commands[command.Name] = command
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
	if command.Pure {
		return
	}

	command.Commands = append(command.Commands, &Command{
		Name:        "help",
		Description: "The command to show help for " + command.Name + " command",
		Action: func(ctx *AppContext) (int, error) {
			if command.RegisterFlags != nil {
				fs, groups := command.RegisterFlags(ctx)
				err := fs.Parse(os.Args)
				if err != nil {
					return 2, err
				}
				command.flagSet = fs
				ctx.FlagSet = fs
				ctx.flagGroups = groups
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

			if len(command.Examples) > 0 {
				res += fmt.Sprintln()
				res += fmt.Sprintln("Examples:")

				for _, example := range command.Examples {
					res += fmt.Sprintf("  $ %s %s %s - %s\n", a.Name, command.Name, example.Line, example.Description)
				}
			}

			if withArgs {
				res += fmt.Sprintln()
				res += fmt.Sprintln("Arguments:")

				buf := bytes.NewBuffer(nil)
				w := tabwriter.NewWriter(buf, 15, 0, 5, ' ', 0)
				for _, arg := range command.Arguments {
					fmt.Fprintf(w, "  %s\t%s\n", arg.Name, arg.Description)
				}
				w.Flush()
				res += buf.String()
			}

			if withFlags {
				res += fmt.Sprintln()
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

	fmt.Println(res)
}

func printCommands(w io.Writer, level int, commands map[string]*Command) {
	commandSlice := make([]*Command, 0, len(commands))
	for _, command := range commands {
		commandSlice = append(commandSlice, command)
	}
	sort.Slice(commandSlice, func(i, j int) bool {
		return commandSlice[i].Name < commandSlice[j].Name
	})

	for _, command := range commandSlice {
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
	os.Args = os.Args[1:]

	a.prepareCommands()

	// Show help when no commandline arguments.
	if len(os.Args) < 1 {
		a.showHelp()
		return 0, nil
	}

	// We do a pre-check by receiving the command only for the first argument.
	command, found := a.getCommandByArgs(os.Args[:1], a.commands)
	if !found {
		a.showHelp()
		return 0, nil
	}

	// And if this command is not pure, then only then do
	// we turn the help flags into subcommand.
	if !command.Pure {
		// replace -help or --help with help subcommand
		for i := range os.Args {
			if strings.HasSuffix(os.Args[i], "help") && strings.HasPrefix(os.Args[i], "-") {
				os.Args[i] = "help"
			}
		}
	}

	command, found = a.getCommandByArgs(os.Args, a.commands)
	if !found {
		a.showHelp()
		return 0, nil
	}

	ctx := &AppContext{
		App:         a,
		MainConfig:  cfg,
		ParsedFlags: ParsedFlags{},
	}

	if command.Pure {
		if command.RegisterFlags != nil {
			return 0, fmt.Errorf("for a pure commands, flags are expected to be registered inside Action")
		}
		return command.Action(ctx)
	}

	var fs *flag.FlagSet
	var groups *FlagsGroups

	if command.RegisterFlags != nil {
		fs, groups = command.RegisterFlags(ctx)
		fs.Usage = nil

		// We don't need any standard output, so we disable it.
		a.disableDefaultFlagsOutput(fs)

		err := fs.Parse(os.Args[1:])
		if err != nil {
			return a.flagNotFound(command.Name, err, ctx)
		}
		command.flagSet = fs
		ctx.ParsedArgs = fs.Args()
		ctx.FlagSet = fs
		ctx.flagGroups = groups
	} else {
		fs = flag.NewFlagSet("empty", flag.ContinueOnError)
		// We don't need any standard output, so we disable it.
		a.disableDefaultFlagsOutput(fs)
	}

	err := fs.Parse(os.Args[1:])
	if err != nil {

		return a.flagNotFound(command.Name, err, ctx)
	}

	command.flagSet = fs
	ctx.ParsedArgs = fs.Args()
	ctx.FlagSet = fs

	return command.Action(ctx)
}

func (a *App) disableDefaultFlagsOutput(fs *flag.FlagSet) {
	dummy := bytes.NewBuffer(nil)
	fs.SetOutput(dummy)
	fs.Usage = func() {}
}

func (a *App) flagNotFound(commandName string, err error, ctx *AppContext) (int, error) {
	if strings.HasPrefix(err.Error(), "flag provided but not defined: ") {
		flagName := strings.TrimPrefix(err.Error(), "flag provided but not defined: ")
		log.Printf("Flag -%s not found for '%s' command\n", flagName, commandName)

		command, found := a.getCommandByArgs([]string{commandName, "help"}, a.commands)
		if found {
			_, _ = command.Action(ctx)
		}
		return 2, nil
	}
	return 2, err
}
