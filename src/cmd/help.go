package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/VKCOM/noverify/src/lintdoc"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
)

func Checkers(ctx *AppContext) (int, error) {
	if ctx.MainConfig.LinterConfig == nil {
		ctx.MainConfig.LinterConfig = linter.NewConfig()
	}

	// `checkers`
	if len(ctx.ParsedArgs) == 0 {
		showAllCheckers(ctx)
		return 0, nil
	}

	checkerName := ctx.ParsedArgs[0]

	// `checkers <name>`
	err := showChecker(ctx.MainConfig.LinterConfig, checkerName)
	if err != nil {
		return 1, err
	}

	return 0, err
}

func showAllCheckers(ctx *AppContext) {
	config := ctx.MainConfig.LinterConfig

	ruleSets, err := AddEmbeddedRules(config.Rules, rules.NewParser(), func(r rules.Rule) bool { return true })
	if err != nil {
		panic(err)
	}

	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	fmt.Println("Usage:")
	fmt.Printf("  $ %s check -allow-checks='<list-checks>' /project/root\n", ctx.App.Name)
	fmt.Println()
	fmt.Println("  NOTE: In order to run the linter with only some checks, the -allow-checks")
	fmt.Println("  flag is used which accepts a comma-separated list of checks that are allowed.")
	fmt.Println()
	fmt.Println("  For other possible options run")
	fmt.Printf("     $ %s check help\n", ctx.App.Name)
	fmt.Println()
	fmt.Println("Checkers:")

	w := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', 0)
	for _, info := range config.Checkers.ListDeclared() {
		fmt.Fprintf(w, "  %s\t%s\t\n", info.Name, info.Comment)
	}
	w.Flush()
}

func showChecker(config *linter.Config, checkerName string) error {
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
