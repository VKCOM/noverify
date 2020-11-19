package cmd

import (
	"flag"
	"fmt"

	"github.com/VKCOM/noverify/src/linter"
)

func cmdCheck(cfg *MainConfig) (int, error) {
	if cfg == nil {
		cfg = &MainConfig{}
	}

	ruleSets, err := ParseRules()
	if err != nil {
		return 1, fmt.Errorf("preload rules: %v", err)
	}
	for _, rset := range ruleSets {
		linter.DeclareRules(rset)
	}

	var args CmdlineArguments
	BindFlags(ruleSets, &args)
	flag.Parse()
	if args.DisableCache {
		linter.CacheDir = ""
	}
	if cfg.AfterFlagParse != nil {
		cfg.AfterFlagParse(InitEnvironment{
			RuleSets: ruleSets,
		})
	}

	return MainNoExit(ruleSets, &args, cfg)
}
