package cmd

import (
	"fmt"
)

func Check(ctx *AppContext) (int, error) {
	ruleSets, err := parseExternalRules(ctx.ParsedFlags.rulesList)
	if err != nil {
		return 1, fmt.Errorf("preload external rules: %v", err)
	}

	for _, rset := range ruleSets {
		ctx.MainConfig.linter.Config().Checkers.DeclareRules(rset)
	}

	ctx.MainConfig.rulesSets = append(ctx.MainConfig.rulesSets, ruleSets...)

	if ctx.ParsedFlags.disableCache {
		ctx.MainConfig.linter.Config().CacheDir = ""
	}

	if ctx.MainConfig.AfterFlagParse != nil {
		ctx.MainConfig.AfterFlagParse(InitEnvironment{
			RuleSets: ctx.MainConfig.rulesSets,
			MetaInfo: ctx.MainConfig.linter.MetaInfo(),
		})
	}

	return mainNoExit(ctx)
}
