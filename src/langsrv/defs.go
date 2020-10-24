package langsrv

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/state"
	"github.com/VKCOM/noverify/src/vscode"
	"go.lsp.dev/uri"
)

type definitionWalker struct {
	st meta.ClassParseState

	position int
	scopes   map[ir.Node]*meta.Scope

	result      []vscode.Location
	foundScopes []*meta.Scope
}

func safeExprType(sc *meta.Scope, cs *meta.ClassParseState, n ir.Node) (res meta.TypesMap) {
	defer func() {
		if r := recover(); r != nil {
			res = meta.NewTypesMap(fmt.Sprintf("Panic: %s", fmt.Sprint(r)))
		}
	}()

	res = solver.ExprType(sc, cs, n)
	return
}

// EnterNode is invoked at every node in hierarchy
func (d *definitionWalker) EnterNode(n ir.Node) bool {
	sc, ok := d.scopes[n]
	if ok {
		d.foundScopes = append(d.foundScopes, sc)
	}

	state.EnterNode(&d.st, n)

	switch n := n.(type) {
	case *ir.FunctionCallExpr:
		return d.processFunctionCallExpr(n)

	case *ir.StaticCallExpr:
		return d.processStaticCallExpr(n)

	case *ir.MethodCallExpr:
		return d.processMethodCallExpr(n)

	case *ir.PropertyFetchExpr:
		return d.processPropertyFetchExpr(n)

	case *ir.ConstFetchExpr:
		return d.processConstFetchExpr(n)

	case *ir.ClassConstFetchExpr:
		return d.processClassConstFetchExpr(n)

	case *ir.Name:
		return d.processName(n)
	}

	return true
}

func (d *definitionWalker) processName(n *ir.Name) bool {
	pos := ir.GetPosition(n)

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return true
	}

	className, ok := solver.GetClassName(&d.st, n)
	if !ok {
		return true
	}

	c, ok := meta.Info.GetClassOrTrait(className)

	if !ok {
		return true
	}

	d.appendResult(c.Pos.Filename, int(c.Pos.Line)-1, int(c.Pos.Line)-1)
	return true
}

func (d *definitionWalker) processClassConstFetchExpr(n *ir.ClassConstFetchExpr) bool {
	if pos := ir.GetPosition(n.ConstantName); d.position > pos.EndPos || d.position < pos.StartPos {
		return true
	}

	constName := n.ConstantName

	if constName.Value == `class` || constName.Value == `CLASS` {
		return false
	}

	className, ok := solver.GetClassName(&d.st, n.Class)
	if !ok {
		return false
	}

	if c, _, ok := solver.FindConstant(className, constName.Value); ok {
		d.appendResult(c.Pos.Filename, int(c.Pos.Line)-1, int(c.Pos.Line)-1)
	}
	return true
}

func (d *definitionWalker) processConstFetchExpr(n *ir.ConstFetchExpr) bool {
	pos := ir.GetPosition(n.Constant)

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return true
	}

	_, c, ok := solver.GetConstant(&d.st, n.Constant)

	if ok {
		d.appendResult(c.Pos.Filename, int(c.Pos.Line)-1, int(c.Pos.Line)-1)
	}
	return true
}

func (d *definitionWalker) processPropertyFetchExpr(n *ir.PropertyFetchExpr) bool {
	pos := ir.GetPosition(n.Property)

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return true
	}

	lintdebug.Send("Property found")

	if len(d.foundScopes) == 0 {
		lintdebug.Send("No scope found")
		return true
	}

	foundScope := d.foundScopes[len(d.foundScopes)-1]

	// not going to resolve $obj->$someProperty
	id, ok := n.Property.(*ir.Identifier)
	if !ok {
		lintdebug.Send("Method is not identifier")
		return true
	}

	types := safeExprType(foundScope, &d.st, n.Variable)

	types.Iterate(func(t string) {
		p, ok := solver.FindProperty(t, id.Value)
		if !ok {
			lintdebug.Send("Could not find property for %s->%s", t, id.Value)
			return
		}

		d.appendResult(p.Info.Pos.Filename, int(p.Info.Pos.Line)-1, int(p.Info.Pos.Line)-1)
	})
	return true
}

func (d *definitionWalker) processMethodCallExpr(n *ir.MethodCallExpr) bool {
	pos := ir.GetPosition(n.Method)

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return true
	}

	lintdebug.Send("Method call found")

	if len(d.foundScopes) == 0 {
		lintdebug.Send("No scope found")
		return true
	}

	foundScope := d.foundScopes[len(d.foundScopes)-1]

	// not going to resolve $obj->$someMethod(); calls
	id, ok := n.Method.(*ir.Identifier)
	if !ok {
		lintdebug.Send("Method is not identifier")
		return true
	}

	types := safeExprType(foundScope, &d.st, n.Variable)

	types.Iterate(func(t string) {
		p, ok := solver.FindMethod(t, id.Value)
		if !ok {
			lintdebug.Send("Could not find method for %s::%s", t, id.Value)
			return
		}

		d.appendResult(p.Info.Pos.Filename, int(p.Info.Pos.Line)-1, int(p.Info.Pos.Line)-1)
	})
	return true
}

func (d *definitionWalker) processFunctionCallExpr(n *ir.FunctionCallExpr) bool {
	pos := ir.GetPosition(n.Function)

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return true
	}

	var fun meta.FuncInfo
	var ok bool
	var nameStr string

	if nm, isName := n.Function.(*ir.Name); isName {
		nameStr = nm.Value
		if nm.IsFullyQualified() {
			fun, ok = meta.Info.GetFunction(nameStr)
		} else {
			fun, ok = meta.Info.GetFunction(d.st.Namespace + `\` + nameStr)
			if !ok && d.st.Namespace != "" {
				fun, ok = meta.Info.GetFunction(`\` + nameStr)
			}
		}
	}

	if ok {
		d.appendResult(fun.Pos.Filename, int(fun.Pos.Line)-1, int(fun.Pos.Line)-1)
	}

	lintdebug.Send("Found function %s: %s:%d", nameStr, fun.Pos.Filename, fun.Pos.Line)
	return true
}

func (d *definitionWalker) processStaticCallExpr(n *ir.StaticCallExpr) bool {
	pos := ir.GetPosition(n.Call)

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return true
	}

	lintdebug.Send("Static call found")

	// not going to resolve $obj->$someMethod(); calls
	id, ok := n.Call.(*ir.Identifier)
	if !ok {
		lintdebug.Send("Static Call is not identifier")
		return true
	}

	className, ok := solver.GetClassName(&d.st, n.Class)
	if !ok {
		return true
	}

	p, ok := solver.FindMethod(className, id.Value)
	if ok {
		d.appendResult(p.Info.Pos.Filename, int(p.Info.Pos.Line)-1, int(p.Info.Pos.Line)-1)
	}
	return true
}

func (d *definitionWalker) appendResult(file string, start, end int) {
	d.result = append(d.result, vscode.Location{
		URI: uri.File(file),
		Range: vscode.Range{
			Start: vscode.Position{Line: start},
			End:   vscode.Position{Line: end},
		},
	})
}

// LeaveNode is invoked after node process
func (d *definitionWalker) LeaveNode(n ir.Node) {
	if d.scopes != nil {
		_, ok := d.scopes[n]
		if ok && len(d.foundScopes) > 0 {
			d.foundScopes = d.foundScopes[0 : len(d.foundScopes)-1]
		}
	}

	state.LeaveNode(&d.st, n)
}
