package linter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/inputs"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/karrick/godirwalk"
	"github.com/quasilyte/regex/syntax"
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
func ParseContents(filename string, contents []byte, lineRanges []git.LineRange) (rootNode node.Node, w *RootWalker, err error) {
	defer func() {
		if r := recover(); r != nil {
			s := fmt.Sprintf("Panic while parsing %s: %s\n\nStack trace: %s", filename, r, dbg.Stack())
			log.Print(s)
			err = errors.New(s)
		}
	}()

	start := time.Now()

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

	waiter := BeforeParse(rd.Size(), filename)
	defer waiter.Finish()

	parser := php7.NewParser(io.TeeReader(rd, b), filename)
	parser.WithFreeFloating()
	parser.Parse()

	atomic.AddInt64(&initParseTime, int64(time.Since(start)))

	bufCopy := append(make([]byte, 0, b.Len()), b.Bytes()...)

	return analyzeFile(filename, bufCopy, parser, lineRanges)
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
			ruleClone := rule
			ruleClone.Matcher = rule.Matcher.Clone()
			res = append(res, ruleClone)
		}
		clone.RulesByKind[i] = res
	}
	return &clone
}

func analyzeFile(filename string, contents []byte, parser *php7.Parser, lineRanges []git.LineRange) (*node.Root, *RootWalker, error) {
	start := time.Now()
	rootNode := parser.GetRootNode()

	if rootNode == nil {
		lintdebug.Send("Could not parse %s at all due to errors", filename)
		return nil, nil, errors.New("Empty root node")
	}

	w := &RootWalker{
		lineRanges: lineRanges,
		st:         &meta.ClassParseState{CurrentFile: filename},

		// We need to clone rules since phpgrep matchers
		// contain mutable state that we don't want to share
		// between goroutines.
		anyRset:   cloneRulesForFile(filename, Rules.Any),
		rootRset:  cloneRulesForFile(filename, Rules.Root),
		localRset: cloneRulesForFile(filename, Rules.Local),

		reSimplifier: &regexpSimplifier{
			parser: syntax.NewParser(&syntax.ParserOptions{
				NoLiterals: true,
			}),
			out: &strings.Builder{},
		},
	}

	w.InitFromParser(contents, parser)
	w.InitCustom()

	rootNode.Walk(w)
	if meta.IsIndexingComplete() {
		AnalyzeFileRootLevel(rootNode, w)
	}
	for _, c := range w.custom {
		c.AfterLeaveFile()
	}

	for _, e := range parser.GetErrors() {
		w.Report(nil, LevelError, "syntax", "Syntax error: "+e.String())
	}

	atomic.AddInt64(&initWalkTime, int64(time.Since(start)))

	return rootNode, w, nil
}

// AnalyzeFileRootLevel does analyze file top-level code.
// This method is exposed for language server use, you usually
// do not need to call it yourself.
func AnalyzeFileRootLevel(rootNode node.Node, d *RootWalker) {
	sc := meta.NewScope()
	sc.AddVarName("argv", meta.NewTypesMap("string[]"), "predefined", true)
	sc.AddVarName("argc", meta.NewTypesMap("int"), "predefined", true)
	b := &BlockWalker{
		ctx:                  &blockContext{sc: sc},
		r:                    d,
		unusedVars:           make(map[string][]node.Node),
		nonLocalVars:         make(map[string]struct{}),
		ignoreFunctionBodies: true,
		rootLevel:            true,
	}

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

// ReadFilenames returns callback that reads filenames into channel
func ReadFilenames(filenames []string, ignoreRegex *regexp.Regexp) ReadCallback {
	return func(ch chan FileInfo) {
		for _, filename := range filenames {
			absFilename, err := filepath.Abs(filename)
			if err == nil {
				filename = absFilename
			}

			// If we use stat here, it will return file info of an entry
			// pointed by a symlink (if filename is a link).
			// lstat is required for a symlink test below to succeed.
			// If we ever want to permit top-level (CLI args) symlinks,
			// caller should resolve them to a files that are pointed by them.
			st, err := os.Lstat(filename)
			if err != nil {
				log.Fatalf("Could not stat file %s: %s", filename, err.Error())
				continue
			}
			if st.Mode()&os.ModeSymlink != 0 {
				// filepath.Walk does not follow symlinks, but it does
				// accept it as a root argument without an error.
				// godirwalk.Walk can traverse symlinks with FollowSymbolicLinks=true,
				// but we don't use it. It will give an error if root is
				// a symlink, so we avoid calling Walk() on them.
				continue
			}

			if !st.IsDir() {
				if ignoreRegex != nil && ignoreRegex.MatchString(filename) {
					continue
				}

				ch <- FileInfo{Filename: filename}
				continue
			}

			err = godirwalk.Walk(filename, &godirwalk.Options{
				Callback: func(path string, de *godirwalk.Dirent) error {
					if de.IsDir() || !isPHPExtension(path) {
						return nil
					}

					if ignoreRegex != nil && ignoreRegex.MatchString(path) {
						return nil
					}

					ch <- FileInfo{Filename: path}
					return nil
				},
				Unsorted: true,
			})

			if err != nil {
				log.Fatalf("Could not walk filepath %s (%v)", filename, err)
			}
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
func ParseFilenames(readFileNamesFunc ReadCallback) []*Report {
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
				rep = append(rep, doParseFile(f, needReports)...)
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

func doParseFile(f FileInfo, needReports bool) (reports []*Report) {
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
		_, w, err = ParseContents(f.Filename, f.Contents, f.LineRanges)
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

// InitStubs parses directory with PHPStorm stubs which has all internal PHP classes and functions declared.
func InitStubs() {
	ParseFilenames(ReadFilenames([]string{StubsDir}, nil))
	meta.Info.InitStubs()
}
