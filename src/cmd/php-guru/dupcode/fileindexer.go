package dupcode

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/normalize"
)

type fileIndexer struct {
	funcs        funcSet
	fileContents []byte
	filename     string
	args         *arguments
	normLevel    normalizationLevel
}

func (indexer *fileIndexer) CollectFuncs(root *ir.Root) {
	indexer.walkStatements(root.Stmts)
}

func (indexer *fileIndexer) walkStatements(list []ir.Node) {
	for _, stmt := range list {
		switch stmt := stmt.(type) {
		case *ir.FunctionStmt:
			indexer.walkFunc(stmt)
		case *ir.ClassStmt:
			indexer.walkClass(stmt)
		case *ir.NamespaceStmt:
			indexer.walkStatements(stmt.Stmts)
		}
	}
}

func (indexer *fileIndexer) walkClass(n *ir.ClassStmt) {
	for _, stmt := range n.Stmts {
		method, ok := stmt.(*ir.ClassMethodStmt)
		if !ok {
			continue
		}
		indexer.walkMethod(n.ClassName.Value, method)
	}
}

func (indexer *fileIndexer) walkMethod(className string, n *ir.ClassMethodStmt) {
	body, ok := n.Stmt.(*ir.StmtList)
	if !ok {
		return
	}
	if !indexer.canAnalyzeMethod(body.Stmts) {
		return
	}
	// Visit private methods only if told to do so.
	if !indexer.args.checkPrivate && hasModifier(n.Modifiers, "private") {
		return
	}
	if !assertComplexity(body.Stmts, indexer.args.minComplexity) {
		return
	}

	indexer.collectFunc(n, className, n.MethodName.Value, n.Params, body.Stmts)
}

func (indexer *fileIndexer) walkFunc(n *ir.FunctionStmt) {
	if !assertComplexity(n.Stmts, indexer.args.minComplexity) {
		return
	}

	indexer.collectFunc(n, "", n.FunctionName.Value, n.Params, n.Stmts)
}

func (indexer *fileIndexer) collectFunc(n ir.Node, className, funcName string, params, stmts []ir.Node) {
	if indexer.normLevel != normNone {
		conf := normalize.Config{
			NormalizeMore: indexer.normLevel > normSafe,
		}
		stmts = normalize.FuncBody(conf, params, stmts)
	}

	pos := ir.GetPosition(n)
	key := calculateCodeHash(stmts)
	newFunc := &funcInfo{
		className: className,
		name:      funcName,
		declPos: position{
			line:     pos.StartLine,
			filename: indexer.filename,
		},
		linesOfCode: pos.EndLine - pos.StartLine,
		code:        indexer.nodeText(n),
	}
	indexer.funcs.AddFunc(key, newFunc)
}

func (indexer *fileIndexer) canAnalyzeMethod(list []ir.Node) bool {
	// Don't waste time on huge methods.
	if len(list) > 25 {
		return false
	}

	for _, stmt := range list {
		cantAnalyze := findNode(stmt, func(n ir.Node) bool {
			switch n := n.(type) {
			case *ir.SimpleVar:
				if n.Name == "this" {
					return true
				}
			case *ir.Name:
				if n.Value == "self" || n.Value == "static" {
					return true
				}
			case *ir.Identifier:
				if n.Value == "self" || n.Value == "static" {
					return true
				}
			}
			return false
		})
		if cantAnalyze {
			return false
		}
	}
	return true
}

func (indexer *fileIndexer) nodeText(n ir.Node) []byte {
	pos := ir.GetPosition(n)
	return indexer.fileContents[pos.StartPos:pos.EndPos]
}
