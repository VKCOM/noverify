package solver

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
)

// GetFuncName resolves func name for the specified func node.
//
// It doesn't handle dynamic function calls where funcNode is
// a variable or some other kind of non-name expression.
//
// The main purpose of this function is to expand a function name to a FQN.
func GetFuncName(cs *meta.ClassParseState, funcNode ir.Node) (funcName string, ok bool) {
	switch nm := funcNode.(type) {
	case *ir.Name:
		if nm.IsFullyQualified() {
			return nm.Value, true
		}

		nameStr := nm.Value
		firstPart := nm.FirstPart()
		if alias, ok := cs.FunctionUses[firstPart]; ok {
			if nm.NumParts() == 1 {
				nameStr = alias
			} else {
				// handle situations like 'use NS\Foo; Foo\Bar::doSomething();'
				nameStr = alias + `\` + nm.RestParts()
			}
			return nameStr, true
		}
		fqName := cs.Namespace + `\` + nameStr
		_, ok := cs.Info.GetFunction(fqName)
		if ok {
			return fqName, true
		}
		return `\` + nameStr, true

	default:
		return "", false
	}
}

// GetClassName resolves class name for specified class node (as used in static calls, property fetch, etc)
func GetClassName(cs *meta.ClassParseState, classNode ir.Node) (className string, ok bool) {
	var firstPart string
	var partsCount int
	var restParts string

	switch nm := classNode.(type) {
	case *ir.Identifier:
		// actually only handles "static::"
		className = nm.Value
		firstPart = nm.Value
		partsCount = 1 // hack for the later if partsCount == 1
	case *ir.Name:
		if nm.IsFullyQualified() {
			return nm.Value, true
		}
		className = nm.Value
		firstPart, restParts = nm.HeadTail()
		partsCount = nm.NumParts()
	default:
		return "", false
	}

	if className == "self" || className == "static" || className == "$this" {
		className = cs.CurrentClass
	} else if className == "parent" {
		className = cs.CurrentParentClass
	} else if alias, ok := cs.Uses[firstPart]; ok {
		if partsCount == 1 {
			className = alias
		} else {
			// handle situations like 'use NS\Foo; Foo\Bar::doSomething();'
			className = alias + `\` + restParts
		}
	} else {
		className = cs.Namespace + `\` + className
	}

	return className, true
}

// GetConstant searches for specified constant in const fetch.
func GetConstant(cs *meta.ClassParseState, constNode ir.Node) (constName string, ci meta.ConstInfo, ok bool) {
	nm, ok := constNode.(*ir.Name)
	if !ok {
		return "", meta.ConstInfo{}, false
	}

	nameStr := nm.Value
	if nm.IsFullyQualified() {
		ci, ok = cs.Info.GetConstant(nameStr)
		if ok {
			return nameStr, ci, true
		}
	}

	nameWithNs := cs.Namespace + `\` + nameStr
	ci, ok = cs.Info.GetConstant(nameWithNs)
	if ok {
		return nameWithNs, ci, true
	}

	if cs.Namespace != "" {
		nameRootNs := `\` + nameStr
		ci, ok = cs.Info.GetConstant(nameRootNs)
		if ok {
			return nameRootNs, ci, ok
		}
	}

	return "", meta.ConstInfo{}, false
}

// Extends reports whether derived class extends the base class.
// It returns false for the derived==base case.
func Extends(info *meta.Info, derived, base string) bool {
	if derived == base {
		return false
	}
	class, ok := info.GetClass(derived)
	if !ok {
		return false
	}
	switch class.Parent {
	case base:
		return true
	case "":
		return false
	default:
		return Extends(info, class.Parent, base)
	}
}
