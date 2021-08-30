package state

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linter/autogen"
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
		if n.NamespaceName != nil {
			st.Namespace = `\` + n.NamespaceName.Value
		}

	case *ir.GroupUseStmt:
		list := &ir.UseListStmt{
			UseType: n.UseType,
			Uses:    n.UseList,
		}
		handleUseList(`\`+n.Prefix.Value, st, list)

	case *ir.UseListStmt:
		handleUseList("", st, n)

	case *ir.AnonClassExpr:
		st.IsTrait = false
		st.IsInterface = false

		name := autogen.GenerateAnonClassName(n, st.CurrentFile)

		st.CurrentClass = st.Namespace + name
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
		if n.Extends != nil {
			st.CurrentParentClass, _ = solver.GetClassName(st, n.Extends.ClassName)
		}

	case *ir.InterfaceStmt:
		st.IsTrait = false
		st.IsInterface = true
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
		st.IsInterface = false
		id := n.ClassName
		st.CurrentClass = st.Namespace + `\` + id.Value
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
		if n.Extends != nil {
			st.CurrentParentClass, _ = solver.GetClassName(st, n.Extends.ClassName)
		}
	case *ir.TraitStmt:
		st.IsTrait = true
		st.IsInterface = false
		st.CurrentClass = st.Namespace + `\` + n.TraitName.Value
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
	}
}

func handleUseList(prefix string, st *meta.ClassParseState, n *ir.UseListStmt) {
	if n.UseType == nil {
		for _, u := range n.Uses {
			if u, ok := u.(*ir.UseStmt); ok {
				handleUseClass(prefix, st, u)
			}
		}
		return
	}

	useType := n.UseType.Value

	if useType == "function" {
		for _, u := range n.Uses {
			if u, ok := u.(*ir.UseStmt); ok {
				handleUseFunction(prefix, st, u)
			}
		}
	}
}

func handleUseClass(prefix string, st *meta.ClassParseState, n *ir.UseStmt) {
	// TODO: there exists groupUse and other stuff
	if st.Uses == nil {
		st.Uses = make(map[string]string)
	}

	var alias string

	if n.Alias != nil {
		alias = n.Alias.Value
	} else {
		alias = n.Use.LastPart()
	}

	st.Uses[alias] = prefix + `\` + n.Use.Value
}

func handleUseFunction(prefix string, st *meta.ClassParseState, n *ir.UseStmt) {
	// TODO: there exists groupUse and other stuff
	if st.FunctionUses == nil {
		st.FunctionUses = make(map[string]string)
	}

	var alias string

	if n.Alias != nil {
		alias = n.Alias.Value
	} else {
		alias = n.Use.LastPart()
	}

	st.FunctionUses[alias] = prefix + `\` + n.Use.Value
}

// LeaveNode must be called upon leaving a node to update current state.
func LeaveNode(st *meta.ClassParseState, n ir.Node) {
	switch n.(type) {
	case *ir.ClassMethodStmt, *ir.FunctionStmt:
		st.CurrentFunction = ""

	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt, *ir.AnonClassExpr:
		st.IsTrait = false
		st.IsInterface = false
		st.CurrentClass = ""
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
	}
}
