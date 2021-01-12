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
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

type linterRunner struct {
	args *cmdlineArguments

	outputFp io.Writer

	filenameFilter *workspace.FilenameFilter

	reportsExcludeChecksSet map[string]bool
	reportsIncludeChecksSet map[string]bool
	reportsCriticalSet      map[string]bool

	allowDisableRegex *regexp.Regexp

	// gitDir is an absolute path to a directory that contains ".git".
	// Empty string if NoVerify is executed in a non-git mode.
	gitDir string
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

func (l *linterRunner) collectGitIgnoreFiles() error {
	l.filenameFilter = workspace.NewFilenameFilter(linter.ExcludeRegex)

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
			linter.DebugMessage("discovered git top level: %s", dir)
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

	linter.ApplyQuickFixes = l.args.fix
	linter.KPHP = l.args.kphp

	if err := l.compileRegexes(); err != nil {
		return err
	}

	if args.misspellList != "" {
		err := LoadMisspellDicts(strings.Split(args.misspellList, ","))
		if err != nil {
			return err
		}
	}

	if args.gitRepo != "" {
		var err error
		// args.gitRepo contains the path to the .git folder,
		// so we need to get the folder above.
		l.gitDir, err = filepath.Abs(args.gitRepo + "/../")
		if err != nil {
			return fmt.Errorf("find git dir: %v", err)
		}
	}

	l.initCheckMappings()
	if err := l.initRules(ruleSets); err != nil {
		return fmt.Errorf("rules: %v", err)
	}
	if err := l.initBaseline(); err != nil {
		return fmt.Errorf("baseline: %v", err)
	}

	return nil
}

func (l *linterRunner) initBaseline() error {
	linter.ConservativeBaseline = l.args.conservativeBaseline
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
	linter.BaselineProfile = profile
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

func LoadMisspellDicts(dicts []string) error {
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
