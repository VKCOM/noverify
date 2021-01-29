package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
)

const allNonMaybe = "<all-non-maybe>"

type cmdlineArguments struct {
	version bool

	pprofHost string

	cpuProfile string
	memProfile string

	maxFileSize int

	fullAnalysisFiles string
	indexOnlyFiles    string

	rulesList string

	output         string
	outputJSON     bool
	outputBaseline bool

	baseline string

	misspellList string

	unusedVarPattern string

	allowAll     bool
	allowChecks  string
	allowDisable string

	reportsExclude       string
	reportsExcludeChecks string
	reportsCritical      string

	phpExtensionsArg string

	gitignore bool

	gitPushArg                 string
	gitAuthorsWhitelist        string
	gitWorkTree                string
	gitSkipFetch               bool
	gitDisableCompensateMaster bool
	gitFullDiff                bool
	gitIncludeUntracked        bool
	gitRepo                    string
	gitRef                     string // TODO: remove? It looks unused

	// These two flags are mutated in prepareGitArgs.
	// This is bad, but it's easier for now than to fix this
	// without introducing other issues.
	mutable struct {
		gitCommitFrom string
		gitCommitTo   string
	}

	disableCache bool
}

func DefaultCacheDir() string {
	defaultCacheDir, err := os.UserCacheDir()
	if err != nil {
		defaultCacheDir = ""
	} else {
		defaultCacheDir = filepath.Join(defaultCacheDir, "noverify-cache")
	}
	return defaultCacheDir
}

func bindFlags(config *linter.Config, ruleSets []*rules.Set, args *cmdlineArguments) {
	var enabledByDefault []string
	declaredChecks := config.Checkers.ListDeclared()
	for _, info := range declaredChecks {
		if info.Default {
			enabledByDefault = append(enabledByDefault, info.Name)
		}
	}
	for _, rset := range ruleSets {
		enabledByDefault = append(enabledByDefault, rset.Names...)
	}

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of noverify:\n")
		fmt.Fprintf(out, "  $ noverify [command] [-stubs-dir=/path/to/phpstorm-stubs] [-cache-dir=$TMPDIR/noverify] /project/root\n")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "\tThe cache directory is optional, by default it is already set to $TMPDIR/noverify.")
		fmt.Fprintln(out, "\tThe phpstorm-stubs directory is optional, the built-in ones are used by default.")
		fmt.Fprintln(out)
		GlobalCmds.WriteHelpPage(out)
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Diagnostics (checks):\n")
		for _, info := range declaredChecks {
			extra := " (disabled by default)"
			if info.Default {
				extra = ""
			}
			fmt.Fprintf(out, "  %s%s\n", info.Name, extra)
			fmt.Fprintf(out, "    \t%s\n", info.Comment)
		}
	}

	flag.StringVar(&args.pprofHost, "pprof", "", "HTTP pprof endpoint (e.g. localhost:8080)")

	flag.StringVar(&args.baseline, "baseline", "",
		"Path to a suppress profile created by -output-baseline")
	flag.BoolVar(&config.ConservativeBaseline, "conservative-baseline", false,
		"If enabled, baseline mode will have less false positive, but more false negatives")

	flag.StringVar(&args.reportsCritical, "critical", allNonMaybe,
		"Comma-separated list of check names that are considered critical (all non-maybe checks by default)")

	flag.StringVar(&args.rulesList, "rules", "",
		"Comma-separated list of rules files")

	flag.BoolVar(&config.ApplyQuickFixes, "fix", false,
		"Apply a quickfix where possible (updates source files)")

	flag.BoolVar(&args.gitignore, "gitignore", false,
		"If enabled, noverify tries to use .gitignore files to exclude matched ignored files from the analysis")
	flag.BoolVar(&config.KPHP, "kphp", false,
		"If enabled, treat the code as KPHP")

	flag.StringVar(&args.gitRepo, "git", "", "Path to git repository to analyze")
	flag.StringVar(&args.mutable.gitCommitFrom, "git-commit-from", "", "Analyze changes between commits <git-commit-from> and <git-commit-to>")
	flag.StringVar(&args.mutable.gitCommitTo, "git-commit-to", "", "")
	flag.StringVar(&args.gitRef, "git-ref", "", "Ref (e.g. branch) that is being pushed")
	flag.StringVar(&args.gitPushArg, "git-push-arg", "", "In {pre,post}-receive hooks a whole line from stdin can be passed")
	flag.StringVar(&args.gitAuthorsWhitelist, "git-author-whitelist", "", "Whitelist (comma-separated) for commit authors, if needed")
	flag.StringVar(&args.gitWorkTree, "git-work-tree", "", "Work tree. If specified, local changes will also be examined")
	flag.BoolVar(&args.gitSkipFetch, "git-skip-fetch", false, "Do not fetch ORIGIN_MASTER (use this option if you already fetch to ORIGIN_MASTER before that)")
	flag.BoolVar(&args.gitDisableCompensateMaster, "git-disable-compensate-master", false, "Do not try to compensate for changes in ORIGIN_MASTER after branch point")
	flag.BoolVar(&args.gitFullDiff, "git-full-diff", false, "Compute full diff: analyze all files, not just changed ones")
	flag.BoolVar(&args.gitIncludeUntracked, "git-include-untracked", true, "Include untracked (new, uncommitted files) into analysis")

	flag.StringVar(&args.reportsExclude, "exclude", "", "Exclude regexp for filenames in reports list")
	flag.StringVar(&args.reportsExcludeChecks, "exclude-checks", "", "Comma-separated list of check names to be excluded")
	flag.StringVar(&args.allowDisable, "allow-disable", "", "Regexp for filenames where '@linter disable' is allowed")
	flag.StringVar(&args.allowChecks, "allow-checks", strings.Join(enabledByDefault, ","),
		"Comma-separated list of check names to be enabled")
	flag.BoolVar(&args.allowAll, "allow-all-checks", false,
		"Enables all checks. Has the same effect as passing all check names to the -allow-checks parameter")
	flag.StringVar(&args.misspellList, "misspell-list", "Eng",
		"Comma-separated list of misspelling dicts; predefined sets are Eng, Eng/US and Eng/UK")

	flag.StringVar(&args.phpExtensionsArg, "php-extensions", "php,inc,php5,phtml", "List of PHP extensions to be recognized")

	flag.StringVar(&args.fullAnalysisFiles, "full-analysis-files", "", "Comma-separated list of files to do full analysis")
	flag.StringVar(&args.indexOnlyFiles, "index-only-files", "", "Comma-separated list of files to do indexing")

	flag.StringVar(&args.output, "output", "", "Output reports to a specified file instead of stderr")
	flag.BoolVar(&args.outputJSON, "output-json", false, "Format output as JSON")
	flag.BoolVar(&args.outputBaseline, "output-baseline", false, "Output a suppression profile instead of reports")

	flag.BoolVar(&config.CheckAutoGenerated, `check-auto-generated`, false, "Whether to lint auto-generated PHP file")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug output")
	flag.DurationVar(&config.DebugParseDuration, "debug-parse-duration", 0, "Print files that took longer than the specified time to analyse")
	flag.IntVar(&args.maxFileSize, "max-sum-filesize", 20*1024*1024, "Max total file size to be parsed concurrently in bytes (limits max memory consumption)")
	flag.IntVar(&config.MaxConcurrency, "cores", runtime.NumCPU(), "Max cores")

	flag.StringVar(&config.StubsDir, "stubs-dir", "", "Directory with phpstorm-stubs")
	flag.StringVar(&config.CacheDir, "cache-dir", DefaultCacheDir(), "Directory for linter cache (greatly improves indexing speed)")
	flag.BoolVar(&args.disableCache, "disable-cache", false, "If set, cache is not used and cache-dir is ignored")

	flag.StringVar(&args.unusedVarPattern, "unused-var-regex", `^_$`,
		"Variables that match such regexp are marked as discarded; not reported as unused, but should not be used as values")

	flag.BoolVar(&args.version, "version", false, "Show version info and exit")

	flag.StringVar(&args.cpuProfile, "cpuprofile", "", "Write cpu profile to `file`")
	flag.StringVar(&args.memProfile, "memprofile", "", "Write memory profile to `file`")

	var encodingUnused string
	flag.StringVar(&encodingUnused, "encoding", "", "Deprecated and unused")
}
