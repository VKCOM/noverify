package cmd

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // it is ok for actually main package
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"

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
