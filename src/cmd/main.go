package cmd

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof" // it is ok for actually main package
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/langsrv"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
)

// Build* заполняются при сборке go build -ldflags
var (
	BuildTime    string
	BuildOSUname string
	BuildCommit  string
)

var (
	outputFp io.Writer = os.Stderr

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

	reportsExclude          string
	reportsExcludeRegex     *regexp.Regexp
	reportsExcludeChecks    string
	reportsExcludeChecksSet map[string]bool
	reportsIncludeChecksSet map[string]bool

	allowChecks       string
	allowDisable      string
	allowDisableRegex *regexp.Regexp

	unusedVarPattern string

	fullAnalysisFiles string

	output string

	version bool

	cpuProfile string
	memProfile string
)

func bindFlags() {
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

	flag.StringVar(&unusedVarPattern, "unused-var-regex", `^_$`,
		"Variables that match such regexp are marked as discarded; not reported as unused, but should not be used as values")

	flag.BoolVar(&version, "version", false, "Show version info and exit")

	flag.StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&memProfile, "memprofile", "", "write memory profile to `file`")
}

func isEnabled(r *linter.Report) bool {
	if !reportsIncludeChecksSet[r.CheckName()] {
		return false // Not enabled by -allow-checks
	}

	if reportsExcludeChecksSet[r.CheckName()] {
		return false // Disabled by -exclude-checks
	}

	if reportsExcludeRegex == nil {
		return true
	}

	// Disabled by a file comment.
	return !reportsExcludeRegex.MatchString(r.GetFilename())
}

// canBeDisabled returns whether or not '@linter disable' can be used for the specified file
func canBeDisabled(filename string) bool {
	if allowDisableRegex == nil {
		return false
	}

	return allowDisableRegex.MatchString(filename)
}

// Main is the actual main function to be run. It is separate from linter so that you can insert your own hooks
// before running main().
//
// Optionally, non-nil config can be passed to customize function behavior.
func Main(cfg *MainConfig) {
	if cfg == nil {
		cfg = &MainConfig{}
	}

	bindFlags()
	flag.Parse()
	if cfg.AfterFlagParse != nil {
		cfg.AfterFlagParse()
	}

	status, err := mainNoExit()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(status)
}

// mainNoExit implements main, but instead of doing log.Fatal or os.Exit it
// returns error or non-zero integer status code to be passed to os.Exit by the caller.
// Note that if error is not nil, integer code will be discarded, so it can be 0.
//
// We don't want os.Exit to be inserted randomly to avoid defer cancellation.
func mainNoExit() (int, error) {
	if version {
		fmt.Printf("PHP Linter\nBuilt on %s\nOS %s\nCommit %s\n", BuildTime, BuildOSUname, BuildCommit)
		return 0, nil
	}

	if pprofHost != "" {
		go http.ListenAndServe(pprofHost, nil)
	}

	// Since this function is expected to be exit-free, it's OK
	// to defer calls here to make required flushes/cleanup.
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return 0, fmt.Errorf("Could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return 0, fmt.Errorf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if memProfile != "" {
		defer func() {
			f, err := os.Create(memProfile)
			if err != nil {
				log.Printf("could not create memory profile: %v", err)
				return
			}
			defer f.Close()
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Printf("could not write memory profile: %v", err)
			}
		}()
	}

	if err := setDiscardVarPredicate(); err != nil {
		return 0, fmt.Errorf("compile unused-var-regex: %v", err)
	}

	linter.PHPExtensions = strings.Split(phpExtensionsArg, ",")
	if err := compileRegexes(); err != nil {
		return 0, err
	}

	buildCheckMappings()

	lintdebug.Register(func(msg string) { linter.DebugMessage("%s", msg) })
	go linter.MemoryLimiterThread()

	if linter.LangServer {
		langsrv.RegisterDebug()
		langsrv.Start()
		return 0, nil
	}

	if output != "" {
		var err error
		outputFp, err = os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return 0, fmt.Errorf("Could not open output file: %v", err)
		}
	}

	log.Printf("Started")
	linter.InitStubs()

	if gitRepo != "" {
		return gitMain()
	}

	linter.AnalysisFiles = flag.Args()

	log.Printf("Indexing %+v", flag.Args())
	linter.ParseFilenames(linter.ReadFilenames(flag.Args(), nil))
	meta.SetIndexingComplete(true)
	log.Printf("Linting")

	filenames := flag.Args()
	if fullAnalysisFiles != "" {
		filenames = strings.Split(fullAnalysisFiles, ",")
	}

	reports := linter.ParseFilenames(linter.ReadFilenames(filenames, reportsExcludeRegex))
	criticalReports := analyzeReports(reports)

	if criticalReports > 0 {
		log.Printf("Found %d critical reports", criticalReports)
		return 2, nil
	}
	return 0, nil
}

func compileRegexes() error {
	var err error

	if reportsExclude != "" {
		reportsExcludeRegex, err = regexp.Compile(reportsExclude)
		if err != nil {
			return fmt.Errorf("Incorrect exclude regex: %v", err)
		}
	}

	if allowDisable != "" {
		allowDisableRegex, err = regexp.Compile(allowDisable)
		if err != nil {
			return fmt.Errorf("Incorrect 'allow disable' regex: %v", err)
		}
	}

	return nil
}

func buildCheckMappings() {
	reportsExcludeChecksSet = make(map[string]bool)
	for _, name := range strings.Split(reportsExcludeChecks, ",") {
		reportsExcludeChecksSet[strings.TrimSpace(name)] = true
	}
	reportsIncludeChecksSet = make(map[string]bool)
	for _, name := range strings.Split(allowChecks, ",") {
		reportsIncludeChecksSet[strings.TrimSpace(name)] = true
	}
}

// Not the best name, and not the best function signature.
// Refactor this function whenever you get the idea how to separate logic better.
func gitRepoComputeReportsFromCommits(logArgs, diffArgs []string) (oldReports, reports []*linter.Report, changes []git.Change, changeLog []git.Commit, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	start := time.Now()
	changeLog, err := git.Log(gitRepo, logArgs)
	if err != nil {
		log.Panicf("Could not get commits in range %+v: %s", logArgs, err.Error())
	}

	if shouldRun := analyzeGitAuthorsWhiteList(changeLog); !shouldRun {
		return nil, nil, nil, nil, false
	}

	changes, err = git.Diff(gitRepo, "", diffArgs)
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if gitFullDiff {
		meta.ResetInfo()
		linter.InitStubs()

		start = time.Now()
		linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitFrom, nil))
		log.Printf("Indexed old commit in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		oldReports = linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitFrom, reportsExcludeRegex))
		log.Printf("Parsed old commit for %s (%d reports)", time.Since(start), len(oldReports))

		meta.ResetInfo()
		linter.InitStubs()

		start = time.Now()
		linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitTo, nil))
		log.Printf("Indexed new commit in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		reports = linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitTo, reportsExcludeRegex))
		log.Printf("Parsed new commit in %s (%d reports)", time.Since(start), len(reports))
	} else {
		start = time.Now()
		linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitTo, nil))
		log.Printf("Indexing complete in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		oldReports = linter.ParseFilenames(linter.ReadOldFilesFromGit(gitRepo, gitCommitFrom, changes))
		log.Printf("Parsed old files versions for %s", time.Since(start))

		start = time.Now()
		meta.SetIndexingComplete(false)
		linter.ParseFilenames(linter.ReadFilesFromGitWithChanges(gitRepo, gitCommitTo, changes))
		meta.SetIndexingComplete(true)
		log.Printf("Indexed files versions for %s", time.Since(start))

		start = time.Now()
		reports = linter.ParseFilenames(linter.ReadFilesFromGitWithChanges(gitRepo, gitCommitTo, changes))
		log.Printf("Parsed new file versions in %s", time.Since(start))
	}

	return oldReports, reports, changes, changeLog, true
}

func gitRepoComputeReportsFromLocalChanges() (oldReports, reports []*linter.Report, changes []git.Change, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	if gitWorkTree == "" {
		return nil, nil, nil, false
	}

	// compute changes for working copy (staged + unstaged changes combined starting with the commit being pushed)
	changes, err := git.Diff(gitRepo, gitWorkTree, []string{gitCommitFrom})
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if len(changes) == 0 {
		return nil, nil, nil, false
	}

	log.Printf("You have changes in your work tree, showing diff between %s and work tree", gitCommitFrom)

	start := time.Now()
	linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitFrom, nil))
	log.Printf("Indexing complete in %s", time.Since(start))

	meta.SetIndexingComplete(true)

	start = time.Now()
	oldReports = linter.ParseFilenames(linter.ReadOldFilesFromGit(gitRepo, gitCommitFrom, changes))
	log.Printf("Parsed old files versions for %s", time.Since(start))

	start = time.Now()
	meta.SetIndexingComplete(false)
	linter.ParseFilenames(linter.ReadChangesFromWorkTree(gitWorkTree, changes))
	meta.SetIndexingComplete(true)
	log.Printf("Indexed new files versions for %s", time.Since(start))

	start = time.Now()
	reports = linter.ParseFilenames(linter.ReadChangesFromWorkTree(gitWorkTree, changes))
	log.Printf("Parsed new file versions in %s", time.Since(start))

	return oldReports, reports, changes, true
}

func gitMain() (int, error) {
	var (
		oldReports, reports []*linter.Report
		diffArgs            []string
		changes             []git.Change
		changeLog           []git.Commit
		ok                  bool
	)

	// prepareGitArgs also populates global variables like fromCommit
	logArgs, diffArgs, err := prepareGitArgs()
	if err != nil {
		return 0, err
	}

	oldReports, reports, changes, ok = gitRepoComputeReportsFromLocalChanges()
	if !ok {
		oldReports, reports, changes, changeLog, ok = gitRepoComputeReportsFromCommits(logArgs, diffArgs)
		if !ok {
			return 0, nil
		}
	}

	start := time.Now()
	diff, err := linter.DiffReports(gitRepo, diffArgs, changes, changeLog, oldReports, reports, 8)
	if err != nil {
		return 0, fmt.Errorf("Could not compute reports diff: %v", err)
	}
	log.Printf("Computed reports diff for %s", time.Since(start))

	criticalReports := analyzeReports(diff)

	if criticalReports > 0 {
		log.Printf("Found %d critical issues, please fix them.", criticalReports)
		return 2, nil
	}
	log.Printf("No critical issues found. Your code is perfect.")
	return 0, nil
}

func analyzeGitAuthorsWhiteList(changeLog []git.Commit) (shouldRun bool) {
	if gitAuthorsWhitelist != "" {
		whiteList := make(map[string]bool)
		for _, name := range strings.Split(gitAuthorsWhitelist, ",") {
			whiteList[name] = true
		}

		for _, commit := range changeLog {
			if !whiteList[commit.Author] {
				log.Printf("Found commit from '%s', PHP linter not running", commit.Author)
				return false
			}
		}
	}

	return true
}

func analyzeReports(diff []*linter.Report) (criticalReports int) {
	for _, r := range diff {
		if !isEnabled(r) {
			continue
		}

		if r.IsDisabledByUser() {
			filename := r.GetFilename()
			if !canBeDisabled(filename) {
				fmt.Fprintf(outputFp, "You are not allowed to disable linter for file '%s'\n", filename)
			} else {
				continue
			}
		}

		if r.IsCritical() {
			criticalReports++
		}

		fmt.Fprintf(outputFp, "%s\n", r)
	}

	return criticalReports
}

func prepareGitArgs() (logArgs, diffArgs []string, err error) {
	if gitPushArg != "" {
		args := strings.Fields(gitPushArg)
		if len(args) != 3 {
			return nil, nil, fmt.Errorf("Unexpected format of push arguments, expected only 3 columns: %s", gitPushArg)
		}
		gitCommitFrom, gitCommitTo, gitRef = args[0], args[1], args[2]
	}

	if gitCommitFrom == git.Zero {
		gitCommitFrom = "master"
	}

	if !gitSkipFetch {
		start := time.Now()
		log.Printf("Fetching origin master to ORIGIN_MASTER")
		if err := git.Fetch(gitRepo, "master", "ORIGIN_MASTER"); err != nil {
			return nil, nil, fmt.Errorf("Could not fetch ORIGIN_MASTER: %v", err.Error())
		}
		log.Printf("Fetched for %s", time.Since(start))
	}

	fromAndMaster, err := git.MergeBase(gitRepo, "ORIGIN_MASTER", gitCommitFrom)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", gitCommitFrom)
	}

	toAndMaster, err := git.MergeBase(gitRepo, "ORIGIN_MASTER", gitCommitTo)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", gitCommitTo)
	}

	// check if master was merged in between the commits
	if fromAndMaster != toAndMaster {
		gitCommitFrom = toAndMaster
	}

	logArgs = []string{gitCommitFrom + ".." + gitCommitTo}
	diffArgs = []string{gitCommitFrom + ".." + gitCommitTo}

	return logArgs, diffArgs, nil
}

func setDiscardVarPredicate() error {
	switch unusedVarPattern {
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
		re, err := regexp.Compile(unusedVarPattern)
		if err != nil {
			return err
		}
		linter.IsDiscardVar = func(s string) bool {
			return re.MatchString(s)
		}
	}

	return nil
}
