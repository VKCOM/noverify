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
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linter/lintapi"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

// Line below implies that we have `https://github.com/VKCOM/phpstorm-stubs.git` cloned
// to the `./src/cmd/stubs/phpstorm-stubs`.
//
//go:generate go-bindata -pkg stubs -nometadata -o ./stubs/phpstorm_stubs.go -ignore=\.idea -ignore=\.git ./stubs/phpstorm-stubs/...

// GlobalCmds is a global map of commands.
var GlobalCmds = NewCommands()

// RegisterDefaultCommands registers default commands for NoVerify.
func RegisterDefaultCommands() {
	GlobalCmds.RegisterCommand(&SubCommand{
		Name:        "check",
		Main:        Check,
		Description: "Lint the entire project",
		Examples: []SubCommandExample{
			{
				Description: "Show command usage",
				Line:        "-help",
			},
			{
				Description: "Run linter with default options",
				Line:        "[options] <analyze-path>",
			},
		},
	})

	GlobalCmds.RegisterCommand(&SubCommand{
		Name:        "help",
		Main:        Help,
		Description: "Print linter documentation based on the subject",
		Examples: []SubCommandExample{
			{
				Description: "Show usage help",
				Line:        "",
			},
			{
				Description: "Show all supported checkers short summary",
				Line:        "checkers",
			},
			{
				Description: "Show <name> checker detailed documentation",
				Line:        "checkers <name>",
			},
		},
	})
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
	RegisterDefaultCommands()

	if cfg.OverriddenCommands != nil {
		GlobalCmds.OverrideCommands(cfg.OverriddenCommands)
	}

	var subcmd *SubCommand
	var found bool
	if len(os.Args) >= 2 {
		commandName := os.Args[1]
		subcmd, found = GlobalCmds.GetCommand(commandName)
		if found {
			subIdx := 1 // [0] is program name
			// Erase sub-command argument (index=1) to make it invisible for
			// sub commands themselves.
			os.Args = append(os.Args[:subIdx], os.Args[subIdx+1:]...)
		} else {
			fmt.Printf("Sub-command '%s' doesn't exist\n\n", commandName)
			GlobalCmds.PrintHelpPage()
			return 0, nil
		}

	}
	if subcmd == nil {
		GlobalCmds.PrintHelpPage()
		return 0, nil
	}

	status, err := subcmd.Main(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return status, nil
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
func mainNoExit(l *linter.Linter, ruleSets []*rules.Set, args *cmdlineArguments, cfg *MainConfig) (int, error) {
	if args.version {
		// Version is already printed. Can exit here.
		return 0, nil
	}

	if args.pprofHost != "" {
		go func() {
			err := http.ListenAndServe(args.pprofHost, nil)
			if err != nil {
				log.Printf("pprof listen and serve: %v", err)
			}
		}()
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

	runner := linterRunner{
		config: l.Config(),
		linter: l,
	}
	if err := runner.Init(ruleSets, args); err != nil {
		return 1, fmt.Errorf("init: %v", err)
	}

	lintdebug.Register(func(msg string) {
		if l.Config().Debug {
			log.Print(msg)
		}
	})
	go linter.MemoryLimiterThread(args.maxFileSize)

	log.Printf("Started")

	if err := initStubs(runner.linter); err != nil {
		return 0, fmt.Errorf("Init stubs: %v", err)
	}

	if args.gitRepo != "" {
		return gitMain(&runner, cfg)
	}

	log.Printf("Indexing %+v", flag.Args())
	runner.linter.AnalyzeFiles(workspace.ReadFilenames(flag.Args(), nil, l.Config().PhpExtensions))
	parseIndexOnlyFiles(&runner)
	runner.linter.MetaInfo().SetIndexingComplete(true)

	filenames := flag.Args()
	if args.fullAnalysisFiles != "" {
		filenames = strings.Split(args.fullAnalysisFiles, ",")
	}

	log.Printf("Linting")
	reports := runner.linter.AnalyzeFiles(workspace.ReadFilenames(filenames, runner.filenameFilter, l.Config().PhpExtensions))
	if args.outputBaseline {
		if err := createBaseline(&runner, cfg, reports); err != nil {
			return 1, fmt.Errorf("write baseline: %v", err)
		}
		return 0, nil
	}
	criticalReports, containsAutofixableReports := analyzeReports(&runner, cfg, reports)

	if containsAutofixableReports {
		log.Println("Some issues are autofixable (try using the `-fix` flag)")
	}

	if criticalReports > 0 {
		log.Printf("Found %d critical reports", criticalReports)
		return 2, nil
	}
	log.Printf("No critical issues found. Your code is perfect.")
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
		if !l.IsEnabledReport(r) {
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

func analyzeReports(l *linterRunner, cfg *MainConfig, diff []*linter.Report) (criticalReports int, containsAutofixableReports bool) {
	filtered := make([]*linter.Report, 0, len(diff))

	for _, r := range diff {
		if cfg.BeforeReport != nil && !cfg.BeforeReport(r) {
			continue
		}
		if !l.IsEnabledReport(r) {
			continue
		}

		filtered = append(filtered, r)

		if l.IsCriticalReport(r) {
			criticalReports++
		}
	}

	containsAutofixableReports = haveAutofixableReports(l.config, filtered)

	if l.args.outputJSON {
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
			if l.IsCriticalReport(r) {
				fmt.Fprintf(l.outputFp, "<critical> %s\n", FormatReport(r))
			} else {
				fmt.Fprintf(l.outputFp, "%s\n", FormatReport(r))
			}
		}
	}

	return criticalReports, containsAutofixableReports
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
