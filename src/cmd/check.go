package cmd

import (
	"flag"
	"fmt"

	"github.com/VKCOM/noverify/src/linter"
)

func Check(cfg *MainConfig) (int, error) {
	if cfg == nil {
		cfg = &MainConfig{}
	}

	config := cfg.LinterConfig
	if config == nil {
		config = linter.NewConfig()
	}
	l := linter.NewLinter(config)

	var args cmdlineArguments
	bindFlags(config, &args)
	flag.Parse()

	ruleSets, err := parseRules(args.rulesList)
	if err != nil {
		return 1, fmt.Errorf("preload rules: %v", err)
	}
	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	if args.disableCache {
		config.CacheDir = ""
	}

	if cfg.AfterFlagParse != nil {
		cfg.AfterFlagParse(InitEnvironment{
			RuleSets: ruleSets,
			MetaInfo: l.MetaInfo(),
		})
	}

	return mainNoExit(l, ruleSets, &args, cfg)
}
