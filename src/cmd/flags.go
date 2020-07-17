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

	fix bool

	fullAnalysisFiles string
	indexOnlyFiles    string

	rulesList string

	output         string
	outputJSON     bool
	outputBaseline bool

	baseline             string
	conservativeBaseline bool

	misspellList string

	unusedVarPattern string

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

func bindFlags(ruleSets []*rules.Set, args *cmdlineArguments) {
	var enabledByDefault []string
	declaredChecks := linter.GetDeclaredChecks()
	for _, info := range declaredChecks {
		if info.Default {
			enabledByDefault = append(enabledByDefault, info.Name)
		}
	}
	for _, rset := range ruleSets {
		for _, name := range rset.Names {
			enabledByDefault = append(enabledByDefault, name)
		}
	}

	defaultCacheDir, err := os.UserCacheDir()
	if err != nil {
		defaultCacheDir = ""
	} else {
		defaultCacheDir = filepath.Join(defaultCacheDir, "noverify-cache")
	}

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of noverify:\n")
		fmt.Fprintf(out, "  $ noverify -stubs-dir=/path/to/phpstorm-stubs -cache-dir=/cache/dir /project/root\n")
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
	flag.BoolVar(&args.conservativeBaseline, "conservative-baseline", false,
		"If enabled, baseline mode will have less false positive, but more false negatives")

	flag.StringVar(&args.reportsCritical, "critical", allNonMaybe,
		"Comma-separated list of check names that are considered critical (all non-maybe checks by default)")

	flag.StringVar(&args.rulesList, "rules", "",
		"Comma-separated list of rules files")

	flag.BoolVar(&args.fix, "fix", false,
		"Apply a quickfix where possible (updates source files)")

	flag.BoolVar(&args.gitignore, "gitignore", false,
		"If enabled, noverify tries to use .gitignore files to exclude matched ignored files from the analysis")

	flag.StringVar(&args.gitRepo, "git", "", "Path to git repository to analyze")
	flag.StringVar(&args.mutable.gitCommitFrom, "git-commit-from", "", "Analyze changes between commits <git-commit-from> and <git-commit-to>")
	flag.StringVar(&args.mutable.gitCommitTo, "git-commit-to", "", "")
	flag.StringVar(&args.gitRef, "git-ref", "", "Ref (e.g. branch) that is being pushed")
	flag.StringVar(&args.gitPushArg, "git-push-arg", "", "In {pre,post}-receive hooks a whole line from stdin can be passed")
	flag.StringVar(&args.gitAuthorsWhitelist, "git-author-whitelist", "", "Whitelist (comma-separated) for commit authors, if needed")
	flag.StringVar(&args.gitWorkTree, "git-work-tree", "", "Work tree. If specified, local changes will also be examined.")
	flag.BoolVar(&args.gitSkipFetch, "git-skip-fetch", false, "Do not fetch ORIGIN_MASTER (use this option if you already fetch to ORIGIN_MASTER before that)")
	flag.BoolVar(&args.gitDisableCompensateMaster, "git-disable-compensate-master", false, "Do not try to compensate for changes in ORIGIN_MASTER after branch point")
	flag.BoolVar(&args.gitFullDiff, "git-full-diff", false, "Compute full diff: analyze all files, not just changed ones")
	flag.BoolVar(&args.gitIncludeUntracked, "git-include-untracked", true, "Include untracked (new, uncommitted files) into analysis")

	flag.StringVar(&args.reportsExclude, "exclude", "", "Exclude regexp for filenames in reports list")
	flag.StringVar(&args.reportsExcludeChecks, "exclude-checks", "", "Comma-separated list of check names to be excluded")
	flag.StringVar(&args.allowDisable, "allow-disable", "", "Regexp for filenames where '@linter disable' is allowed")
	flag.StringVar(&args.allowChecks, "allow-checks", strings.Join(enabledByDefault, ","),
		"Comma-separated list of check names to be enabled")
	flag.StringVar(&args.misspellList, "misspell-list", "Eng",
		"Comma-separated list of misspelling dicts; predefined sets are Eng, Eng/US and Eng/UK")

	flag.StringVar(&args.phpExtensionsArg, "php-extensions", "php,inc,php5,phtml,inc", "List of PHP extensions to be recognized")

	flag.StringVar(&args.fullAnalysisFiles, "full-analysis-files", "", "Comma-separated list of files to do full analysis")
	flag.StringVar(&args.indexOnlyFiles, "index-only-files", "", "Comma-separated list of files to do indexing")

	flag.StringVar(&args.output, "output", "", "Output reports to a specified file instead of stderr")
	flag.BoolVar(&args.outputJSON, "output-json", false, "Format output as JSON")
	flag.BoolVar(&args.outputBaseline, "output-baseline", false, "Output a suppression profile instead of reports")

	flag.BoolVar(&linter.CheckAutoGenerated, `check-auto-generated`, false, "whether to lint auto-generated PHP file")
	flag.BoolVar(&linter.Debug, "debug", false, "Enable debug output")
	flag.DurationVar(&linter.DebugParseDuration, "debug-parse-duration", 0, "Print files that took longer than the specified time to analyse")
	flag.IntVar(&linter.MaxFileSize, "max-sum-filesize", 20*1024*1024, "max total file size to be parsed concurrently in bytes (limits max memory consumption)")
	flag.IntVar(&linter.MaxConcurrency, "cores", runtime.NumCPU(), "max cores")
	flag.BoolVar(&linter.LangServer, "lang-server", false, "Run language server for VS Code")

	flag.StringVar(&linter.StubsDir, "stubs-dir", "", "phpstorm-stubs directory")
	flag.StringVar(&linter.CacheDir, "cache-dir", defaultCacheDir, "Directory for linter cache (greatly improves indexing speed)")
	flag.BoolVar(&args.disableCache, "disable-cache", false, "If set, cache is not used and cache-dir is ignored")

	flag.StringVar(&args.unusedVarPattern, "unused-var-regex", `^_$`,
		"Variables that match such regexp are marked as discarded; not reported as unused, but should not be used as values")

	flag.BoolVar(&args.version, "version", false, "Show version info and exit")

	flag.StringVar(&args.cpuProfile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&args.memProfile, "memprofile", "", "write memory profile to `file`")

	var encodingUnused string
	flag.StringVar(&encodingUnused, "encoding", "", "deprecated and unused")
}
