package cmd

import (
	"flag"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/VKCOM/noverify/src/linter"
)

type severityLevelValue struct {
	value linter.SeverityLevel
}

func (f *severityLevelValue) String() string {
	return linter.GetSeverityLevelName(f.value)

}

func (f *severityLevelValue) Set(s string) error {
	level, err := linter.ParseSeverityLevel(s)
	if err != nil {
		return err
	}
	f.value = level
	return nil
}

var (
	gitRepo    string
	isGitLocal bool

	pprofHost string

	gitCommitFrom       string
	gitCommitTo         string
	gitRef              string
	gitPushArg          string
	gitAuthorsWhitelist string
	gitWorkTree         string
	gitSkipFetch        bool
	gitFullDiff         bool

	phpExtensionsArg string

	reportsMinSeverityLevel severityLevelValue = severityLevelValue{linter.LevelSyntax}
	reportsExclude          string
	reportsExcludeRegex     *regexp.Regexp
	reportsExcludeChecks    string
	reportsExcludeChecksSet map[string]bool
	reportsIncludeChecksSet map[string]bool

	allowChecks       string
	allowDisable      string
	allowDisableRegex *regexp.Regexp

	fullAnalysisFiles string

	output string

	version bool

	cpuProfile string
	memProfile string
)

func parseFlags() {
	var enabledByDefault []string
	declaredChecks := linter.GetDeclaredChecks()
	for _, info := range declaredChecks {
		if info.Default {
			enabledByDefault = append(enabledByDefault, info.Name)
		}
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

	flag.StringVar(&pprofHost, "pprof", "", "HTTP pprof endpoint (e.g. localhost:8080)")

	flag.StringVar(&gitRepo, "git", "", "Path to git repository to analyze")
	flag.BoolVar(&isGitLocal, "git-local", false, "Analyze local changes in git (everything not yet pushed)")
	flag.StringVar(&gitCommitFrom, "git-commit-from", "", "Analyze changes between commits <git-commit-from> and <git-commit-to>")
	flag.StringVar(&gitCommitTo, "git-commit-to", "", "")
	flag.StringVar(&gitRef, "git-ref", "", "Ref (e.g. branch) that is being pushed")
	flag.StringVar(&gitPushArg, "git-push-arg", "", "In {pre,post}-receive hooks a whole line from stdin can be passed")
	flag.StringVar(&gitAuthorsWhitelist, "git-author-whitelist", "", "Whitelist (comma-separated) for commit authors, if needed")
	flag.StringVar(&gitWorkTree, "git-work-tree", "", "Work tree. If specified, local changes will also be examined.")
	flag.BoolVar(&gitSkipFetch, "git-skip-fetch", false, "Do not fetch ORIGIN_MASTER (use this option if you already fetch to ORIGIN_MASTER before that)")
	flag.BoolVar(&gitFullDiff, "git-full-diff", false, "Compute full diff: analyze all files, not just changed ones")

	flag.Var(&reportsMinSeverityLevel, "min-level", "Set name of the minimal severity level to report")
	flag.StringVar(&reportsExclude, "exclude", "", "Exclude regexp for filenames in reports list")
	flag.StringVar(&reportsExcludeChecks, "exclude-checks", "", "Comma-separated list of check names to be excluded")
	flag.StringVar(&allowDisable, "allow-disable", "", "Regexp for filenames where '@linter disable' is allowed")
	flag.StringVar(&allowChecks, "allow-checks", strings.Join(enabledByDefault, ","),
		"Comma-separated list of check names to be enabled")

	flag.StringVar(&phpExtensionsArg, "php-extensions", "php,inc,php5,phtml,inc", "List of PHP extensions to be recognized")

	flag.StringVar(&fullAnalysisFiles, "full-analysis-files", "", "Comma-separated list of files to do full analysis")

	flag.StringVar(&output, "output", "", "Output reports to a specified file instead of stderr")

	flag.BoolVar(&linter.Debug, "debug", false, "Enable debug output")
	flag.IntVar(&linter.MaxFileSize, "max-sum-filesize", 20*1024*1024, "max total file size to be parsed concurrently in bytes (limits max memory consumption)")
	flag.IntVar(&linter.MaxConcurrency, "cores", runtime.NumCPU(), "max cores")
	flag.BoolVar(&linter.LangServer, "lang-server", false, "Run language server for VS Code")
	flag.StringVar(&linter.DefaultEncoding, "encoding", "UTF-8", "Default encoding. Only UTF-8 and windows-1251 are supported")
	flag.StringVar(&linter.StubsDir, "stubs-dir", "/path/to/phpstorm-stubs", "phpstorm-stubs directory")
	flag.StringVar(&linter.CacheDir, "cache-dir", "", "Directory for linter cache (greatly improves indexing speed)")

	flag.BoolVar(&version, "version", false, "Show version info and exit")

	flag.StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&memProfile, "memprofile", "", "write memory profile to `file`")

	flag.Parse()
}
