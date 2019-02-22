package langsrv

import (
	"strings"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/state"
	"github.com/VKCOM/noverify/src/vscode"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/php7"
	"github.com/z7zmey/php-parser/position"
	"github.com/z7zmey/php-parser/walker"
)

func findFunctionReferences(funcName string) []vscode.Location {
	substr := funcName
	if idx := strings.LastIndexByte(funcName, '\\'); idx >= 0 {
		substr = funcName[idx+1:]
	}

	return findReferences(substr, func(filename string, rootNode node.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &funcCallVisitor{
			funcName:  funcName,
			positions: parser.GetPositions(),
			filename:  filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findStaticMethodReferences(className string, methodName string) []vscode.Location {
	return findReferences(methodName, func(filename string, rootNode node.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &staticMethodCallVisitor{
			className:  className,
			methodName: methodName,
			positions:  parser.GetPositions(),
			filename:   filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findConstantsReferences(constName string) []vscode.Location {
	return findReferences(constName, func(filename string, rootNode node.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &constVisitor{
			constName: constName,
			positions: parser.GetPositions(),
			filename:  filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findClassConstantsReferences(className string, constName string) []vscode.Location {
	return findReferences(constName, func(filename string, rootNode node.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		v := &classConstVisitor{
			className: className,
			constName: constName,
			positions: parser.GetPositions(),
			filename:  filename,
		}
		rootNode.Walk(v)
		return v.found
	})
}

func findMethodReferences(className string, methodName string) []vscode.Location {
	return findReferences(methodName, func(filename string, rootNode node.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		var found []vscode.Location

		rootWalker := linter.NewWalkerForReferencesSearcher(
			filename,
			func(ctx linter.BlockContext) linter.BlockChecker {
				return &blockMethodCallVisitor{
					ctx:        ctx,
					className:  className,
					methodName: methodName,
					filename:   filename,
					positions:  parser.GetPositions(),
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
	return findReferences(propName, func(filename string, rootNode node.Node, contents []byte, parser *php7.Parser) []vscode.Location {
		var found []vscode.Location

		rootWalker := linter.NewWalkerForReferencesSearcher(
			filename,
			func(ctx linter.BlockContext) linter.BlockChecker {
				return &blockPropertyVisitor{
					ctx:       ctx,
					className: className,
					propName:  propName,
					filename:  filename,
					positions: parser.GetPositions(),
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
	st        meta.ClassParseState
	funcName  string
	positions position.Positions
	filename  string

	found []vscode.Location
}

func (d *funcCallVisitor) GetFoundLocations() []vscode.Location {
	return d.found
}

// EnterNode is invoked at every node in hierarchy
func (d *funcCallVisitor) EnterNode(w walker.Walkable) bool {
	state.EnterNode(&d.st, w)

	switch n := w.(type) {
	case *expr.FunctionCall:
		_, nameStr, ok := getFunction(&d.st, n)
		if ok && nameStr == d.funcName {
			if pos, ok := d.positions[n]; ok {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *funcCallVisitor) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *funcCallVisitor) LeaveNode(w walker.Walkable) {
	state.LeaveNode(&d.st, w)
}

type staticMethodCallVisitor struct {
	// params
	className  string
	methodName string
	positions  position.Positions
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
func (d *staticMethodCallVisitor) EnterNode(w walker.Walkable) bool {
	state.EnterNode(&d.st, w)

	switch n := w.(type) {
	case *expr.StaticCall:
		id, ok := n.Call.(*node.Identifier)
		if !ok {
			return true
		}
		className, ok := solver.GetClassName(&d.st, n.Class)
		if !ok {
			return true
		}
		_, realClassName, ok := solver.FindMethod(className, id.Value)

		if ok && realClassName == d.className && id.Value == d.methodName {
			if pos, ok := d.positions[n]; ok {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *staticMethodCallVisitor) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *staticMethodCallVisitor) LeaveNode(w walker.Walkable) {
	state.LeaveNode(&d.st, w)
}

type constVisitor struct {
	// params
	constName string
	positions position.Positions
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
func (d *constVisitor) EnterNode(w walker.Walkable) bool {
	state.EnterNode(&d.st, w)

	switch n := w.(type) {
	case *expr.ConstFetch:
		constName, _, ok := solver.GetConstant(&d.st, n.Constant)

		if ok && constName == d.constName {
			if pos, ok := d.positions[n]; ok {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *constVisitor) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *constVisitor) LeaveNode(w walker.Walkable) {
	state.LeaveNode(&d.st, w)
}

type classConstVisitor struct {
	// params
	className string
	constName string
	positions position.Positions
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
func (d *classConstVisitor) EnterNode(w walker.Walkable) bool {
	state.EnterNode(&d.st, w)

	switch n := w.(type) {
	case *expr.ClassConstFetch:
		constName, ok := n.ConstantName.(*node.Identifier)
		if !ok {
			return true
		}

		if constName.Value == `class` || constName.Value == `CLASS` {
			return true
		}

		className, ok := solver.GetClassName(&d.st, n.Class)
		if !ok {
			return true
		}

		_, implClassName, ok := solver.FindConstant(className, constName.Value)

		if ok && constName.Value == d.constName && implClassName == d.className {
			if pos, ok := d.positions[n]; ok {
				d.found = append(d.found, refPosition(d.filename, pos))
			}
		}
	}

	return true
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *classConstVisitor) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *classConstVisitor) LeaveNode(w walker.Walkable) {
	state.LeaveNode(&d.st, w)
}

type blockMethodCallVisitor struct {
	ctx linter.BlockContext

	className  string
	methodName string

	filename  string
	positions position.Positions

	addFound func(f vscode.Location)
}

func (d *blockMethodCallVisitor) BeforeEnterNode(w walker.Walkable) {
	switch n := w.(type) {
	case *expr.MethodCall:
		var methodName string
		switch id := n.Method.(type) {
		case *node.Identifier:
			methodName = id.Value
		default:
			return
		}

		if methodName != d.methodName {
			return
		}

		exprType := solver.ExprType(d.ctx.Scope(), d.ctx.ClassParseState(), n.Variable)

		exprType.Iterate(func(typ string) {
			_, realClassName, ok := solver.FindMethod(typ, methodName)

			if ok && realClassName == d.className {
				if pos, ok := d.positions[n]; ok {
					d.addFound(refPosition(d.filename, pos))
				}
			}
		})
	}
}

func (d *blockMethodCallVisitor) AfterEnterNode(w walker.Walkable)  {}
func (d *blockMethodCallVisitor) BeforeLeaveNode(w walker.Walkable) {}
func (d *blockMethodCallVisitor) AfterLeaveNode(w walker.Walkable)  {}

type blockPropertyVisitor struct {
	ctx linter.BlockContext

	className string
	propName  string

	filename  string
	positions position.Positions

	addFound func(f vscode.Location)
}

func (d *blockPropertyVisitor) BeforeEnterNode(w walker.Walkable) {
	switch n := w.(type) {
	case *assign.Assign:
		// Linter handles assignment separately so we need too :(
		if f, ok := n.Variable.(*expr.PropertyFetch); ok {
			d.handlePropertyFetch(f)
		}
	case *expr.PropertyFetch:
		d.handlePropertyFetch(n)
	}
}

func (d *blockPropertyVisitor) handlePropertyFetch(n *expr.PropertyFetch) {
	id, ok := n.Property.(*node.Identifier)
	if !ok {
		return
	}

	if id.Value != d.propName {
		return
	}

	exprType := solver.ExprType(d.ctx.Scope(), d.ctx.ClassParseState(), n.Variable)
	exprType.Iterate(func(className string) {
		_, realClassName, ok := solver.FindProperty(className, id.Value)

		if ok && realClassName == d.className {
			if pos, ok := d.positions[n]; ok {
				d.addFound(refPosition(d.filename, pos))
			}
		}
	})
}

func (d *blockPropertyVisitor) AfterEnterNode(w walker.Walkable)  {}
func (d *blockPropertyVisitor) BeforeLeaveNode(w walker.Walkable) {}
func (d *blockPropertyVisitor) AfterLeaveNode(w walker.Walkable)  {}
