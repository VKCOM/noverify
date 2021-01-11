package cmd

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/VKCOM/noverify/src/linter"
)

func Check(cfg *MainConfig) (int, error) {
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
	if args.gitRepo != "" {
		gitDir, err := filepath.Abs(args.gitRepo + "/../")
		if err != nil {
			log.Fatalf("Find git dir: %v", err)
		}
		linter.GitDir = gitDir
	}
	if cfg.AfterFlagParse != nil {
		cfg.AfterFlagParse(InitEnvironment{
			RuleSets: ruleSets,
		})
	}

	return mainNoExit(ruleSets, &args, cfg)
}
