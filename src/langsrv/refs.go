package langsrv

import (
	"bytes"
	"io/ioutil"
	"sync"

	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/state"
	"github.com/VKCOM/noverify/src/vscode"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/php7"
	"github.com/z7zmey/php-parser/position"
	"github.com/z7zmey/php-parser/walker"
)

type referencesWalker struct {
	st meta.ClassParseState

	position  int
	positions position.Positions
	scopes    map[node.Node]*meta.Scope

	result      []vscode.Location
	foundScopes []*meta.Scope
}

func getFunction(st *meta.ClassParseState, n *expr.FunctionCall) (fun meta.FuncInfo, nameStr string, ok bool) {
	switch nm := n.Function.(type) {
	case *name.Name:
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
	case *name.FullyQualified:
		nameStr = meta.FullyQualifiedToString(nm)
		fun, ok = meta.Info.GetFunction(nameStr)
	}

	return fun, nameStr, ok
}

// EnterNode is invoked at every node in hierarchy
func (d *referencesWalker) EnterNode(w walker.Walkable) bool {
	n := w.(node.Node)

	sc, ok := d.scopes[n]
	if ok {
		d.foundScopes = append(d.foundScopes, sc)
	}

	state.EnterNode(&d.st, n)

	switch n := w.(type) {
	case *expr.FunctionCall:
		if pos := d.positions[n.Function]; d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		_, nameStr, ok := getFunction(&d.st, n)
		if ok {
			d.result = findFunctionReferences(nameStr)
		}
	case *expr.StaticCall:
		if pos := d.positions[n.Call]; d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		// not going to resolve $obj->$someMethod(); calls
		id, ok := n.Call.(*node.Identifier)
		if !ok {
			return true
		}

		className, ok := solver.GetClassName(&d.st, n.Class)
		if !ok {
			return true
		}

		_, realClassName, ok := solver.FindMethod(className, id.Value)
		if ok {
			d.result = findStaticMethodReferences(realClassName, id.Value)
		}
	case *stmt.Function:
		if pos := d.positions[n.FunctionName]; d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		d.result = findFunctionReferences(d.st.Namespace + `\` + n.FunctionName.(*node.Identifier).Value)
	case *stmt.ClassMethod:
		if pos := d.positions[n.MethodName]; d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		isStatic := false

		for _, m := range n.Modifiers {
			switch v := m.(*node.Identifier).Value; v {
			case "static":
				isStatic = true
			}
		}

		if isStatic {
			d.result = findStaticMethodReferences(d.st.CurrentClass, n.MethodName.(*node.Identifier).Value)
		} else {
			d.result = findMethodReferences(d.st.CurrentClass, n.MethodName.(*node.Identifier).Value)
		}
	case *stmt.Property:
		if pos := d.positions[n]; d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		d.result = findPropertyReferences(d.st.CurrentClass, n.Variable.(*expr.Variable).VarName.(*node.Identifier).Value)
	case *stmt.Constant:
		if pos := d.positions[n.ConstantName]; d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		if d.st.CurrentClass == "" {
			d.result = findConstantsReferences(d.st.Namespace + `\` + n.ConstantName.(*node.Identifier).Value)
		} else {
			d.result = findClassConstantsReferences(d.st.CurrentClass, n.ConstantName.(*node.Identifier).Value)
		}
	}

	return true
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *referencesWalker) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *referencesWalker) LeaveNode(w walker.Walkable) {
	n := w.(node.Node)

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

func readFile(openMapCopy map[string]string, filename string) (contents []byte, needCharsetConvertion bool, err error) {
	if cont, ok := openMapCopy[filename]; ok {
		return []byte(cont), false, nil
	}

	contents, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, false, err
	}

	return contents, true, nil
}

// Very simple computation for now, it needs to be improved for refactoring purposes :)
func refPosition(filename string, pos *position.Position) vscode.Location {
	endLine := int(pos.EndLine) - 1
	if pos.StartLine == pos.EndLine {
		endLine++
	}

	return vscode.Location{
		URI: "file://" + filename,
		Range: vscode.Range{
			Start: vscode.Position{Line: int(pos.StartLine) - 1},
			End:   vscode.Position{Line: endLine},
		},
	}
}

type parseFn func(filename string, rootNode node.Node, contents []byte, parser *php7.Parser) []vscode.Location

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
				contents, needCharsetConvertion, err := readFile(openMapCopy, fi.Filename)
				if err == nil && bytes.Contains(contents, substrBytes) {
					if needCharsetConvertion {
						contents = convertEncodingIfNeeded(contents)
					}

					func() {
						waiter := linter.BeforeParse(len(contents), fi.Filename)
						defer waiter.Finish()

						parser := php7.NewParser(bytes.NewReader(contents), fi.Filename)
						parser.Parse()

						rootNode := parser.GetRootNode()
						if rootNode != nil {
							var found []vscode.Location
							func() {
								defer func() {
									if r := recover(); r != nil {
										lintdebug.Send("Panic while processing %s: %v", fi.Filename, r)
									}
								}()

								found = parse(fi.Filename, rootNode, contents, parser)
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
