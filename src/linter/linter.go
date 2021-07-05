package linter

import (
	"regexp"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/workspace"
)

type Linter struct {
	config *Config

	info   *meta.Info
	checks *CheckersFilter
}

func NewLinter(config *Config) *Linter {
	return &Linter{
		config: config,
		info:   meta.NewInfo(),
		checks: NewCheckersFilterWithEnabledAll(),
	}
}

func (l *Linter) Config() *Config {
	return l.config
}

func (l *Linter) MetaInfo() *meta.Info {
	return l.info
}

func (l *Linter) UseCheckersFilter(checks *CheckersFilter) {
	l.checks = checks
}

func (l *Linter) NewLintingWorker(id int) *Worker {
	w := newWorker(l.config, l.info, id, l.checks)
	w.needReports = true
	return w
}

func (l *Linter) NewIndexingWorker(id int) *Worker {
	w := newWorker(l.config, l.info, id, l.checks)
	w.needReports = false
	return w
}

// AnalyzeFiles runs linter on the files that are provided by the readFileNamesFunc function.
func (l *Linter) AnalyzeFiles(readFileNamesFunc workspace.ReadCallback) []*Report {
	return l.analyzeFiles(readFileNamesFunc, l.config.AllowDisable)
}

func (l *Linter) analyzeFiles(readFileNamesFunc workspace.ReadCallback, allowDisable *regexp.Regexp) []*Report {
	start := time.Now()

	defer func() {
		lintdebug.Send("Processing time: %s", time.Since(start))

		l.info.Lock()
		defer l.info.Unlock()

		lintdebug.Send("Funcs: %d, consts: %d, files: %d",
			l.info.NumFunctions(), l.info.NumConstants(), l.info.NumFilesWithFunctions())
	}()

	needReports := l.info.IsIndexingComplete()

	lintdebug.Send("Parsing using %d cores", l.config.MaxConcurrency)

	filenamesCh := make(chan workspace.FileInfo, 512)

	go func() {
		readFileNamesFunc(filenamesCh)
		close(filenamesCh)
	}()

	var wg sync.WaitGroup
	reportsCh := make(chan []*Report, l.config.MaxConcurrency)

	wg.Add(l.config.MaxConcurrency)
	for i := 0; i < l.config.MaxConcurrency; i++ {
		go func(id int) {
			var w *Worker
			if needReports {
				w = l.NewLintingWorker(id)
			} else {
				w = l.NewIndexingWorker(id)
			}
			w.AllowDisable = allowDisable
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
	for i := 0; i < l.config.MaxConcurrency; i++ {
		allReports = append(allReports, <-reportsCh...)
	}

	return allReports
}

func (l *Linter) InitStubs(readFileNamesFunc workspace.ReadCallback) {
	l.info.SetLoadingStubs(true)
	l.analyzeFiles(readFileNamesFunc, nil)
	l.info.InitStubs()
	if l.config.KPHP {
		l.info.InitKphpStubs()
	}
	l.info.SetLoadingStubs(false)
}

// InitStubsFromDir parses directory with PHPStorm stubs which has all internal PHP classes and functions declared.
func (l *Linter) InitStubsFromDir(dir string) {
	l.InitStubs(workspace.ReadFilenames([]string{dir}, nil, l.config.PhpExtensions))
}
