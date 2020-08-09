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
		pos := ir.GetPosition(n.Function)

		if d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		var fun meta.FuncInfo
		var ok bool
		var nameStr string

		switch nm := n.Function.(type) {
		case *ir.Name:
			nameStr = meta.NameToString(nm)
			fun, ok = meta.Info.GetFunction(d.st.Namespace + `\` + nameStr)
			if !ok && d.st.Namespace != "" {
				fun, ok = meta.Info.GetFunction(`\` + nameStr)
			}
		case *ir.FullyQualifiedName:
			nameStr = meta.FullyQualifiedToString(nm)
			fun, ok = meta.Info.GetFunction(nameStr)
		}

		if ok {
			d.result = append(d.result, vscode.Location{
				URI: uri.File(fun.Pos.Filename),
				Range: vscode.Range{
					Start: vscode.Position{Line: int(fun.Pos.Line) - 1},
					End:   vscode.Position{Line: int(fun.Pos.Line) - 1},
				},
			})
		}

		lintdebug.Send("Found function %s: %s:%d", nameStr, fun.Pos.Filename, fun.Pos.Line)
	case *ir.StaticCallExpr:
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

		m, ok := solver.FindMethod(className, id.Value)
		if ok {
			d.result = append(d.result, vscode.Location{
				URI: uri.File(m.Info.Pos.Filename),
				Range: vscode.Range{
					Start: vscode.Position{Line: int(m.Info.Pos.Line) - 1},
					End:   vscode.Position{Line: int(m.Info.Pos.Line) - 1},
				},
			})
		}
	case *ir.MethodCallExpr:
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
			m, ok := solver.FindMethod(t, id.Value)
			if !ok {
				lintdebug.Send("Could not find method for %s::%s", t, id.Value)
				return
			}

			d.result = append(d.result, vscode.Location{
				URI: uri.File(m.Info.Pos.Filename),
				Range: vscode.Range{
					Start: vscode.Position{Line: int(m.Info.Pos.Line) - 1},
					End:   vscode.Position{Line: int(m.Info.Pos.Line) - 1},
				},
			})
		})
	case *ir.PropertyFetchExpr:
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

			d.result = append(d.result, vscode.Location{
				URI: uri.File(p.Info.Pos.Filename),
				Range: vscode.Range{
					Start: vscode.Position{Line: int(p.Info.Pos.Line) - 1},
					End:   vscode.Position{Line: int(p.Info.Pos.Line) - 1},
				},
			})
		})
	case *ir.ConstFetchExpr:
		pos := ir.GetPosition(n.Constant)

		if d.position > pos.EndPos || d.position < pos.StartPos {
			return true
		}

		_, c, ok := solver.GetConstant(&d.st, n.Constant)

		if ok {
			d.result = append(d.result, vscode.Location{
				URI: uri.File(c.Pos.Filename),
				Range: vscode.Range{
					Start: vscode.Position{Line: int(c.Pos.Line) - 1},
					End:   vscode.Position{Line: int(c.Pos.Line) - 1},
				},
			})
		}
	case *ir.ClassConstFetchExpr:
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
			d.result = append(d.result, vscode.Location{
				URI: uri.File(c.Pos.Filename),
				Range: vscode.Range{
					Start: vscode.Position{Line: int(c.Pos.Line) - 1},
					End:   vscode.Position{Line: int(c.Pos.Line) - 1},
				},
			})
		}

	case *ir.Name:
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

		d.result = append(d.result, vscode.Location{
			URI: uri.File(c.Pos.Filename),
			Range: vscode.Range{
				Start: vscode.Position{Line: int(c.Pos.Line) - 1},
				End:   vscode.Position{Line: int(c.Pos.Line) - 1},
			},
		})
	case *ir.FullyQualifiedName:
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

		lintdebug.Send("name:%s , uri:%s", c.Pos.Filename, uri.File(c.Pos.Filename))

		d.result = append(d.result, vscode.Location{
			URI: uri.File(c.Pos.Filename),
			Range: vscode.Range{
				Start: vscode.Position{Line: int(c.Pos.Line) - 1},
				End:   vscode.Position{Line: int(c.Pos.Line) - 1},
			},
		})
	}

	return true
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
