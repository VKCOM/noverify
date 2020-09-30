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

	"github.com/VKCOM/noverify/src/git"
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
func (w *Worker) ParseContents(filename string, contents []byte, lineRanges []git.LineRange) (root *ir.Root, walker *RootWalker, err error) {
	defer func() {
		if r := recover(); r != nil {
			s := fmt.Sprintf("Panic while parsing %s: %s\n\nStack trace: %s", filename, r, dbg.Stack())
			log.Print(s)
			err = errors.New(s)
		}
	}()

	start := time.Now()

	// TODO: Ragel lexer can handle non-UTF8 input.
	// We can simplify code below and read from files directly.

	var rd inputs.ReadCloseSizer
	if contents == nil {
		rd, err = SrcInput.NewReader(filename)
	} else {
		rd, err = SrcInput.NewBytesReader(filename, contents)
	}
	if err != nil {
		log.Panicf("open source input: %v", err)
	}
	defer rd.Close()

	b := w.ctx.scratchBuf
	b.Reset()
	b.ReadFrom(rd)
	contents = append(make([]byte, 0, b.Len()), b.Bytes()...)

	waiter := BeforeParse(len(contents), filename)
	defer waiter.Finish()

	parser := php7.NewParser(contents)
	parser.WithFreeFloating()
	parser.Parse()

	atomic.AddInt64(&initParseTime, int64(time.Since(start)))

	return w.analyzeFile(filename, contents, parser, lineRanges)
}

// IndexFile parses the file and fills in the meta info. Can use cache.
func (w *Worker) IndexFile(filename string, contents []byte) error {
	if CacheDir == "" {
		_, w, err := w.ParseContents(filename, contents, nil)
		if w != nil {
			updateMetaInfo(filename, &w.meta)
		}
		return err
	}

	h := md5.New()

	if contents == nil {
		start := time.Now()
		fp, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fp.Close()
		if _, err := io.Copy(h, fp); err != nil {
			return err
		}
		atomic.AddInt64(&initFileReadTime, int64(time.Since(start)))
	} else {
		h.Write(contents)
	}

	contentsHash := fmt.Sprintf("%x", h.Sum(nil))

	cacheFilenamePart := filename

	volumeName := filepath.VolumeName(filename)

	// windows user supplied full path to directory to be analyzed,
	// but windows paths does not support ":" in the middle
	if len(volumeName) == 2 && volumeName[1] == ':' {
		cacheFilenamePart = filename[0:1] + "_" + filename[2:]
	}

	cacheFile := filepath.Join(CacheDir, cacheFilenamePart+"."+contentsHash)

	start := time.Now()
	fp, err := os.Open(cacheFile)
	if err != nil {
		_, w, err := w.ParseContents(filename, contents, nil)
		if err != nil {
			return err
		}

		return createMetaCacheFile(filename, cacheFile, w)
	}
	defer fp.Close()

	if err := restoreMetaFromCache(filename, fp); err != nil {
		// do not really care about why exactly reading from cache failed
		os.Remove(cacheFile)

		_, w, err := w.ParseContents(filename, contents, nil)
		if err != nil {
			return err
		}

		return createMetaCacheFile(filename, cacheFile, w)
	}

	atomic.AddInt64(&initCacheReadTime, int64(time.Since(start)))
	return nil
}

func (w *Worker) doParseFile(f workspace.FileInfo) []*Report {
	var err error

	if DebugParseDuration > 0 {
		start := time.Now()
		defer func() {
			if dur := time.Since(start); dur > DebugParseDuration {
				log.Printf("Parsing of %s took %s", f.Filename, dur)
			}
		}()
	}

	var reports []*Report

	if w.needReports {
		var walker *RootWalker
		_, walker, err = w.ParseContents(f.Filename, f.Contents, f.LineRanges)
		if err == nil {
			reports = walker.GetReports()
		}
	} else {
		err = w.IndexFile(f.Filename, f.Contents)
	}

	if err != nil {
		log.Printf("Failed parsing %s: %s", f.Filename, err.Error())
		lintdebug.Send("Failed parsing %s: %s", f.Filename, err.Error())
	}

	return reports
}

func (w *Worker) analyzeFile(filename string, contents []byte, parser *php7.Parser, lineRanges []git.LineRange) (*ir.Root, *RootWalker, error) {
	start := time.Now()
	rootNode := parser.GetRootNode()

	if rootNode == nil {
		lintdebug.Send("Could not parse %s at all due to errors", filename)
		return nil, nil, errors.New("Empty root node")
	}

	rootIR := w.irconv.ConvertRoot(rootNode)

	st := &meta.ClassParseState{CurrentFile: filename}
	walker := &RootWalker{
		lineRanges: lineRanges,
		ctx:        newRootContext(w.ctx, st),

		// We clone rules sets to remove all rules that
		// should not be applied to this file because of the @path.
		anyRset:   cloneRulesForFile(filename, Rules.Any),
		rootRset:  cloneRulesForFile(filename, Rules.Root),
		localRset: cloneRulesForFile(filename, Rules.Local),

		reVet: &regexpVet{
			parser: w.reParser,
		},
		reSimplifier: &regexpSimplifier{
			parser: w.reParserNoLiterals,
			out:    &strings.Builder{},
		},

		allowDisabledRegexp: w.AllowDisable,
	}

	walker.InitFromParser(contents, parser)
	walker.InitCustom()

	walker.beforeEnterFile()
	rootIR.Walk(walker)
	if meta.IsIndexingComplete() {
		AnalyzeFileRootLevel(rootIR, walker)
	}
	walker.afterLeaveFile()

	if len(walker.ctx.fixes) != 0 {
		if err := quickfix.Apply(filename, contents, walker.ctx.fixes); err != nil {
			linterError(filename, "apply quickfix: %v", err)
		}
	}

	for _, e := range parser.GetErrors() {
		walker.Report(nil, LevelError, "syntax", "Syntax error: "+e.String())
	}

	atomic.AddInt64(&initWalkTime, int64(time.Since(start)))

	return rootIR, walker, nil
}
