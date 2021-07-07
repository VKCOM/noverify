package linter

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	dbg "runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/quasilyte/regex/syntax"
	"github.com/z7zmey/php-parser/pkg/ast"
	"github.com/z7zmey/php-parser/pkg/conf"
	phperrors "github.com/z7zmey/php-parser/pkg/errors"
	"github.com/z7zmey/php-parser/pkg/parser"

	"github.com/VKCOM/noverify/src/inputs"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/workspace"
)

type ParseResult struct {
	RootNode *ir.Root
	Reports  []*Report

	walker *rootWalker
}

// Worker is a linter handle that is expected to be executed in a single goroutine context.
//
// It's not thread-safe and contains the state that will be re-used between the linter API calls.
//
// See NewLintingWorker and NewIndexingWorker.
type Worker struct {
	id  int
	ctx *WorkerContext

	irconv *irconv.Converter

	reParserNoLiterals *syntax.Parser
	reParser           *syntax.Parser

	needReports bool

	AllowDisable *regexp.Regexp

	config         *Config
	checkersFilter *CheckersFilter
	info           *meta.Info
}

func newWorker(config *Config, info *meta.Info, id int, checkersFilter *CheckersFilter) *Worker {
	ctx := NewWorkerContext()
	irConverter := irconv.NewConverter(ctx.phpdocTypeParser)
	return &Worker{
		config: config,
		info:   info,
		id:     id,
		ctx:    ctx,
		irconv: irConverter,
		reParserNoLiterals: syntax.NewParser(&syntax.ParserOptions{
			NoLiterals: true,
		}),
		reParser: syntax.NewParser(&syntax.ParserOptions{
			NoLiterals: false,
		}),
		checkersFilter: checkersFilter,
	}
}

func (w *Worker) ID() int { return w.id }

func (w *Worker) MetaInfo() *meta.Info { return w.info }

// ParseContents parses specified contents (or file) and returns *RootWalker.
// Function does not update global meta.
func (w *Worker) ParseContents(fileInfo workspace.FileInfo) (result ParseResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			s := fmt.Sprintf("Panic while parsing %s: %s\n\nStack trace: %s", fileInfo.Name, r, dbg.Stack())
			log.Print(s)
			err = errors.New(s)
		}
	}()

	// TODO: Ragel lexer can handle non-UTF8 input.
	// We can simplify code below and read from files directly.

	var rd inputs.ReadCloseSizer
	if fileInfo.Contents == nil {
		rd, err = w.config.SrcInput.NewReader(fileInfo.Name)
	} else {
		rd, err = w.config.SrcInput.NewBytesReader(fileInfo.Name, fileInfo.Contents)
	}
	if err != nil {
		log.Panicf("open source input: %v", err)
	}
	defer rd.Close()

	b := w.ctx.scratchBuf
	b.Reset()
	if _, err := b.ReadFrom(rd); err != nil {
		return result, err
	}
	contents := append(make([]byte, 0, b.Len()), b.Bytes()...)

	waiter := beforeParse(len(contents), fileInfo.Name)
	defer waiter.Finish()

	var parserErrors []*phperrors.Error
	rootNode, err := parser.Parse(contents, conf.Config{
		Version: w.config.PhpVersion,
		ErrorHandlerFunc: func(e *phperrors.Error) {
			parserErrors = append(parserErrors, e)
		},
	})
	if err != nil {
		return ParseResult{}, fmt.Errorf("parse error: %v", err.Error())
	}

	if rootNode == nil {
		return result, fmt.Errorf("file has incorrect syntax and cannot be parsed")
	}

	rootIR := w.irconv.ConvertRoot(rootNode.(*ast.Root))

	file := workspace.NewFile(fileInfo.Name, contents)
	walker, err := w.analyzeFile(file, rootIR)
	if err != nil {
		return result, err
	}

	for _, e := range parserErrors {
		walker.Report(nil, LevelError, "syntax", "Syntax error: "+e.String())
	}

	result = ParseResult{
		RootNode: rootIR,
		Reports:  walker.reports,
		walker:   walker,
	}
	return result, nil
}

func (w *Worker) parseWithCache(cacheFilename string, file workspace.FileInfo) error {
	result, err := w.ParseContents(file)
	if err != nil {
		return err
	}
	return createMetaCacheFile(file.Name, cacheFilename, result.walker)
}

// IndexFile parses the file and fills in the meta info. Can use cache.
func (w *Worker) IndexFile(file workspace.FileInfo) error {
	if w.config.CacheDir == "" {
		result, err := w.ParseContents(file)
		if err != nil {
			return err
		}
		if w != nil {
			updateMetaInfo(w.info, file.Name, &result.walker.meta)
		}
		return nil
	}

	h := md5.New()

	if file.Contents == nil {
		start := time.Now()
		fp, err := os.Open(file.Name)
		if err != nil {
			return err
		}
		defer fp.Close()
		if _, err := io.Copy(h, fp); err != nil {
			return err
		}
		atomic.AddInt64(&initFileReadTime, int64(time.Since(start)))
	} else if _, err := h.Write(file.Contents); err != nil {
		return err
	}

	contentsHash := fmt.Sprintf("%x", h.Sum(nil))

	cacheFilenamePart := file.Name

	volumeName := filepath.VolumeName(file.Name)

	// windows user supplied full path to directory to be analyzed,
	// but windows paths does not support ":" in the middle
	if len(volumeName) == 2 && volumeName[1] == ':' {
		cacheFilenamePart = file.Name[0:1] + "_" + file.Name[2:]
	}

	cacheFile := filepath.Join(w.config.CacheDir, cacheFilenamePart+"."+contentsHash)

	start := time.Now()

	fp, err := os.Open(cacheFile)
	if err != nil {
		return w.parseWithCache(cacheFile, file)
	}
	defer fp.Close()

	if err := restoreMetaFromCache(w.info, w.config.Checkers.cachers, file.Name, fp); err != nil {
		// do not really care about why exactly reading from cache failed
		os.Remove(cacheFile)
		return w.parseWithCache(cacheFile, file)
	}

	atomic.AddInt64(&initCacheReadTime, int64(time.Since(start)))
	return nil
}

func (w *Worker) doParseFile(f workspace.FileInfo) []*Report {
	var err error

	if w.config.DebugParseDuration > 0 {
		start := time.Now()
		defer func() {
			if dur := time.Since(start); dur > w.config.DebugParseDuration {
				log.Printf("Parsing of %s took %s", f.Name, dur)
			}
		}()
	}

	var reports []*Report

	if w.needReports {
		var result ParseResult
		result, err = w.ParseContents(f)
		if err == nil {
			reports = result.Reports
		}
	} else {
		err = w.IndexFile(f)
	}

	if err != nil {
		log.Printf("Failed parsing %s: %s", f.Name, err.Error())
		lintdebug.Send("Failed parsing %s: %s", f.Name, err.Error())
	}

	return reports
}

func (w *Worker) analyzeFile(file *workspace.File, rootNode *ir.Root) (*rootWalker, error) {
	if rootNode == nil {
		lintdebug.Send("Could not parse %s at all due to errors", file.Name())
		return nil, errors.New("empty root node")
	}

	st := &meta.ClassParseState{Info: w.info, CurrentFile: file.Name()}
	walker := &rootWalker{
		config: w.config,
		file:   file,
		ctx:    newRootContext(w.config, w.ctx, st),

		// We clone rules sets to remove all rules that
		// should not be applied to this file because of the @path.
		anyRset:   cloneRulesForFile(file.Name(), w.config.Rules.Any),
		rootRset:  cloneRulesForFile(file.Name(), w.config.Rules.Root),
		localRset: cloneRulesForFile(file.Name(), w.config.Rules.Local),

		reVet: &regexpVet{
			parser: w.reParser,
		},
		reSimplifier: &regexpSimplifier{
			parser: w.reParserNoLiterals,
			out:    &strings.Builder{},
		},

		allowDisabledRegexp: w.AllowDisable,
		checkersFilter:      w.checkersFilter,
	}

	walker.InitCustom()

	walker.beforeEnterFile()
	rootNode.Walk(walker)
	if w.info.IsIndexingComplete() {
		analyzeFileRootLevel(rootNode, walker)
	}
	walker.afterLeaveFile()

	if len(walker.ctx.fixes) != 0 {
		needApplyFixes := !file.AutoGenerated() || w.config.CheckAutoGenerated

		if needApplyFixes {
			if err := quickfix.Apply(file.Name(), file.Contents(), walker.ctx.fixes); err != nil {
				linterError(file.Name(), "apply quickfix: %v", err)
			}
		}
	}

	return walker, nil
}

// analyzeFileRootLevel does analyze file top-level code.
// This method is exposed for language server use, you usually
// do not need to call it yourself.
func analyzeFileRootLevel(rootNode ir.Node, d *rootWalker) {
	sc := meta.NewScope()
	sc.AddVarName("argv", types.NewMap("string[]"), "predefined", meta.VarAlwaysDefined)
	sc.AddVarName("argc", types.NewMap("int"), "predefined", meta.VarAlwaysDefined)

	b := newBlockWalker(d, sc)
	b.ignoreFunctionBodies = true
	b.rootLevel = true

	for _, createFn := range d.customBlock {
		b.custom = append(b.custom, createFn(&BlockContext{w: b}))
	}

	rootNode.Walk(b)
}
