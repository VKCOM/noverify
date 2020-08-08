package state

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

// EnterNode must be called upon entering new node to update current state.
func EnterNode(st *meta.ClassParseState, n ir.Node) {
	switch n := n.(type) {
	case *ir.FunctionStmt:
		st.CurrentFunction = n.FunctionName.Value
	case *ir.ClassMethodStmt:
		st.CurrentFunction = n.MethodName.Value

	case *ir.NamespaceStmt:
		// TODO: handle another namespace syntax:
		// namespace NS { ... }
		nm, ok := n.NamespaceName.(*ir.Name)
		if ok {
			st.Namespace = `\` + meta.NameToString(nm)
		}
	case *ir.UseListStmt:
		if n.UseType == nil {
			for _, u := range n.Uses {
				if u, ok := u.(*ir.UseStmt); ok {
					handleUseClass(st, u)
				}
			}
		} else if id, ok := n.UseType.(*ir.Identifier); ok && id.Value == "function" {
			for _, u := range n.Uses {
				if u, ok := u.(*ir.UseStmt); ok {
					handleUseFunction(st, u)
				}
			}
		}
	case *ir.InterfaceStmt:
		st.IsTrait = false
		st.CurrentClass = st.Namespace + `\` + n.InterfaceName.Value
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
		if n.Extends != nil {
			for _, iface := range n.Extends.InterfaceNames {
				ifaceName, ok := solver.GetClassName(st, iface)
				if ok {
					st.CurrentParentInterfaces = append(st.CurrentParentInterfaces, ifaceName)
				}
			}
		}

	case *ir.ClassStmt:
		// TODO: handle anonymous classes (they can be nested as well)
		st.IsTrait = false
		id := n.ClassName
		st.CurrentClass = st.Namespace + `\` + id.Value
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
		if n.Extends != nil {
			st.CurrentParentClass, _ = solver.GetClassName(st, n.Extends.ClassName)
		}
	case *ir.TraitStmt:
		st.IsTrait = true
		st.CurrentClass = st.Namespace + `\` + n.TraitName.Value
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
	}
}

func handleUseClass(st *meta.ClassParseState, n *ir.UseStmt) {
	// TODO: there exists groupUse and other stuff
	if st.Uses == nil {
		st.Uses = make(map[string]string)
	}

	parts := n.Use.(*ir.Name).Parts
	var alias string

	if n.Alias != nil {
		alias = n.Alias.Value
	} else {
		alias = parts[len(parts)-1].(*ir.NamePart).Value
	}

	st.Uses[alias] = `\` + meta.NameToString(n.Use.(*ir.Name))
}

func handleUseFunction(st *meta.ClassParseState, n *ir.UseStmt) {
	// TODO: there exists groupUse and other stuff
	if st.FunctionUses == nil {
		st.FunctionUses = make(map[string]string)
	}

	parts := n.Use.(*ir.Name).Parts
	var alias string

	if n.Alias != nil {
		alias = n.Alias.Value
	} else {
		alias = parts[len(parts)-1].(*ir.NamePart).Value
	}

	st.FunctionUses[alias] = `\` + meta.NameToString(n.Use.(*ir.Name))
}

// LeaveNode must be called upon leaving a node to update current state.
func LeaveNode(st *meta.ClassParseState, n ir.Node) {
	switch n.(type) {
	case *ir.ClassMethodStmt, *ir.FunctionStmt:
		st.CurrentFunction = ""

	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		st.IsTrait = false
		st.CurrentClass = ""
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
	}
}
