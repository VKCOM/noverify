package solver

import (
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
)

// ResolveType resolves function calls, method calls and global variables
func ResolveType(typ string, visitedMap map[string]struct{}) map[string]struct{} {
	if _, ok := visitedMap[typ]; ok {
		return nil
	}

	if len(typ) == 0 || typ[0] >= meta.WMax {
		return identityType(typ)
	}

	res := make(map[string]struct{})
	visitedMap[typ] = struct{}{}

	switch typ[0] {
	case meta.WGlobal:
		varTyp, ok := meta.Info.GetVarNameType(meta.UnwrapGlobal(typ))
		if ok {
			for tt := range ResolveTypes(varTyp, visitedMap) {
				res[tt] = struct{}{}
			}
		}
	case meta.WConstant:
		ci, ok := meta.Info.GetConstant(meta.UnwrapConstant(typ))
		if ok {
			for tt := range ResolveTypes(ci.Typ, visitedMap) {
				res[tt] = struct{}{}
			}
		}
	case meta.WArrayOf:
		for tt := range ResolveType(meta.UnwrapArrayOf(typ), visitedMap) {
			res[tt+"[]"] = struct{}{}
		}
	case meta.WElemOf:
		for tt := range ResolveType(meta.UnwrapElemOf(typ), visitedMap) {
			if strings.HasSuffix(tt, "[]") {
				res[strings.TrimSuffix(tt, "[]")] = struct{}{}
			} else if tt == "mixed" {
				res["mixed"] = struct{}{}
			}
		}
	case meta.WFunctionCall:
		nm := meta.UnwrapFunctionCall(typ)
		fn, ok := meta.Info.GetFunction(nm)
		// functions can fall back to root namespace
		if !ok && strings.Count(nm, `\`) > 1 {
			fn, ok = meta.Info.GetFunction(nm[strings.LastIndex(nm, `\`):])
		}

		if ok {
			return ResolveTypes(fn.Typ, visitedMap)
		}
	case meta.WInstanceMethodCall:
		expr, methodName := meta.UnwrapInstanceMethodCall(typ)

		for className := range ResolveType(expr, visitedMap) {
			info, _, ok := FindMethod(className, methodName)
			if ok {
				for tt := range ResolveTypes(info.Typ, visitedMap) {
					res[tt] = struct{}{}
				}
			}
		}
	case meta.WInstancePropertyFetch:
		expr, propertyName := meta.UnwrapInstancePropertyFetch(typ)

		for className := range ResolveType(expr, visitedMap) {
			info, _, ok := FindProperty(className, propertyName)
			if ok {
				for tt := range ResolveTypes(info.Typ, visitedMap) {
					res[tt] = struct{}{}
				}
			}
		}
	case meta.WStaticMethodCall:
		className, methodName := meta.UnwrapStaticMethodCall(typ)
		info, _, ok := FindMethod(className, methodName)
		if ok {
			return ResolveTypes(info.Typ, visitedMap)
		}
	case meta.WStaticPropertyFetch:
		className, propertyName := meta.UnwrapStaticPropertyFetch(typ)
		info, _, ok := FindProperty(className, propertyName)
		if ok {
			return ResolveTypes(info.Typ, visitedMap)
		}
	default:
		panic(fmt.Sprintf("Unexpected type: %d", typ[0]))
	}

	return res
}

// FindMethod searches for a method in specified class. meta.Info.RLock() must be held
func FindMethod(className string, methodName string) (res meta.FuncInfo, implClassName string, ok bool) {
	return findMethod(className, methodName, make(map[string]struct{}))
}

func findMethod(className string, methodName string, visitedMap map[string]struct{}) (res meta.FuncInfo, implClassName string, ok bool) {
	for {
		if _, ok := visitedMap[className]; ok {
			return res, "", false
		}
		visitedMap[className] = struct{}{}

		class, ok := meta.Info.GetClass(className)
		if !ok {
			class, ok = meta.Info.GetTrait(className)
			if !ok {
				return res, "", false
			}
		}

		res, ok = class.Methods[methodName]
		if ok {
			return res, className, ok
		}

		for trait := range class.Traits {
			res, implClassName, ok = findMethod(trait, methodName, visitedMap)
			if ok {
				return res, implClassName, ok
			}
		}

		// interfaces support multiple inheritance and I use a separate property for that for now
		for _, parentIfaceName := range class.ParentInterfaces {
			res, implClassName, ok = findMethod(parentIfaceName, methodName, visitedMap)
			if ok {
				return res, implClassName, ok
			}
		}

		if class.Parent == "" {
			return res, "", false
		}

		className = class.Parent
	}
}

func FindProperty(className string, propertyName string) (res meta.PropertyInfo, implClassName string, ok bool) {
	for {
		class, ok := meta.Info.GetClass(className)
		if !ok {
			return res, "", false
		}

		res, ok = class.Properties[propertyName]
		if ok {
			return res, className, ok
		}

		if class.Parent == "" {
			return res, "", false
		}

		className = class.Parent
	}
}

// Implements checks if className implements interfaceName
func Implements(className string, interfaceName string) bool {
	visited := make(map[string]struct{}, 8)

	for {
		class, ok := meta.Info.GetClass(className)
		if !ok {
			return false
		}

		_, ok = class.Interfaces[interfaceName]
		if ok {
			return true
		}

		for iface := range class.Interfaces {
			if interfaceExtends(iface, interfaceName, visited) {
				return true
			}
		}

		if class.Parent == "" {
			return false
		}

		className = class.Parent
	}
}

// interfaceExtends checks if interface orig extends interface parent
func interfaceExtends(orig string, parent string, visited map[string]struct{}) bool {
	if _, ok := visited[orig]; ok {
		return false
	}

	visited[orig] = struct{}{}

	class, ok := meta.Info.GetClass(orig)
	if !ok {
		return false
	}

	for _, iface := range class.ParentInterfaces {
		if iface == parent {
			return true
		}

		if interfaceExtends(iface, parent, visited) {
			return true
		}
	}

	return false
}

// FindConstant searches for a costant in specified class and returns actual class that contains the constant.
func FindConstant(className string, constName string) (res meta.ConstantInfo, implClassName string, ok bool) {
	visitedClasses := make(map[string]struct{}, 8) // expecting to be not so many inheritance levels
	return findConstant(className, constName, visitedClasses)
}

func findConstant(className string, constName string, visitedClasses map[string]struct{}) (res meta.ConstantInfo, implClassName string, ok bool) {
	for {
		// check for inheritance loops
		if _, ok := visitedClasses[className]; ok {
			return res, "", false
		}

		visitedClasses[className] = struct{}{}

		class, ok := meta.Info.GetClass(className)
		if !ok {
			return res, "", false
		}

		// inferfaces can have constants...
		for ifaceName := range class.Interfaces {
			res, implClassName, ok = findConstant(ifaceName, constName, visitedClasses)
			if ok {
				return res, implClassName, ok
			}
		}

		res, ok = class.Constants[constName]
		if ok {
			return res, className, ok
		}

		// interfaces support multiple inheritance and I use a separate property for that for now
		for _, parentIfaceName := range class.ParentInterfaces {
			res, implClassName, ok = findConstant(parentIfaceName, constName, visitedClasses)
			if ok {
				return res, implClassName, ok
			}
		}

		if class.Parent == "" {
			return res, "", false
		}

		className = class.Parent
	}
}

func identityType(typ string) map[string]struct{} {
	res := make(map[string]struct{})
	res[typ] = struct{}{}
	return res
}

func ResolveTypes(m *meta.TypesMap, visitedMap map[string]struct{}) map[string]struct{} {
	res := make(map[string]struct{}, m.Len())

	m.Iterate(func(t string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic during parsing '%s'", meta.NewTypesMap(t))
				panic(r)
			}
		}()

		for tt := range ResolveType(t, visitedMap) {
			res[tt] = struct{}{}
		}
	})

	return res
}
