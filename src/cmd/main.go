package cmd

import (
	"encoding/json"
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
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linter/lintapi"
	"github.com/VKCOM/noverify/src/workspace"
)

// Line below implies that we have `https://github.com/VKCOM/phpstorm-stubs.git` cloned
// to the `./src/cmd/stubs/phpstorm-stubs`.
//
//go:generate go-bindata -pkg stubs -nometadata -o ./stubs/phpstorm_stubs.go -ignore=\.idea -ignore=\.git ./stubs/phpstorm-stubs/...

func registerMainApp() *App {
	return &App{
		Name:        "noverify",
		Description: "Pretty fast linter (static analysis tool) for PHP",
		Commands: []*Command{
			{
				Name:        "check",
				Description: "The command to lint files",
				Action:      Check,
				Arguments: []*Argument{
					{
						Name:        "folders/files",
						Description: "Folders and/or files for check",
					},
				},
				RegisterFlags: RegisterCheckFlags,
			},
			{
				Name:        "checkers",
				Description: "The command to show list of checkers",
				Arguments: []*Argument{
					{
						Name:        "checker-name",
						Description: "Show info for a certain <checker-name> checker ",
					},
				},
				Action: Checkers,
			},

			{
				Name:        "version",
				Description: "The command to output the tool version",
				Action: func(ctx *AppContext) (int, error) {
					printVersion()
					return 0, nil
				},
			},
		},
	}
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

	if cfg.LinterVersion == "" {
		cfg.LinterVersion = BuildCommit
	}

	config := cfg.LinterConfig
	if config == nil {
		config = linter.NewConfig()
		cfg.LinterConfig = config
	}

	cfg.linter = linter.NewLinter(config)

	ruleSets, err := parseEmbeddedRules()
	if err != nil {
		return 1, fmt.Errorf("preload embedded rules: %v", err)
	}

	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	cfg.rulesSets = append(cfg.rulesSets, ruleSets...)

	if cfg.RegisterCheckers != nil {
		for _, checker := range cfg.RegisterCheckers() {
			cfg.linter.Config().Checkers.DeclareChecker(checker)
		}
	}

	app := registerMainApp()
	if cfg.ModifyApp != nil {
		cfg.ModifyApp(app)
	}
	return app.Run(cfg)
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
func mainNoExit(ctx *AppContext) (int, error) {
	if ctx.ParsedFlags.PprofHost != "" {
		go func() {
			err := http.ListenAndServe(ctx.ParsedFlags.PprofHost, nil)
			if err != nil {
				log.Printf("pprof listen and serve: %v", err)
			}
		}()
	}

	// Since this function is expected to be exit-free, it's OK
	// to defer calls here to make required flushes/cleanup.
	if ctx.ParsedFlags.CPUProfile != "" {
		f, err := os.Create(ctx.ParsedFlags.CPUProfile)
		if err != nil {
			return 0, fmt.Errorf("Could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return 0, fmt.Errorf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if ctx.ParsedFlags.MemProfile != "" {
		defer func() {
			f, err := os.Create(ctx.ParsedFlags.MemProfile)
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

	lint := ctx.MainConfig.linter
	ruleSets := ctx.MainConfig.rulesSets

	runner := linterRunner{
		config:         lint.Config(),
		linter:         lint,
		checkersFilter: linter.NewCheckersFilter(),
	}
	if err := runner.Init(ruleSets, &ctx.ParsedFlags); err != nil {
		return 1, fmt.Errorf("init: %v", err)
	}

	lintdebug.Register(func(msg string) {
		if lint.Config().Debug {
			log.Print(msg)
		}
	})
	go linter.MemoryLimiterThread(ctx.ParsedFlags.MaxFileSize)

	log.Printf("Started")

	if err := initStubs(runner.linter); err != nil {
		return 0, fmt.Errorf("Init stubs: %v", err)
	}

	if ctx.ParsedFlags.GitRepo != "" {
		return gitMain(&runner, ctx.MainConfig)
	}

	filenames := ctx.ParsedArgs

	log.Printf("Indexing %+v", filenames)
	runner.linter.AnalyzeFiles(workspace.ReadFilenames(filenames, nil, lint.Config().PhpExtensions))
	parseIndexOnlyFiles(&runner)
	runner.linter.MetaInfo().SetIndexingComplete(true)

	if ctx.ParsedFlags.FullAnalysisFiles != "" {
		filenames = strings.Split(ctx.ParsedFlags.FullAnalysisFiles, ",")
	}

	log.Printf("Linting")
	reports := runner.linter.AnalyzeFiles(workspace.ReadFilenames(filenames, runner.filenameFilter, lint.Config().PhpExtensions))
	if ctx.ParsedFlags.OutputBaseline {
		if err := createBaseline(&runner, ctx.MainConfig, reports); err != nil {
			return 1, fmt.Errorf("write baseline: %v", err)
		}
		return 0, nil
	}
	criticalReports, minorReports, containsAutofixableReports := analyzeReports(&runner, ctx.MainConfig, reports)

	if containsAutofixableReports && !runner.config.ApplyQuickFixes {
		log.Println("Some issues are autofixable (try using the `-fix` flag)")
	}

	if criticalReports > 0 {
		log.Printf("Found %d critical and %d minor reports", criticalReports, minorReports)
		return 2, nil
	}
	if !ctx.MainConfig.DisableCriticalIssuesLog {
		if minorReports == 0 {
			log.Printf("No issues found. Your code is perfect.")
		} else {
			log.Printf("Found %d minor issues.", minorReports)
		}
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
		if !l.checkersFilter.IsEnabledReport(r.CheckName, r.Filename) {
			continue
		}

		stats.CountTotal++
		stats.CountPerCheck[r.CheckName]++
		filename := filepath.Base(r.Filename)
		f, ok := files[filename]
		if !ok {
			f.Filename = filename
			f.Reports = make(map[uint64]baseline.Report)
		}
		info := f.Reports[r.Hash]
		info.Count++
		info.Hash = r.Hash
		f.Reports[r.Hash] = info
		files[filename] = f
	}

	profile := &baseline.Profile{
		LinterVersion: cfg.LinterVersion,
		CreatedAt:     time.Now().Unix(),
		Files:         files,
	}
	return baseline.WriteProfile(l.outputFp, profile, &stats)
}

func FormatReport(r *linter.Report) string {
	msg := r.Message
	if r.CheckName != "" {
		msg = r.CheckName + ": " + msg
	}

	// No context line for security-level warnings.
	if r.Level == lintapi.LevelSecurity {
		return fmt.Sprintf("%-7s %s at %s:%d", r.Severity(), msg, r.Filename, r.Line)
	}

	cursor := strings.Builder{}
	for i, ch := range r.Context {
		if i == r.StartChar {
			break
		}
		if ch == '\t' {
			cursor.WriteRune('\t')
		} else {
			cursor.WriteByte(' ')
		}
	}

	if r.EndChar > r.StartChar {
		cursor.WriteString(strings.Repeat("^", r.EndChar-r.StartChar))
	}

	return fmt.Sprintf("%-7s %s at %s:%d\n%s\n%s",
		r.Severity(), msg, r.Filename, r.Line, r.Context, cursor.String())
}

func haveAutofixableReports(config *linter.Config, reports []*linter.Report) bool {
	if len(reports) == 0 {
		return false
	}

	declaredChecks := config.Checkers.ListDeclared()
	checksWithQuickfix := make(map[string]struct{})

	for _, check := range declaredChecks {
		if !check.Quickfix {
			continue
		}

		checksWithQuickfix[check.Name] = struct{}{}
	}

	for _, r := range reports {
		if _, ok := checksWithQuickfix[r.CheckName]; ok {
			return true
		}
	}

	return false
}

func analyzeReports(l *linterRunner, cfg *MainConfig, diff []*linter.Report) (criticalReports, minorReports int, containsAutofixableReports bool) {
	filtered := make([]*linter.Report, 0, len(diff))

	for _, r := range diff {
		if cfg.BeforeReport != nil && !cfg.BeforeReport(r) {
			continue
		}

		filtered = append(filtered, r)

		if l.checkersFilter.IsCriticalReport(r) {
			criticalReports++
		} else {
			minorReports++
		}
	}

	containsAutofixableReports = haveAutofixableReports(l.config, filtered)

	if l.flags.OutputJSON {
		type reportList struct {
			Reports []*linter.Report
			Errors  []string
		}
		list := &reportList{
			Reports: filtered,
		}
		d := json.NewEncoder(l.outputFp)
		if err := d.Encode(list); err != nil {
			// Should never fail to marshal our own reports.
			panic(fmt.Sprintf("report list marshaling failed: %v", err))
		}
	} else {
		for _, r := range filtered {
			if l.checkersFilter.IsCriticalReport(r) {
				fmt.Fprintf(l.outputFp, "<critical> %s\n", FormatReport(r))
			} else {
				fmt.Fprintf(l.outputFp, "%s\n", FormatReport(r))
			}
		}
	}

	return criticalReports, minorReports, containsAutofixableReports
}

func initStubs(l *linter.Linter) error {
	if l.Config().StubsDir != "" {
		l.InitStubsFromDir(l.Config().StubsDir)
		return nil
	}

	// Try to use embedded stubs (from stubs/phpstorm_stubs.go).
	if err := loadEmbeddedStubs(l); err != nil {
		return fmt.Errorf("failed to load embedded stubs: %v", err)
	}

	return nil
}

func LoadEmbeddedStubs(l *linter.Linter, filenames []string) error {
	var errorsCount int64

	readStubs := func(ch chan workspace.FileInfo) {
		for _, filename := range filenames {
			contents, err := stubs.Asset(filename)
			if err != nil {
				log.Printf("Failed to read embedded %q file: %v", filename, err)
				atomic.AddInt64(&errorsCount, 1)
				continue
			}
			ch <- workspace.FileInfo{
				Name:     filename,
				Contents: contents,
			}
		}
	}

	l.InitStubs(readStubs)

	// Using atomic here for consistency.
	if atomic.LoadInt64(&errorsCount) != 0 {
		return fmt.Errorf("failed to load %d embedded files", errorsCount)
	}

	return nil
}

func loadEmbeddedStubs(l *linter.Linter) error {
	var filenames []string
	// NOVERIFYDEBUG_LOAD_STUBS is used in golden tests to specify
	// the test dependencies that need to be loaded.
	if list := os.Getenv("NOVERIFYDEBUG_LOAD_STUBS"); list != "" {
		filenames = strings.Split(list, ",")
	} else {
		filenames = stubs.AssetNames()
	}
	if len(filenames) == 0 {
		return fmt.Errorf("empty file list")
	}
	return LoadEmbeddedStubs(l, filenames)
}
