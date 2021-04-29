package cmd

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"

	"github.com/VKCOM/noverify/src/linter"
)

const allNonNoticeChecks = "<all-non-notice>"
const allChecks = "<all>"

type ParsedFlags struct {
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

func RegisterCheckFlags(config *linter.Config, parsedFlags *ParsedFlags) *flag.FlagSet {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)

	fs.StringVar(&parsedFlags.pprofHost, "pprof", "", "HTTP pprof endpoint (e.g. localhost:8080)")

	fs.StringVar(&parsedFlags.baseline, "baseline", "",
		"Path to a suppress profile created by -output-baseline")
	fs.BoolVar(&config.ConservativeBaseline, "conservative-baseline", false,
		"If enabled, baseline mode will have less false positive, but more false negatives")

	fs.StringVar(&parsedFlags.reportsCritical, "critical", allNonNoticeChecks,
		"Comma-separated list of check names that are considered critical (all non-notice checks by default)")

	fs.StringVar(&parsedFlags.rulesList, "rules", "",
		"Comma-separated list of rules files")

	fs.BoolVar(&config.ApplyQuickFixes, "fix", false,
		"Apply a quickfix where possible (updates source files)")

	fs.BoolVar(&parsedFlags.gitignore, "gitignore", false,
		"If enabled, noverify tries to use .gitignore files to exclude matched ignored files from the analysis")
	fs.BoolVar(&config.KPHP, "kphp", false,
		"If enabled, treat the code as KPHP")

	fs.StringVar(&parsedFlags.gitRepo, "git", "", "Path to git repository to analyze")
	fs.StringVar(&parsedFlags.mutable.gitCommitFrom, "git-commit-from", "", "Analyze changes between commits <git-commit-from> and <git-commit-to>")
	fs.StringVar(&parsedFlags.mutable.gitCommitTo, "git-commit-to", "", "Analyze changes between commits <git-commit-from> and <git-commit-to>")
	fs.StringVar(&parsedFlags.gitRef, "git-ref", "", "Ref (e.g. branch) that is being pushed")
	fs.StringVar(&parsedFlags.gitPushArg, "git-push-arg", "", "In {pre,post}-receive hooks a whole line from stdin can be passed")
	fs.StringVar(&parsedFlags.gitAuthorsWhitelist, "git-author-whitelist", "", "Whitelist (comma-separated) for commit authors, if needed")
	fs.StringVar(&parsedFlags.gitWorkTree, "git-work-tree", "", "Work tree. If specified, local changes will also be examined")
	fs.BoolVar(&parsedFlags.gitSkipFetch, "git-skip-fetch", false, "Do not fetch ORIGIN_MASTER (use this option if you already fetch to ORIGIN_MASTER before that)")
	fs.BoolVar(&parsedFlags.gitDisableCompensateMaster, "git-disable-compensate-master", false, "Do not try to compensate for changes in ORIGIN_MASTER after branch point")
	fs.BoolVar(&parsedFlags.gitFullDiff, "git-full-diff", false, "Compute full diff: analyze all files, not just changed ones")
	fs.BoolVar(&parsedFlags.gitIncludeUntracked, "git-include-untracked", true, "Include untracked (new, uncommitted files) into analysis")

	fs.StringVar(&parsedFlags.reportsExclude, "exclude", "", "Exclude regexp for filenames in reports list")
	fs.StringVar(&parsedFlags.reportsExcludeChecks, "exclude-checks", "", "Comma-separated list of check names to be excluded")
	fs.StringVar(&parsedFlags.allowDisable, "allow-disable", "", "Regexp for filenames where '@linter disable' is allowed")
	fs.StringVar(&parsedFlags.allowChecks, "allow-checks", allChecks,
		"Comma-separated list of check names to be enabled")
	fs.BoolVar(&parsedFlags.allowAll, "allow-all-checks", false,
		"Enables all checks. Has the same effect as passing '<all>' to the -allow-checks parameter")
	fs.StringVar(&parsedFlags.misspellList, "misspell-list", "Eng",
		"Comma-separated list of misspelling dicts; predefined sets are Eng, Eng/US and Eng/UK")

	fs.StringVar(&parsedFlags.phpExtensionsArg, "php-extensions", "php,inc,php5,phtml", "List of PHP extensions to be recognized")

	fs.StringVar(&parsedFlags.fullAnalysisFiles, "full-analysis-files", "", "Comma-separated list of files to do full analysis")
	fs.StringVar(&parsedFlags.indexOnlyFiles, "index-only-files", "", "Comma-separated list of files to do indexing")

	fs.StringVar(&parsedFlags.output, "output", "", "Output reports to a specified file instead of stderr")
	fs.BoolVar(&parsedFlags.outputJSON, "output-json", false, "Format output as JSON")
	fs.BoolVar(&parsedFlags.outputBaseline, "output-baseline", false, "Output a suppression profile instead of reports")

	fs.BoolVar(&config.CheckAutoGenerated, `check-auto-generated`, false, "Whether to lint auto-generated PHP file")
	fs.BoolVar(&config.Debug, "debug", false, "Enable debug output")
	fs.DurationVar(&config.DebugParseDuration, "debug-parse-duration", 0, "Print files that took longer than the specified time to analyse")
	fs.IntVar(&parsedFlags.maxFileSize, "max-sum-filesize", 20*1024*1024, "Max total file size to be parsed concurrently in bytes (limits max memory consumption)")
	fs.IntVar(&config.MaxConcurrency, "cores", runtime.NumCPU(), "Max cores")

	fs.StringVar(&config.StubsDir, "stubs-dir", "", "Directory with phpstorm-stubs")
	fs.StringVar(&config.CacheDir, "cache-dir", DefaultCacheDir(), "Directory for linter cache (greatly improves indexing speed)")
	fs.BoolVar(&parsedFlags.disableCache, "disable-cache", false, "If set, cache is not used and cache-dir is ignored")
	fs.BoolVar(&config.IgnoreTriggerError, "ignore-trigger-error", false, "If set, trigger_error control flow will be ignored")

	fs.StringVar(&parsedFlags.unusedVarPattern, "unused-var-regex", `^_$`,
		"Variables that match such regexp are marked as discarded; not reported as unused, but should not be used as values")
	fs.BoolVar(&parsedFlags.version, "version", false, "Show version info and exit")

	fs.StringVar(&parsedFlags.cpuProfile, "cpuprofile", "", "Write cpu profile to `file`")
	fs.StringVar(&parsedFlags.memProfile, "memprofile", "", "Write memory profile to `file`")

	var encodingUnused string
	fs.StringVar(&encodingUnused, "encoding", "", "Deprecated and unused")

	return fs
}
