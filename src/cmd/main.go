package cmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof" // it is ok for actually main package
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync/atomic"

	"github.com/VKCOM/noverify/src/cmd/stubs"
	"github.com/VKCOM/noverify/src/langsrv"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/client9/misspell"
)

// Line below implies that we have `https://github.com/VKCOM/phpstorm-stubs.git` cloned
// to the `./src/cmd/stubs/phpstorm-stubs`.
//
//go:generate go-bindata -pkg stubs -nometadata -o ./stubs/phpstorm_stubs.go -ignore=\.idea -ignore=\.git ./stubs/phpstorm-stubs/...

func isCritical(r *linter.Report) bool {
	if len(reportsCriticalSet) != 0 {
		return reportsCriticalSet[r.CheckName()]
	}
	return r.IsCritical()
}

func isEnabled(r *linter.Report) bool {
	if !reportsIncludeChecksSet[r.CheckName()] {
		return false // Not enabled by -allow-checks
	}

	if reportsExcludeChecksSet[r.CheckName()] {
		return false // Disabled by -exclude-checks
	}

	if linter.ExcludeRegex == nil {
		return true
	}

	// Disabled by a file comment.
	return !linter.ExcludeRegex.MatchString(r.GetFilename())
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

func loadMisspellDicts(dicts []string) error {
	linter.TypoFixer = &misspell.Replacer{}

	for _, d := range dicts {
		d = strings.TrimSpace(d)
		switch {
		case d == "Eng":
			linter.TypoFixer.AddRuleList(misspell.DictMain)
		case d == "Eng/US":
			linter.TypoFixer.AddRuleList(misspell.DictAmerican)
		case d == "Eng/UK" || d == "Eng/GB":
			linter.TypoFixer.AddRuleList(misspell.DictBritish)
		default:
			return fmt.Errorf("unsupported %s misspell-list entry", d)
		}
	}

	linter.TypoFixer.Compile()
	return nil
}

// mainNoExit implements main, but instead of doing log.Fatal or os.Exit it
// returns error or non-zero integer status code to be passed to os.Exit by the caller.
// Note that if error is not nil, integer code will be discarded, so it can be 0.
//
// We don't want os.Exit to be inserted randomly to avoid defer cancellation.
func mainNoExit() (int, error) {
	if version {
		// Version is already printed. Can exit here.
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

	if misspellList != "" {
		err := loadMisspellDicts(strings.Split(misspellList, ","))
		if err != nil {
			return 0, err
		}
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

	if err := initStubs(); err != nil {
		return 0, fmt.Errorf("Init stubs: %v", err)
	}

	if err := initRules(); err != nil {
		return 0, fmt.Errorf("Init rules: %v", err)
	}

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

	reports := linter.ParseFilenames(linter.ReadFilenames(filenames, linter.ExcludeRegex))
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
		linter.ExcludeRegex, err = regexp.Compile(reportsExclude)
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
	stringToSet := func(s string) map[string]bool {
		set := make(map[string]bool)
		for _, name := range strings.Split(s, ",") {
			set[strings.TrimSpace(name)] = true
		}
		return set
	}

	reportsExcludeChecksSet = stringToSet(reportsExcludeChecks)
	reportsIncludeChecksSet = stringToSet(allowChecks)
	if reportsCritical != allNonMaybe {
		reportsCriticalSet = stringToSet(reportsCritical)
	}
}

func analyzeReports(diff []*linter.Report) (criticalReports int) {
	filtered := make([]*linter.Report, 0, len(diff))
	var linterErrors []string
	for _, r := range diff {
		if !isEnabled(r) {
			continue
		}

		if r.IsDisabledByUser() {
			filename := r.GetFilename()
			if !canBeDisabled(filename) {
				linterErrors = append(linterErrors, fmt.Sprintf("You are not allowed to disable linter for file '%s'", filename))
			} else {
				continue
			}
		}

		filtered = append(filtered, r)

		if isCritical(r) {
			criticalReports++
		}
	}

	if outputJSON {
		type reportList struct {
			Reports []*linter.Report
			Errors  []string
		}
		list := &reportList{
			Reports: filtered,
			Errors:  linterErrors,
		}
		d := json.NewEncoder(outputFp)
		if err := d.Encode(list); err != nil {
			// Should never fail to marshal our own reports.
			panic(fmt.Sprintf("report list marshaling failed: %v", err))
		}
	} else {
		for _, err := range linterErrors {
			fmt.Fprintf(outputFp, "%s\n", err)
		}
		for _, r := range filtered {
			if isCritical(r) {
				fmt.Fprintf(outputFp, "<critical> %s\n", r.String())
			} else {
				fmt.Fprintf(outputFp, "%s\n", r.String())
			}
		}
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
		linter.IsDiscardVar = re.MatchString
	}

	return nil
}

func initRules() error {
	if rulesList == "" {
		return nil
	}

	appendRules := func(dst, src *rules.ScopedSet) {
		for i, list := range src.RulesByKind {
			dst.RulesByKind[i] = append(dst.RulesByKind[i], list...)
		}
	}

	p := rules.NewParser()
	linter.Rules = rules.NewSet()
	for _, filename := range strings.Split(rulesList, ",") {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		rset, err := p.Parse(filename, bytes.NewReader(data))
		if err != nil {
			return err
		}

		appendRules(linter.Rules.Any, rset.Any)
		appendRules(linter.Rules.Root, rset.Root)
		appendRules(linter.Rules.Local, rset.Local)

		for _, name := range rset.AlwaysAllowed {
			reportsIncludeChecksSet[name] = true
		}
		for _, name := range rset.AlwaysCritical {
			reportsCriticalSet[name] = true
		}
	}

	return nil
}

func initStubs() error {
	if linter.StubsDir != "" {
		linter.InitStubs()
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

	linter.ParseFilenames(readStubs)
	meta.Info.InitStubs()

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
