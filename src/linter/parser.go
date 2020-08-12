package linter

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	dbg "runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/karrick/godirwalk"
	"github.com/monochromegane/go-gitignore"
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
)

type FileInfo struct {
	Filename   string
	Contents   []byte
	LineRanges []git.LineRange
}

func isPHPExtension(filename string) bool {
	fileExt := filepath.Ext(filename)
	if fileExt == "" {
		return false
	}

	fileExt = fileExt[1:] // cut "." in the beginning

	for _, ext := range PHPExtensions {
		if fileExt == ext {
			return true
		}
	}

	return false
}

func makePHPExtensionSuffixes() [][]byte {
	res := make([][]byte, 0, len(PHPExtensions))
	for _, ext := range PHPExtensions {
		res = append(res, []byte("."+ext))
	}
	return res
}

func isPHPExtensionBytes(filename []byte, suffixes [][]byte) bool {
	for _, suffix := range suffixes {
		if bytes.HasSuffix(filename, suffix) {
			return true
		}
	}

	return false
}

type ReadCallback func(ch chan FileInfo)

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

// ParseGitignoreFromDir tries to parse a gitignore file at path/.gitignore.
// If no such file exists, <nil, nil> is returned.
func ParseGitignoreFromDir(path string) (gitignore.IgnoreMatcher, error) {
	f, err := os.Open(filepath.Join(path, ".gitignore"))
	switch {
	case os.IsNotExist(err):
		return nil, nil // No gitignore file, not an error
	case err != nil:
		return nil, err // Some unexpected error (e.g. access failure)
	}
	defer f.Close()
	matcher := gitignore.NewGitIgnoreFromReader(path, f)
	return matcher, nil
}

func readFilenames(ch chan<- FileInfo, filename string, filter *FilenameFilter) {
	absFilename, err := filepath.Abs(filename)
	if err == nil {
		filename = absFilename
	}

	if filter == nil {
		// No-op filter that doesn't track gitignore files.
		filter = &FilenameFilter{}
	}

	// If we use stat here, it will return file info of an entry
	// pointed by a symlink (if filename is a link).
	// lstat is required for a symlink test below to succeed.
	// If we ever want to permit top-level (CLI args) symlinks,
	// caller should resolve them to a files that are pointed by them.
	st, err := os.Lstat(filename)
	if err != nil {
		log.Fatalf("Could not stat file %s: %s", filename, err.Error())
	}
	if st.Mode()&os.ModeSymlink != 0 {
		// filepath.Walk does not follow symlinks, but it does
		// accept it as a root argument without an error.
		// godirwalk.Walk can traverse symlinks with FollowSymbolicLinks=true,
		// but we don't use it. It will give an error if root is
		// a symlink, so we avoid calling Walk() on them.
		return
	}

	if !st.IsDir() {
		if filter.IgnoreFile(filename) {
			return
		}

		ch <- FileInfo{Filename: filename}
		return
	}

	// Start with a sentinel "" path to make last(gitignorePaths) safe
	// without a length check.
	gitignorePaths := []string{""}

	walkOptions := &godirwalk.Options{
		Unsorted: true,

		Callback: func(path string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				if filter.IgnoreDir(path) {
					return filepath.SkipDir
				}
				// During indexing phase and with -gitignore=false
				// we don't want to do extra FS operations.
				if !filter.GitignoreIsEnabled() {
					return nil
				}

				matcher, err := ParseGitignoreFromDir(path)
				if err != nil {
					linterError(path, "read .gitignore: %v", err)
				}
				if matcher != nil {
					gitignorePaths = append(gitignorePaths, path)
					filter.GitignorePush(path, matcher)
				}
				return nil
			}

			if !isPHPExtension(path) {
				return nil
			}
			if filter.IgnoreFile(path) {
				return nil
			}

			ch <- FileInfo{Filename: path}
			return nil
		},
	}

	if filter.GitignoreIsEnabled() {
		walkOptions.PostChildrenCallback = func(path string, de *godirwalk.Dirent) error {
			topGitignorePath := gitignorePaths[len(gitignorePaths)-1]
			if topGitignorePath == path {
				gitignorePaths = gitignorePaths[:len(gitignorePaths)-1]
				filter.GitignorePop(path)
			}
			return nil
		}
	}

	if err := godirwalk.Walk(filename, walkOptions); err != nil {
		log.Fatalf("Could not walk filepath %s (%v)", filename, err)
	}
}

// ReadFilenames returns callback that reads filenames into channel
func ReadFilenames(filenames []string, filter *FilenameFilter) ReadCallback {
	return func(ch chan FileInfo) {
		for _, filename := range filenames {
			readFilenames(ch, filename, filter)
		}
	}
}

// ReadChangesFromWorkTree returns callback that reads files from workTree dir that are changed
func ReadChangesFromWorkTree(dir string, changes []git.Change) ReadCallback {
	return func(ch chan FileInfo) {
		for _, c := range changes {
			if c.Type == git.Deleted {
				continue
			}

			if !isPHPExtension(c.NewName) {
				continue
			}

			filename := filepath.Join(dir, c.NewName)

			contents, err := ioutil.ReadFile(filename)
			if err != nil {
				log.Fatalf("Could not read file %s: %s", filename, err.Error())
			}

			ch <- FileInfo{
				Filename: filename,
				Contents: contents,
			}
		}
	}
}

// ReadFilesFromGit parses file contents in the specified commit
func ReadFilesFromGit(repo, commitSHA1 string, ignoreRegex *regexp.Regexp) ReadCallback {
	catter, err := git.NewCatter(repo)
	if err != nil {
		log.Fatalf("Could not start catter: %s", err.Error())
	}

	tree, err := git.GetTreeSHA1(catter, commitSHA1)
	if err != nil {
		log.Fatalf("Could not get tree sha1: %s", err.Error())
	}

	suffixes := makePHPExtensionSuffixes()

	return func(ch chan FileInfo) {
		start := time.Now()
		idx := 0

		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				return isPHPExtensionBytes(filename, suffixes)
			},
			func(filename string, contents []byte) {
				idx++
				if time.Since(start) >= 2*time.Second {
					start = time.Now()
					action := "Indexed"
					if meta.IsIndexingComplete() {
						action = "Analyzed"
					}
					log.Printf("%s %d files from git", action, idx)
				}

				if ignoreRegex != nil && ignoreRegex.MatchString(filename) {
					return
				}

				ch <- FileInfo{
					Filename: filename,
					Contents: contents,
				}
			},
		)

		if err != nil {
			log.Fatalf("Could not walk: %s", err.Error())
		}
	}
}

// ReadOldFilesFromGit parses file contents in the specified commit, the old version
func ReadOldFilesFromGit(repo, commitSHA1 string, changes []git.Change) ReadCallback {
	changedMap := make(map[string][]git.LineRange, len(changes))
	for _, ch := range changes {
		if ch.Type == git.Added {
			continue
		}
		changedMap[ch.OldName] = append(changedMap[ch.OldName], ch.OldLineRanges...)
	}

	catter, err := git.NewCatter(repo)
	if err != nil {
		log.Fatalf("Could not start catter: %s", err.Error())
	}

	tree, err := git.GetTreeSHA1(catter, commitSHA1)
	if err != nil {
		log.Fatalf("Could not get tree sha1: %s", err.Error())
	}

	suffixes := makePHPExtensionSuffixes()

	return func(ch chan FileInfo) {
		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				if !isPHPExtensionBytes(filename, suffixes) {
					return false
				}

				_, ok := changedMap[string(filename)]
				return ok
			},
			func(filename string, contents []byte) {
				ch <- FileInfo{
					Filename:   filename,
					Contents:   contents,
					LineRanges: changedMap[filename],
				}
			},
		)

		if err != nil {
			log.Fatalf("Could not walk: %s", err.Error())
		}
	}
}

// ReadFilesFromGitWithChanges parses file contents in the specified commit, but only specified ranges
func ReadFilesFromGitWithChanges(repo, commitSHA1 string, changes []git.Change) ReadCallback {
	changedMap := make(map[string][]git.LineRange, len(changes))
	for _, ch := range changes {
		if ch.Type == git.Deleted {
			// TODO: actually support deletes too
			continue
		}

		changedMap[ch.NewName] = append(changedMap[ch.NewName], ch.LineRanges...)
	}

	catter, err := git.NewCatter(repo)
	if err != nil {
		log.Fatalf("Could not start catter: %s", err.Error())
	}

	tree, err := git.GetTreeSHA1(catter, commitSHA1)
	if err != nil {
		log.Fatalf("Could not get tree sha1: %s", err.Error())
	}

	suffixes := makePHPExtensionSuffixes()

	return func(ch chan FileInfo) {
		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				if !isPHPExtensionBytes(filename, suffixes) {
					return false
				}

				_, ok := changedMap[string(filename)]
				return ok
			},
			func(filename string, contents []byte) {
				ch <- FileInfo{
					Filename:   filename,
					Contents:   contents,
					LineRanges: changedMap[filename],
				}
			},
		)

		if err != nil {
			log.Fatalf("Could not walk: %s", err.Error())
		}
	}
}

// ParseFilenames is used to do initial parsing of files.
func ParseFilenames(readFileNamesFunc ReadCallback, allowDisabled *regexp.Regexp) []*Report {
	start := time.Now()
	defer func() {
		lintdebug.Send("Processing time: %s", time.Since(start))

		meta.Info.Lock()
		defer meta.Info.Unlock()

		lintdebug.Send("Funcs: %d, consts: %d, files: %d", meta.Info.NumFunctions(), meta.Info.NumConstants(), meta.Info.NumFilesWithFunctions())
	}()

	needReports := meta.IsIndexingComplete()

	lintdebug.Send("Parsing using %d cores", MaxConcurrency)

	filenamesCh := make(chan FileInfo, 512)

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

func doParseFile(f FileInfo, needReports bool, allowDisabled *regexp.Regexp) (reports []*Report) {
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

func InitStubs(readFileNamesFunc ReadCallback) {
	meta.SetLoadingStubs(true)
	ParseFilenames(readFileNamesFunc, nil)
	meta.Info.InitStubs()
	meta.SetLoadingStubs(false)
}

// InitStubsFromDir parses directory with PHPStorm stubs which has all internal PHP classes and functions declared.
func InitStubsFromDir(dir string) {
	InitStubs(ReadFilenames([]string{dir}, nil))
}
