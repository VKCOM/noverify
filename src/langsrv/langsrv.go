package langsrv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	dbg "runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/vscode"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/php7"
)

const maxLength = 16 << 20

type baseRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      *int   `json:"id"`
	Method  string `json:"method"`
	Params  json.RawMessage
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id"`
	Result  interface{} `json:"result"`
}

type methodCall struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

func (response) IMessage()   {}
func (methodCall) IMessage() {}

var (
	respMutex sync.Mutex
	connWr    io.Writer
)

// RegisterDebug starts listening for debug events
func RegisterDebug() {
	lintdebug.Register(writeLog)
}

func writeLog(msg string) {
	writeMessage(&methodCall{
		JSONRPC: "2.0",
		Method:  "window/logMessage",
		Params: map[string]interface{}{
			"type":    3,
			"message": msg,
		},
	})
}

func writeMessage(message interface{ IMessage() }) error {
	respMutex.Lock()
	defer respMutex.Unlock()

	_, err := connWr.Write([]byte("Content-Type: application/vscode-jsonrpc; charset=utf8\r\n"))
	if err != nil {
		return err
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(connWr, "Content-Length: %d\r\n\r\n", len(data))
	if err != nil {
		return err
	}

	_, err = connWr.Write(data)
	return err
}

func handleMessage(buf []byte) error {
	defer func() {
		if r := recover(); r != nil {
			lintdebug.Send("Panic ocurred: %s, trace: %s", r, dbg.Stack())
		}
	}()

	var req baseRequest
	err := json.Unmarshal(buf, &req)
	if err != nil {
		return err
	}

	switch req.Method {
	case "initialize":
		return handleInitialize(&req)
	case "textDocument/didOpen":
		return handleTextDocumentDidOpen(&req)
	case "textDocument/didChange":
		return handleTextDocumentDidChange(&req)
	case "textDocument/definition":
		return handleTextDocumentDefinition(&req)
	case "textDocument/references":
		return handleTextDocumentReferences(&req)
	case "textDocument/didClose":
		return handleTextDocumentDidClose(&req)
	case "textDocument/completion":
		return handleTextDocumentCompletion(&req)
	case "textDocument/hover":
		return handleTextDocumentHover(&req)
	case "textDocument/documentSymbol":
		return handleTextDocumentSymbol(&req)
	case "workspace/didChangeWatchedFiles":
		return handleChangeWatchedFiles(&req)
	default:
		lintdebug.Send("Got %s, data: %s", req.Method, req.Params)
	}

	if req.ID == nil {
		return nil
	}

	return writeMessage(&response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result:  map[string]interface{}{},
	})
}

func handleInitialize(req *baseRequest) error {
	var params vscode.InitializeParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	lintdebug.Send("Root dir: %s", params.RootPath)

	go func() {
		linter.AnalysisFiles = []string{params.RootPath}

		linter.ParseFilenames(linter.ReadFilenames(linter.AnalysisFiles, nil))

		meta.SetIndexingComplete(true)

		// fully analyze all opened files
		// other files are not analyzed fully at all
		openMapMutex.Lock()
		for filename, op := range openMap {
			go openFile(filename, op.contents)
		}
		openMapMutex.Unlock()
	}()

	return writeMessage(&response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result: map[string]interface{}{
			"capabilities": map[string]interface{}{
				"codeActionProvider":               nil,
				"codeLensProvider":                 nil,
				"textDocumentSync":                 1, // FULL
				"documentSymbolProvider":           true,
				"workspaceSymbolProvider":          true,
				"definitionProvider":               true,
				"dependenciesProvider":             nil,
				"documentFormattingProvider":       nil,
				"documentHighlightProvider":        nil,
				"documentOnTypeFormattingProvider": nil,
				"documentRangeFormattingProvider":  nil,
				"referencesProvider":               true,
				"hoverProvider":                    true,
				"completionProvider": map[string]interface{}{
					"resolveProvider":   true,
					"triggerCharacters": []string{"$", ">", "\\"},
				},
				"renameProvider": nil,
				"signatureHelpProvider": map[string]interface{}{
					"triggerCharacters": []string{"(", ","},
				},
				"xworkspaceReferencesProvider": true,
				"xdefinitionProvider":          true,
				"xdependenciesProvider":        true,
			},
		},
	})
}

func handleTextDocumentDidOpen(req *baseRequest) error {
	var params vscode.TextDocumentDidOpenParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	uri := params.TextDocument.URI
	lintdebug.Send("Open text document %s", uri)

	if strings.HasPrefix(uri, "file://") {
		openFile(strings.TrimPrefix(uri, "file://"), params.TextDocument.Text)
	}

	return nil
}

func handleTextDocumentDidClose(req *baseRequest) error {
	var params vscode.TextDocumentDidOpenParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	uri := params.TextDocument.URI
	lintdebug.Send("Close text document %s", uri)

	if strings.HasPrefix(uri, "file://") {
		closeFile(strings.TrimPrefix(uri, "file://"))
	}

	return nil
}

func handleTextDocumentDidChange(req *baseRequest) error {
	var params vscode.TextDocumentDidChangeParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	if len(params.ContentChanges) != 1 {
		lintdebug.Send("Unexpected number of content changes: %d", len(params.ContentChanges))
		return nil
	}

	uri := params.TextDocument.URI

	if strings.HasPrefix(uri, "file://") {
		changeFile(strings.TrimPrefix(uri, "file://"), params.ContentChanges[0].Text)
	}

	return nil
}

func baseSymbolName(s string) string {
	if idx := strings.LastIndexByte(s, '\\'); idx >= 0 && len(s) > idx {
		return s[idx+1:]
	}
	return s
}

func handleTextDocumentSymbol(req *baseRequest) error {
	var params vscode.TextDocumentDidOpenParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	// TODO: make it actually safe

	meta.OnIndexingComplete(func() {
		uri := params.TextDocument.URI

		var result []vscode.SymbolInformation

		if strings.HasPrefix(uri, "file://") {
			filename := strings.TrimPrefix(uri, "file://")
			res := meta.Info.GetMetaForFile(filename)

			for className, classInfo := range res.Classes {
				result = append(result, vscode.SymbolInformation{
					Name:     baseSymbolName(className),
					Kind:     vscode.CompletionKindClass,
					Location: posToLocation(classInfo.Pos),
				})

				for methodName, info := range classInfo.Methods {
					result = append(result, vscode.SymbolInformation{
						Name:     methodName,
						Kind:     vscode.CompletionKindMethod,
						Location: posToLocation(info.Pos),
					})
				}

				for propName, info := range classInfo.Properties {
					result = append(result, vscode.SymbolInformation{
						Name:     propName,
						Kind:     vscode.CompletionKindProperty,
						Location: posToLocation(info.Pos),
					})
				}

				for constName, info := range classInfo.Constants {
					result = append(result, vscode.SymbolInformation{
						Name:     constName,
						Kind:     vscode.CompletionKindEnum,
						Location: posToLocation(info.Pos),
					})
				}
			}

			for funcName, info := range res.Functions {
				result = append(result, vscode.SymbolInformation{
					Name:     baseSymbolName(funcName),
					Kind:     vscode.CompletionKindFunction,
					Location: posToLocation(info.Pos),
				})
			}

			for constName, info := range res.Constants {
				result = append(result, vscode.SymbolInformation{
					Name:     baseSymbolName(constName),
					Kind:     vscode.CompletionKindEnum,
					Location: posToLocation(info.Pos),
				})
			}
		}

		writeMessage(&response{
			JSONRPC: req.JSONRPC,
			ID:      req.ID,
			Result:  result,
		})
	})

	return nil
}

// very simple convertion
func posToLocation(pos meta.ElementPosition) vscode.Location {
	return vscode.Location{
		URI: "file://" + pos.Filename,
		Range: vscode.Range{
			Start: vscode.Position{Line: int(pos.Line) - 1},
			End:   vscode.Position{Line: int(pos.EndLine) - 1},
		},
	}
}

func handleTextDocumentDefinition(req *baseRequest) error {
	changingMutex.Lock()
	defer changingMutex.Unlock()

	var params vscode.DefinitionParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	filename := strings.TrimPrefix(params.TextDocument.URI, "file://")
	openMapMutex.Lock()
	f, ok := openMap[filename]
	openMapMutex.Unlock()

	if !ok {
		lintdebug.Send("File is not opened, but definition requested: %s", filename)
		return nil
	}

	result := make([]vscode.Location, 0)

	if params.Position.Line < len(f.linesPositions) {
		w := &definitionWalker{
			position:  f.linesPositions[params.Position.Line] + params.Position.Character,
			positions: f.positions,
			scopes:    f.scopes,
		}
		f.rootNode.Walk(w)
		if len(w.result) > 0 {
			result = w.result
		}
	}

	return writeMessage(&response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result:  result,
	})
}

func handleTextDocumentReferences(req *baseRequest) error {
	changingMutex.Lock()
	defer changingMutex.Unlock()

	var params vscode.ReferencesParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	filename := strings.TrimPrefix(params.TextDocument.URI, "file://")
	openMapMutex.Lock()
	f, ok := openMap[filename]
	openMapMutex.Unlock()

	if !ok {
		lintdebug.Send("File is not opened, but references requested: %s", filename)
		return nil
	}

	result := make([]vscode.Location, 0)

	if params.Position.Line < len(f.linesPositions) {
		w := &referencesWalker{
			position:  f.linesPositions[params.Position.Line] + params.Position.Character,
			positions: f.positions,
		}
		f.rootNode.Walk(w)
		if len(w.result) > 0 {
			result = w.result
		}
	}

	return writeMessage(&response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result:  result,
	})
}

func resolveTypesSafe(m *meta.TypesMap, visitedMap map[string]struct{}) (res map[string]struct{}) {
	defer func() {
		if r := recover(); r != nil {
			res = make(map[string]struct{})
			res["panic: "+fmt.Sprint(r)] = struct{}{}
			res["orig: "+m.String()] = struct{}{}
		}
	}()

	res = solver.ResolveTypes(m, visitedMap)
	return
}

func handleTextDocumentHover(req *baseRequest) error {
	changingMutex.Lock()
	defer changingMutex.Unlock()

	var contents string

	defer func() {
		writeMessage(&response{
			JSONRPC: req.JSONRPC,
			ID:      req.ID,
			Result: map[string]interface{}{
				"contents": contents,
			},
		})
	}()

	var params vscode.DefinitionParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	filename := strings.TrimPrefix(params.TextDocument.URI, "file://")
	openMapMutex.Lock()
	f, ok := openMap[filename]
	openMapMutex.Unlock()

	if !ok {
		return nil
	}

	lnPos := params.Position.Line
	chPos := params.Position.Character - 1

	if lnPos >= len(f.lines) {
		lintdebug.Send("Line out of range for file %s: %d", filename, lnPos)
		return nil
	}

	ln := f.lines[lnPos]

	if chPos < 0 || chPos >= len(ln) {
		lintdebug.Send("Char out of range for file %s: line '%s', char %d", filename, ln, chPos)
		return nil
	}

	compl := &completionWalker{
		position:  f.linesPositions[params.Position.Line] + params.Position.Character,
		positions: f.positions,
		scopes:    f.scopes,
	}

	f.rootNode.Walk(compl)

	if compl.foundScope == nil {
		return nil
	}

	hover := &hoverWalker{
		position:  compl.position,
		positions: compl.positions,
	}

	f.rootNode.Walk(hover)

	if hover.n == nil {
		return nil
	}

	contents = getHoverForNode(hover.n, compl.foundScope, &hover.st)

	return nil
}

func getHoverForNode(n node.Node, sc *meta.Scope, cs *meta.ClassParseState) string {
	switch n := n.(type) {
	case *expr.Variable:
		return getHoverForVariable(n, sc)
	case *expr.FunctionCall:
		return getHoverForFunctionCall(n, sc, cs)
	case *expr.MethodCall:
		return getHoverForMethodCall(n, sc, cs)
	case *expr.StaticCall:
		return getHoverForStaticCall(n, sc, cs)
	}

	return ""
}

func getHoverForVariable(n *expr.Variable, sc *meta.Scope) string {
	id, ok := n.VarName.(*node.Identifier)
	if !ok {
		return ""
	}

	name := id.Value

	typ, _ := sc.GetVarNameType(name)
	newM := meta.NewTypesMapFromMap(resolveTypesSafe(typ, make(map[string]struct{})))
	return newM.String() + " $" + name
}

func getHoverForFunctionCall(n *expr.FunctionCall, sc *meta.Scope, cs *meta.ClassParseState) string {
	var fun meta.FuncInfo
	var ok bool
	var nameStr string

	switch nm := n.Function.(type) {
	case *name.Name:
		nameStr = meta.NameToString(nm)
		fun, ok = meta.Info.GetFunction(cs.Namespace + `\` + nameStr)
		if !ok && cs.Namespace != "" {
			fun, ok = meta.Info.GetFunction(`\` + nameStr)
		}
	case *name.FullyQualified:
		nameStr = meta.FullyQualifiedToString(nm)
		fun, ok = meta.Info.GetFunction(nameStr)
	}

	return linter.FlagsToString(fun.ExitFlags)
}

func getHoverForMethodCall(n *expr.MethodCall, sc *meta.Scope, cs *meta.ClassParseState) string {
	id, ok := n.Method.(*node.Identifier)
	if !ok {
		return ""
	}

	types := safeExprType(sc, cs, n.Variable)

	var fun meta.FuncInfo
	ok = false

	types.Iterate(func(t string) {
		if ok {
			return
		}
		fun, _, ok = solver.FindMethod(t, id.Value)
	})

	return linter.FlagsToString(fun.ExitFlags)
}

func getHoverForStaticCall(n *expr.StaticCall, sc *meta.Scope, cs *meta.ClassParseState) string {
	id, ok := n.Call.(*node.Identifier)
	if !ok {
		return ""
	}

	className, ok := solver.GetClassName(cs, n.Class)
	if !ok {
		return ""
	}

	fun, _, ok := solver.FindMethod(className, id.Value)
	if !ok {
		return ""
	}

	return linter.FlagsToString(fun.ExitFlags)
}

func handleTextDocumentCompletion(req *baseRequest) error {
	changingMutex.Lock()
	defer changingMutex.Unlock()

	start := time.Now()
	defer func() { lintdebug.Send("Completion took %s", time.Since(start)) }()

	var params vscode.DefinitionParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	filename := strings.TrimPrefix(params.TextDocument.URI, "file://")
	openMapMutex.Lock()
	f, ok := openMap[filename]
	openMapMutex.Unlock()

	if !ok {
		lintdebug.Send("File is not opened, but completion requested: %s", filename)
		return nil
	}

	lnPos := params.Position.Line
	chPos := params.Position.Character - 1

	var ln []byte
	var position int

	if lnPos >= len(f.lines) {
		lintdebug.Send("Line out of range for file %s: %d", filename, lnPos)
		return nil
	}

	ln = f.lines[lnPos]

	if chPos < 0 || chPos >= len(ln) {
		lintdebug.Send("Char out of range for file %s: line '%s', char %d", filename, ln, chPos)
		return nil
	}

	position = f.linesPositions[params.Position.Line] + params.Position.Character

	compl := &completionWalker{
		position:  position,
		positions: f.positions,
		scopes:    f.scopes,
	}

	f.rootNode.Walk(compl)

	chBytes := []byte{}
	for i := chPos; i >= 0; i-- {
		curCh := ln[i]
		if curCh >= 'A' && curCh <= 'Z' || curCh >= 'a' && curCh <= 'z' || curCh >= '0' && curCh <= '9' || curCh == '_' || curCh == '$' || curCh == '>' || curCh == '\\' || curCh == '-' && len(ln) >= i+2 && ln[i+1] == '>' {
			chBytes = append(chBytes, curCh)
		} else {
			break
		}
	}

	for i, j := 0, len(chBytes)-1; i < j; i, j = i+1, j-1 {
		chBytes[i], chBytes[j] = chBytes[j], chBytes[i]
	}

	chStr := string(chBytes)

	lintdebug.Send("Ch str: %s, have scope: %v", chStr, compl.foundScope != nil)

	var result []vscode.CompletionItem

	if compl.foundScope != nil && strings.HasPrefix(chStr, "$") {
		lintdebug.Send("Var str: %s", chStr)

		if strings.HasSuffix(chStr, "->") {
			result = append(result, getMethodCompletionItems(&compl.st, chStr, compl.foundScope)...)
		} else {
			compl.foundScope.Iterate(func(varName string, typ *meta.TypesMap, alwaysDefined bool) {
				result = append(result, vscode.CompletionItem{
					Kind:  vscode.CompletionKindVariable,
					Label: "$" + varName,
				})
			})
		}
	} else {
		var funcs []string
		var funcsNs []string
		var constants []string
		var constantsNs []string

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			funcStr := `\` + chStr

			funcs = meta.Info.FindFunctions(funcStr)
			sort.Strings(funcs)

			wg.Done()
		}()

		if compl.st.Namespace != "" {
			wg.Add(1)
			go func() {
				funcStr := compl.st.Namespace + `\` + chStr

				funcsNs = meta.Info.FindFunctions(funcStr)
				sort.Strings(funcsNs)

				wg.Done()
			}()
		}

		wg.Add(1)
		go func() {
			constStr := `\` + chStr

			constants = meta.Info.FindConstants(constStr)
			sort.Strings(constants)

			wg.Done()
		}()

		if compl.st.Namespace != "" {
			wg.Add(1)
			go func() {
				constStr := compl.st.Namespace + `\` + chStr

				constantsNs = meta.Info.FindConstants(constStr)
				sort.Strings(constantsNs)

				wg.Done()
			}()
		}

		wg.Wait()

		for _, f := range funcsNs {
			result = append(result, vscode.CompletionItem{
				Kind:       vscode.CompletionKindFunction,
				Label:      f,
				InsertText: strings.TrimPrefix(f, `\`),
			})
		}

		for _, f := range funcs {
			result = append(result, vscode.CompletionItem{
				Kind:       vscode.CompletionKindFunction,
				Label:      f,
				InsertText: strings.TrimPrefix(f, `\`),
			})
		}

		for _, f := range constantsNs {
			result = append(result, vscode.CompletionItem{
				Kind:       vscode.CompletionKindEnum,
				Label:      f,
				InsertText: strings.TrimPrefix(f, `\`),
			})
		}

		for _, f := range constants {
			result = append(result, vscode.CompletionItem{
				Kind:       vscode.CompletionKindEnum,
				Label:      f,
				InsertText: strings.TrimPrefix(f, `\`),
			})
		}
	}

	return writeMessage(&response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result: map[string]interface{}{
			"isIncomplete": false,
			"items":        result,
		},
	})
}

func getMethods(className string) (res []string) {
	for {
		class, ok := meta.Info.GetClass(className)
		if !ok {
			return res
		}

		for m := range class.Methods {
			res = append(res, m)
		}

		className = class.Parent
		if className == "" {
			return res
		}
	}
}

func getInstanceProperties(className string) (res []string) {
	for {
		class, ok := meta.Info.GetClass(className)
		if !ok {
			return res
		}

		for m := range class.Properties {
			if !strings.HasPrefix(m, "$") {
				res = append(res, m)
			}
		}

		className = class.Parent
		if className == "" {
			return res
		}
	}
}

func getMethodCompletionItems(st *meta.ClassParseState, str string, sc *meta.Scope) (result []vscode.CompletionItem) {
	strTemp := "<?php " + strings.TrimSuffix(str, "->") + ";"
	parser := php7.NewParser(strings.NewReader(strTemp), "temp")
	parser.Parse()

	tempNode := parser.GetRootNode()
	if tempNode == nil {
		lintdebug.Send("Could not parse %s", strTemp)
		return nil
	}

	stmtLst, ok := tempNode.(*stmt.StmtList)
	if !ok || len(stmtLst.Stmts) == 0 {
		return nil
	}

	s, ok := stmtLst.Stmts[0].(*stmt.Expression)
	if !ok {
		return nil
	}

	var methodList []string
	var propList []string
	methodDedup := map[string]struct{}{}
	propDedup := map[string]struct{}{}

	safeExprType(sc, st, s.Expr).Iterate(func(t string) {
		for _, m := range getMethods(t) {
			if _, ok := methodDedup[m]; ok {
				continue
			}
			methodList = append(methodList, m)
			methodDedup[m] = struct{}{}
		}

		for _, m := range getInstanceProperties(t) {
			if _, ok := propDedup[m]; ok {
				continue
			}
			propList = append(propList, m)
			propDedup[m] = struct{}{}
		}
	})

	sort.Strings(methodList)
	sort.Strings(propList)

	for _, m := range propList {
		result = append(result, vscode.CompletionItem{
			Kind:  vscode.CompletionKindProperty,
			Label: m,
		})
	}

	for _, m := range methodList {
		result = append(result, vscode.CompletionItem{
			Kind:  vscode.CompletionKindMethod,
			Label: m,
		})
	}

	return result
}

func handleChangeWatchedFiles(req *baseRequest) error {
	var params vscode.DidChangeWatchedFilesParams
	if err := json.Unmarshal([]byte(req.Params), &params); err != nil {
		return err
	}

	externalChanges(params.Changes)

	return nil
}

// Start starts Microsoft LSP server with stdin/stdout I/O.
func Start() {
	rd := bufio.NewReader(os.Stdin)
	connWr = os.Stdout

	linter.InitStubs()

	for {
		ln, err := rd.ReadString('\n')
		if err != nil {
			log.Fatalf("Could not read: %s", err.Error())
		}

		if !strings.HasPrefix(ln, "Content-Length: ") {
			log.Fatalf("Wrong line: expected 'Content-Length:', got '%s'", ln)
		}

		length, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(ln, "Content-Length: ")))
		if err != nil {
			log.Fatalf("Could not parse content length: %s", err.Error())
		}

		// should be empty line
		rd.ReadString('\n')

		if length > maxLength {
			log.Fatalf("Length too high: %d, max: %d", length, maxLength)
		}

		buf := make([]byte, length)
		if _, err = io.ReadFull(rd, buf); err != nil {
			log.Fatalf("Could not read message: %s", err.Error())
		}

		if err := handleMessage(buf); err != nil {
			log.Fatalf("Could not write message: %s", err.Error())
		}
	}
}
