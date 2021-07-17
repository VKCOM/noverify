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

type LinterRunner struct {
	flags *ParsedFlags

	linter         *linter.Linter
	config         *linter.Config
	checkersFilter *linter.CheckersFilter

	outputFp io.Writer

	filenameFilter *workspace.FilenameFilter
}

func (l *LinterRunner) collectGitIgnoreFiles() error {
	l.filenameFilter = workspace.NewFilenameFilter(l.config.ExcludeRegex)

	if !l.flags.Gitignore {
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

func (l *LinterRunner) Init(ruleSets []*rules.Set, flags *ParsedFlags) error {
	l.flags = flags

	if err := l.collectGitIgnoreFiles(); err != nil {
		return fmt.Errorf("collect gitignore files: %v", err)
	}

	l.outputFp = os.Stderr
	if flags.Output != "" {
		outputFp, err := os.OpenFile(flags.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return fmt.Errorf("-output=%s: %v", flags.Output, err)
		}
		l.outputFp = outputFp
	}

	if err := l.compileRegexes(); err != nil {
		return err
	}

	l.config.PhpExtensions = strings.Split(flags.PhpExtensionsArg, ",")

	l.config.ComputeBaselineHashes = l.flags.Baseline != "" || l.flags.OutputBaseline

	if flags.MisspellList != "" {
		err := LoadMisspellDicts(l.config, strings.Split(flags.MisspellList, ","))
		if err != nil {
			return err
		}
	}

	l.addVendorFolderToIndex(flags)

	l.checkersFilter = l.initCheckMappings(ruleSets)

	if err := l.initRules(ruleSets); err != nil {
		return fmt.Errorf("rules: %v", err)
	}
	if err := l.initBaseline(); err != nil {
		return fmt.Errorf("baseline: %v", err)
	}

	l.linter.UseCheckersFilter(l.checkersFilter)

	return nil
}

func (l *LinterRunner) addVendorFolderToIndex(flags *ParsedFlags) {
	if flags.IgnoreVendor {
		return
	}

	alreadyContainsVendor := false
	parts := strings.Split(flags.IndexOnlyFiles, ",")
	for _, part := range parts {
		part = strings.TrimLeft(filepath.ToSlash(part), " ./")
		if part == "vendor" || strings.HasSuffix(part, "/vendor") {
			alreadyContainsVendor = true
			break
		}
	}
	if alreadyContainsVendor {
		return
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return
	}

	vendorPath := filepath.Join(workingDir, "vendor")
	_, err = os.Stat(vendorPath)
	if os.IsNotExist(err) {
		// If such a folder does not exist, then nothing needs to be done.
		return
	}

	if flags.IndexOnlyFiles == "" {
		flags.IndexOnlyFiles = "./vendor"
	} else {
		flags.IndexOnlyFiles += ",./vendor"
	}
}

func (l *LinterRunner) initBaseline() error {
	if l.flags.Baseline == "" {
		return nil
	}

	f, err := os.Open(l.flags.Baseline)
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

func (l *LinterRunner) compileRegexes() error {
	if l.flags.ReportsExclude != "" {
		var err error
		l.config.ExcludeRegex, err = regexp.Compile(l.flags.ReportsExclude)
		if err != nil {
			return fmt.Errorf("incorrect exclude regex: %v", err)
		}
	}

	if l.flags.AllowDisable != "" {
		allowDisableRegex, err := regexp.Compile(l.flags.AllowDisable)
		if err != nil {
			return fmt.Errorf("incorrect 'allow disable' regex: %v", err)
		}
		l.config.AllowDisable = allowDisableRegex
	}

	switch l.flags.UnusedVarPattern {
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
		re, err := regexp.Compile(l.flags.UnusedVarPattern)
		if err != nil {
			return fmt.Errorf("incorrect unused-var-regex regex: %v", err)
		}
		l.config.IsDiscardVar = re.MatchString
	}

	return nil
}

func (l *LinterRunner) initCheckMappings(ruleSets []*rules.Set) *linter.CheckersFilter {
	stringToSet := func(s string) map[string]bool {
		set := make(map[string]bool)
		for _, name := range strings.Split(s, ",") {
			set[strings.TrimSpace(name)] = true
		}
		return set
	}

	l.checkersFilter.All = l.config.Checkers.ListDeclared()
	l.checkersFilter.EnableAll = l.flags.AllowAll
	l.checkersFilter.ExcludeFileRegexp = l.config.ExcludeRegex

	l.checkersFilter.Excluded = stringToSet(l.flags.ReportsExcludeChecks)

	if l.flags.AllowChecks == AllChecks {
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

		l.checkersFilter.Allowed = set
	} else {
		l.checkersFilter.Allowed = stringToSet(l.flags.AllowChecks)
	}

	if l.flags.ReportsCritical != AllNonNoticeChecks {
		l.checkersFilter.Critical = stringToSet(l.flags.ReportsCritical)
	}

	return l.checkersFilter
}

func (l *LinterRunner) initRules(ruleSets []*rules.Set) error {
	ruleFilter := func(r rules.Rule) bool {
		return l.checkersFilter.IsEnabledCheck(r.Name)
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
