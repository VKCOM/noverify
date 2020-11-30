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
func resolveType(curStaticClass, typ string, visitedMap ResolverMap) (result map[string]struct{}) {
	r := resolver{visited: visitedMap}
	return r.resolveType(curStaticClass, typ)
}

// ResolveTypes resolves function calls, method calls and global variables.
//   curStaticClass is current class name (if inside the class, otherwise "")
func ResolveTypes(curStaticClass string, m meta.TypesMap, visitedMap ResolverMap) map[string]struct{} {
	r := resolver{visited: visitedMap}
	return r.resolveTypes(curStaticClass, m)
}

type ResolverMap map[string]map[string]struct{}

type resolver struct {
	visited ResolverMap
}

func (r *resolver) collectMethodCallTypes(out, possibleTypes map[string]struct{}, methodName string) map[string]struct{} {
	for className := range possibleTypes {
		m, ok := FindMethod(className, methodName)
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

	if len(typ) == 0 || typ[0] >= meta.WMax {
		return identityType(typ)
	}

	res := make(map[string]struct{})
	visitedMap[typ] = nil // Nil guards against unbound recursion

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
	case meta.WElemOfKey:
		arrayType, key := meta.UnwrapElemOfKey(typ)
		for tt := range r.resolveType(class, arrayType) {
			if meta.IsShapeType(tt) {
				res = r.solveElemOfShape(class, tt, key, res)
			} else {
				res = r.solveElemOf(tt, res)
			}
		}
	case meta.WElemOf:
		for tt := range r.resolveType(class, meta.UnwrapElemOf(typ)) {
			res = r.solveElemOf(tt, res)
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
			p, ok := FindProperty(className, propertyName)
			if ok {
				for tt := range r.resolveTypes(class, p.Info.Typ) {
					res[tt] = struct{}{}
				}
			} else {
				// If there is a __get method, it might have
				// a @return annotation that will help to
				// get appropriate type for dynamic property lookup.
				m, ok := FindMethod(className, "__get")
				if ok {
					return r.resolveTypes(class, m.Info.Typ)
				}
			}
		}
	case meta.WBaseMethodParam:
		return solveBaseMethodParam(class, typ, visitedMap, res)
	case meta.WStaticMethodCall:
		className, methodName := meta.UnwrapStaticMethodCall(typ)
		m, ok := FindMethod(className, methodName)
		if ok {
			res = r.resolveTypes(className, m.Info.Typ)
			if m.TraitName != "" {
				res = replaceTraitName(res, m.TraitName, m.ClassName)
			}
			return res
		}
		m, ok = FindMethod(className, "__callStatic")
		if ok {
			// Should probably run replaceTraitName here on the result
			// as well, but I don't have a good __callStatic trait method
			// example, so I hesitate.
			return r.resolveTypes(className, m.Info.Typ)
		}

	case meta.WStaticPropertyFetch:
		className, propertyName := meta.UnwrapStaticPropertyFetch(typ)
		p, ok := FindProperty(className, propertyName)
		if ok {
			res = r.resolveTypes(className, p.Info.Typ)
			if p.TraitName != "" {
				res = replaceTraitName(res, p.TraitName, p.ClassName)
			}
			return res
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

func solveBaseMethodParam(curStaticClass, typ string, visitedMap ResolverMap, res map[string]struct{}) map[string]struct{} {
	index, className, methodName := meta.UnwrapBaseMethodParam(typ)
	class, ok := meta.Info.GetClass(className)
	if ok {
		// TODO(quasilyte): walk parent interfaces as well?
		for ifaceName := range class.Interfaces {
			iface, ok := meta.Info.GetClass(ifaceName)
			if !ok {
				continue
			}
			fn, ok := iface.Methods.Get(methodName)
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

func (r *resolver) solveElemOfShape(class, shapeName, key string, res map[string]struct{}) map[string]struct{} {
	shape, ok := meta.Info.GetClass(shapeName)
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
	case meta.IsArrayType(tt):
		res[strings.TrimSuffix(tt, "[]")] = struct{}{}
	case tt == "mixed":
		res["mixed"] = struct{}{}
	case Implements(tt, `\ArrayAccess`):
		m, ok := FindMethod(tt, "offsetGet")
		if ok {
			for tt := range r.resolveTypes(tt, m.Info.Typ) {
				res[tt] = struct{}{}
			}
		}
	case Implements(tt, `\Traversable`):
		m, ok := FindMethod(tt, "current")
		if ok {
			for tt := range r.resolveTypes(tt, m.Info.Typ) {
				res[tt] = struct{}{}
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
			if meta.IsArrayType(tt) {
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
func FindMethod(className string, methodName string) (FindMethodResult, bool) {
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

	return findMethod(className, methodName, make(map[string]struct{}))
}

func peekImplemented(a, b FindMethodResult) FindMethodResult {
	if a.Implemented {
		return a
	}
	return b
}

func findMethod(className string, methodName string, visitedMap map[string]struct{}) (FindMethodResult, bool) {
	var result FindMethodResult
	found := false

	for {
		if _, ok := visitedMap[className]; ok {
			break
		}
		visitedMap[className] = struct{}{}

		class, ok := getClassOrTrait(className)
		if !ok {
			break
		}

		info, ok := class.Methods.Get(methodName)
		if ok {
			found = true
			result = peekImplemented(result, FindMethodResult{
				Info:        info,
				ClassName:   className,
				Implemented: !info.IsAbstract(),
			})
			if result.Implemented {
				return result, true
			}
		}

		for trait := range class.Traits {
			m, ok := findMethod(trait, methodName, visitedMap)
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
			m, ok := findMethod(parentIfaceName, methodName, visitedMap)
			if ok {
				m.Implemented = false
				return m, true
			}
		}

		for _, mixin := range class.Mixins {
			_, ok := getClassOrTrait(mixin)
			if !ok {
				continue
			}

			result, ok := findMethod(mixin, methodName, visitedMap)
			if ok {
				return result, true
			}
		}

		for ifaceName := range class.Interfaces {
			m, ok := findMethod(ifaceName, methodName, visitedMap)
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
func FindProperty(className string, propertyName string) (FindPropertyResult, bool) {
	return findProperty(className, propertyName, make(map[string]struct{}))
}

func findProperty(className string, propertyName string, visitedMap map[string]struct{}) (FindPropertyResult, bool) {
	var result FindPropertyResult
	for {
		if _, ok := visitedMap[className]; ok {
			return result, false
		}
		visitedMap[className] = struct{}{}

		class, ok := getClassOrTrait(className)
		if !ok || class.IsShape() {
			return result, false
		}

		info, ok := class.Properties[propertyName]
		if ok {
			result.Info = info
			result.ClassName = className
			return result, true
		}

		for trait := range class.Traits {
			p, ok := findProperty(trait, propertyName, visitedMap)
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
func Implements(className string, interfaceName string) bool {
	visited := make(map[string]struct{}, 8)
	return implements(className, interfaceName, visited)
}

func implements(className string, interfaceName string, visited map[string]struct{}) bool {
	if className == interfaceName {
		return true
	}

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

		for _, iface := range class.ParentInterfaces {
			if implements(iface, interfaceName, visited) {
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
func FindConstant(className string, constName string) (res meta.ConstInfo, implClassName string, ok bool) {
	visitedClasses := make(map[string]struct{}, 8) // expecting to be not so many inheritance levels
	return findConstant(className, constName, visitedClasses)
}

func findConstant(className string, constName string, visitedClasses map[string]struct{}) (res meta.ConstInfo, implClassName string, ok bool) {
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

func getClassOrTrait(typeName string) (meta.ClassInfo, bool) {
	class, ok := meta.Info.GetClass(typeName)
	if ok {
		return class, true
	}
	trait, ok := meta.Info.GetTrait(typeName)
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
