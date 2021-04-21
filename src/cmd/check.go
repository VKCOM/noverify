package cmd

import (
	"fmt"

	"github.com/VKCOM/noverify/src/linter"
)

func Check(ctx *AppContext) (int, error) {
	if ctx.MainConfig == nil {
		ctx.MainConfig = &MainConfig{}
	}

	config := ctx.MainConfig.LinterConfig
	if config == nil {
		config = linter.NewConfig()
	}
	l := linter.NewLinter(config)

	ruleSets, err := parseRules(ctx.ParsedFlags.rulesList)
	if err != nil {
		return 1, fmt.Errorf("preload rules: %v", err)
	}
	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	if ctx.ParsedFlags.disableCache {
		config.CacheDir = ""
	}

	if ctx.MainConfig.AfterFlagParse != nil {
		ctx.MainConfig.AfterFlagParse(InitEnvironment{
			RuleSets: ruleSets,
			MetaInfo: l.MetaInfo(),
		})
	}

	return mainNoExit(l, ruleSets, ctx)
}
