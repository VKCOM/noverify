package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/client9/misspell"

	"github.com/VKCOM/noverify/src/baseline"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

type linterRunner struct {
	args *cmdlineArguments

	linter *linter.Linter

	config *linter.Config

	outputFp io.Writer

	filenameFilter *workspace.FilenameFilter

	reportsExcludeChecksSet map[string]bool
	reportsIncludeChecksSet map[string]bool
	reportsCriticalSet      map[string]bool
}

func (l *linterRunner) IsEnabledByFlags(checkName string) bool {
	if !l.args.allowAll && !l.reportsIncludeChecksSet[checkName] {
		return false // Not enabled by -allow-checks
	}

	if l.reportsExcludeChecksSet[checkName] {
		return false // Disabled by -exclude-checks
	}

	return true
}

func (l *linterRunner) IsCriticalReport(r *linter.Report) bool {
	if len(l.reportsCriticalSet) != 0 {
		return l.reportsCriticalSet[r.CheckName]
	}
	return r.IsCritical()
}

func (l *linterRunner) IsEnabledReport(r *linter.Report) bool {
	if !l.IsEnabledByFlags(r.CheckName) {
		return false
	}

	if l.config.ExcludeRegex == nil {
		return true
	}

	// Disabled by a file comment.
	return !l.config.ExcludeRegex.MatchString(r.Filename)
}

func (l *linterRunner) collectGitIgnoreFiles() error {
	l.filenameFilter = workspace.NewFilenameFilter(l.config.ExcludeRegex)

	if !l.args.gitignore {
		return nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %v", err)
	}

	l.filenameFilter.EnableGitignore()

	// Walk wd up until we find .git (good) or FS root (bad).
	// We collect all gitignore files along the way.
	dir := workingDir
	for {
		m, err := workspace.ParseGitignoreFromDir(dir)
		if err != nil {
			return fmt.Errorf("read .gitignore: %v", err)
		}
		if m != nil {
			l.filenameFilter.InitialGitignorePush(dir, m)
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			lintdebug.Send("discovered git top level: %s", dir)
			break
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return fmt.Errorf("not a git repository (don't use -gitignore for non-git projects)")
		}
		dir = parentDir
	}

	return nil
}

func (l *linterRunner) Init(ruleSets []*rules.Set, args *cmdlineArguments) error {
	l.args = args

	if err := l.collectGitIgnoreFiles(); err != nil {
		return fmt.Errorf("collect gitignore files: %v", err)
	}

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

	l.config.PhpExtensions = strings.Split(args.phpExtensionsArg, ",")

	l.config.ComputeBaselineHashes = l.args.baseline != "" || l.args.outputBaseline

	if args.misspellList != "" {
		err := LoadMisspellDicts(l.config, strings.Split(args.misspellList, ","))
		if err != nil {
			return err
		}
	}

	l.initCheckMappings(ruleSets)
	if err := l.initRules(ruleSets); err != nil {
		return fmt.Errorf("rules: %v", err)
	}
	if err := l.initBaseline(); err != nil {
		return fmt.Errorf("baseline: %v", err)
	}

	return nil
}

func (l *linterRunner) initBaseline() error {
	if l.args.baseline == "" {
		return nil
	}

	f, err := os.Open(l.args.baseline)
	if err != nil {
		return err
	}
	defer f.Close()
	profile, _, err := baseline.ReadProfile(f)
	if err != nil {
		return err
	}
	l.config.BaselineProfile = profile
	return nil
}

func (l *linterRunner) compileRegexes() error {
	if l.args.reportsExclude != "" {
		var err error
		l.config.ExcludeRegex, err = regexp.Compile(l.args.reportsExclude)
		if err != nil {
			return fmt.Errorf("incorrect exclude regex: %v", err)
		}
	}

	if l.args.allowDisable != "" {
		allowDisableRegex, err := regexp.Compile(l.args.allowDisable)
		if err != nil {
			return fmt.Errorf("incorrect 'allow disable' regex: %v", err)
		}
		l.config.AllowDisable = allowDisableRegex
	}

	switch l.args.unusedVarPattern {
	case "^_$":
		// Default pattern, only $_ is allowed.
		// Don't change anything.
	case "^_.*$":
		// Leading underscore plus anything after it.
		// Recognize as quite common pattern.
		l.config.IsDiscardVar = func(s string) bool {
			return strings.HasPrefix(s, "_")
		}
	default:
		re, err := regexp.Compile(l.args.unusedVarPattern)
		if err != nil {
			return fmt.Errorf("incorrect unused-var-regex regex: %v", err)
		}
		l.config.IsDiscardVar = re.MatchString
	}

	return nil
}

func (l *linterRunner) initCheckMappings(ruleSets []*rules.Set) {
	stringToSet := func(s string) map[string]bool {
		set := make(map[string]bool)
		for _, name := range strings.Split(s, ",") {
			set[strings.TrimSpace(name)] = true
		}
		return set
	}

	l.reportsExcludeChecksSet = stringToSet(l.args.reportsExcludeChecks)

	if l.args.allowChecks == allChecks {
		set := make(map[string]bool)

		declaredChecks := l.config.Checkers.ListDeclared()
		for _, info := range declaredChecks {
			if info.Default {
				set[info.Name] = true
			}
		}
		for _, ruleSet := range ruleSets {
			for _, name := range ruleSet.Names {
				set[name] = true
			}
		}

		l.reportsIncludeChecksSet = set
	} else {
		l.reportsIncludeChecksSet = stringToSet(l.args.allowChecks)
	}

	if l.args.reportsCritical != allNonNoticeChecks {
		l.reportsCriticalSet = stringToSet(l.args.reportsCritical)
	}
}

func (l *linterRunner) initRules(ruleSets []*rules.Set) error {
	ruleFilter := func(r rules.Rule) bool {
		return l.IsEnabledByFlags(r.Name)
	}

	for _, rset := range ruleSets {
		appendRuleSet(l.config.Rules, rset, ruleFilter)
	}

	return nil
}

func LoadMisspellDicts(config *linter.Config, dicts []string) error {
	config.TypoFixer = &misspell.Replacer{}

	for _, d := range dicts {
		d = strings.TrimSpace(d)
		switch {
		case d == "Eng":
			config.TypoFixer.AddRuleList(misspell.DictMain)
		case d == "Eng/US":
			config.TypoFixer.AddRuleList(misspell.DictAmerican)
		case d == "Eng/UK" || d == "Eng/GB":
			config.TypoFixer.AddRuleList(misspell.DictBritish)
		default:
			return fmt.Errorf("unsupported %s misspell-list entry", d)
		}
	}

	config.TypoFixer.Compile()
	return nil
}
