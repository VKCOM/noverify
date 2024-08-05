//go:build wasm
// +build wasm

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"syscall/js"
	"time"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

type ReportSeverity uint8

const (
	Error ReportSeverity = iota
	Warning
	Notice
)

type Position struct {
	Line      int `json:"line"`
	Character int `json:"ch"`
}

type Report struct {
	Message   string         `json:"message"`
	CheckName string         `json:"check_name"`
	Severity  ReportSeverity `json:"severity"`
	From      Position       `json:"from"`
	To        Position       `json:"to"`
}

type PlaygroundConfig struct {
	EnableNotice  bool   `json:"enable_notice"`
	EnableWarning bool   `json:"enable_warning"`
	EnableError   bool   `json:"enable_error"`
	StrictMixed   bool   `json:"strict_mixed"`
	PHP7          bool   `json:"enable_php7"`
	UnusedVar     string `json:"unused_var"`
}

type PlaygroundLinter struct {
	Rules     []*rules.Set
	StubsInfo *meta.Info
	Flags     cmd.ParsedFlags
	Config    PlaygroundConfig
}

func parseFile(worker *linter.Worker, file workspace.FileInfo) linter.ParseResult {
	var err error
	var result linter.ParseResult
	if worker.MetaInfo().IsIndexingComplete() {
		result, err = worker.ParseContents(file)
	} else {
		err = worker.IndexFile(file)
	}
	if err != nil {
		// log.Fatalf("could not parse %s: %v", file.Name, err.Error())
	}

	return result
}

func (pl *PlaygroundLinter) parse(filename string, contents string) []*linter.Report {
	version := "8.1"
	if pl.Config.PHP7 {
		version = "7.4"
	}
	mainLinter := linter.NewLinterWithInfo(linter.NewConfig(version), pl.StubsInfo.Clone())

	mainLinter.Config().ApplyQuickFixes = pl.Flags.ApplyQuickFixes
	mainLinter.Config().StrictMixed = pl.Config.StrictMixed

	runner := cmd.NewLinterRunner(mainLinter, linter.NewCheckersFilter())
	if err := runner.Init(pl.Rules, &pl.Flags); err != nil {
		log.Fatalf("init: %v", err)
	}

	indexing := mainLinter.NewIndexingWorker(0)

	file := workspace.FileInfo{
		Name:     filename,
		Contents: []byte(contents),
	}

	parseFile(indexing, file)

	mainLinter.MetaInfo().SetIndexingComplete(true)

	linting := mainLinter.NewLintingWorker(0)

	result := parseFile(linting, file)

	var reports []*linter.Report
	reports = append(reports, result.Reports...)

	return reports
}

func (pl *PlaygroundLinter) getReports(contents string) (reports []Report, err error) {
	rawReports := pl.parse(`demo.php`, contents)

	for _, report := range rawReports {
		reports = append(reports, Report{
			Message:   report.Message,
			CheckName: report.CheckName,
			Severity:  ReportSeverity(report.Level - 1),
			From: Position{
				Line:      report.Line - 1,
				Character: report.StartChar,
			},
			To: Position{
				Line:      report.Line - 1,
				Character: report.EndChar,
			},
		})
	}

	return reports, err
}

func main() {
	go linter.MemoryLimiterThread(20 * 1024 * 1024)

	stubsLinter := linter.NewLinter(linter.NewConfig("8.1"))

	fmt.Println("Load stubs")

	stubs := []string{
		"stubs/phpstorm-stubs/Core/Core.php",
		"stubs/phpstorm-stubs/Core/Core_c.php",
		"stubs/phpstorm-stubs/Core/Core_d.php",

		"stubs/phpstorm-stubs/standard/standard_0.php",
		"stubs/phpstorm-stubs/standard/standard_1.php",
		"stubs/phpstorm-stubs/standard/standard_2.php",
		"stubs/phpstorm-stubs/standard/standard_3.php",
		"stubs/phpstorm-stubs/standard/standard_4.php",
		"stubs/phpstorm-stubs/standard/standard_5.php",
		"stubs/phpstorm-stubs/standard/standard_6.php",
		"stubs/phpstorm-stubs/standard/standard_7.php",
		"stubs/phpstorm-stubs/standard/standard_8.php",
		"stubs/phpstorm-stubs/standard/standard_9.php",
		"stubs/phpstorm-stubs/standard/standard_defines.php",

		"stubs/phpstorm-stubs/pcntl/pcntl.php",
		"stubs/phpstorm-stubs/pcov/pcov.php",
		"stubs/phpstorm-stubs/pcre/pcre.php",

		"stubs/phpstorm-stubs/PDO/PDO.php",
		"stubs/phpstorm-stubs/bcmath/bcmath.php",

		"stubs/phpstorm-stubs/crypto/crypto.php",
		"stubs/phpstorm-stubs/ctype/ctype.php",
		"stubs/phpstorm-stubs/curl/curl.php",
		"stubs/phpstorm-stubs/curl/curl_d.php",
		"stubs/phpstorm-stubs/date/date.php",
		"stubs/phpstorm-stubs/date/date_c.php",
		"stubs/phpstorm-stubs/date/date_d.php",
	}

	now := time.Now()
	err := cmd.LoadEmbeddedStubs(stubsLinter, stubs)
	if err != nil {
		log.Printf("Init stubs error: %v", err)
	}
	fmt.Println("Time: ", time.Since(now))

	fmt.Println("Load rules")
	now = time.Now()
	ruleSets, err := cmd.ParseEmbeddedRules()
	if err != nil {
		log.Fatalf("preload embedded rules: %v", err)
	}
	for _, rset := range ruleSets {
		stubsLinter.Config().Checkers.DeclareRules(rset)
	}
	fmt.Println("Time: ", time.Since(now))

	pl := &PlaygroundLinter{
		Rules:     ruleSets,
		StubsInfo: stubsLinter.MetaInfo(),
		Flags: cmd.ParsedFlags{
			AllowAll:         true,
			UnusedVarPattern: "^_$",
			ApplyQuickFixes:  false,
		},
		Config: PlaygroundConfig{
			EnableNotice:  true,
			EnableWarning: true,
			EnableError:   true,
			StrictMixed:   false,
		},
	}

	playgroundInstance := js.Global().Get("playground")
	wasmInstance := js.Global().Get("wasm")
	wasmProps := wasmInstance.Get("props")

	wasmProps.Set("analyzeCallback", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		config := PlaygroundConfig{}
		configJSON := wasmProps.Get("configJson").String()
		fmt.Println("Config JSON: ", configJSON)
		err := json.Unmarshal([]byte(configJSON), &config)
		if err != nil {
			fmt.Printf("Fail parse config %s: %v", configJSON, err)
		} else {
			pl.Config = config
		}

		text := playgroundInstance.Call("getCode").String()

		if pl.Config.UnusedVar != "" {
			fmt.Println("New unused var: ", pl.Config.UnusedVar)
			pl.Flags.UnusedVarPattern = pl.Config.UnusedVar
		}

		reports, err := pl.getReports(text)

		filtered := make([]Report, 0, len(reports))
		for _, report := range reports {
			if !pl.Config.EnableNotice && report.Severity == Notice {
				continue
			}
			if !pl.Config.EnableWarning && report.Severity == Warning {
				continue
			}
			if !pl.Config.EnableError && report.Severity == Error {
				continue
			}

			filtered = append(filtered, report)
		}

		var value string
		if err != nil {
			value = "Error: " + err.Error()
		} else {
			m, _ := json.Marshal(filtered)
			value = string(m)
		}

		wasmProps.Set("reportsJson", value)
		return nil
	}))

	select {}
}
