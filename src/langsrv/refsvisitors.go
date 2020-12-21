package langsrv

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/state"
	"github.com/VKCOM/noverify/src/vscode"
)

func findFunctionReferences(funcName string) []vscode.Location {
	substr := funcName
	if idx := strings.LastIndexByte(funcName, '\\'); idx >= 0 {
		substr = funcName[idx+1:]
	}

	return findReferences(substr, func(filename string, rootNode ir.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &funcCallVisitor{
			funcName: funcName,
			filename: filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findStaticMethodReferences(className string, methodName string) []vscode.Location {
	return findReferences(methodName, func(filename string, rootNode ir.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &staticMethodCallVisitor{
			className:  className,
			methodName: methodName,
			filename:   filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findConstantsReferences(constName string) []vscode.Location {
	return findReferences(constName, func(filename string, rootNode ir.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &constVisitor{
			constName: constName,
			filename:  filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findClassConstantsReferences(className string, constName string) []vscode.Location {
	return findReferences(constName, func(filename string, rootNode ir.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &classConstVisitor{
			className: className,
			constName: constName,
			filename:  filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findMethodReferences(className string, methodName string) []vscode.Location {
	return findReferences(methodName, func(filename string, rootNode ir.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		var found []vscode.Location

		rootWalker := linter.NewWalkerForReferencesSearcher(
			linter.NewWorkerContext(),
			filename,
			func(ctx *linter.BlockContext) linter.BlockChecker {
				return &blockMethodCallVisitor{
					ctx:        ctx,
					className:  className,
					methodName: methodName,
					filename:   filename,
					addFound:   func(f vscode.Location) { found = append(found, f) },
				}
			},
		)

		rootWalker.InitFromParser(contents, parser)

		rootNode.Walk(rootWalker)
		linter.AnalyzeFileRootLevel(rootNode, rootWalker)

		return found
	})
}

func findPropertyReferences(className string, propName string) []vscode.Location {
	return findReferences(propName, func(filename string, rootNode ir.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		var found []vscode.Location

		rootWalker := linter.NewWalkerForReferencesSearcher(
			linter.NewWorkerContext(),
			filename,
			func(ctx *linter.BlockContext) linter.BlockChecker {
				return &blockPropertyVisitor{
					ctx:       ctx,
					className: className,
					propName:  propName,
					filename:  filename,
					addFound:  func(f vscode.Location) { found = append(found, f) },
				}
			},
		)

		rootWalker.InitFromParser(contents, parser)

		rootNode.Walk(rootWalker)
		linter.AnalyzeFileRootLevel(rootNode, rootWalker)

		return found
	})
}

type funcCallVisitor struct {
	st       meta.ClassParseState
	funcName string
	filename string

	found []vscode.Location
}

func (d *funcCallVisitor) GetFoundLocations() []vscode.Location {
	return d.found
}

// EnterNode is invoked at every node in hierarchy
func (d *funcCallVisitor) EnterNode(w ir.Node) bool {
	state.EnterNode(&d.st, w)

	if n, ok := w.(*ir.FunctionCallExpr); ok {
		_, nameStr, ok := getFunction(&d.st, n)
		if ok && nameStr == d.funcName {
			if pos := ir.GetPosition(n); pos != nil {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// LeaveNode is invoked after node process
func (d *funcCallVisitor) LeaveNode(w ir.Node) {
	state.LeaveNode(&d.st, w)
}

type staticMethodCallVisitor struct {
	// params
	className  string
	methodName string
	filename   string

	// output
	found []vscode.Location

	// state
	st meta.ClassParseState
}

func (d *staticMethodCallVisitor) GetFoundLocations() []vscode.Location {
	return d.found
}

// EnterNode is invoked at every node in hierarchy
func (d *staticMethodCallVisitor) EnterNode(n ir.Node) bool {
	state.EnterNode(&d.st, n)

	if n, ok := n.(*ir.StaticCallExpr); ok {
		id, ok := n.Call.(*ir.Identifier)
		if !ok {
			return true
		}
		className, ok := solver.GetClassName(&d.st, n.Class)
		if !ok {
			return true
		}
		m, ok := solver.FindMethod(className, id.Value)
		realClassName := m.ImplName()

		if ok && realClassName == d.className && id.Value == d.methodName {
			if pos := ir.GetPosition(n); pos != nil {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// LeaveNode is invoked after node process
func (d *staticMethodCallVisitor) LeaveNode(w ir.Node) {
	state.LeaveNode(&d.st, w)
}

type constVisitor struct {
	// params
	constName string
	filename  string

	// output
	found []vscode.Location

	// state
	st meta.ClassParseState
}

func (d *constVisitor) GetFoundLocations() []vscode.Location {
	return d.found
}

// EnterNode is invoked at every node in hierarchy
func (d *constVisitor) EnterNode(n ir.Node) bool {
	state.EnterNode(&d.st, n)

	if n, ok := n.(*ir.ConstFetchExpr); ok {
		constName, _, ok := solver.GetConstant(&d.st, n.Constant)

		if ok && constName == d.constName {
			if pos := ir.GetPosition(n); pos != nil {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// LeaveNode is invoked after node process
func (d *constVisitor) LeaveNode(w ir.Node) {
	state.LeaveNode(&d.st, w)
}

type classConstVisitor struct {
	// params
	className string
	constName string
	filename  string

	// output
	found []vscode.Location

	// state
	st meta.ClassParseState
}

func (d *classConstVisitor) GetFoundLocations() []vscode.Location {
	return d.found
}

// EnterNode is invoked at every node in hierarchy
func (d *classConstVisitor) EnterNode(n ir.Node) bool {
	state.EnterNode(&d.st, n)

	if n, ok := n.(*ir.ClassConstFetchExpr); ok {
		constName := n.ConstantName
		if constName.Value == `class` || constName.Value == `CLASS` {
			return true
		}

		className, ok := solver.GetClassName(&d.st, n.Class)
		if !ok {
			return true
		}

		_, implClassName, ok := solver.FindConstant(className, constName.Value)

		if ok && constName.Value == d.constName && implClassName == d.className {
			if pos := ir.GetPosition(n); pos != nil {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// LeaveNode is invoked after node process
func (d *classConstVisitor) LeaveNode(w ir.Node) {
	state.LeaveNode(&d.st, w)
}

type blockMethodCallVisitor struct {
	ctx *linter.BlockContext

	className  string
	methodName string

	filename string

	addFound func(f vscode.Location)
}

func (d *blockMethodCallVisitor) BeforeEnterNode(n ir.Node) {
	call, ok := n.(*ir.MethodCallExpr)
	if !ok {
		return
	}
	var methodName string
	switch id := call.Method.(type) {
	case *ir.Identifier:
		methodName = id.Value
	default:
		return
	}

	if methodName != d.methodName {
		return
	}

	exprType := solver.ExprType(d.ctx.Scope(), d.ctx.ClassParseState(), call.Variable)

	exprType.Iterate(func(typ meta.Type) {
		className := typ.String()
		m, ok := solver.FindMethod(className, methodName)
		realClassName := m.ImplName()

		if ok && realClassName == d.className {
			if pos := ir.GetPosition(call); pos != nil {
				d.addFound(refPosition(d.filename, pos))
			}
		}
	})
}

func (d *blockMethodCallVisitor) AfterEnterNode(n ir.Node)  {}
func (d *blockMethodCallVisitor) BeforeLeaveNode(n ir.Node) {}
func (d *blockMethodCallVisitor) AfterLeaveNode(n ir.Node)  {}

type blockPropertyVisitor struct {
	ctx *linter.BlockContext

	className string
	propName  string

	filename string

	addFound func(f vscode.Location)
}

func (d *blockPropertyVisitor) BeforeEnterNode(n ir.Node) {
	switch n := n.(type) {
	case *ir.Assign:
		// Linter handles assignment separately so we need too :(
		if f, ok := n.Variable.(*ir.PropertyFetchExpr); ok {
			d.handlePropertyFetch(f)
		}
	case *ir.PropertyFetchExpr:
		d.handlePropertyFetch(n)
	}
}

func (d *blockPropertyVisitor) handlePropertyFetch(n *ir.PropertyFetchExpr) {
	id, ok := n.Property.(*ir.Identifier)
	if !ok {
		return
	}

	if id.Value != d.propName {
		return
	}

	exprType := solver.ExprType(d.ctx.Scope(), d.ctx.ClassParseState(), n.Variable)
	exprType.Iterate(func(typ meta.Type) {
		className := typ.String()
		prop, ok := solver.FindProperty(className, id.Value)
		realClassName := prop.ImplName()

		if ok && realClassName == d.className {
			if pos := ir.GetPosition(n); pos != nil {
				d.addFound(refPosition(d.filename, pos))
			}
		}
	})
}

func (d *blockPropertyVisitor) AfterEnterNode(n ir.Node)  {}
func (d *blockPropertyVisitor) BeforeLeaveNode(n ir.Node) {}
func (d *blockPropertyVisitor) AfterLeaveNode(n ir.Node)  {}
