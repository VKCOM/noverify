package state

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

// EnterNode must be called upon entering new node to update current state.
func EnterNode(st *meta.ClassParseState, n walker.Walkable) {
	switch n := n.(type) {
	case *stmt.Namespace:
		// TODO: handle another namespace syntax:
		// namespace NS { ... }
		nm, ok := n.NamespaceName.(*name.Name)
		if ok {
			st.Namespace = `\` + meta.NameToString(nm)
		}
	case *stmt.UseList:
		if n.UseType == nil {
			for _, u := range n.Uses {
				if u, ok := u.(*stmt.Use); ok {
					handleUseClass(st, u)
				}
			}
		} else if id, ok := n.UseType.(*node.Identifier); ok && id.Value == "function" {
			for _, u := range n.Uses {
				if u, ok := u.(*stmt.Use); ok {
					handleUseFunction(st, u)
				}
			}
		}
	case *stmt.Interface:
		st.IsTrait = false
		st.CurrentClass = st.Namespace + `\` + n.InterfaceName.(*node.Identifier).Value
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
		for _, iface := range n.Extends {
			ifaceName, ok := solver.GetClassName(st, iface)
			if ok {
				st.CurrentParentInterfaces = append(st.CurrentParentInterfaces, ifaceName)
			}
		}

	case *stmt.Class:
		st.IsTrait = false
		id, ok := n.ClassName.(*node.Identifier)
		// TODO: handle anonymous classes (they can be nested as well)
		if ok {
			st.CurrentClass = st.Namespace + `\` + id.Value
		}
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
		if n.Extends != nil {
			st.CurrentParentClass, _ = solver.GetClassName(st, n.Extends)
		}
	case *stmt.Trait:
		st.IsTrait = true
		st.CurrentClass = st.Namespace + `\` + n.TraitName.(*node.Identifier).Value
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
	}
}

func handleUseClass(st *meta.ClassParseState, n *stmt.Use) {
	// TODO: there exists groupUse and other stuff
	if st.Uses == nil {
		st.Uses = make(map[string]string)
	}

	parts := n.Use.(*name.Name).Parts
	var alias string

	if n.Alias != nil {
		alias = n.Alias.(*node.Identifier).Value
	} else {
		alias = parts[len(parts)-1].(*name.NamePart).Value
	}

	st.Uses[alias] = `\` + meta.NameToString(n.Use.(*name.Name))
}

func handleUseFunction(st *meta.ClassParseState, n *stmt.Use) {
	// TODO: there exists groupUse and other stuff
	if st.FunctionUses == nil {
		st.FunctionUses = make(map[string]string)
	}

	parts := n.Use.(*name.Name).Parts
	var alias string

	if n.Alias != nil {
		alias = n.Alias.(*node.Identifier).Value
	} else {
		alias = parts[len(parts)-1].(*name.NamePart).Value
	}

	st.FunctionUses[alias] = `\` + meta.NameToString(n.Use.(*name.Name))
}

// LeaveNode must be called upon leaving a node to update current state.
func LeaveNode(st *meta.ClassParseState, n walker.Walkable) {
	switch n.(type) {
	case *stmt.Class, *stmt.Interface, *stmt.Trait:
		st.IsTrait = false
		st.CurrentClass = ""
		st.CurrentParentClass = ""
		st.CurrentParentInterfaces = nil
	}
}
