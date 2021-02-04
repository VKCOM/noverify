package linter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/z7zmey/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

// This file contains methods that were defined inside BlockWalker
// but in fact they can be separated and used in other contexts.

func findMethod(info *meta.Info, className, methodName string) (res solver.FindMethodResult, magic, ok bool) {
	m, ok := solver.FindMethod(info, className, methodName)
	if ok {
		return m, false, true
	}
	m, ok = solver.FindMethod(info, className, `__call`)
	if ok {
		return m, true, true
	}
	return m, false, false
}

func findProperty(info *meta.Info, className, propName string) (res solver.FindPropertyResult, magic, ok bool) {
	p, ok := solver.FindProperty(info, className, propName)
	if ok {
		return p, false, true
	}
	m, ok := solver.FindMethod(info, className, `__get`)
	if ok {
		// Construct a dummy property from the magic method.
		p.ClassName = m.ClassName
		p.TraitName = m.TraitName
		p.Info = meta.PropertyInfo{
			Pos:         m.Info.Pos,
			Typ:         m.Info.Typ,
			AccessLevel: m.Info.AccessLevel,
		}
		return p, true, true
	}
	return p, false, false
}

func classDistance(st *meta.ClassParseState, class string) int {
	if class == st.CurrentClass {
		return 0
	}
	if class == st.CurrentParentClass {
		return 1
	}
	// TODO: traverse the class hierarchy?
	// It looks like a quite rare corner case, so lets
	// not introduce another map allocating loop per
	// every property/method lookup.
	return 2
}

func enoughArgs(args []ir.Node, fn meta.FuncInfo) bool {
	if len(args) < fn.MinParamsCnt {
		// If the last argument is ...$arg, then assume it is an array with
		// sufficient values for the parameters
		if len(args) == 0 || !args[len(args)-1].(*ir.Argument).Variadic {
			return false
		}
	}
	return true
}

// checks whether or not we can access to className::method/property/constant/etc from this context
func canAccess(st *meta.ClassParseState, className string, accessLevel meta.AccessLevel) bool {
	switch accessLevel {
	case meta.Private:
		return st.CurrentClass == className
	case meta.Protected:
		if st.CurrentClass == className {
			return true
		}

		// TODO: perhaps shpuld extract this common logic with visited map somewhere
		visited := make(map[string]struct{}, 8)
		parent := st.CurrentParentClass
		for parent != "" {
			if _, ok := visited[parent]; ok {
				return false
			}

			visited[parent] = struct{}{}

			if parent == className {
				return true
			}

			class, ok := st.Info.GetClass(parent)
			if !ok {
				return false
			}

			parent = class.Parent
		}

		return false
	case meta.Public:
		return true
	}

	panic("Invalid access level")
}

func nodeEqual(st *meta.ClassParseState, x, y ir.Node) bool {
	if x == nil || y == nil {
		return x == y
	}
	switch x := x.(type) {
	case *ir.ConstFetchExpr:
		y, ok := y.(*ir.ConstFetchExpr)
		if !ok {
			return false
		}
		_, info1, ok := solver.GetConstant(st, x.Constant)
		if !ok {
			return false
		}
		_, info2, ok := solver.GetConstant(st, y.Constant)
		if !ok {
			return false
		}

		return info1.Value.IsEqual(info2.Value)

	default:
		return irutil.NodeEqual(x, y)
	}
}

func getCaseStmts(c ir.Node) (cond ir.Node, list []ir.Node) {
	switch c := c.(type) {
	case *ir.CaseStmt:
		cond = c.Cond
		list = c.Stmts
	case *ir.DefaultStmt:
		list = c.Stmts
	default:
		panic(fmt.Errorf("Unexpected type in switch statement: %T", c))
	}

	return cond, list
}

var fallthroughMarkerRegex = func() *regexp.Regexp {
	markers := []string{
		"fallthrough",
		"fall through",
		"falls through",
		"no break",
	}

	pattern := `(?:/\*|//)\s?(?:` + strings.Join(markers, `|`) + `)`
	return regexp.MustCompile(pattern)
}()

func caseHasFallthroughComment(n ir.Node) bool {
	var docTkn *token.Token

	switch n := n.(type) {
	case *ir.CaseStmt:
		docTkn = n.CaseTkn
	case *ir.DefaultStmt:
		docTkn = n.DefaultTkn
	default:
		return false
	}

	for _, tok := range docTkn.FreeFloating {
		if tok.ID != token.T_COMMENT {
			continue
		}

		if fallthroughMarkerRegex.Match(tok.Value) {
			return true
		}
	}
	return false
}
