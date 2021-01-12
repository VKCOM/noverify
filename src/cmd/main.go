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
	"github.com/VKCOM/noverify/src/linter/lintapi"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

// GlobalCmds is a global map of commands.
var GlobalCmds = NewCommands()

// RegisterDefaultCommands registers default commands for NoVerify.
func RegisterDefaultCommands() {
	GlobalCmds.RegisterCommand(&SubCommand{
		Name:        "check",
		Main:        Check,
		Description: "lint the entire project",
		Examples: []SubCommandExample{
			{
				Description: "show subcommand usage",
				Line:        "-help",
			},
			{
				Description: "run linter with default options",
				Line:        "<analyze-path>",
			},
		},
	})

	GlobalCmds.RegisterCommand(&SubCommand{
		Name:        "help",
		Main:        Help,
		Description: "print linter documentation based on the subject",
		Examples: []SubCommandExample{
			{
				Description: "show supported sub-subCommands",
				Line:        "",
			},
			{
				Description: "print all supported checkers short summary",
				Line:        "checkers",
			},
			{
				Description: "print dupSubExpr checker detailed documentation",
				Line:        "checkers dupSubExpr",
			},
		},
	})
}

// Line below implies that we have `https://github.com/VKCOM/phpstorm-stubs.git` cloned
// to the `./src/cmd/stubs/phpstorm-stubs`.
//
//go:generate go-bindata -pkg stubs -nometadata -o ./stubs/phpstorm_stubs.go -ignore=\.idea -ignore=\.git ./stubs/phpstorm-stubs/...

//go:generate go-bindata -pkg embeddedrules -nometadata -o ./embeddedrules/rules.go ./embeddedrules/rules.php

func isCritical(l *linterRunner, r *linter.Report) bool {
	if len(l.reportsCriticalSet) != 0 {
		return l.reportsCriticalSet[r.CheckName]
	}
	return r.IsCritical()
}

func isEnabled(l *linterRunner, r *linter.Report) bool {
	if !l.IsEnabledByFlags(r.CheckName) {
		return false
	}

	if linter.ExcludeRegex == nil {
		return true
	}

	// Disabled by a file comment.
	return !linter.ExcludeRegex.MatchString(r.Filename)
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
		} else if looksLikeCommandName(commandName) {
			fmt.Printf("Sub-command %s doesn't exist\n\n", commandName)
			GlobalCmds.PrintHelpPage()
			return 0, nil
		}

	}
	if subcmd == nil {
		log.Print(`

NoVerify migrates to the new CLI using commands, launching in this way is still possible, but is already deprecated.
Use 'noverify help' for more information.

`)
		subcmd, _ = GlobalCmds.GetCommand("check")
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

	workspace.PHPExtensions = strings.Split(args.phpExtensionsArg, ",")

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
	linter.ParseFilenames(workspace.ReadFilenames(flag.Args(), nil), l.allowDisableRegex)
	parseIndexOnlyFiles(&l)
	meta.SetIndexingComplete(true)

	filenames := flag.Args()
	if args.fullAnalysisFiles != "" {
		filenames = strings.Split(args.fullAnalysisFiles, ",")
	}

	log.Printf("Linting")
	reports := linter.ParseFilenames(workspace.ReadFilenames(filenames, l.filenameFilter), l.allowDisableRegex)
	if args.outputBaseline {
		if err := createBaseline(&l, cfg, reports); err != nil {
			return 1, fmt.Errorf("write baseline: %v", err)
		}
		return 0, nil
	}
	criticalReports, containsAutofixableReports := analyzeReports(&l, cfg, reports)

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
		if !isEnabled(l, r) {
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

func haveAutofixableReports(reports []*linter.Report) bool {
	if len(reports) == 0 {
		return false
	}

	declaredChecks := linter.GetDeclaredChecks()
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
		if !isEnabled(l, r) {
			continue
		}

		filtered = append(filtered, r)

		if isCritical(l, r) {
			criticalReports++
		}
	}

	containsAutofixableReports = haveAutofixableReports(filtered)

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
			if isCritical(l, r) {
				fmt.Fprintf(l.outputFp, "<critical> %s\n", FormatReport(r))
			} else {
				fmt.Fprintf(l.outputFp, "%s\n", FormatReport(r))
			}
		}
	}

	return criticalReports, containsAutofixableReports
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

	readStubs := func(ch chan workspace.FileInfo) {
		for _, filename := range filenames {
			data, err := stubs.Asset(filename)
			if err != nil {
				log.Printf("Failed to read embedded %q file: %v", filename, err)
				atomic.AddInt64(&errorsCount, 1)
				continue
			}
			ch <- workspace.FileInfo{
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
	return LoadEmbeddedStubs(filenames)
}
