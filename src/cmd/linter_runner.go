package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/client9/misspell"
)

type linterRunner struct {
	args *cmdlineArguments

	outputFp io.Writer

	reportsExcludeChecksSet map[string]bool
	reportsIncludeChecksSet map[string]bool
	reportsCriticalSet      map[string]bool

	allowDisableRegex *regexp.Regexp
}

func (l *linterRunner) IsEnabledByFlags(checkName string) bool {
	if !l.reportsIncludeChecksSet[checkName] {
		return false // Not enabled by -allow-checks
	}

	if l.reportsExcludeChecksSet[checkName] {
		return false // Disabled by -exclude-checks
	}

	return true
}

func (l *linterRunner) Init(ruleSets []*rules.Set, args *cmdlineArguments) error {
	l.args = args

	l.outputFp = os.Stderr
	if args.output != "" {
		outputFp, err := os.OpenFile(args.output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return fmt.Errorf("-output=%s: %v", args.output, err)
		}
		l.outputFp = outputFp
	}

	if err := l.compileRegexes(); err != nil {
		return err
	}

	if args.misspellList != "" {
		err := l.loadMisspellDicts(strings.Split(args.misspellList, ","))
		if err != nil {
			return err
		}
	}

	l.initCheckMappings()
	l.initRules(ruleSets)

	return nil
}

func (l *linterRunner) loadMisspellDicts(dicts []string) error {
	linter.TypoFixer = &misspell.Replacer{}

	for _, d := range dicts {
		d = strings.TrimSpace(d)
		switch {
		case d == "Eng":
			linter.TypoFixer.AddRuleList(misspell.DictMain)
		case d == "Eng/US":
			linter.TypoFixer.AddRuleList(misspell.DictAmerican)
		case d == "Eng/UK" || d == "Eng/GB":
			linter.TypoFixer.AddRuleList(misspell.DictBritish)
		default:
			return fmt.Errorf("unsupported %s misspell-list entry", d)
		}
	}

	linter.TypoFixer.Compile()
	return nil
}

func (l *linterRunner) compileRegexes() error {
	if l.args.reportsExclude != "" {
		var err error
		linter.ExcludeRegex, err = regexp.Compile(l.args.reportsExclude)
		if err != nil {
			return fmt.Errorf("incorrect exclude regex: %v", err)
		}
	}

	if l.args.allowDisable != "" {
		allowDisableRegex, err := regexp.Compile(l.args.allowDisable)
		if err != nil {
			return fmt.Errorf("incorrect 'allow disable' regex: %v", err)
		}
		l.allowDisableRegex = allowDisableRegex
	}

	switch l.args.unusedVarPattern {
	case "^_$":
		// Default pattern, only $_ is allowed.
		// Don't change anything.
	case "^_.*$":
		// Leading underscore plus anything after it.
		// Recognize as quite common pattern.
		linter.IsDiscardVar = func(s string) bool {
			return strings.HasPrefix(s, "_")
		}
	default:
		re, err := regexp.Compile(l.args.unusedVarPattern)
		if err != nil {
			return fmt.Errorf("incorrect unused-var-regex regex: %v", err)
		}
		linter.IsDiscardVar = re.MatchString
	}

	return nil
}

func (l *linterRunner) initCheckMappings() {
	stringToSet := func(s string) map[string]bool {
		set := make(map[string]bool)
		for _, name := range strings.Split(s, ",") {
			set[strings.TrimSpace(name)] = true
		}
		return set
	}

	l.reportsExcludeChecksSet = stringToSet(l.args.reportsExcludeChecks)
	l.reportsIncludeChecksSet = stringToSet(l.args.allowChecks)
	if l.args.reportsCritical != allNonMaybe {
		l.reportsCriticalSet = stringToSet(l.args.reportsCritical)
	}
}

func (l *linterRunner) initRules(ruleSets []*rules.Set) error {
	ruleFilter := func(r rules.Rule) bool {
		return l.IsEnabledByFlags(r.Name)
	}

	linter.Rules = rules.NewSet()
	for _, rset := range ruleSets {
		appendRuleSet(rset, ruleFilter)
	}

	return nil
}
