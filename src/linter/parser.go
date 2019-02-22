package linter

import (
	"bufio"
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
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/php7"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type FileInfo struct {
	Filename   string
	Contents   []byte
	LineRanges []git.LineRange
}

type ReadCallback func(ch chan FileInfo)

// ParseContents parses specified contents (or file) and returns *RootWalker.
// Function does not update global meta.
func ParseContents(filename string, contents []byte, encoding string, lineRanges []git.LineRange) (rootNode node.Node, w *RootWalker, err error) {
	defer func() {
		if r := recover(); r != nil {
			s := fmt.Sprintf("Panic while parsing %s: %s\n\nStack trace: %s", filename, r, dbg.Stack())
			log.Print(s)
			err = errors.New(s)
		}
	}()

	start := time.Now()

	var rd io.Reader
	var size int

	if contents == nil {
		fp, err := os.Open(filename)
		if err != nil {
			log.Fatalf("Could not open file %s: %s", filename, err.Error())
		}

		st, err := fp.Stat()
		if err != nil {
			log.Fatalf("Could not stat file %s: %s", filename, err.Error())
		}

		size = int(st.Size())
		rd = fp

		defer fp.Close()
	} else {
		rd = bytes.NewReader(contents)
		size = len(contents)
	}

	b := bytesBufPool.Get().(*bytes.Buffer)
	b.Reset()
	defer bytesBufPool.Put(b)

	if encoding == "windows-1251" {
		bufRd := bufPool.Get().(*bufio.Reader)
		bufRd.Reset(rd)
		rd = transform.NewReader(bufRd, charmap.Windows1251.NewDecoder())
		defer bufPool.Put(bufRd)
	}

	waiter := BeforeParse(size, filename)
	defer waiter.Finish()

	parser := php7.NewParser(io.TeeReader(rd, b), filename)
	parser.Parse()

	atomic.AddInt64(&initParseTime, int64(time.Since(start)))

	bufCopy := append(make([]byte, 0, b.Len()), b.Bytes()...)

	return analyzeFile(filename, bufCopy, parser, lineRanges)
}

func analyzeFile(filename string, contents []byte, parser *php7.Parser, lineRanges []git.LineRange) (rootNode node.Node, w *RootWalker, err error) {
	start := time.Now()
	rootNode = parser.GetRootNode()

	if rootNode == nil {
		lintdebug.Send("Could not parse %s at all due to errors", filename)
		return nil, nil, errors.New("Empty root node")
	}

	w = &RootWalker{
		filename:   filename,
		lineRanges: lineRanges,
		st:         &meta.ClassParseState{},
	}

	w.InitFromParser(contents, parser)
	w.InitCustom()

	rootNode.Walk(w)
	if meta.IsIndexingComplete() {
		AnalyzeFileRootLevel(rootNode, w)
	}

	for _, e := range parser.GetErrors() {
		w.Report(nil, LevelError, "Syntax error: "+e.String())
	}

	atomic.AddInt64(&initWalkTime, int64(time.Since(start)))

	return rootNode, w, nil
}

// AnalyzeFileRootLevel does analyze file top-level code.
// This method is exposed for language server use, you usually
// do not need to call it yourself.
func AnalyzeFileRootLevel(rootNode node.Node, d *RootWalker) {
	b := &BlockWalker{
		sc:                   meta.NewScope(),
		r:                    d,
		unusedVars:           make(map[string][]node.Node),
		nonLocalVars:         make(map[string]struct{}),
		ignoreFunctionBodies: true,
		rootLevel:            true,
	}

	for _, createFn := range d.customBlock {
		b.custom = append(b.custom, createFn(b))
	}

	rootNode.Walk(b)
}

var bufPool = sync.Pool{
	New: func() interface{} { return bufio.NewReaderSize(nil, 256<<10) },
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

			st, err := os.Stat(filename)
			if err != nil {
				log.Fatalf("Could not stat file %s: %s", filename, err.Error())
				continue
			}

			if !st.IsDir() {
				if ignoreRegex != nil && ignoreRegex.MatchString(filename) {
					continue
				}

				ch <- FileInfo{Filename: filename}
				continue
			}

			err = filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
				if !strings.HasSuffix(path, ".php") || info.IsDir() {
					return nil
				}

				if ignoreRegex != nil && ignoreRegex.MatchString(path) {
					return nil
				}

				ch <- FileInfo{Filename: path}
				return nil
			})

			if err != nil {
				log.Fatalf("Could not walk filepath %s", filename)
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

			if !strings.HasSuffix(c.NewName, ".php") {
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

	dotPHPBytes := []byte(".php")

	return func(ch chan FileInfo) {
		start := time.Now()
		idx := 0

		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				return bytes.HasSuffix(filename, dotPHPBytes)
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

	dotPHPBytes := []byte(".php")

	return func(ch chan FileInfo) {
		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				if !bytes.HasSuffix(filename, dotPHPBytes) {
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

	dotPHPBytes := []byte(".php")

	return func(ch chan FileInfo) {
		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				if !bytes.HasSuffix(filename, dotPHPBytes) {
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

	filenamesCh := make(chan FileInfo)

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

	if needReports {
		var w *RootWalker
		_, w, err = ParseContents(f.Filename, f.Contents, DefaultEncoding, f.LineRanges)
		if err == nil {
			reports = w.GetReports()
		}
	} else {
		err = Parse(f.Filename, f.Contents, DefaultEncoding)
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
