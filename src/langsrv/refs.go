package langsrv

import (
	"bytes"
	"sync"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/irgen"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/state"
	"github.com/VKCOM/noverify/src/vscode"
	"go.lsp.dev/uri"
)

type referencesWalker struct {
	st meta.ClassParseState

	position int
	scopes   map[ir.Node]*meta.Scope

	result      []vscode.Location
	foundScopes []*meta.Scope
}

func getFunction(st *meta.ClassParseState, n *ir.FunctionCallExpr) (fun meta.FuncInfo, nameStr string, ok bool) {
	switch nm := n.Function.(type) {
	case *ir.Name:
		nameStr = meta.NameToString(nm)
		tryStr := st.Namespace + `\` + nameStr

		fun, ok = meta.Info.GetFunction(tryStr)
		if ok {
			return fun, tryStr, true
		}

		if !ok && st.Namespace != "" {
			tryStr := `\` + nameStr
			fun, ok = meta.Info.GetFunction(`\` + nameStr)
			if ok {
				return fun, tryStr, true
			}
		}
	case *ir.FullyQualifiedName:
		nameStr = meta.FullyQualifiedToString(nm)
		fun, ok = meta.Info.GetFunction(nameStr)
	}

	return fun, nameStr, ok
}

// EnterNode is invoked at every node in hierarchy
func (d *referencesWalker) EnterNode(n ir.Node) bool {
	sc, ok := d.scopes[n]
	if ok {
		d.foundScopes = append(d.foundScopes, sc)
	}

	state.EnterNode(&d.st, n)

	switch n := n.(type) {
	case *ir.FunctionCallExpr:
		if pos := ir.GetPosition(n.Function); d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		_, nameStr, ok := getFunction(&d.st, n)
		if ok {
			d.result = findFunctionReferences(nameStr)
		}
	case *ir.StaticCallExpr:
		if pos := ir.GetPosition(n.Call); d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		// not going to resolve $obj->$someMethod(); calls
		id, ok := n.Call.(*ir.Identifier)
		if !ok {
			return true
		}

		className, ok := solver.GetClassName(&d.st, n.Class)
		if !ok {
			return true
		}

		m, ok := solver.FindMethod(className, id.Value)
		if ok {
			realClassName := m.ImplName()
			d.result = findStaticMethodReferences(realClassName, id.Value)
		}
	case *ir.FunctionStmt:
		if pos := ir.GetPosition(n.FunctionName); d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		d.result = findFunctionReferences(d.st.Namespace + `\` + n.FunctionName.Value)
	case *ir.ClassMethodStmt:
		if pos := ir.GetPosition(n.MethodName); d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		isStatic := false

		for _, m := range n.Modifiers {
			if m.Value == "static" {
				isStatic = true
			}
		}

		if isStatic {
			d.result = findStaticMethodReferences(d.st.CurrentClass, n.MethodName.Value)
		} else {
			d.result = findMethodReferences(d.st.CurrentClass, n.MethodName.Value)
		}
	case *ir.PropertyStmt:
		if pos := ir.GetPosition(n); d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		d.result = findPropertyReferences(d.st.CurrentClass, n.Variable.Name)
	case *ir.ConstantStmt:
		if pos := ir.GetPosition(n.ConstantName); d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		if d.st.CurrentClass == "" {
			d.result = findConstantsReferences(d.st.Namespace + `\` + n.ConstantName.Value)
		} else {
			d.result = findClassConstantsReferences(d.st.CurrentClass, n.ConstantName.Value)
		}
	}

	return true
}

// LeaveNode is invoked after node process
func (d *referencesWalker) LeaveNode(n ir.Node) {
	if d.scopes != nil {
		_, ok := d.scopes[n]
		if ok && len(d.foundScopes) > 0 {
			d.foundScopes = d.foundScopes[0 : len(d.foundScopes)-1]
		}
	}

	state.LeaveNode(&d.st, n)
}

// copyOpenMap returns map[filename]contents
func copyOpenMap() map[string]string {
	openMapMutex.Lock()
	res := make(map[string]string, len(openMap))
	for filename, info := range openMap {
		res[filename] = info.contents
	}
	openMapMutex.Unlock()

	return res
}

func readFile(openMapCopy map[string]string, filename string) (contents []byte, err error) {
	if cont, ok := openMapCopy[filename]; ok {
		return []byte(cont), nil
	}

	return getFileContents(filename)
}

// Very simple computation for now, it needs to be improved for refactoring purposes :)
func refPosition(filename string, pos *position.Position) vscode.Location {
	endLine := pos.EndLine - 1
	if pos.StartLine == pos.EndLine {
		endLine++
	}

	return vscode.Location{
		URI: uri.File(filename),
		Range: vscode.Range{
			Start: vscode.Position{Line: pos.StartLine - 1},
			End:   vscode.Position{Line: endLine},
		},
	}
}

type parseFn func(filename string, rootNode ir.Node, contents []byte, parser *php7.Parser) []vscode.Location

func findReferences(substr string, parse parseFn) []vscode.Location {
	cb := linter.ReadFilenames(linter.AnalysisFiles, nil)
	ch := make(chan linter.FileInfo)
	go func() {
		cb(ch)
		close(ch)
	}()

	substrBytes := []byte(substr)

	var (
		resultMutex sync.Mutex
		result      []vscode.Location
		wg          sync.WaitGroup
	)

	openMapCopy := copyOpenMap()

	for i := 0; i < linter.MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			for fi := range ch {
				contents, err := readFile(openMapCopy, fi.Filename)
				if err == nil && bytes.Contains(contents, substrBytes) {
					func() {
						waiter := linter.BeforeParse(len(contents), fi.Filename)
						defer waiter.Finish()

						parser := php7.NewParser(contents)
						parser.Parse()

						rootNode := parser.GetRootNode()
						rootIR := irgen.ConvertRoot(rootNode)
						if rootNode != nil {
							var found []vscode.Location
							func() {
								defer func() {
									if r := recover(); r != nil {
										lintdebug.Send("Panic while processing %s: %v", fi.Filename, r)
									}
								}()

								found = parse(fi.Filename, rootIR, contents, parser)
							}()
							resultMutex.Lock()
							result = append(result, found...)
							resultMutex.Unlock()
						}
					}()
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()

	return result
}
