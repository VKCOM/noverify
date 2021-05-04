package linter

import (
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
)

// FlagsToString is designed for debugging flags.
func FlagsToString(f int) string {
	var res []string

	if (f & FlagReturn) == FlagReturn {
		res = append(res, "Return")
	}

	if (f & FlagDie) == FlagDie {
		res = append(res, "Die")
	}

	if (f & FlagThrow) == FlagThrow {
		res = append(res, "Throw")
	}

	if (f & FlagBreak) == FlagBreak {
		res = append(res, "Break")
	}

	return "Exit flags: [" + strings.Join(res, ", ") + "], digits: " + fmt.Sprintf("%d", f)
}

func haveMagicMethod(info *meta.Info, class, methodName string) bool {
	_, ok := solver.FindMethod(info, class, methodName)
	return ok
}

func isQuote(r rune) bool {
	return r == '"' || r == '\''
}

// walkNode is a convenience wrapper for EnterNode-only traversals.
// It gives a way to traverse a node without defining a new kind of walker.
//
// enterNode function is called in place where EnterNode method would be called.
// If n is nil, no traversal is performed.
func walkNode(n ir.Node, enterNode func(ir.Node) bool) {
	if n == nil {
		return
	}
	v := nodeVisitor{enterNode: enterNode}
	n.Walk(v)
}

type nodeVisitor struct {
	enterNode func(ir.Node) bool
}

func (v nodeVisitor) LeaveNode(n ir.Node) {}

func (v nodeVisitor) EnterNode(n ir.Node) bool {
	return v.enterNode(n)
}

func varToString(v ir.Node) string {
	switch t := v.(type) {
	case *ir.SimpleVar:
		return t.Name
	case *ir.Var:
		return "$" + varToString(t.Expr)
	case *ir.FunctionCallExpr:
		// TODO: support function calls here :)
		return ""
	case *ir.String:
		// Things like ${"x"}
		return "${" + t.Value + "}"
	default:
		return ""
	}
}

func typesMapToTypeExpr(p *phpdoc.TypeParser, m types.Map) phpdoc.Type {
	typeString := m.String()
	return p.Parse(typeString)
}

// mergeTypeMaps merges two typesmaps without losing information.
// So merging int[] and array will give int[], and Foo and object will give Foo.
func mergeTypeMaps(left types.Map, right types.Map) types.Map {
	var hasAtLeastOneArray bool
	var hasAtLeastOneClass bool

	merged := make(map[string]struct{}, left.Len()+right.Len())

	left.Iterate(func(typ string) {
		if typ == "" {
			return
		}

		if typ[0] == types.WArrayOf {
			hasAtLeastOneArray = true
		}
		if typ[0] == '\\' {
			hasAtLeastOneClass = true
		}
		merged[typ] = struct{}{}
	})

	right.Iterate(func(typ string) {
		if typ == "" {
			return
		}

		if typ[0] == types.WArrayOf && types.UnwrapArrayOf(typ) == "mixed" && hasAtLeastOneArray {
			return
		}
		if typ == "object" && hasAtLeastOneClass {
			return
		}
		merged[typ] = struct{}{}
	})

	return types.NewMapFromMap(merged)
}

// functionReturnType returns the return type of a function over computed types
// according to the convention below:
//
// The types are inferred as follows:
// 1. If there is an @return annotation, then its value becomes the return type;
//
// 2. If there is a type hint, then it is added to the types from the @return.
//    If the @return is empty, then the type matches the type hint itself;
//
// 3. If there is no @return annotation and type hint, then the return type is equal to
//    the union of the types that are returned from the function by return.
func functionReturnType(phpdocReturnType types.Map, hintReturnType types.Map, actualReturnTypes types.Map) types.Map {
	var returnTypes types.Map
	if !phpdocReturnType.Empty() || !hintReturnType.Empty() {
		returnTypes = mergeTypeMaps(phpdocReturnType, hintReturnType)
	} else {
		returnTypes = actualReturnTypes
	}

	if returnTypes.Empty() {
		returnTypes = types.VoidType
	}

	return returnTypes
}

type funcCallInfo struct {
	funcName   string
	info       meta.FuncInfo
	isFound    bool
	canAnalyze bool
}

// TODO: bundle type solving params somehow.
// We usually need ClassParseState+Scope+[]CustomType.
func resolveFunctionCall(sc *meta.Scope, st *meta.ClassParseState, customTypes []solver.CustomType, call *ir.FunctionCallExpr) funcCallInfo {
	var res funcCallInfo
	if !st.Info.IsIndexingComplete() {
		return res
	}
	res.canAnalyze = true

	fqName, ok := solver.GetFuncName(st, call.Function)
	if ok {
		res.funcName = fqName
		res.info, res.isFound = st.Info.GetFunction(fqName)
	} else {
		solver.ExprTypeCustom(sc, st, call.Function, customTypes).Iterate(func(typ string) {
			if res.isFound {
				return
			}
			m, ok := solver.FindMethod(st.Info, typ, `__invoke`)
			res.info = m.Info
			res.isFound = ok
		})
		if !res.isFound {
			res.canAnalyze = false
		}
	}

	return res
}

type methodCallInfo struct {
	methodName       string
	className        string
	info             meta.FuncInfo
	methodCallerType types.Map
	isFound          bool
	isMagic          bool
	canAnalyze       bool
}

func resolveMethodCall(sc *meta.Scope, st *meta.ClassParseState, customTypes []solver.CustomType, e *ir.MethodCallExpr) methodCallInfo {
	if !st.Info.IsIndexingComplete() {
		return methodCallInfo{canAnalyze: false}
	}

	var methodName string

	switch id := e.Method.(type) {
	case *ir.Identifier:
		methodName = id.Value
	default:
		return methodCallInfo{canAnalyze: false}
	}

	var (
		matchDist   = math.MaxInt32
		foundMethod bool
		magic       bool
		fn          meta.FuncInfo
		className   string
	)

	methodCallerType := solver.ExprTypeCustom(sc, st, e.Variable, customTypes)

	methodCallerType.Find(func(typ string) bool {
		m, isMagic, ok := findMethod(st.Info, typ, methodName)
		if !ok {
			return false
		}
		foundMethod = true
		if dist := classDistance(st, typ); dist < matchDist {
			matchDist = dist
			fn = m.Info
			className = m.ClassName
			magic = isMagic
		}
		return matchDist == 0 // Stop if found inside the current class
	})

	return methodCallInfo{
		methodName:       methodName,
		className:        className,
		isFound:          foundMethod,
		isMagic:          magic,
		info:             fn,
		methodCallerType: methodCallerType,
		canAnalyze:       true,
	}
}

type staticMethodCallInfo struct {
	methodName               string
	className                string
	methodInfo               solver.FindMethodResult
	isParentCall             bool
	isMagic                  bool
	isFound                  bool
	isCallsParentConstructor bool
	canAnalyze               bool
}

func resolveStaticMethodCall(st *meta.ClassParseState, e *ir.StaticCallExpr) staticMethodCallInfo {
	if !st.Info.IsIndexingComplete() {
		return staticMethodCallInfo{canAnalyze: false}
	}

	var methodName string

	switch id := e.Call.(type) {
	case *ir.Identifier:
		methodName = id.Value
	default:
		return staticMethodCallInfo{canAnalyze: false}
	}

	classNameNode, ok := e.Class.(*ir.Name)
	parentCall := ok && classNameNode.Value == "parent"
	var callsParentConstructor bool
	if parentCall && methodName == "__construct" {
		callsParentConstructor = true
	}

	className, ok := solver.GetClassName(st, e.Class)
	if !ok {
		return staticMethodCallInfo{canAnalyze: false}
	}

	m, found := solver.FindMethod(st.Info, className, methodName)
	isMagic := haveMagicMethod(st.Info, className, `__callStatic`)

	return staticMethodCallInfo{
		methodName:               methodName,
		className:                className,
		methodInfo:               m,
		isMagic:                  isMagic,
		isParentCall:             parentCall,
		isFound:                  found,
		isCallsParentConstructor: callsParentConstructor,
		canAnalyze:               true,
	}
}

type propertyFetchInfo struct {
	className         string
	info              meta.PropertyInfo
	propertyFetchType types.Map
	propertyNode      *ir.Identifier
	isFound           bool
	isMagic           bool
	canAnalyze        bool
}

func resolvePropertyFetch(sc *meta.Scope, st *meta.ClassParseState, customTypes []solver.CustomType, e *ir.PropertyFetchExpr) propertyFetchInfo {
	propertyNode, ok := e.Property.(*ir.Identifier)
	if !ok {
		return propertyFetchInfo{canAnalyze: false}
	}

	var found bool
	var magic bool
	var matchDist = math.MaxInt32
	var className string
	var info meta.PropertyInfo

	propertyFetchType := solver.ExprTypeCustom(sc, st, e.Variable, customTypes)
	propertyFetchType.Find(func(typ string) bool {
		p, isMagic, ok := findProperty(st.Info, typ, propertyNode.Value)
		if !ok {
			return false
		}
		found = true
		if dist := classDistance(st, typ); dist < matchDist {
			matchDist = dist
			info = p.Info
			className = p.ClassName
			magic = isMagic
		}
		return matchDist == 0 // Stop if found inside the current class
	})

	return propertyFetchInfo{
		className:         className,
		isFound:           found,
		isMagic:           magic,
		info:              info,
		propertyFetchType: propertyFetchType,
		propertyNode:      propertyNode,
		canAnalyze:        true,
	}
}

type propertyStaticFetchInfo struct {
	className       string
	propertyName    string
	info            solver.FindPropertyResult
	isFound         bool
	needHandleAsVar bool
	canAnalyze      bool
}

func resolveStaticPropertyFetch(st *meta.ClassParseState, e *ir.StaticPropertyFetchExpr) propertyStaticFetchInfo {
	if !st.Info.IsIndexingComplete() {
		return propertyStaticFetchInfo{canAnalyze: false}
	}

	propertyNode, ok := e.Property.(*ir.SimpleVar)
	if !ok {
		return propertyStaticFetchInfo{needHandleAsVar: true, canAnalyze: false}
	}

	className, ok := solver.GetClassName(st, e.Class)
	if !ok {
		return propertyStaticFetchInfo{canAnalyze: false}
	}

	property, found := solver.FindProperty(st.Info, className, "$"+propertyNode.Name)

	return propertyStaticFetchInfo{
		className:    className,
		propertyName: propertyNode.Name,
		info:         property,
		isFound:      found,
		canAnalyze:   true,
	}
}

type classPropertyFetchInfo struct {
	constName     string
	className     string
	implClassName string
	info          meta.ConstInfo
	isFound       bool
	canAnalyze    bool
}

func resolveClassConstFetch(st *meta.ClassParseState, e *ir.ClassConstFetchExpr) classPropertyFetchInfo {
	if !st.Info.IsIndexingComplete() {
		return classPropertyFetchInfo{canAnalyze: false}
	}

	constName := e.ConstantName
	if constName.Value == `class` || constName.Value == `CLASS` {
		return classPropertyFetchInfo{canAnalyze: false}
	}

	className, ok := solver.GetClassName(st, e.Class)
	if !ok {
		return classPropertyFetchInfo{canAnalyze: false}
	}

	info, implClass, found := solver.FindConstant(st.Info, className, constName.Value)

	return classPropertyFetchInfo{
		constName:     constName.Value,
		className:     className,
		implClassName: implClass,
		info:          info,
		isFound:       found,
		canAnalyze:    true,
	}
}

// isCapitalized reports whether s starts with an upper case letter.
func isCapitalized(s string) bool {
	ch, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(ch)
}

// findVarNode returns expression variable node root.
// If expression doesn't start from a variable, returns nil.
func findVarNode(n ir.Node) ir.Node {
	switch n := n.(type) {
	case *ir.Var, *ir.SimpleVar:
		return n
	case *ir.PropertyFetchExpr:
		return findVarNode(n.Variable)
	case *ir.ArrayDimFetchExpr:
		return findVarNode(n.Variable)
	default:
		return nil
	}
}

func classHasProp(st *meta.ClassParseState, className, propName string) bool {
	var nameWithDollar string
	var nameWithoutDollar string
	if strings.HasPrefix(propName, "$") {
		nameWithDollar = propName
		nameWithoutDollar = strings.TrimPrefix(propName, "$")
	} else {
		nameWithDollar = "$" + propName
		nameWithoutDollar = propName
	}

	// Static props stored with leading "$".
	if _, ok := solver.FindProperty(st.Info, className, nameWithDollar); ok {
		return true
	}
	_, ok := solver.FindProperty(st.Info, className, nameWithoutDollar)
	return ok
}

func getConstValue(c meta.ConstValue) string {
	if c.Type == meta.Undefined {
		return ""
	}

	return fmt.Sprintf("%v", c.Value)
}

func binaryOpString(n ir.Node) string {
	switch n.(type) {
	case *ir.BitwiseAndExpr:
		return "&"
	case *ir.BitwiseOrExpr:
		return "|"
	case *ir.BitwiseXorExpr:
		return "^"
	case *ir.LogicalAndExpr:
		return "and"
	case *ir.BooleanAndExpr:
		return "&&"
	case *ir.LogicalOrExpr:
		return "or"
	case *ir.BooleanOrExpr:
		return "||"
	case *ir.LogicalXorExpr:
		return "xor"
	case *ir.PlusExpr:
		return "+"
	case *ir.MinusExpr:
		return "-"
	case *ir.MulExpr:
		return "*"
	case *ir.DivExpr:
		return "/"
	case *ir.ModExpr:
		return "%"
	case *ir.PowExpr:
		return "**"
	case *ir.EqualExpr:
		return "=="
	case *ir.NotEqualExpr:
		return "!="
	case *ir.IdenticalExpr:
		return "==="
	case *ir.NotIdenticalExpr:
		return "!=="
	case *ir.SmallerExpr:
		return "<"
	case *ir.SmallerOrEqualExpr:
		return "<="
	case *ir.GreaterExpr:
		return ">"
	case *ir.GreaterOrEqualExpr:
		return ">="
	case *ir.SpaceshipExpr:
		return "<=>"

	default:
		return ""
	}
}

func cloneRulesForFile(filename string, ruleSet *rules.ScopedSet) *rules.ScopedSet {
	if ruleSet.CountRules == 0 {
		return nil
	}

	var clone rules.ScopedSet
	for kind, ruleByKind := range &ruleSet.RulesByKind {
		res := make([]rules.Rule, 0, len(ruleByKind))
		for _, rule := range ruleByKind {
			if !strings.Contains(filename, rule.Path) {
				continue
			}
			res = append(res, rule)
		}
		clone.Set(ir.NodeKind(kind), res)
	}
	return &clone
}

// List taken from https://wiki.php.net/rfc/context_sensitive_lexer
var phpKeywords = map[string]bool{
	"callable":     true,
	"class":        true,
	"trait":        true,
	"extends":      true,
	"implements":   true,
	"static":       true,
	"abstract":     true,
	"final":        true,
	"public":       true,
	"protected":    true,
	"private":      true,
	"const":        true,
	"enddeclare":   true,
	"endfor":       true,
	"endforeach":   true,
	"endif":        true,
	"endwhile":     true,
	"and":          true,
	"global":       true,
	"goto":         true,
	"instanceof":   true,
	"insteadof":    true,
	"interface":    true,
	"namespace":    true,
	"new":          true,
	"or":           true,
	"xor":          true,
	"try":          true,
	"use":          true,
	"var":          true,
	"exit":         true,
	"list":         true,
	"clone":        true,
	"include":      true,
	"include_once": true,
	"throw":        true,
	"array":        true,
	"print":        true,
	"echo":         true,
	"require":      true,
	"require_once": true,
	"return":       true,
	"else":         true,
	"elseif":       true,
	"default":      true,
	"break":        true,
	"continue":     true,
	"switch":       true,
	"yield":        true,
	"function":     true,
	"if":           true,
	"endswitch":    true,
	"finally":      true,
	"for":          true,
	"foreach":      true,
	"declare":      true,
	"case":         true,
	"do":           true,
	"while":        true,
	"as":           true,
	"catch":        true,
	"die":          true,
	"self":         true,
	"parent":       true,
}
