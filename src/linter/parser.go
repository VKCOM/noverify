package linter

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

func cloneRulesForFile(filename string, ruleSet *rules.ScopedSet) *rules.ScopedSet {
	if ruleSet == nil {
		return nil
	}

	var clone rules.ScopedSet
	for i, list := range &ruleSet.RulesByKind {
		res := make([]rules.Rule, 0, len(list))
		for _, rule := range list {
			if !strings.Contains(filename, rule.Path) {
				continue
			}
			res = append(res, rule)
		}
		clone.RulesByKind[i] = res
	}
	return &clone
}

// ParseFilenames is used to do initial parsing of files.
func ParseFilenames(readFileNamesFunc workspace.ReadCallback, allowDisabled *regexp.Regexp) []*Report {
	start := time.Now()
	defer func() {
		lintdebug.Send("Processing time: %s", time.Since(start))

		meta.Info.Lock()
		defer meta.Info.Unlock()

		lintdebug.Send("Funcs: %d, consts: %d, files: %d", meta.Info.NumFunctions(), meta.Info.NumConstants(), meta.Info.NumFilesWithFunctions())
	}()

	needReports := meta.IsIndexingComplete()

	lintdebug.Send("Parsing using %d cores", MaxConcurrency)

	filenamesCh := make(chan workspace.FileInfo, 512)

	go func() {
		readFileNamesFunc(filenamesCh)
		close(filenamesCh)
	}()

	var wg sync.WaitGroup
	reportsCh := make(chan []*Report, MaxConcurrency)

	wg.Add(MaxConcurrency)
	for i := 0; i < MaxConcurrency; i++ {
		go func(id int) {
			var w *Worker
			if needReports {
				w = NewLintingWorker(id)
			} else {
				w = NewIndexingWorker(id)
			}
			w.AllowDisable = allowDisabled
			var rep []*Report
			for f := range filenamesCh {
				rep = append(rep, w.doParseFile(f)...)
			}
			reportsCh <- rep
			wg.Done()
		}(i)
	}
	wg.Wait()

	var allReports []*Report
	for i := 0; i < MaxConcurrency; i++ {
		allReports = append(allReports, <-reportsCh...)
	}

	return allReports
}

func InitStubs(readFileNamesFunc workspace.ReadCallback) {
	meta.SetLoadingStubs(true)
	ParseFilenames(readFileNamesFunc, nil)
	meta.Info.InitStubs()
	if KPHP {
		meta.Info.InitKphpStubs()
	}
	meta.SetLoadingStubs(false)
}

// InitStubsFromDir parses directory with PHPStorm stubs which has all internal PHP classes and functions declared.
func InitStubsFromDir(dir string) {
	InitStubs(workspace.ReadFilenames([]string{dir}, nil))
}
