package cmd

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/lintdoc"
	"github.com/VKCOM/noverify/src/linter"
)

func CheckersDocumentation(ctx *AppContext) (int, error) {
	config := ctx.MainConfig.linter.Config()

	fmt.Println("# Checkers")

	countEnabledDefault, countDisabledDefault, countAutofixable := checkersStat(config.Checkers.ListDeclared())
	fmt.Println()
	fmt.Println("## Brief statistics")
	fmt.Printf(`
| Total checks | Checks enabled by default | Disabled checks by default | Autofix checks |
| ------------ | ------------------------- | -------------------------- | -------------- |
| %d           | %d                        | %d                         | %d             |
`, len(config.Checkers.ListDeclared()), countEnabledDefault, countDisabledDefault, countAutofixable)

	fmt.Println("## Table of contents")

	fmt.Println(" - Enabled by default")
	for _, info := range config.Checkers.ListDeclared() {
		if !info.Default {
			continue
		}
		fmt.Printf("   - [`%[1]s` checker](#-%[1]s--checker)\n", info.Name)
	}

	fmt.Println(" - Disabled by default")
	for _, info := range config.Checkers.ListDeclared() {
		if info.Default {
			continue
		}
		fmt.Printf("   - [`%[1]s` checker](#-%[1]s--checker)\n", info.Name)
	}

	fmt.Println("## Enabled")

	for _, info := range config.Checkers.ListDeclared() {
		if !info.Default {
			continue
		}

		err := showMarkdownCheckerInfo(ctx.MainConfig.linter.Config(), info.Name)
		if err != nil {
			return 2, nil
		}
	}

	fmt.Println("<p><br></p>")
	fmt.Println("## Disabled")

	for _, info := range config.Checkers.ListDeclared() {
		if info.Default {
			continue
		}

		err := showMarkdownCheckerInfo(ctx.MainConfig.linter.Config(), info.Name)
		if err != nil {
			return 2, nil
		}
	}

	return 0, nil
}

func checkersStat(checkers []linter.CheckerInfo) (countEnabledDefault int, countDisabledDefault int, countAutofixable int) {
	for _, info := range checkers {
		if info.Default {
			countEnabledDefault++
		} else {
			countDisabledDefault++
		}
		if info.Quickfix {
			countAutofixable++
		}
	}

	return countEnabledDefault, countDisabledDefault, countAutofixable
}

func showMarkdownCheckerInfo(config *linter.Config, checkerName string) error {
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
	if err := lintdoc.RenderMarkdownCheckDocumentation(&buf, info); err != nil {
		return err
	}
	fmt.Println(buf.String())
	return nil
}
