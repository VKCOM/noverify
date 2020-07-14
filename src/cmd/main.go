package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // it is ok for actually main package
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync/atomic"
	"time"

	"github.com/VKCOM/noverify/src/baseline"
	"github.com/VKCOM/noverify/src/cmd/stubs"
	"github.com/VKCOM/noverify/src/langsrv"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/rules"
)

// Line below implies that we have `https://github.com/VKCOM/phpstorm-stubs.git` cloned
// to the `./src/cmd/stubs/phpstorm-stubs`.
//
//go:generate go-bindata -pkg stubs -nometadata -o ./stubs/phpstorm_stubs.go -ignore=\.idea -ignore=\.git ./stubs/phpstorm-stubs/...

//go:generate go-bindata -pkg embeddedrules -nometadata -o ./embeddedrules/rules.go ./embeddedrules/rules.php

func isCritical(l *linterRunner, r *linter.Report) bool {
	if len(l.reportsCriticalSet) != 0 {
		return l.reportsCriticalSet[r.CheckName()]
	}
	return r.IsCritical()
}

func isEnabled(l *linterRunner, r *linter.Report) bool {
	if !l.IsEnabledByFlags(r.CheckName()) {
		return false
	}

	if linter.ExcludeRegex == nil {
		return true
	}

	// Disabled by a file comment.
	return !linter.ExcludeRegex.MatchString(r.GetFilename())
}

// canBeDisabled returns whether or not '@linter disable' can be used for the specified file
func canBeDisabled(l *linterRunner, filename string) bool {
	if l.allowDisableRegex == nil {
		return false
	}

	return l.allowDisableRegex.MatchString(filename)
}

// Run executes linter main function.
//
// It is separate from linter so that you can insert your own hooks
// before running Run().
//
// It returns a status code to be used for os.Exit() and
// initialization error (if any).
//
// Optionally, non-nil config can be passed to customize function behavior.
func Run(cfg *MainConfig) (int, error) {
	if cfg == nil {
		cfg = &MainConfig{}
	}

	ruleSets, err := parseRules()
	if err != nil {
		return 1, fmt.Errorf("preload rules: %v", err)
	}

	var args cmdlineArguments
	bindFlags(ruleSets, &args)
	flag.Parse()
	if args.disableCache {
		linter.CacheDir = ""
	}
	if cfg.AfterFlagParse != nil {
		cfg.AfterFlagParse(InitEnvironment{
			RuleSets: ruleSets,
		})
	}

	return mainNoExit(ruleSets, &args, cfg)
}

// Main is like Run(), but it calls os.Exit() and does not return.
func Main(cfg *MainConfig) {
	status, err := Run(cfg)
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
func mainNoExit(ruleSets []*rules.Set, args *cmdlineArguments, cfg *MainConfig) (int, error) {
	if args.version {
		// Version is already printed. Can exit here.
		return 0, nil
	}

	if args.pprofHost != "" {
		go http.ListenAndServe(args.pprofHost, nil)
	}

	// Since this function is expected to be exit-free, it's OK
	// to defer calls here to make required flushes/cleanup.
	if args.cpuProfile != "" {
		f, err := os.Create(args.cpuProfile)
		if err != nil {
			return 0, fmt.Errorf("Could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return 0, fmt.Errorf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if args.memProfile != "" {
		defer func() {
			f, err := os.Create(args.memProfile)
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

	linter.PHPExtensions = strings.Split(args.phpExtensionsArg, ",")

	var l linterRunner
	if err := l.Init(ruleSets, args); err != nil {
		return 1, fmt.Errorf("init: %v", err)
	}

	lintdebug.Register(func(msg string) { linter.DebugMessage("%s", msg) })
	go linter.MemoryLimiterThread()

	if linter.LangServer {
		langsrv.RegisterDebug()
		langsrv.Start()
		return 0, nil
	}

	log.Printf("Started")

	if err := initStubs(); err != nil {
		return 0, fmt.Errorf("Init stubs: %v", err)
	}

	if args.gitRepo != "" {
		return gitMain(&l, cfg)
	}

	linter.AnalysisFiles = flag.Args()

	log.Printf("Indexing %+v", flag.Args())
	linter.ParseFilenames(linter.ReadFilenames(flag.Args(), nil))
	parseIndexOnlyFiles(&l)
	meta.SetIndexingComplete(true)
	log.Printf("Linting")

	filenames := flag.Args()
	if args.fullAnalysisFiles != "" {
		filenames = strings.Split(args.fullAnalysisFiles, ",")
	}

	reports := linter.ParseFilenames(linter.ReadFilenames(filenames, linter.ExcludeRegex))
	if args.outputBaseline {
		if err := createBaseline(&l, cfg, reports); err != nil {
			return 1, fmt.Errorf("write baseline: %v", err)
		}
		return 0, nil
	}
	criticalReports := analyzeReports(&l, cfg, reports)

	if criticalReports > 0 {
		log.Printf("Found %d critical reports", criticalReports)
		return 2, nil
	}
	return 0, nil
}

func createBaseline(l *linterRunner, cfg *MainConfig, reports []*linter.Report) error {
	var stats baseline.Stats
	stats.CountPerCheck = make(map[string]int)

	files := make(map[string]baseline.FileProfile)
	for _, r := range reports {
		if cfg.BeforeReport != nil && !cfg.BeforeReport(r) {
			continue
		}
		if !isEnabled(l, r) {
			continue
		}

		stats.CountTotal++
		stats.CountPerCheck[r.CheckName()]++
		filename := filepath.Base(r.GetFilename())
		f, ok := files[filename]
		if !ok {
			f.Filename = filename
			f.Reports = make(map[uint64]baseline.Report)
		}
		info := f.Reports[r.Hash()]
		info.Count++
		info.Hash = r.Hash()
		f.Reports[r.Hash()] = info
		files[filename] = f
	}

	profile := &baseline.Profile{
		LinterVersion: cfg.LinterVersion,
		CreatedAt:     time.Now().Unix(),
		Files:         files,
	}
	return baseline.WriteProfile(l.outputFp, profile, &stats)
}

func analyzeReports(l *linterRunner, cfg *MainConfig, diff []*linter.Report) (criticalReports int) {
	filtered := make([]*linter.Report, 0, len(diff))
	var linterErrors []string

	handeledFiles := map[string]bool{}

	for _, r := range diff {
		if cfg.BeforeReport != nil && !cfg.BeforeReport(r) {
			continue
		}
		if !isEnabled(l, r) {
			continue
		}

		if r.IsDisabledByUser() {
			filename := r.GetFilename()
			if !canBeDisabled(l, filename) {

				if !handeledFiles[filename] {
					linterErrors = append(linterErrors, fmt.Sprintf("You are not allowed to disable linter for file '%s'", filename))
					handeledFiles[filename] = true
				}

			} else {
				continue
			}
		}

		filtered = append(filtered, r)

		if isCritical(l, r) {
			criticalReports++
		}
	}

	if l.args.outputJSON {
		type reportList struct {
			Reports []*linter.Report
			Errors  []string
		}
		list := &reportList{
			Reports: filtered,
			Errors:  linterErrors,
		}
		d := json.NewEncoder(l.outputFp)
		if err := d.Encode(list); err != nil {
			// Should never fail to marshal our own reports.
			panic(fmt.Sprintf("report list marshaling failed: %v", err))
		}
	} else {
		for _, err := range linterErrors {
			fmt.Fprintf(l.outputFp, "%s\n", err)
		}
		for _, r := range filtered {
			if isCritical(l, r) {
				fmt.Fprintf(l.outputFp, "<critical> %s\n", r.String())
			} else {
				fmt.Fprintf(l.outputFp, "%s\n", r.String())
			}
		}
	}

	return criticalReports
}

func initStubs() error {
	if linter.StubsDir != "" {
		linter.InitStubsFromDir(linter.StubsDir)
		return nil
	}

	// Try to use embedded stubs (from stubs/phpstorm_stubs.go).
	if err := loadEmbeddedStubs(); err != nil {
		return fmt.Errorf("failed to load embedded stubs: %v", err)
	}

	return nil
}

func LoadEmbeddedStubs(filenames []string) error {
	var errorsCount int64

	readStubs := func(ch chan linter.FileInfo) {
		for _, filename := range filenames {
			data, err := stubs.Asset(filename)
			if err != nil {
				log.Printf("Failed to read embedded %q file: %v", filename, err)
				atomic.AddInt64(&errorsCount, 1)
				continue
			}
			ch <- linter.FileInfo{
				Filename: filename,
				Contents: data,
			}
		}
	}

	linter.InitStubs(readStubs)

	// Using atomic here for consistency.
	if atomic.LoadInt64(&errorsCount) != 0 {
		return fmt.Errorf("failed to load %d embedded files", errorsCount)
	}

	return nil
}

func loadEmbeddedStubs() error {
	filenames := stubs.AssetNames()
	if len(filenames) == 0 {
		return fmt.Errorf("empty file list")
	}
	return LoadEmbeddedStubs(filenames)
}
