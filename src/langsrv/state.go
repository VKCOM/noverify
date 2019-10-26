package langsrv

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/vscode"
)

type openedFile struct {
	rootNode       node.Node
	contents       string
	scopes         map[node.Node]*meta.Scope
	lines          [][]byte
	linesPositions []int
}

var (
	openMapMutex sync.Mutex
	openMap      = make(map[string]openedFile)

	changingMutex sync.Mutex
)

func openFile(filename, contents string) {
	changingMutex.Lock()
	defer changingMutex.Unlock()

	if meta.IsIndexingComplete() {
		changeFileNonLocked(filename, contents)
		return
	}

	// just parse file, do not fully analyze it as indexing is not yet done
	rootNode, _, err := linter.ParseContents(filename, []byte(contents), nil)
	if err != nil {
		log.Printf("Could not parse %s: %s", filename, err.Error())
		lintdebug.Send("Could not parse %s: %s", filename, err.Error())
		return
	}

	openMapMutex.Lock()
	openMap[filename] = openedFile{rootNode: rootNode, contents: contents}
	openMapMutex.Unlock()
}

// Handle changed contents of a file in the editor
func changeFile(filename, contents string) {
	changingMutex.Lock()
	defer changingMutex.Unlock()

	changeFileNonLocked(filename, contents)
}

func changeFileNonLocked(filename, contents string) {
	if !meta.IsIndexingComplete() {
		return
	}

	// parse file, update index for it, and then generate diagnostics based on new index
	meta.SetIndexingComplete(false)

	rootNode, w, err := linter.ParseContents(filename, []byte(contents), nil)
	if err != nil {
		log.Printf("Could not parse %s: %s", filename, err.Error())
		lintdebug.Send("Could not parse %s: %s", filename, err.Error())
		return
	}

	w.UpdateMetaInfo()

	meta.SetIndexingComplete(true)

	newWalker := linter.NewWalkerForLangServer(w)

	newWalker.InitCustom()
	rootNode.Walk(newWalker)
	linter.AnalyzeFileRootLevel(rootNode, newWalker)

	openMapMutex.Lock()
	f := openedFile{rootNode, contents, w.Scopes, w.Lines, w.LinesPositions}
	openMap[filename] = f
	openMapMutex.Unlock()

	flushReports(filename, newWalker)
}

// parse creations and changes of files concurrently
// changingMutex must be held
func concurrentParseChanges(changes []vscode.FileEvent) {
	filenamesCh := make(chan string)

	go func() {
		for _, ev := range changes {
			switch ev.Type {
			case vscode.Created, vscode.Changed:
				filenamesCh <- strings.TrimPrefix(ev.URI, "file://")
			}
		}
		close(filenamesCh)
	}()

	var wg sync.WaitGroup

	for i := 0; i < linter.MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			for filename := range filenamesCh {
				err := linter.IndexFile(filename, nil)
				if err != nil {
					lintdebug.Send("Could not parse %s: %s", filename, err.Error())
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func externalChanges(changes []vscode.FileEvent) {
	changingMutex.Lock()

	start := time.Now()
	lintdebug.Send("Started processing external changes %+v", changes)

	meta.SetIndexingComplete(false)

	meta.Info.Lock()
	for _, ev := range changes {
		switch ev.Type {
		case vscode.Deleted:
			meta.Info.DeleteMetaForFileNonLocked(strings.TrimPrefix(ev.URI, "file://"))
		}
	}
	meta.Info.Unlock()

	concurrentParseChanges(changes)

	changingMutex.Unlock()
	meta.SetIndexingComplete(true)

	// update currently opened files if needed
	for _, ev := range changes {
		filename := strings.TrimPrefix(ev.URI, "file://")
		switch ev.Type {
		case vscode.Created, vscode.Changed:
			openMapMutex.Lock()
			_, ok := openMap[filename]
			openMapMutex.Unlock()

			if !ok {
				break
			}

			contents, err := getFileContents(filename)
			if err != nil {
				lintdebug.Send("Could not read %s: %s", filename, err.Error())
				break
			}

			changeFile(filename, string(contents))
		}
	}

	lintdebug.Send("Finished processing %d external changes in %s", len(changes), time.Since(start))
}

// getFileContents reads specified file and returns UTF-8 encoded bytes.
func getFileContents(filename string) ([]byte, error) {
	r, err := linter.SrcInput.NewReader(filename)
	if err != nil {
		return nil, fmt.Errorf("open input: %v", err)
	}
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read input: %v", err)
	}
	return contents, nil
}

func flushReports(filename string, d *linter.RootWalker) {
	diag := d.Diagnostics
	if len(diag) == 0 && diag == nil {
		diag = make([]vscode.Diagnostic, 0)
	}

	writeMessage(&methodCall{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: &vscode.PublishDiagnosticsParams{
			URI:         "file://" + filename,
			Diagnostics: diag,
		},
	})
}

func closeFile(filename string) {
	openMapMutex.Lock()
	delete(openMap, filename)
	openMapMutex.Unlock()
}
