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

	ruleSets, err := parseRules()
	if err != nil {
		return 1, fmt.Errorf("preload rules: %v", err)
	}
	for _, rset := range ruleSets {
		linter.DeclareRules(rset)
	}

	var args cmdlineArguments
	bindFlags(ruleSets, &args)
	flag.Parse()
	if args.disableCache {
		linter.CacheDir = ""
	}
	if cfg.AfterFlagParse != nil {
		cfg.AfterFlagParse(InitEnvironment{
			RuleSets: ruleSets,
		})
	}

	return mainNoExit(ruleSets, &args, cfg)
}
