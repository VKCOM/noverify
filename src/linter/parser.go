package linter

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	dbg "runtime/debug"
	"strings"
	"sync"
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
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

// ParseContents parses specified contents (or file) and returns *RootWalker.
// Function does not update global meta.
func ParseContents(filename string, contents []byte, lineRanges []git.LineRange, allowDisabled *regexp.Regexp) (rootNode *ir.Root, w *RootWalker, err error) {
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

	b := bytesBufPool.Get().(*bytes.Buffer)
	b.Reset()
	defer bytesBufPool.Put(b)

	b.ReadFrom(rd)
	contents = append(make([]byte, 0, b.Len()), b.Bytes()...)

	waiter := BeforeParse(len(contents), filename)
	defer waiter.Finish()

	parser := php7.NewParser(contents)
	parser.WithFreeFloating()
	parser.Parse()

	atomic.AddInt64(&initParseTime, int64(time.Since(start)))

	return analyzeFile(filename, contents, parser, lineRanges, allowDisabled)
}

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

func analyzeFile(filename string, contents []byte, parser *php7.Parser, lineRanges []git.LineRange, allowedDisabled *regexp.Regexp) (*ir.Root, *RootWalker, error) {
	start := time.Now()
	rootNode := parser.GetRootNode()

	if rootNode == nil {
		lintdebug.Send("Could not parse %s at all due to errors", filename)
		return nil, nil, errors.New("Empty root node")
	}

	rootIR := irconv.ConvertRoot(rootNode)

	st := &meta.ClassParseState{CurrentFile: filename}
	w := &RootWalker{
		lineRanges: lineRanges,
		ctx:        newRootContext(st),

		// We clone rules sets to remove all rules that
		// should not be applied to this file because of the @path.
		anyRset:   cloneRulesForFile(filename, Rules.Any),
		rootRset:  cloneRulesForFile(filename, Rules.Root),
		localRset: cloneRulesForFile(filename, Rules.Local),

		reVet: &regexpVet{
			parser: syntax.NewParser(&syntax.ParserOptions{
				NoLiterals: false,
			}),
		},
		reSimplifier: &regexpSimplifier{
			parser: syntax.NewParser(&syntax.ParserOptions{
				NoLiterals: true,
			}),
			out: &strings.Builder{},
		},

		allowDisabledRegexp: allowedDisabled,
	}

	w.InitFromParser(contents, parser)
	w.InitCustom()

	rootIR.Walk(w)
	if meta.IsIndexingComplete() {
		AnalyzeFileRootLevel(rootIR, w)
	}
	w.afterLeaveFile()

	if len(w.ctx.fixes) != 0 {
		if err := quickfix.Apply(filename, contents, w.ctx.fixes); err != nil {
			linterError(filename, "apply quickfix: %v", err)
		}
	}

	for _, e := range parser.GetErrors() {
		w.Report(nil, LevelError, "syntax", "Syntax error: "+e.String())
	}

	atomic.AddInt64(&initWalkTime, int64(time.Since(start)))

	return rootIR, w, nil
}

// AnalyzeFileRootLevel does analyze file top-level code.
// This method is exposed for language server use, you usually
// do not need to call it yourself.
func AnalyzeFileRootLevel(rootNode ir.Node, d *RootWalker) {
	sc := meta.NewScope()
	sc.AddVarName("argv", meta.NewTypesMap("string[]"), "predefined", meta.VarAlwaysDefined)
	sc.AddVarName("argc", meta.NewTypesMap("int"), "predefined", meta.VarAlwaysDefined)

	b := newBlockWalker(d, sc)
	b.ignoreFunctionBodies = true
	b.rootLevel = true

	for _, createFn := range d.customBlock {
		b.custom = append(b.custom, createFn(&BlockContext{w: b}))
	}

	rootNode.Walk(b)
}

var bytesBufPool = sync.Pool{
	New: func() interface{} { return &bytes.Buffer{} },
}

// DebugMessage is used to actually print debug messages.
func DebugMessage(msg string, args ...interface{}) {
	if Debug {
		log.Printf(msg, args...)
	}
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

	for i := 0; i < MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			var rep []*Report
			for f := range filenamesCh {
				rep = append(rep, doParseFile(f, needReports, allowDisabled)...)
			}
			reportsCh <- rep
			wg.Done()
		}()
	}
	wg.Wait()

	var allReports []*Report
	for i := 0; i < MaxConcurrency; i++ {
		allReports = append(allReports, (<-reportsCh)...)
	}

	return allReports
}

func doParseFile(f workspace.FileInfo, needReports bool, allowDisabled *regexp.Regexp) (reports []*Report) {
	var err error

	if DebugParseDuration > 0 {
		start := time.Now()
		defer func() {
			if dur := time.Since(start); dur > DebugParseDuration {
				log.Printf("Parsing of %s took %s", f.Filename, dur)
			}
		}()
	}

	if needReports {
		var w *RootWalker
		_, w, err = ParseContents(f.Filename, f.Contents, f.LineRanges, allowDisabled)
		if err == nil {
			reports = w.GetReports()
		}
	} else {
		err = IndexFile(f.Filename, f.Contents)
	}

	if err != nil {
		log.Printf("Failed parsing %s: %s", f.Filename, err.Error())
		lintdebug.Send("Failed parsing %s: %s", f.Filename, err.Error())
	}

	return reports
}

func InitStubs(readFileNamesFunc workspace.ReadCallback) {
	meta.SetLoadingStubs(true)
	ParseFilenames(readFileNamesFunc, nil)
	meta.Info.InitStubs()
	meta.SetLoadingStubs(false)
}

// InitStubsFromDir parses directory with PHPStorm stubs which has all internal PHP classes and functions declared.
func InitStubsFromDir(dir string) {
	InitStubs(workspace.ReadFilenames([]string{dir}, nil))
}
