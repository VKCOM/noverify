package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

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
	fmt.Println("Usage:")
	fmt.Println("  $ noverify check -allow-checks='<list-checks>' /project/root")
	fmt.Println()
	fmt.Println("  NOTE: In order to run the linter with only some checks, the -allow-checks")
	fmt.Println("  flag is used which accepts a comma-separated list of checks that are allowed.")
	fmt.Println()
	fmt.Println("  For other possible options run")
	fmt.Println("     $ noverify check -help")
	fmt.Println()
	fmt.Println("Checkers:")

	w := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', 0)
	for _, info := range config.Checkers.ListDeclared() {
		fmt.Fprintf(w, "  %s\t%s\t\n", info.Name, info.Comment)
	}
	w.Flush()
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
