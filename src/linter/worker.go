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

	"github.com/VKCOM/noverify/src/inputs"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/workspace"
)

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
}

func NewLintingWorker(id int) *Worker {
	w := newWorker(id)
	w.needReports = true
	return w
}

func NewIndexingWorker(id int) *Worker {
	w := newWorker(id)
	w.needReports = false
	return w
}

func newWorker(id int) *Worker {
	ctx := NewWorkerContext()
	irConverter := irconv.NewConverter(ctx.phpdocTypeParser)
	return &Worker{
		id:     id,
		ctx:    ctx,
		irconv: irConverter,
		reParserNoLiterals: syntax.NewParser(&syntax.ParserOptions{
			NoLiterals: true,
		}),
		reParser: syntax.NewParser(&syntax.ParserOptions{
			NoLiterals: false,
		}),
	}
}

func (w *Worker) ID() int { return w.id }

// ParseContents parses specified contents (or file) and returns *RootWalker.
// Function does not update global meta.
func (w *Worker) ParseContents(fileInfo *workspace.FileInfo) (root *ir.Root, walker *RootWalker, err error) {
	defer func() {
		if r := recover(); r != nil {
			s := fmt.Sprintf("Panic while parsing %s: %s\n\nStack trace: %s", fileInfo.Name, r, dbg.Stack())
			log.Print(s)
			err = errors.New(s)
		}
	}()

	start := time.Now()

	// TODO: Ragel lexer can handle non-UTF8 input.
	// We can simplify code below and read from files directly.

	var rd inputs.ReadCloseSizer
	if fileInfo.Contents == nil {
		rd, err = SrcInput.NewReader(fileInfo.Name)
	} else {
		rd, err = SrcInput.NewBytesReader(fileInfo.Name, fileInfo.Contents)
	}
	if err != nil {
		log.Panicf("open source input: %v", err)
	}
	defer rd.Close()

	b := w.ctx.scratchBuf
	b.Reset()
	if _, err := b.ReadFrom(rd); err != nil {
		return nil, nil, err
	}
	contents := append(make([]byte, 0, b.Len()), b.Bytes()...)

	waiter := BeforeParse(len(contents), fileInfo.Name)
	defer waiter.Finish()

	parser := php7.NewParser(contents)
	parser.WithFreeFloating()
	parser.Parse()

	atomic.AddInt64(&initParseTime, int64(time.Since(start)))

	file := workspace.NewFile(fileInfo.Name, contents)
	return w.analyzeFile(file, parser)
}

// IndexFile parses the file and fills in the meta info. Can use cache.
func (w *Worker) IndexFile(file *workspace.FileInfo) error {
	if CacheDir == "" {
		_, w, err := w.ParseContents(file)
		if w != nil {
			updateMetaInfo(file.Name, &w.meta)
		}
		return err
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
	} else if _, err := h.Write(contents); err != nil {
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

	cacheFile := filepath.Join(CacheDir, cacheFilenamePart+"."+contentsHash)

	start := time.Now()
	fp, err := os.Open(cacheFile)
	if err != nil {
		_, w, err := w.ParseContents(file)
		if err != nil {
			return err
		}

		return createMetaCacheFile(file.Name, cacheFile, w)
	}
	defer fp.Close()

	if err := restoreMetaFromCache(file.Name, fp); err != nil {
		// do not really care about why exactly reading from cache failed
		os.Remove(cacheFile)

		_, w, err := w.ParseContents(file)
		if err != nil {
			return err
		}

		return createMetaCacheFile(file.Name, cacheFile, w)
	}

	atomic.AddInt64(&initCacheReadTime, int64(time.Since(start)))
	return nil
}

func (w *Worker) doParseFile(f *workspace.FileInfo) []*Report {
	var err error

	if DebugParseDuration > 0 {
		start := time.Now()
		defer func() {
			if dur := time.Since(start); dur > DebugParseDuration {
				log.Printf("Parsing of %s took %s", f.Name, dur)
			}
		}()
	}

	var reports []*Report

	if w.needReports {
		var walker *RootWalker
		_, walker, err = w.ParseContents(f)
		if err == nil {
			reports = walker.GetReports()
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

func (w *Worker) analyzeFile(file *workspace.File, parser *php7.Parser) (*ir.Root, *RootWalker, error) {
	start := time.Now()
	rootNode := parser.GetRootNode()

	if rootNode == nil {
		lintdebug.Send("Could not parse %s at all due to errors", file.Name())
		return nil, nil, errors.New("empty root node")
	}

	rootIR := w.irconv.ConvertRoot(rootNode)

	st := &meta.ClassParseState{CurrentFile: file.Name()}
	walker := &RootWalker{
		file: file,
		ctx:  newRootContext(w.ctx, st),

		// We clone rules sets to remove all rules that
		// should not be applied to this file because of the @path.
		anyRset:   cloneRulesForFile(file.Name(), Rules.Any),
		rootRset:  cloneRulesForFile(file.Name(), Rules.Root),
		localRset: cloneRulesForFile(file.Name(), Rules.Local),

		reVet: &regexpVet{
			parser: w.reParser,
		},
		reSimplifier: &regexpSimplifier{
			parser: w.reParserNoLiterals,
			out:    &strings.Builder{},
		},

		allowDisabledRegexp: w.AllowDisable,
	}

	walker.InitCustom()

	walker.beforeEnterFile()
	rootIR.Walk(walker)
	if meta.IsIndexingComplete() {
		AnalyzeFileRootLevel(rootIR, walker)
	}
	walker.afterLeaveFile()

	if len(walker.ctx.fixes) != 0 {
		if err := quickfix.Apply(file.Name(), file.Contents(), walker.ctx.fixes); err != nil {
			linterError(file.Name(), "apply quickfix: %v", err)
		}
	}

	for _, e := range parser.GetErrors() {
		walker.Report(nil, LevelError, "syntax", "Syntax error: "+e.String())
	}

	atomic.AddInt64(&initWalkTime, int64(time.Since(start)))

	return rootIR, walker, nil
}
