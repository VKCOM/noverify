package solver

import (
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/types"
)

var sharedMixedType = map[string]struct{}{"mixed": {}}

func mixedType() map[string]struct{} {
	if len(sharedMixedType) != 1 {
		// At least until we're 100% sure this is safe.
		panic(fmt.Sprintf("mixed type map was modified: %v", sharedMixedType))
	}
	return sharedMixedType
}

// resolveType resolves function calls, method calls and global variables.
//   curStaticClass is current class name (if inside the class, otherwise "")
func resolveType(info *meta.Info, curStaticClass, typ string, visitedMap ResolverMap) (result map[string]struct{}) {
	r := resolver{info: info, visited: visitedMap}
	return r.resolveType(curStaticClass, typ)
}

// ResolveTypes resolves function calls, method calls and global variables.
//   curStaticClass is current class name (if inside the class, otherwise "")
func ResolveTypes(info *meta.Info, curStaticClass string, m types.Map, visitedMap ResolverMap) map[string]struct{} {
	r := resolver{info: info, visited: visitedMap}
	return r.resolveTypes(curStaticClass, m)
}

// TODO: can we make it unexported?
type ResolverMap map[string]map[string]struct{}

type resolver struct {
	info    *meta.Info
	visited ResolverMap
}

func (r *resolver) collectMethodCallTypes(out, possibleTypes map[string]struct{}, methodName string) map[string]struct{} {
	for className := range possibleTypes {
		m, ok := FindMethod(r.info, className, methodName)
		if ok {
			for tt := range r.resolveTypes(className, m.Info.Typ) {
				out[tt] = struct{}{}
			}
		}
	}
	return out
}

func (r *resolver) resolveType(class, typ string) map[string]struct{} {
	res := r.resolveTypeNoLateStaticBinding(class, typ)
	if typ == "static" {
		r.visited[typ+class] = res
	} else {
		r.visited[typ] = res
	}

	if _, ok := res["static"]; ok {
		delete(res, "static")
		res[class] = struct{}{}
	}

	return res
}

func (r *resolver) resolveTypeNoLateStaticBinding(class, typ string) map[string]struct{} {
	visitedMap := r.visited

	if result, ok := visitedMap[typ]; ok {
		return result
	}

	if typ == "" || typ[0] >= types.WMax {
		return identityType(typ)
	}

	res := make(map[string]struct{})
	visitedMap[typ] = nil // Nil guards against unbound recursion

	switch typ[0] {
	case types.WGlobal:
		varTyp, ok := r.info.GetVarNameType(types.UnwrapGlobal(typ))
		if ok {
			for tt := range r.resolveTypes(class, varTyp) {
				res[tt] = struct{}{}
			}
		}
	case types.WConstant:
		ci, ok := r.info.GetConstant(types.UnwrapConstant(typ))
		if ok {
			for tt := range r.resolveTypes(class, ci.Typ) {
				res[tt] = struct{}{}
			}
		}
	case types.WArrayOf:
		for tt := range r.resolveType(class, types.UnwrapArrayOf(typ)) {
			res[tt+"[]"] = struct{}{}
		}
	case types.WElemOfKey:
		arrayType, key := types.UnwrapElemOfKey(typ)
		for tt := range r.resolveType(class, arrayType) {
			if types.IsShape(tt) {
				res = r.solveElemOfShape(class, tt, key, res)
			} else {
				res = r.solveElemOf(tt, res)
			}
		}
	case types.WElemOf:
		for tt := range r.resolveType(class, types.UnwrapElemOf(typ)) {
			res = r.solveElemOf(tt, res)
		}
	case types.WFunctionCall:
		nm := types.UnwrapFunctionCall(typ)
		fn, ok := r.info.GetFunction(nm)
		// functions can fall back to root namespace
		if !ok && strings.Count(nm, `\`) > 1 {
			fn, ok = r.info.GetFunction(nm[strings.LastIndex(nm, `\`):])
		}

		if ok {
			return r.resolveTypes(class, fn.Typ)
		}
	case types.WInstanceMethodCall:
		expr, methodName := types.UnwrapInstanceMethodCall(typ)

		instanceTypes := r.resolveType(class, expr)
		res = r.collectMethodCallTypes(res, instanceTypes, methodName)
		if len(res) == 0 {
			res = r.collectMethodCallTypes(res, instanceTypes, "__call")
		}

	case types.WInstancePropertyFetch:
		expr, propertyName := types.UnwrapInstancePropertyFetch(typ)

		for className := range r.resolveType(class, expr) {
			p, ok := FindProperty(r.info, className, propertyName)
			if ok {
				for tt := range r.resolveTypes(class, p.Info.Typ) {
					res[tt] = struct{}{}
				}
			} else {
				// If there is a __get method, it might have
				// a @return annotation that will help to
				// get appropriate type for dynamic property lookup.
				m, ok := FindMethod(r.info, className, "__get")
				if ok {
					return r.resolveTypes(class, m.Info.Typ)
				}
			}
		}
	case types.WBaseMethodParam:
		return r.solveBaseMethodParam(class, typ, visitedMap, res)
	case types.WStaticMethodCall:
		className, methodName := types.UnwrapStaticMethodCall(typ)
		m, ok := FindMethod(r.info, className, methodName)
		if ok {
			res = r.resolveTypes(className, m.Info.Typ)
			if m.TraitName != "" {
				res = replaceTraitName(res, m.TraitName, m.ClassName)
			}
			return res
		}
		m, ok = FindMethod(r.info, className, "__callStatic")
		if ok {
			// Should probably run replaceTraitName here on the result
			// as well, but I don't have a good __callStatic trait method
			// example, so I hesitate.
			return r.resolveTypes(className, m.Info.Typ)
		}

	case types.WStaticPropertyFetch:
		className, propertyName := types.UnwrapStaticPropertyFetch(typ)
		p, ok := FindProperty(r.info, className, propertyName)
		if ok {
			res = r.resolveTypes(className, p.Info.Typ)
			if p.TraitName != "" {
				res = replaceTraitName(res, p.TraitName, p.ClassName)
			}
			return res
		}
	case types.WClassConstFetch:
		className, constName := types.UnwrapClassConstFetch(typ)
		info, _, ok := FindConstant(r.info, className, constName)
		if ok {
			return r.resolveTypes(class, info.Typ)
		}
	default:
		panic(fmt.Sprintf("Unexpected type: %d", typ[0]))
	}

	return res
}

func (r *resolver) solveBaseMethodParam(curStaticClass, typ string, visitedMap ResolverMap, res map[string]struct{}) map[string]struct{} {
	index, className, methodName := types.UnwrapBaseMethodParam(typ)
	class, ok := r.info.GetClass(className)
	if ok {
		// TODO(quasilyte): walk parent interfaces as well?
		for ifaceName := range class.Interfaces {
			iface, ok := r.info.GetClass(ifaceName)
			if !ok {
				continue
			}
			fn, ok := iface.Methods.Get(methodName)
			if !ok {
				continue
			}
			if len(fn.Params) > int(index) {
				return ResolveTypes(r.info, curStaticClass, fn.Params[index].Typ, visitedMap)
			}
		}
	}
	return res
}

func (r *resolver) solveElemOfShape(class, shapeName, key string, res map[string]struct{}) map[string]struct{} {
	shape, ok := r.info.GetClass(shapeName)
	if !ok {
		return res
	}
	p, ok := shape.Properties[key]
	if ok {
		for tt := range r.resolveTypes(class, p.Typ) {
			res[tt] = struct{}{}
		}
	}
	return res
}

func (r *resolver) solveElemOf(tt string, res map[string]struct{}) map[string]struct{} {
	switch {
	case types.IsArray(tt):
		res[types.ArrayType(tt)] = struct{}{}
	case tt == "mixed":
		res["mixed"] = struct{}{}
	case Implements(r.info, tt, `\ArrayAccess`):
		m, ok := FindMethod(r.info, tt, "offsetGet")
		if ok {
			for tt := range r.resolveTypes(tt, m.Info.Typ) {
				res[tt] = struct{}{}
			}
		}
	case Implements(r.info, tt, `\Traversable`):
		m, ok := FindMethod(r.info, tt, "current")
		if ok {
			for tt := range r.resolveTypes(tt, m.Info.Typ) {
				res[tt] = struct{}{}
			}
		}
	}
	return res
}

func (r *resolver) resolveTypes(class string, m types.Map) map[string]struct{} {
	res := make(map[string]struct{}, m.Len())

	m.Iterate(func(t string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic during parsing '%s'", types.NewMap(t))
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
			if types.IsArray(tt) {
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

type FindMethodResult struct {
	Info        meta.FuncInfo
	ClassName   string
	TraitName   string
	Implemented bool
}

func (m FindMethodResult) ImplName() string {
	if m.TraitName != "" {
		return m.TraitName
	}
	return m.ClassName
}

// FindMethod searches for a method in specified class
func FindMethod(info *meta.Info, className, methodName string) (FindMethodResult, bool) {
	// We do 2 lookup attempts.
	//
	// The first round ignores interfaces inside hierarchy tree.
	// The second round processes leftovers, interfaces.
	//
	// This way, we will return concrete implementation even if
	// it's deeper inside a type tree.
	//
	// Suppose we do FindMethod("C", "a") for this type tree:
	//	interface A { function a(); }
	//	class Base1 { function a(); }
	//	class Base2 extends Base1 {}
	//	class C extends Base2 implements A {}
	//
	// If we would process interfaces right away, a() would be returned
	// from the A interface, but we want to get Base1.

	return findMethod(info, className, methodName, make(map[string]struct{}))
}

func peekImplemented(a, b FindMethodResult) FindMethodResult {
	if a.Implemented {
		return a
	}
	return b
}

func findMethod(info *meta.Info, className, methodName string, visitedMap map[string]struct{}) (FindMethodResult, bool) {
	var result FindMethodResult
	found := false

	for {
		if _, ok := visitedMap[className]; ok {
			break
		}
		visitedMap[className] = struct{}{}

		class, ok := getClassOrTrait(info, className)
		if !ok {
			break
		}

		methodInfo, ok := class.Methods.Get(methodName)
		if ok {
			found = true
			result = peekImplemented(result, FindMethodResult{
				Info:        methodInfo,
				ClassName:   className,
				Implemented: !methodInfo.IsAbstract(),
			})
			if result.Implemented {
				return result, true
			}
		}

		for trait := range class.Traits {
			m, ok := findMethod(info, trait, methodName, visitedMap)
			if ok {
				found = true
				result = peekImplemented(result, FindMethodResult{
					Info:        m.Info,
					ClassName:   className,
					TraitName:   trait,
					Implemented: !m.Info.IsAbstract(),
				})
				if result.Implemented {
					return result, true
				}
			}
		}

		// interfaces support multiple inheritance and I use a separate property for that for now.
		// This loop is executed *only* when we're searching a method with interface
		// as a root, so we don't need to check whether a method is implemented.
		for _, parentIfaceName := range class.ParentInterfaces {
			m, ok := findMethod(info, parentIfaceName, methodName, visitedMap)
			if ok {
				m.Implemented = false
				return m, true
			}
		}

		for _, mixin := range class.Mixins {
			_, ok := getClassOrTrait(info, mixin)
			if !ok {
				continue
			}

			result, ok := findMethod(info, mixin, methodName, visitedMap)
			if ok {
				return result, true
			}
		}

		for ifaceName := range class.Interfaces {
			m, ok := findMethod(info, ifaceName, methodName, visitedMap)
			if ok {
				found = true
				m.Implemented = false
				result = peekImplemented(result, m)
				break // No point in searching other interfaces
			}
		}

		if class.Parent == "" {
			break
		}

		className = class.Parent
	}

	return result, found
}

type FindPropertyResult struct {
	Info      meta.PropertyInfo
	ClassName string
	TraitName string
}

func (p FindPropertyResult) ImplName() string {
	if p.TraitName != "" {
		return p.TraitName
	}
	return p.ClassName
}

// FindProperty searches for a property in specified class (both static and instance properties)
func FindProperty(info *meta.Info, className, propertyName string) (FindPropertyResult, bool) {
	return findProperty(info, className, propertyName, make(map[string]struct{}))
}

func findProperty(info *meta.Info, className, propertyName string, visitedMap map[string]struct{}) (FindPropertyResult, bool) {
	var result FindPropertyResult
	for {
		if _, ok := visitedMap[className]; ok {
			return result, false
		}
		visitedMap[className] = struct{}{}

		class, ok := getClassOrTrait(info, className)
		if !ok || class.IsShape() {
			return result, false
		}

		propInfo, ok := class.Properties[propertyName]
		if ok {
			result.Info = propInfo
			result.ClassName = className
			return result, true
		}

		for trait := range class.Traits {
			p, ok := findProperty(info, trait, propertyName, visitedMap)
			if ok {
				result.Info = p.Info
				result.ClassName = className
				result.TraitName = trait
				return result, true
			}
		}

		if class.Parent == "" {
			return result, false
		}

		className = class.Parent
	}
}

// Implements checks if className implements interfaceName
//
// Does not perform the actual method set comparison.
func Implements(info *meta.Info, className, interfaceName string) bool {
	visited := make(map[string]struct{}, 8)
	return implements(info, className, interfaceName, visited)
}

func implements(info *meta.Info, className, interfaceName string, visited map[string]struct{}) bool {
	if className == interfaceName {
		return true
	}

	for {
		class, ok := info.GetClass(className)
		if !ok {
			return false
		}

		_, ok = class.Interfaces[interfaceName]
		if ok {
			return true
		}

		for iface := range class.Interfaces {
			if interfaceExtends(info, iface, interfaceName, visited) {
				return true
			}
		}

		for _, iface := range class.ParentInterfaces {
			if implements(info, iface, interfaceName, visited) {
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
func interfaceExtends(info *meta.Info, orig, parent string, visited map[string]struct{}) bool {
	if _, ok := visited[orig]; ok {
		return false
	}

	visited[orig] = struct{}{}

	class, ok := info.GetClass(orig)
	if !ok {
		return false
	}

	for _, iface := range class.ParentInterfaces {
		if iface == parent {
			return true
		}

		if interfaceExtends(info, iface, parent, visited) {
			return true
		}
	}

	return false
}

// FindConstant searches for a costant in specified class and returns actual class that contains the constant.
func FindConstant(info *meta.Info, className string, constName string) (res meta.ConstInfo, implClassName string, ok bool) {
	visitedClasses := make(map[string]struct{}, 8) // expecting to be not so many inheritance levels
	return findConstant(info, className, constName, visitedClasses)
}

func findConstant(info *meta.Info, className, constName string, visitedClasses map[string]struct{}) (res meta.ConstInfo, implClassName string, ok bool) {
	for {
		// check for inheritance loops
		if _, ok := visitedClasses[className]; ok {
			return res, "", false
		}

		visitedClasses[className] = struct{}{}

		class, ok := info.GetClass(className)
		if !ok {
			return res, "", false
		}

		// inferfaces can have constants...
		for ifaceName := range class.Interfaces {
			res, implClassName, ok = findConstant(info, ifaceName, constName, visitedClasses)
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
			res, implClassName, ok = findConstant(info, parentIfaceName, constName, visitedClasses)
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

func getClassOrTrait(info *meta.Info, typeName string) (meta.ClassInfo, bool) {
	class, ok := info.GetClass(typeName)
	if ok {
		return class, true
	}
	trait, ok := info.GetTrait(typeName)
	if ok {
		return trait, true
	}
	return class, false
}

// replaceTraitName replaces traitName with className inside res.
func replaceTraitName(res map[string]struct{}, traitName, className string) map[string]struct{} {
	_, ok := res[traitName]
	if !ok {
		return res
	}
	delete(res, traitName)
	res[className] = struct{}{}
	return res
}
