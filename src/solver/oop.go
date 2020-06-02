package solver

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
)

// GetFuncName resolves func name for the specified func node.
//
// It doesn't handle dynamic function calls where funcNode is
// a variable or some other kind of non-name expression.
//
// The main purpose of this function is to expand a function name to a FQN.
func GetFuncName(cs *meta.ClassParseState, funcNode node.Node) (funcName string, ok bool) {
	switch nm := funcNode.(type) {
	case *name.Name:
		nameStr := meta.NameToString(nm)
		firstPart := nm.Parts[0].(*name.NamePart).Value
		if alias, ok := cs.FunctionUses[firstPart]; ok {
			if len(nm.Parts) == 1 {
				nameStr = alias
			} else {
				// handle situations like 'use NS\Foo; Foo\Bar::doSomething();'
				nameStr = alias + `\` + meta.NamePartsToString(nm.Parts[1:])
			}
			return nameStr, true
		}
		fqName := cs.Namespace + `\` + nameStr
		_, ok := meta.Info.GetFunction(fqName)
		if ok {
			return fqName, true
		}
		return `\` + nameStr, true

	case *name.FullyQualified:
		return meta.FullyQualifiedToString(nm), true

	default:
		return "", false
	}
}

// GetClassName resolves class name for specified class node (as used in static calls, property fetch, etc)
func GetClassName(cs *meta.ClassParseState, classNode node.Node) (className string, ok bool) {
	if nm, ok := classNode.(*name.FullyQualified); ok {
		return meta.FullyQualifiedToString(nm), true
	}

	var firstPart string
	var parts []node.Node
	var partsCount int

	switch nm := classNode.(type) {
	case *node.Identifier:
		// actually only handles "static::"
		className = nm.Value
		firstPart = nm.Value
		partsCount = 1 // hack for the later if partsCount == 1
	case *name.Name:
		className = meta.NameToString(nm)
		firstPart = nm.Parts[0].(*name.NamePart).Value
		parts = nm.Parts
		partsCount = len(parts)
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
			className = alias + `\` + meta.NamePartsToString(parts[1:])
		}
	} else {
		className = cs.Namespace + `\` + className
	}

	return className, true
}

// GetConstant searches for specified constant in const fetch.
func GetConstant(cs *meta.ClassParseState, constNode node.Node) (constName string, ci meta.ConstantInfo, ok bool) {
	switch nm := constNode.(type) {
	case *name.Name:
		nameStr := meta.NameToString(nm)
		nameWithNs := cs.Namespace + `\` + nameStr
		ci, ok = meta.Info.GetConstant(nameWithNs)
		if ok {
			return nameWithNs, ci, true
		}

		if cs.Namespace != "" {
			nameRootNs := `\` + nameStr
			ci, ok = meta.Info.GetConstant(nameRootNs)
			if ok {
				return nameRootNs, ci, ok
			}
		}
	case *name.FullyQualified:
		nameStr := meta.FullyQualifiedToString(nm)
		ci, ok = meta.Info.GetConstant(nameStr)
		if ok {
			return nameStr, ci, true
		}
	}

	return "", meta.ConstantInfo{}, false
}
