package cmd

func getSubCommands(cfg *MainConfig) []*subCommand {
	subCommands := []*subCommand{
		{
			name:    "check",
			main:    cmdCheck,
			summary: "lint the entire project",
			examples: []subCommandExample{
				{
					comment: "show subcommand usage",
					line:    "-help",
				},
				{
					comment: "run linter with default options",
					line:    "<analyze-path>",
				},
			},
		},

		{
			name:    "help",
			main:    cmdHelp,
			summary: "print linter documentation based on the subject",
			examples: []subCommandExample{
				{
					comment: "show supported sub-subCommands",
					line:    "",
				},
				{
					comment: "print all supported checkers short summary",
					line:    "checkers",
				},
				{
					comment: "print dupSubExpr checker detailed documentation",
					line:    "checkers dupSubExpr",
				},
			},
		},
	}

	return subCommands
}
