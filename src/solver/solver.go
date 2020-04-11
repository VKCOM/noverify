package solver

import (
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
)

var sharedMixedType = map[string]struct{}{"mixed": {}}

func mixedType() map[string]struct{} {
	if len(sharedMixedType) != 1 {
		// At least until we're 100% sure this is safe.
		panic(fmt.Sprintf("mixed type map was modified: %v", sharedMixedType))
	}
	return sharedMixedType
}

// ResolveType resolves function calls, method calls and global variables.
//   curStaticClass is current class name (if inside the class, otherwise "")
func resolveType(curStaticClass, typ string, visitedMap map[string]struct{}) (result map[string]struct{}) {
	r := resolver{visited: visitedMap}
	return r.resolveType(curStaticClass, typ)
}

// ResolveTypes resolves function calls, method calls and global variables.
//   curStaticClass is current class name (if inside the class, otherwise "")
func ResolveTypes(curStaticClass string, m meta.TypesMap, visitedMap map[string]struct{}) map[string]struct{} {
	r := resolver{visited: visitedMap}
	return r.resolveTypes(curStaticClass, m)
}

type resolver struct {
	visited map[string]struct{}
}

func (r *resolver) collectMethodCallTypes(out, possibleTypes map[string]struct{}, methodName string) map[string]struct{} {
	for className := range possibleTypes {
		info, _, ok := FindMethod(className, methodName)
		if ok {
			for tt := range r.resolveTypes(className, info.Typ) {
				out[tt] = struct{}{}
			}
		}
	}
	return out
}

func (r *resolver) resolveType(class, typ string) map[string]struct{} {
	res := r.resolveTypeNoLateStaticBinding(class, typ)

	if _, ok := res["static"]; ok {
		delete(res, "static")
		res[class] = struct{}{}
	}

	return res
}

func (r *resolver) resolveTypeNoLateStaticBinding(class, typ string) map[string]struct{} {
	visitedMap := r.visited

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
			for tt := range r.resolveTypes(class, varTyp) {
				res[tt] = struct{}{}
			}
		}
	case meta.WConstant:
		ci, ok := meta.Info.GetConstant(meta.UnwrapConstant(typ))
		if ok {
			for tt := range r.resolveTypes(class, ci.Typ) {
				res[tt] = struct{}{}
			}
		}
	case meta.WArrayOf:
		for tt := range r.resolveType(class, meta.UnwrapArrayOf(typ)) {
			res[tt+"[]"] = struct{}{}
		}
	case meta.WElemOf:
		for tt := range r.resolveType(class, meta.UnwrapElemOf(typ)) {
			switch {
			case strings.HasSuffix(tt, "[]"):
				res[strings.TrimSuffix(tt, "[]")] = struct{}{}
			case tt == "mixed":
				res["mixed"] = struct{}{}
			case Implements(tt, `\ArrayAccess`):
				offsetGet, _, ok := FindMethod(tt, "offsetGet")
				if ok {
					for tt := range r.resolveTypes(tt, offsetGet.Typ) {
						res[tt] = struct{}{}
					}
				}
			case Implements(tt, `\Traversable`):
				current, _, ok := FindMethod(tt, "current")
				if ok {
					for tt := range r.resolveTypes(tt, current.Typ) {
						res[tt] = struct{}{}
					}
				}
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
			return r.resolveTypes(class, fn.Typ)
		}
	case meta.WInstanceMethodCall:
		expr, methodName := meta.UnwrapInstanceMethodCall(typ)

		instanceTypes := r.resolveType(class, expr)
		res = r.collectMethodCallTypes(res, instanceTypes, methodName)
		if len(res) == 0 {
			res = r.collectMethodCallTypes(res, instanceTypes, "__call")
		}

	case meta.WInstancePropertyFetch:
		expr, propertyName := meta.UnwrapInstancePropertyFetch(typ)

		for className := range r.resolveType(class, expr) {
			info, _, ok := FindProperty(className, propertyName)
			if ok {
				for tt := range r.resolveTypes(class, info.Typ) {
					res[tt] = struct{}{}
				}
			} else {
				// If there is a __get method, it might have
				// a @return annotation that will help to
				// get appropriate type for dynamic property lookup.
				get, _, ok := FindMethod(className, "__get")
				if ok {
					return r.resolveTypes(class, get.Typ)
				}
			}
		}
	case meta.WBaseMethodParam:
		return solveBaseMethodParam(class, typ, visitedMap, res)
	case meta.WStaticMethodCall:
		className, methodName := meta.UnwrapStaticMethodCall(typ)
		info, _, ok := FindMethod(className, methodName)
		if ok {
			return r.resolveTypes(className, info.Typ)
		}
		info, _, ok = FindMethod(className, "__callStatic")
		if ok {
			return r.resolveTypes(className, info.Typ)
		}

	case meta.WStaticPropertyFetch:
		className, propertyName := meta.UnwrapStaticPropertyFetch(typ)
		info, _, ok := FindProperty(className, propertyName)
		if ok {
			return r.resolveTypes(class, info.Typ)
		}
	case meta.WClassConstFetch:
		className, constName := meta.UnwrapClassConstFetch(typ)
		info, _, ok := FindConstant(className, constName)
		if ok {
			return r.resolveTypes(class, info.Typ)
		}
	default:
		panic(fmt.Sprintf("Unexpected type: %d", typ[0]))
	}

	return res
}

func solveBaseMethodParam(curStaticClass, typ string, visitedMap, res map[string]struct{}) map[string]struct{} {
	index, className, methodName := meta.UnwrapBaseMethodParam(typ)
	class, ok := meta.Info.GetClass(className)
	if ok {
		// TODO(quasilyte): walk parent interfaces as well?
		for ifaceName := range class.Interfaces {
			iface, ok := meta.Info.GetClass(ifaceName)
			if !ok {
				continue
			}
			fn, ok := iface.Methods[methodName]
			if !ok {
				continue
			}
			if len(fn.Params) > int(index) {
				return ResolveTypes(curStaticClass, fn.Params[index].Typ, visitedMap)
			}
		}
	}
	return res
}

func (r *resolver) resolveTypes(class string, m meta.TypesMap) map[string]struct{} {
	res := make(map[string]struct{}, m.Len())

	m.Iterate(func(t string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic during parsing '%s'", meta.NewTypesMap(t))
				panic(r)
			}
		}()

		for tt := range r.resolveType(class, t) {
			res[tt] = struct{}{}
		}
	})

	if len(res) == 0 {
		return mixedType()
	}

	if _, ok := res["empty_array"]; ok {
		delete(res, "empty_array")
		specialized := false
		for tt := range res {
			if strings.HasSuffix(tt, "[]") {
				specialized = true
				break
			}
		}
		if !specialized {
			res["mixed[]"] = struct{}{}
		}
	}

	return res
}

// FindMethod searches for a method in specified class
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

		// For the purposes of finding a method info, we do use interfaces
		// method sets. If looked up method is found there, we return
		// an empty string to clarify that the method has no actual implementation.
		for ifaceName := range class.Interfaces {
			res, _, ok := findMethod(ifaceName, methodName, visitedMap)
			if ok {
				return res, "", ok
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

// FindProperty searches for a property in specified class (both static and instance properties)
func FindProperty(className string, propertyName string) (res meta.PropertyInfo, implClassName string, ok bool) {
	return findProperty(className, propertyName, make(map[string]struct{}))
}

func findProperty(className string, propertyName string, visitedMap map[string]struct{}) (res meta.PropertyInfo, implClassName string, ok bool) {
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

		res, ok = class.Properties[propertyName]
		if ok {
			return res, className, ok
		}

		for trait := range class.Traits {
			res, implClassName, ok = findProperty(trait, propertyName, visitedMap)
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
