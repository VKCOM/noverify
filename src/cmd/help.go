package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/lintdoc"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
)

func Help(*MainConfig) (int, error) {
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Printf("Usage of noverify:\n")
		fmt.Printf("  $ noverify [command] -stubs-dir=/path/to/phpstorm-stubs -cache-dir=/cache/dir /project/root\n\n")
		GlobalCmds.PrintHelpPage()
		return 0, nil
	}

	mainSubject := args[0]
	switch mainSubject {
	case "checkers":
		config := declareRules()
		var subSubject string
		if len(args) > 1 {
			subSubject = args[1]
		}
		if subSubject == "" {
			helpAllCheckers(config)
			return 0, nil
		}
		err := helpChecker(config, subSubject)
		if err != nil {
			return 1, err
		}
		return 0, err
	default:
		return 1, fmt.Errorf("unknown subject: %s", mainSubject)
	}
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

func helpAllCheckers(config *linter.Config) {
	for _, info := range config.Checkers.ListDeclared() {
		fmt.Printf("%s: %s\n", info.Name, info.Comment)
	}
}

func helpChecker(config *linter.Config, checkerName string) error {
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
