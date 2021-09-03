package attributes

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

func Name(attr *ir.Attribute, state *meta.ClassParseState) string {
	fqn, ok := solver.GetClassName(state, attr.Name)
	if !ok {
		return ""
	}
	return fqn
}

func NamedArgument(attr *ir.Attribute, name string) (ir.Node, bool) {
	for i := range attr.Args {
		arg := attr.Arg(i)
		if arg.Name != nil && arg.Name.Value == name {
			return arg.Expr, true
		}
	}

	return nil, false
}

func NamedStringArgument(attr *ir.Attribute, name string) string {
	expr, ok := NamedArgument(attr, name)
	if !ok {
		return ""
	}

	str, ok := expr.(*ir.String)
	if !ok {
		return ""
	}

	return str.Value
}

func Each(groups []*ir.AttributeGroup, cb func(attr *ir.Attribute) bool) {
	for _, group := range groups {
		if group == nil {
			continue
		}

		for _, attr := range group.Attrs {
			if !cb(attr) {
				return
			}
		}
	}
}
