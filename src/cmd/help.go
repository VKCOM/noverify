package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gookit/color"
	"github.com/gosuri/uitable"
	"github.com/i582/cfmt"

	"github.com/VKCOM/noverify/src/lintdoc"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
)

func Help(*MainConfig) (int, error) {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		GlobalCmds.PrintHelpPage()
		return 0, nil
	}

	mainSubject := args[0]
	if mainSubject == "checkers" {
		return handleCheckersHelp(args[1:])
	}

	return 1, fmt.Errorf("unknown subject: %s", mainSubject)
}

func handleCheckersHelp(args []string) (int, error) {
	config := declareRules()

	// `help checkers`
	if len(args) == 0 {
		showHelpAllCheckers(config)
		return 0, nil
	}

	checkerName := args[0]

	// `help checkers <name>`
	err := showHelpChecker(config, checkerName)
	if err != nil {
		return 1, err
	}

	return 0, err
}

func declareRules() *linter.Config {
	p := rules.NewParser()
	config := linter.NewConfig()

	ruleSets, err := AddEmbeddedRules(config.Rules, p, func(r rules.Rule) bool { return true })
	if err != nil {
		panic(err)
	}

	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	return config
}

func showHelpAllCheckers(config *linter.Config) {
	table := uitable.New()
	for _, info := range config.Checkers.ListDeclared() {
		table.AddRow(color.Green.Sprintf("  %s", info.Name), info.Comment)
	}

	cfmt.Println("{{Usage:}}::yellow")
	cfmt.Println("  {{$}}::gray noverify {{check}}::green {{-allow-checks}}::yellow='{{<list-checks>}}::underline' /project/root")
	fmt.Println()

	cfmt.Println("  {{NOTE:}}::gray In order to run the linter with only some checks, the {{-allow-checks}}::yellow")
	cfmt.Println("  flag is used which accepts a comma-separated list of checks that are allowed.")
	cfmt.Println()
	cfmt.Println("  For other possible options run")
	cfmt.Println("     {{$}}::gray noverify {{check}}::green -help")

	fmt.Println()

	cfmt.Println("{{Checkers:}}::yellow")
	fmt.Println(table.String())
}

func showHelpChecker(config *linter.Config, checkerName string) error {
	var info linter.CheckerInfo
	checks := config.Checkers.ListDeclared()
	for i := range checks {
		if checks[i].Name == checkerName {
			info = checks[i]
		}
	}
	if info.Name == "" {
		return fmt.Errorf("checker %s not found", checkerName)
	}
	var buf strings.Builder
	if err := lintdoc.RenderCheckDocumentation(&buf, info); err != nil {
		return err
	}
	fmt.Println(buf.String())
	return nil
}
