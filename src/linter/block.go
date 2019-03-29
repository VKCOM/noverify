package linter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/z7zmey/php-parser/comment"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

const (
	// FlagReturn shows whether or not block has "return"
	FlagReturn = 1 << iota
	FlagBreak
	FlagContinue
	FlagThrow
	FlagDie
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

	return "Exit flags: " + strings.Join(res, ", ") + ", digits: " + fmt.Sprintf("%d", f)
}

// BlockWalker is used to process function/method contents.
//
// Current list of annotated checks:
//	- accessLevel
//	- argCount
//	- arrayAccess
//	- arrayKeys
//	- arraySyntax
//	- bareTry
//	- caseBreak
//	- deadCode
//	- phpdoc
//	- undefined
//	- unused
type BlockWalker struct {
	sc *meta.Scope
	r  *RootWalker

	custom []BlockChecker

	ignoreFunctionBodies bool
	rootLevel            bool // analysing root-level code

	// state
	statements  map[node.Node]struct{}
	customTypes []solver.CustomType

	// shared state between all blocks
	unusedVars   map[string][]node.Node
	nonLocalVars map[string]struct{} // static, global and other vars that have complex control flow

	// inferred return types if any
	returnTypes *meta.TypesMap

	isLoopBody bool

	// block flags
	exitFlags         int // if block always breaks code flow then there will be exitFlags
	containsExitFlags int // if block sometimes breaks code flow then there will be containsExitFlags

	// analyzer state
	deadCodeReported bool
}

// Scope returns block-level variable scope if it exists.
func (b *BlockWalker) Scope() *meta.Scope {
	return b.sc
}

// PrematureExitFlags returns information about whether or not all code branches have exit/return/throw/etc.
// You need to check what exactly you expect from a block to have (or not to have) by checking Flag* bits.
func (b *BlockWalker) PrematureExitFlags() int {
	return b.exitFlags
}

// RootState returns state that was stored in root context (if any) for use in custom hooks.
func (b *BlockWalker) RootState() map[string]interface{} {
	return b.r.customState
}

// IsRootLevel returns whether or not we currently analyze root level code.
func (b *BlockWalker) IsRootLevel() bool {
	return b.rootLevel
}

// Report registers a single report message about some found problem.
func (b *BlockWalker) Report(n node.Node, level int, checkName, msg string, args ...interface{}) {
	b.r.Report(n, level, checkName, msg, args...)
}

// ClassParseState returns class parse state (namespace, current class, etc)
func (b *BlockWalker) ClassParseState() *meta.ClassParseState {
	return b.r.st
}

// IsStatement checks whether or not the specified node is a top-level or a block-level statement.
func (b *BlockWalker) IsStatement(n node.Node) bool {
	_, ok := b.statements[n]
	return ok
}

func (b *BlockWalker) addStatement(n node.Node) {
	if b.statements == nil {
		b.statements = make(map[node.Node]struct{})
	}
	b.statements[n] = struct{}{}

	// A hack for assignment-used-as-expression checks to work
	expr, ok := n.(*stmt.Expression)
	if !ok {
		return
	}

	assignment, ok := expr.Expr.(*assign.Assign)
	if !ok {
		return
	}

	b.statements[assignment] = struct{}{}
}

func (b *BlockWalker) copy() *BlockWalker {
	bCopy := &BlockWalker{
		sc:                   b.sc.Clone(),
		r:                    b.r,
		isLoopBody:           b.isLoopBody,
		unusedVars:           b.unusedVars,
		nonLocalVars:         b.nonLocalVars,
		ignoreFunctionBodies: b.ignoreFunctionBodies,
	}
	for _, createFn := range b.r.customBlock {
		bCopy.custom = append(bCopy.custom, createFn(&BlockContext{w: bCopy}))
	}

	for _, c := range b.customTypes {
		bCopy.customTypes = append(bCopy.customTypes, c)
	}

	return bCopy
}

func (b *BlockWalker) reportDeadCode(n node.Node) {
	if b.deadCodeReported {
		return
	}

	switch n.(type) {
	case *stmt.Break, *stmt.Return, *expr.Die, *expr.Exit, *stmt.Throw:
		// Allow to break code flow more than once.
		// This is useful in situation like this:
		//
		//    callSomeFuncThatExits(); exit;
		//
		// You can explicitly mark that function exits unconditionally for code clarity.
		return
	case *stmt.Function, *stmt.Class, *stmt.ConstList, *stmt.Interface, *stmt.Trait:
		// when we analyze root scope, function definions and other things are parsed even after exit, throw, etc
		if b.ignoreFunctionBodies {
			return
		}
	}

	b.deadCodeReported = true
	b.r.Report(n, LevelInformation, "deadCode", "Unreachable code")
}

func varToString(v *expr.Variable) string {
	switch t := v.VarName.(type) {
	case *node.Identifier:
		return t.Value
	case *expr.Variable:
		return "$" + varToString(t)
	case *expr.FunctionCall:
		// TODO: support function calls here :)
		return "WTF_FUNCTION_CALL"
	case *scalar.String:
		// Things like ${"x"}
		return "${" + t.Value + "}"
	default:
		panic(fmt.Errorf("Unexpected variable VarName type: %T", t))
	}
}

// EnterNode is called before walking to inner nodes.
func (b *BlockWalker) EnterNode(w walker.Walkable) (res bool) {
	res = true

	for _, c := range b.custom {
		c.BeforeEnterNode(w)
	}

	n := w.(node.Node)

	if b.exitFlags != 0 {
		b.reportDeadCode(n)
	}

	for _, c := range b.r.comments[n] {
		b.parseComment(c)
	}

	switch s := w.(type) {
	case *stmt.Global:
		for _, v := range s.Vars {
			ev := v.(*expr.Variable)

			// TODO: when varToString will handle Encapsed,
			// remove this check. Encapsed is a string with potentially
			// complex interpolation expressions.
			if _, ok := ev.VarName.(*scalar.Encapsed); ok {
				continue // Otherwise varToString would panic
			}

			b.addVar(ev, meta.NewTypesMap(meta.WrapGlobal(varToString(ev))), "global", true)
			b.addNonLocalVar(ev)
		}
		res = false
	case *stmt.Static:
		for _, vv := range s.Vars {
			v := vv.(*stmt.StaticVar)
			ev := v.Variable.(*expr.Variable)
			b.addVar(ev, solver.ExprTypeLocalCustom(b.sc, b.r.st, v.Expr, b.customTypes), "static", true)
			b.addNonLocalVar(ev)
			if v.Expr != nil {
				v.Expr.Walk(b)
			}
		}
		res = false
	case *stmt.StmtList:
		for _, st := range s.Stmts {
			b.addStatement(st)
		}
	// TODO: analyze control flow in try blocks separately and account for the fact that some functions or operations can
	// throw exceptions
	case *stmt.Try:
		res = b.handleTry(s)
	case *assign.Assign:
		// TODO: only accept first assignment, not all of them
		// e.g. if there is a condition like ($a = 10) || ($b = 5)
		// we must only accept $a = 10 as condition that is always executed
		res = b.handleAssign(s)
	case *assign.Reference:
		res = b.handleAssignReference(s)
	case *expr.Array:
		res = b.handleArray(s)
	case *expr.ShortArray:
		res = b.handleArrayItems(s, s.Items)
	case *stmt.Foreach:
		res = b.handleForeach(s)
	case *stmt.For:
		res = b.handleFor(s)
	case *stmt.While:
		res = b.handleWhile(s)
	case *stmt.Do:
		res = b.handleDo(s)
	case *stmt.If:
		// TODO: handle constant if expressions
		// TODO: maybe try to handle when variables are defined and used with the same condition
		res = b.handleIf(s)
	case *stmt.Switch:
		res = b.handleSwitch(s)
	case *expr.FunctionCall:
		res = b.handleFunctionCall(s)
	case *expr.MethodCall:
		res = b.handleMethodCall(s)
	case *expr.StaticCall:
		res = b.handleStaticCall(s)
	case *expr.PropertyFetch:
		res = b.handlePropertyFetch(s)
	case *expr.StaticPropertyFetch:
		res = b.handleStaticPropertyFetch(s)
	case *expr.ClassConstFetch:
		res = b.handleClassConstFetch(s)
	case *expr.ConstFetch:
		res = b.handleConstFetch(s)
	case *expr.New:
		res = b.handleNew(s)
	case *stmt.Unset:
		res = b.handleUnset(s)
	case *expr.Isset:
		res = b.handleIsset(s)
	case *expr.Empty:
		res = b.handleEmpty(s)
	case *expr.Variable:
		res = b.handleVariable(s)
	case *expr.ArrayDimFetch:
		b.checkArrayDimFetch(s)
	case *stmt.Function:
		if b.ignoreFunctionBodies {
			res = false
		} else {
			res = b.r.enterFunction(s)
		}
	case *stmt.Class:
		if b.ignoreFunctionBodies {
			res = false
		}
	case *stmt.Interface:
		if b.ignoreFunctionBodies {
			res = false
		}
	case *stmt.Trait:
		if b.ignoreFunctionBodies {
			res = false
		}
	case *expr.Closure:
		var typ *meta.TypesMap
		isInstance := b.sc.IsInInstanceMethod()
		if isInstance {
			typ, _ = b.sc.GetVarNameType("this")
		}
		res = b.enterClosure(s, isInstance, typ)
	case *stmt.Return:
		solver.ExprTypeLocalCustom(b.sc, b.r.st, s.Expr, b.customTypes).Iterate(func(t string) {
			b.returnTypes = b.returnTypes.AppendString(t)
		})
	default:
		// b.d.debug(`  Statement: %T`, s)
	}

	if res {
		for _, c := range b.custom {
			c.AfterEnterNode(w)
		}
	}

	return res
}

func (b *BlockWalker) addNonLocalVar(v *expr.Variable) {
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return
	}

	b.nonLocalVars[name.Value] = struct{}{}
}

// replaceVar must be used to track assignments to conrete var nodes if they are available
func (b *BlockWalker) replaceVar(v *expr.Variable, typ *meta.TypesMap, reason string, alwaysDefined bool) {
	b.sc.ReplaceVar(v, typ, reason, alwaysDefined)
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return
	}

	// Writes to non-local variables do count as usages
	if _, ok := b.nonLocalVars[name.Value]; ok {
		delete(b.unusedVars, name.Value)
		return
	}

	// Writes to variables that are done in a loop should not count as unused variables
	// because they can be read on the next iteration (ideally we should check for that too :))
	if !b.isLoopBody {
		b.unusedVars[name.Value] = append(b.unusedVars[name.Value], v)
	}
}

// addVar must be used to track assignments to conrete var nodes if they are available
func (b *BlockWalker) addVar(v *expr.Variable, typ *meta.TypesMap, reason string, alwaysDefined bool) {
	b.sc.AddVar(v, typ, reason, alwaysDefined)
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return
	}

	// Writes to non-local variables do count as usages
	if _, ok := b.nonLocalVars[name.Value]; ok {
		delete(b.unusedVars, name.Value)
		return
	}

	// Writes to variables that are done in a loop should not count as unused variables
	// because they can be read on the next iteration (ideally we should check for that too :))
	if !b.isLoopBody {
		b.unusedVars[name.Value] = append(b.unusedVars[name.Value], v)
	}
}

func (b *BlockWalker) parseComment(c comment.Comment) {
	str := c.String()

	if !phpdoc.IsPHPDoc(str) {
		return
	}

	for _, p := range phpdoc.Parse(str) {
		if p.Name != "var" {
			continue
		}

		if len(p.Params) < 2 {
			continue
		}

		varName, typ := p.Params[0], p.Params[1]
		if !strings.HasPrefix(varName, "$") && strings.HasPrefix(typ, "$") {
			varName, typ = typ, varName
		}

		if !strings.HasPrefix(varName, "$") {
			// TODO: report something about bad @var syntax
			continue
		}

		m := meta.NewTypesMap(b.r.maybeAddNamespace(typ))
		b.sc.AddVarFromPHPDoc(strings.TrimPrefix(varName, "$"), m, "@var")
	}
}

func (b *BlockWalker) handleUnset(s *stmt.Unset) bool {
	for _, v := range s.Vars {
		switch v := v.(type) {
		case *expr.Variable:
			if id, ok := v.VarName.(*node.Identifier); ok {
				delete(b.unusedVars, id.Value)
			}
			b.sc.DelVar(v, "unset")
		case *expr.ArrayDimFetch:
			b.handleIssetDimFetch(v) // unset($a["something"]) does not unset $a itself, so no delVar here
		default:
			if v != nil {
				v.Walk(b)
			}
		}
	}

	return false
}

func (b *BlockWalker) handleIsset(s *expr.Isset) bool {
	for _, v := range s.Variables {
		switch v := v.(type) {
		case *expr.Variable:
			if id, ok := v.VarName.(*node.Identifier); ok {
				delete(b.unusedVars, id.Value)
			}
		case *expr.ArrayDimFetch:
			b.handleIssetDimFetch(v)
		default:
			if v != nil {
				v.Walk(b)
			}
		}
	}

	return false
}

func (b *BlockWalker) handleEmpty(s *expr.Empty) bool {
	switch v := s.Expr.(type) {
	case *expr.Variable:
		if id, ok := v.VarName.(*node.Identifier); ok {
			delete(b.unusedVars, id.Value)
		}
	case *expr.ArrayDimFetch:
		b.handleIssetDimFetch(v)
	default:
		if v != nil {
			v.Walk(b)
		}
	}

	return false
}

func (b *BlockWalker) handleTry(s *stmt.Try) bool {
	if len(s.Catches) == 0 && s.Finally == nil {
		b.r.Report(s, LevelError, "bareTry", "At least one catch or finally block must be present")
	}

	contexts := make([]*BlockWalker, 0, len(s.Catches)+1)

	// Assume that no code in try{} block has executed because exceptions can be thrown from anywhere.
	// So we handle catches and finally blocks first.
	for _, c := range s.Catches {
		bCopy := b.copy()
		contexts = append(contexts, bCopy)
		b.r.addScope(c, bCopy.sc)
		cc := c.(*stmt.Catch)
		for _, s := range cc.Stmts {
			bCopy.addStatement(s)
		}
		bCopy.handleCatch(cc)
	}

	if s.Finally != nil {
		bCopy := b.copy()
		contexts = append(contexts, bCopy)
		b.r.addScope(s.Finally, bCopy.sc)
		cc := s.Finally.(*stmt.Finally)
		for _, s := range cc.Stmts {
			bCopy.addStatement(s)
		}
		s.Finally.Walk(bCopy)
	}

	// whether or not all other catches and finallies exit ("return", "throw", etc)
	othersExit := true
	prematureExitFlags := 0

	for _, ctx := range contexts {
		if ctx.exitFlags == 0 {
			othersExit = false
		} else {
			prematureExitFlags |= ctx.exitFlags
		}

		b.containsExitFlags |= ctx.containsExitFlags
	}

	tryB := b.copy()
	for _, s := range s.Stmts {
		tryB.addStatement(s)
		s.Walk(tryB)
		b.r.addScope(s, tryB.sc)
	}

	tryB.sc.Iterate(func(varName string, typ *meta.TypesMap, alwaysDefined bool) {
		b.sc.AddVarName(varName, typ, "try var", alwaysDefined && othersExit)
	})

	if othersExit && tryB.exitFlags != 0 {
		b.exitFlags |= prematureExitFlags
		b.exitFlags |= tryB.exitFlags
	}

	b.containsExitFlags |= tryB.containsExitFlags

	return false
}

func (b *BlockWalker) handleCatch(s *stmt.Catch) bool {
	m := meta.NewEmptyTypesMap(len(s.Types))
	for _, t := range s.Types {
		typ, ok := solver.GetClassName(b.r.st, t)
		if !ok {
			continue
		}
		m = m.AppendString(typ)
	}

	b.handleVariableNode(s.Variable, m, "catch")

	for _, stmt := range s.Stmts {
		if stmt != nil {
			b.addStatement(stmt)
			stmt.Walk(b)
		}
	}

	return false
}

// We still need to analyze expressions in isset()/unset()/empty() statements
func (b *BlockWalker) handleIssetDimFetch(e *expr.ArrayDimFetch) {
	b.checkArrayDimFetch(e)

	switch v := e.Variable.(type) {
	case *expr.Variable:
		if id, ok := v.VarName.(*node.Identifier); ok {
			delete(b.unusedVars, id.Value)
		}
	case *expr.ArrayDimFetch:
		b.handleIssetDimFetch(v)
	default:
		if v != nil {
			v.Walk(b)
		}
	}

	if e.Dim != nil {
		e.Dim.Walk(b)
	}
}

func (b *BlockWalker) checkArrayDimFetch(s *expr.ArrayDimFetch) {
	if !meta.IsIndexingComplete() {
		return
	}

	typ := solver.ExprType(b.sc, b.r.st, s.Variable)

	var (
		maybeHaveClasses bool
		haveArrayAccess  bool
	)

	typ.Iterate(func(t string) {
		// FullyQualified class name will have "\" in the beginning
		if len(t) > 0 && t[0] == '\\' && !strings.HasSuffix(t, "[]") {
			maybeHaveClasses = true

			if !haveArrayAccess && solver.Implements(t, `\ArrayAccess`) {
				haveArrayAccess = true
			}
		}
	})

	if maybeHaveClasses && !haveArrayAccess {
		b.r.Report(s.Variable, LevelDoNotReject, "arrayAccess", "Array access to non-array type %s", typ)
	}
}

func (b *BlockWalker) handleCallArgs(n node.Node, args []node.Node, fn meta.FuncInfo) {
	if len(args) < fn.MinParamsCnt {
		b.r.Report(n, LevelWarning, "argCount", "Too few arguments for %s", meta.NameNodeToString(n))
	}

	for i, arg := range args {
		if i >= len(fn.Params) {
			arg.Walk(b)
			continue
		}

		ref := fn.Params[i].IsRef

		switch a := arg.(*node.Argument).Expr.(type) {
		case *expr.Variable:
			if ref {
				b.addNonLocalVar(a)
				b.addVar(a, fn.Params[i].Typ, "call_with_ref", true /* TODO: variable may actually not be set by ref */)
				break
			}
			a.Walk(b)
		case *expr.ArrayDimFetch:
			if ref {
				b.handleDimFetchLValue(a, "call_with_ref", meta.NewTypesMap("array"))
				break
			}
			a.Walk(b)
		default:
			a.Walk(b)
		}
	}
}

func (b *BlockWalker) handleFunctionCall(e *expr.FunctionCall) bool {
	var fn meta.FuncInfo

	if meta.IsIndexingComplete() {
		defined := true
		canAnalyze := true

		switch nm := e.Function.(type) {
		case *name.Name:
			nameStr := meta.NameToString(nm)
			firstPart := nm.Parts[0].(*name.NamePart).Value
			if alias, ok := b.r.st.FunctionUses[firstPart]; ok {
				if len(nm.Parts) == 1 {
					nameStr = alias
				} else {
					// handle situations like 'use NS\Foo; Foo\Bar::doSomething();'
					nameStr = alias + `\` + meta.NamePartsToString(nm.Parts[1:])
				}
				fn, defined = meta.Info.GetFunction(nameStr)
			} else {
				fn, defined = meta.Info.GetFunction(b.r.st.Namespace + `\` + nameStr)
				if !defined && b.r.st.Namespace != "" {
					fn, defined = meta.Info.GetFunction(`\` + nameStr)
				}
			}

		case *name.FullyQualified:
			fn, defined = meta.Info.GetFunction(meta.FullyQualifiedToString(nm))
		default:
			defined = false

			solver.ExprTypeCustom(b.sc, b.r.st, nm, b.customTypes).Iterate(func(typ string) {
				if defined {
					return
				}
				fn, _, defined = solver.FindMethod(typ, `__invoke`)
			})

			if !defined {
				canAnalyze = false
			}
		}

		if !canAnalyze {
			return true
		}

		if !defined {
			b.r.Report(e.Function, LevelError, "undefined", "Call to undefined function %s", meta.NameNodeToString(e.Function))
		}
	}

	e.Function.Walk(b)

	b.handleCallArgs(e.Function, e.Arguments, fn)
	b.exitFlags |= fn.ExitFlags

	return false
}

// checks whether or not we can access to className::method/property/constant/etc from this context
func (b *BlockWalker) canAccess(className string, accessLevel meta.AccessLevel) bool {
	switch accessLevel {
	case meta.Private:
		return b.r.st.CurrentClass == className
	case meta.Protected:
		if b.r.st.CurrentClass == className {
			return true
		}

		// TODO: perhaps shpuld extract this common logic with visited map somewhere
		visited := make(map[string]struct{}, 8)
		parent := b.r.st.CurrentParentClass
		for parent != "" {
			if _, ok := visited[parent]; ok {
				return false
			}

			visited[parent] = struct{}{}

			if parent == className {
				return true
			}

			class, ok := meta.Info.GetClass(parent)
			if !ok {
				return false
			}

			parent = class.Parent
		}

		return false
	case meta.Public:
		return true
	}

	panic("Invalid access level")
}

func (b *BlockWalker) handleMethodCall(e *expr.MethodCall) bool {
	if !meta.IsIndexingComplete() {
		return true
	}

	var methodName string

	switch id := e.Method.(type) {
	case *node.Identifier:
		methodName = id.Value
	default:
		return true
	}

	var (
		foundMethod bool
		magic       bool
		fn          meta.FuncInfo
		implClass   string
	)

	exprType := solver.ExprTypeCustom(b.sc, b.r.st, e.Variable, b.customTypes)

	exprType.Iterate(func(typ string) {
		if foundMethod || magic {
			return
		}

		fn, implClass, foundMethod = solver.FindMethod(typ, methodName)
		magic = haveMagicMethod(typ, `__call`)
	})

	e.Variable.Walk(b)
	e.Method.Walk(b)

	if !foundMethod && !magic && !b.r.st.IsTrait && !b.isThisInsideClosure(e.Variable) {
		b.r.Report(e.Method, LevelError, "undefined", "Call to undefined method {%s}->%s()", exprType, methodName)
	}

	if foundMethod && !b.canAccess(implClass, fn.AccessLevel) {
		b.r.Report(e.Method, LevelError, "accessLevel", "Cannot access %s method %s->%s()", fn.AccessLevel, implClass, methodName)
	}

	b.handleCallArgs(e.Method, e.Arguments, fn)
	b.exitFlags |= fn.ExitFlags

	return false
}

func (b *BlockWalker) handleStaticCall(e *expr.StaticCall) bool {
	if !meta.IsIndexingComplete() {
		return true
	}

	var methodName string

	switch id := e.Call.(type) {
	case *node.Identifier:
		methodName = id.Value
	default:
		return true
	}

	className, ok := solver.GetClassName(b.r.st, e.Class)
	if !ok {
		return true
	}

	fn, implClass, ok := solver.FindMethod(className, methodName)

	e.Class.Walk(b)
	e.Call.Walk(b)

	if !ok && !haveMagicMethod(className, `__callStatic`) && !b.r.st.IsTrait {
		b.r.Report(e.Call, LevelError, "undefined", "Call to undefined method %s::%s()", className, methodName)
	}

	if ok && !b.canAccess(implClass, fn.AccessLevel) {
		b.r.Report(e.Call, LevelError, "accessLevel", "Cannot access %s method %s::%s()", fn.AccessLevel, implClass, methodName)
	}

	b.handleCallArgs(e.Call, e.Arguments, fn)
	b.exitFlags |= fn.ExitFlags

	return false
}

func (b *BlockWalker) isThisInsideClosure(varNode node.Node) bool {
	if !b.sc.IsInClosure() {
		return false
	}

	variable, ok := varNode.(*expr.Variable)
	if !ok {
		return false
	}

	if varName, ok := variable.VarName.(*node.Identifier); ok && varName.Value == `this` {
		return true
	}

	return false
}

func (b *BlockWalker) handlePropertyFetch(e *expr.PropertyFetch) bool {
	e.Variable.Walk(b)

	if !meta.IsIndexingComplete() {
		return false
	}

	id, ok := e.Property.(*node.Identifier)
	if !ok {
		return false
	}

	found := false
	magic := false
	var implClass string
	var info meta.PropertyInfo

	typ := solver.ExprTypeCustom(b.sc, b.r.st, e.Variable, b.customTypes)
	typ.Iterate(func(className string) {
		if found || magic {
			return
		}
		info, implClass, found = solver.FindProperty(className, id.Value)
		magic = haveMagicMethod(className, `__get`)
	})

	if !found && !magic && !b.r.st.IsTrait && !b.isThisInsideClosure(e.Variable) {
		b.r.Report(e.Property, LevelError, "undefined", "Property {%s}->%s does not exist", typ, id.Value)
	}

	if found && !b.canAccess(implClass, info.AccessLevel) {
		b.r.Report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s->%s", info.AccessLevel, implClass, id.Value)
	}

	return false
}

func (b *BlockWalker) handleStaticPropertyFetch(e *expr.StaticPropertyFetch) bool {
	e.Class.Walk(b)

	if !meta.IsIndexingComplete() {
		return false
	}

	varExpr, ok := e.Property.(*expr.Variable)
	if !ok {
		return false
	}

	varName, ok := varExpr.VarName.(*node.Identifier)
	if !ok {
		return false
	}

	className, ok := solver.GetClassName(b.r.st, e.Class)
	if !ok {
		return false
	}

	info, implClass, ok := solver.FindProperty(className, "$"+varName.Value)
	if !ok && !b.r.st.IsTrait {
		b.r.Report(e.Property, LevelError, "undefined", "Property %s::$%s does not exist", className, varName.Value)
	}

	if ok && !b.canAccess(implClass, info.AccessLevel) {
		b.r.Report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s::$%s", info.AccessLevel, implClass, varName.Value)
	}

	return false
}

func (b *BlockWalker) handleArray(arr *expr.Array) bool {
	b.r.Report(arr, LevelDoNotReject, "arraySyntax", "Use of old array syntax (use short form instead)")
	return b.handleArrayItems(arr, arr.Items)
}

func (b *BlockWalker) handleArrayItems(arr node.Node, items []node.Node) bool {
	haveKeys := false
	haveImplicitKeys := false
	keys := make(map[string]struct{}, len(items))

	for _, itemNode := range items {
		item, ok := itemNode.(*expr.ArrayItem)
		if !ok {
			// TODO: it is possible at all here?
			continue
		}

		if item.Key == nil {
			haveImplicitKeys = true
			continue
		}

		haveKeys = true

		var key string
		var constKey bool

		switch k := item.Key.(type) {
		case *scalar.String:
			key = strings.TrimFunc(k.Value, isQuote)
			constKey = true
		case *scalar.Lnumber:
			key = k.Value
			constKey = true
		}

		if !constKey {
			continue
		}

		if _, ok := keys[key]; ok {
			b.Report(item.Key, LevelWarning, "arrayKeys", "Duplicate array key '%s'", key)
		}

		keys[key] = struct{}{}
	}

	if haveImplicitKeys && haveKeys {
		b.Report(arr, LevelWarning, "arrayKeys", "Mixing implicit and explicit array keys")
	}

	return true
}

func haveMagicMethod(class string, methodName string) bool {
	_, _, ok := solver.FindMethod(class, methodName)
	return ok
}

func (b *BlockWalker) handleClassConstFetch(e *expr.ClassConstFetch) bool {
	if !meta.IsIndexingComplete() {
		return true
	}

	constName, ok := e.ConstantName.(*node.Identifier)
	if !ok {
		return false
	}

	if constName.Value == `class` || constName.Value == `CLASS` {
		return false
	}

	className, ok := solver.GetClassName(b.r.st, e.Class)
	if !ok {
		return false
	}

	info, implClass, ok := solver.FindConstant(className, constName.Value)

	e.Class.Walk(b)

	if !ok && !b.r.st.IsTrait {
		b.r.Report(e.ConstantName, LevelError, "undefined", "Class constant %s::%s does not exist", className, constName.Value)
	}

	if ok && !b.canAccess(implClass, info.AccessLevel) {
		b.r.Report(e.ConstantName, LevelError, "accessLevel", "Cannot access %s constant %s::%s", info.AccessLevel, implClass, constName.Value)
	}

	return false
}

func (b *BlockWalker) handleConstFetch(e *expr.ConstFetch) bool {
	if !meta.IsIndexingComplete() {
		return true
	}

	_, _, defined := solver.GetConstant(b.r.st, e.Constant)

	if !defined {
		b.r.Report(e.Constant, LevelError, "undefined", "Undefined constant %s", meta.NameNodeToString(e.Constant))
	}

	return true
}

func (b *BlockWalker) handleNew(e *expr.New) bool {
	if !meta.IsIndexingComplete() {
		return true
	}

	className, ok := solver.GetClassName(b.r.st, e.Class)
	if !ok {
		// perhaps something like 'new $class', cannot check this
		return true
	}

	if _, ok := meta.Info.GetClass(className); !ok {
		b.r.Report(e.Class, LevelError, "undefined", "Class not found %s", className)
	}

	return true
}

func (b *BlockWalker) handleForeach(s *stmt.Foreach) bool {
	// TODO: add reference semantics to foreach analyze as well

	b.handleVariableNode(s.Key, meta.NewTypesMap("foreach_key"), "foreach_key")
	b.handleVariableNode(s.Variable, meta.NewTypesMap("foreach_value"), "foreach_value")

	// expression is always executed and is executed in base context
	if s.Expr != nil {
		s.Expr.Walk(b)
		solver.ExprTypeLocalCustom(b.sc, b.r.st, s.Expr, b.customTypes).Iterate(func(typ string) {
			b.handleVariableNode(s.Variable, meta.NewTypesMap(meta.WrapElemOf(typ)), "foreach_value")
		})
	}

	// foreach body can do 0 cycles so we need a separate context for that
	if s.Stmt != nil {
		bCopy := b.copy()
		bCopy.isLoopBody = true

		if _, ok := s.Stmt.(*stmt.StmtList); !ok {
			bCopy.addStatement(s.Stmt)
		}

		s.Stmt.Walk(bCopy)
		b.maybeAddAllVars(bCopy.sc, "foreach body")
		if !bCopy.returnTypes.IsEmpty() {
			b.returnTypes = b.returnTypes.Append(bCopy.returnTypes)
		}
	}

	return false
}

func (b *BlockWalker) handleFor(s *stmt.For) bool {
	for _, v := range s.Init {
		b.addStatement(v)
		v.Walk(b)
	}

	for _, v := range s.Cond {
		v.Walk(b)
	}

	for _, v := range s.Loop {
		b.addStatement(v)
		v.Walk(b)
	}

	// for body can do 0 cycles so we need a separate context for that
	if s.Stmt != nil {
		bCopy := b.copy()
		bCopy.isLoopBody = true
		s.Stmt.Walk(bCopy)
		b.maybeAddAllVars(bCopy.sc, "while body")
		if !bCopy.returnTypes.IsEmpty() {
			b.returnTypes = b.returnTypes.Append(bCopy.returnTypes)
		}
	}

	return false
}

func (b *BlockWalker) enterClosure(fun *expr.Closure, haveThis bool, thisType *meta.TypesMap) bool {
	sc := meta.NewScope()
	sc.SetInClosure(true)

	if haveThis {
		sc.AddVarName("this", thisType, "closure inside instance method", true)
	} else {
		sc.AddVarName("this", meta.NewTypesMap("possibly_late_bound"), "possibly late bound $this", true)
	}

	_, phpDocParamTypes, phpDocError := b.r.parsePHPDoc(fun.PhpDocComment, fun.Params)

	if phpDocError != "" {
		b.r.Report(fun, LevelInformation, "phpdoc", "PHPDoc is incorrect: %s", phpDocError)
	}

	for _, useExpr := range fun.Uses {
		u := useExpr.(*expr.ClosureUse)
		v := u.Variable.(*expr.Variable)
		varName := v.VarName.(*node.Identifier).Value

		if !b.sc.HaveVar(v) && !u.ByRef {
			b.r.Report(v, LevelWarning, "undefined", "Undefined variable %s", varName)
		}

		typ, ok := b.sc.GetVarNameType(varName)
		if ok {
			sc.AddVarName(varName, typ, "use", true)
		}

		delete(b.unusedVars, varName)
	}

	params, _ := b.r.parseFuncArgs(fun.Params, phpDocParamTypes, sc)

	b.r.handleFuncStmts(params, fun.Uses, fun.Stmts, sc)
	b.r.addScope(fun, sc)

	return false
}

func (b *BlockWalker) maybeAddAllVars(sc *meta.Scope, reason string) {
	sc.Iterate(func(varName string, typ *meta.TypesMap, alwaysDefined bool) {
		b.sc.AddVarName(varName, typ, reason, false)
	})
}

func (b *BlockWalker) handleWhile(s *stmt.While) bool {
	if s.Cond != nil {
		s.Cond.Walk(b)
	}

	// while body can do 0 cycles so we need a separate context for that
	if s.Stmt != nil {
		bCopy := b.copy()
		bCopy.isLoopBody = true
		s.Stmt.Walk(bCopy)
		b.maybeAddAllVars(bCopy.sc, "while body")
		if !bCopy.returnTypes.IsEmpty() {
			b.returnTypes = b.returnTypes.Append(bCopy.returnTypes)
		}
	}

	return false
}

func (b *BlockWalker) handleDo(s *stmt.Do) bool {
	if s.Stmt != nil {
		oldIsLoopBody := b.isLoopBody
		b.isLoopBody = true
		s.Stmt.Walk(b)
		b.isLoopBody = oldIsLoopBody
	}

	if s.Cond != nil {
		s.Cond.Walk(b)
	}

	return false
}

// Propagate premature exit flags from visited branches ("contexts").
func (b *BlockWalker) propagateFlagsFromBranches(contexts []*BlockWalker, linksCount int) {
	allExit := false
	prematureExitFlags := 0

	for _, ctx := range contexts {
		b.containsExitFlags |= ctx.containsExitFlags
	}

	if len(contexts) > 0 && linksCount == 0 {
		allExit = true

		for _, ctx := range contexts {
			if ctx.exitFlags == 0 {
				allExit = false
			} else {
				prematureExitFlags |= ctx.exitFlags
			}
		}
	}

	if allExit {
		b.exitFlags |= prematureExitFlags
	}
}

// andWalker walks through all expressions with && and does not enter deeper
type andWalker struct {
	issets      []*expr.Isset
	instanceOfs []*expr.InstanceOf
}

func (a *andWalker) EnterNode(w walker.Walkable) (res bool) {
	switch n := w.(type) {
	case *binary.BooleanAnd:
		return true
	case *expr.Isset:
		a.issets = append(a.issets, n)
	case *expr.InstanceOf:
		a.instanceOfs = append(a.instanceOfs, n)
	}
	return false
}

func (a *andWalker) GetChildrenVisitor(key string) walker.Visitor { return a }
func (a *andWalker) LeaveNode(w walker.Walkable)                  {}

func (b *BlockWalker) handleVariable(v *expr.Variable) bool {
	if !b.sc.HaveVar(v) {
		b.r.reportUndefinedVariable(v, b.sc.MaybeHaveVar(v))
		b.sc.AddVar(v, meta.NewTypesMap("undefined"), "undefined", true)
	} else if id, ok := v.VarName.(*node.Identifier); ok {
		delete(b.unusedVars, id.Value)
	}
	return false
}

func (b *BlockWalker) handleIf(s *stmt.If) bool {
	// first condition is always executed, so run it in base context
	if s.Cond != nil {
		a := &andWalker{}

		s.Cond.Walk(b)
		s.Cond.Walk(a)

		for _, isset := range a.issets {
			for _, v := range isset.Variables {
				if v, ok := v.(*expr.Variable); ok {
					if b.sc.HaveVar(v) {
						continue
					}
					switch vn := v.VarName.(type) {
					case *node.Identifier:
						b.addVar(v, meta.NewTypesMap("isset_$"+vn.Value), "isset", true)
						defer b.sc.DelVar(v, "isset")
					case *expr.Variable:
						b.handleVariable(vn)
						name, ok := vn.VarName.(*node.Identifier)
						if !ok {
							continue
						}
						b.addVar(v, meta.NewTypesMap("isset_$$"+name.Value), "isset", true)
						defer b.sc.DelVar(v, "isset")
					}
				}
			}
		}

		for _, instanceof := range a.instanceOfs {
			if className, ok := solver.GetClassName(b.r.st, instanceof.Class); ok {
				if v, ok := instanceof.Expr.(*expr.Variable); ok {
					b.addVar(v, meta.NewTypesMap(className), "instanceof", false)
				} else {
					b.customTypes = append(b.customTypes, solver.CustomType{
						Node: instanceof.Expr,
						Typ:  meta.NewTypesMap(className),
					})
				}
				// TODO: actually this needs to be present inside if body only
			}
		}
	}

	var contexts []*BlockWalker

	walk := func(n node.Node) (links int) {
		bCopy := b.copy()
		contexts = append(contexts, bCopy)

		// handle if (...) smth(); else other_thing(); // without braces
		if els, ok := n.(*stmt.Else); ok {
			bCopy.addStatement(els.Stmt)
		} else if elsif, ok := n.(*stmt.ElseIf); ok {
			bCopy.addStatement(elsif.Stmt)
		} else {
			bCopy.addStatement(n)
		}

		n.Walk(bCopy)

		b.r.addScope(n, bCopy.sc)

		if bCopy.exitFlags != 0 {
			return 0
		}

		return 1
	}

	linksCount := 0

	if s.Stmt != nil {
		linksCount += walk(s.Stmt)
	} else {
		linksCount++
	}

	for _, n := range s.ElseIf {
		linksCount += walk(n)
	}

	if s.Else != nil {
		linksCount += walk(s.Else)
	} else {
		linksCount++
	}

	b.propagateFlagsFromBranches(contexts, linksCount)

	varTypes := make(map[string]*meta.TypesMap, b.sc.Len())
	defCounts := make(map[string]int, b.sc.Len())

	for _, ctx := range contexts {
		if ctx.exitFlags != 0 {
			b.returnTypes = b.returnTypes.Append(ctx.returnTypes)
			continue
		}

		ctx.sc.Iterate(func(nm string, typ *meta.TypesMap, alwaysDefined bool) {
			varTypes[nm] = varTypes[nm].Append(typ)
			if alwaysDefined {
				defCounts[nm]++
			}
		})
	}

	for nm, types := range varTypes {
		b.sc.AddVarName(nm, types, "all branches", defCounts[nm] == linksCount)
	}

	return false
}

func (b *BlockWalker) getCaseStmts(c node.Node) (list []node.Node, isDefault bool) {
	switch c := c.(type) {
	case *stmt.Case:
		list = c.Stmts
	case *stmt.Default:
		list = c.Stmts
		if len(list) > 0 {
			isDefault = true
		}
	default:
		panic(fmt.Errorf("Unexpected type in switch statement: %T", c))
	}

	return list, isDefault
}

func (b *BlockWalker) iterateNextCases(cases []node.Node, startIdx int) {
	for i := startIdx; i < len(cases); i++ {
		list, _ := b.getCaseStmts(cases[i])

		for _, stmt := range list {
			if stmt != nil {
				b.addStatement(stmt)
				stmt.Walk(b)
				if b.exitFlags != 0 {
					return
				}
			}
		}
	}
}

func (b *BlockWalker) handleSwitch(s *stmt.Switch) bool {
	// first condition is always executed, so run it in base context
	if s.Cond != nil {
		s.Cond.Walk(b)
	}

	var contexts []*BlockWalker

	linksCount := 0
	haveDefault := false
	breakFlags := FlagBreak | FlagContinue

	for idx, c := range s.Cases {
		var list []node.Node

		list, isDefault := b.getCaseStmts(c)
		if isDefault {
			haveDefault = true
		}

		// allow empty case body without "break;"
		// nothing new can be defined here so we just skip it
		if len(list) == 0 {
			continue
		}

		bCopy := b.copy()
		contexts = append(contexts, bCopy)

		for _, stmt := range list {
			if stmt != nil {
				bCopy.addStatement(stmt)
				stmt.Walk(bCopy)
			}
		}

		// allow to omit "break;" in the final statement
		if idx != len(s.Cases)-1 && bCopy.exitFlags == 0 {
			// allow the fallthrough if appropriate comment is present
			nextCase := s.Cases[idx+1]
			if !b.caseHasFallthroughComment(nextCase) {
				b.r.Report(c, LevelInformation, "caseBreak", "Add break or '// fallthrough' to the end of the case")
			}
		}

		if (bCopy.exitFlags & (^breakFlags)) == 0 {
			linksCount++

			if bCopy.exitFlags == 0 {
				bCopy.iterateNextCases(s.Cases, idx+1)
			}
		}
	}

	if !haveDefault {
		linksCount++
	}

	// whether or not all branches exit (return, throw, etc)
	allExit := false
	prematureExitFlags := 0

	if len(contexts) > 0 && haveDefault {
		allExit = true

		for _, ctx := range contexts {
			cleanFlags := ctx.exitFlags & (^breakFlags)
			if cleanFlags == 0 {
				allExit = false
			} else {
				prematureExitFlags |= cleanFlags
			}
			b.containsExitFlags |= ctx.containsExitFlags
		}
	}

	if allExit {
		b.exitFlags |= prematureExitFlags
	}

	varTypes := make(map[string]*meta.TypesMap, b.sc.Len())
	defCounts := make(map[string]int, b.sc.Len())

	for _, ctx := range contexts {
		cleanFlags := ctx.exitFlags & (^breakFlags)
		if cleanFlags != 0 {
			b.returnTypes = b.returnTypes.Append(ctx.returnTypes)
			continue
		}

		ctx.sc.Iterate(func(nm string, typ *meta.TypesMap, alwaysDefined bool) {
			varTypes[nm] = varTypes[nm].Append(typ)
			if alwaysDefined {
				defCounts[nm]++
			}
		})
	}

	for nm, types := range varTypes {
		b.sc.AddVarName(nm, types, "all cases", defCounts[nm] == linksCount)
	}

	return false
}

// if $a was previously undefined,
// handle case when doing assignment like '$a[] = 4;'
// or call to function that accepts like exec("command", $a)
func (b *BlockWalker) handleDimFetchLValue(e *expr.ArrayDimFetch, reason string, typ *meta.TypesMap) {
	b.checkArrayDimFetch(e)

	switch v := e.Variable.(type) {
	case *expr.Variable:
		arrTyp := meta.NewEmptyTypesMap(typ.Len())
		typ.Iterate(func(t string) {
			arrTyp = arrTyp.AppendString(meta.WrapArrayOf(t))
		})
		b.addVar(v, arrTyp, reason, true)
	case *expr.ArrayDimFetch:
		b.handleDimFetchLValue(v, reason, meta.NewTypesMap("array"))
	default:
		// probably not assignable?
		v.Walk(b)
	}

	if e.Dim != nil {
		e.Dim.Walk(b)
	}
}

// some day, perhaps, there will be some difference between handleAssignReference and handleAssign
func (b *BlockWalker) handleAssignReference(a *assign.Reference) bool {
	switch v := a.Variable.(type) {
	case *expr.ArrayDimFetch:
		b.handleDimFetchLValue(v, "assign_array", meta.NewTypesMap("array"))
		a.Expression.Walk(b)
		return false
	case *expr.Variable:
		b.addVar(v, solver.ExprTypeLocal(b.sc, b.r.st, a.Expression), "assign", true)
	case *expr.List:
		for _, item := range v.Items {
			arrayItem, ok := item.(*expr.ArrayItem)
			if !ok {
				continue
			}

			b.handleVariableNode(arrayItem.Val, meta.NewTypesMap("unknown_from_list"), "assign")
		}
	default:
		a.Variable.Walk(b)
	}

	a.Expression.Walk(b)
	return false
}

func (b *BlockWalker) handleAssignList(items []node.Node) {
	for _, item := range items {
		arrayItem, ok := item.(*expr.ArrayItem)
		if !ok {
			continue
		}

		b.handleVariableNode(arrayItem.Val, meta.NewTypesMap("unknown_from_list"), "assign")
	}
}

func (b *BlockWalker) handleAssign(a *assign.Assign) bool {
	a.Expression.Walk(b)

	switch v := a.Variable.(type) {
	case *expr.ArrayDimFetch:
		typ := solver.ExprTypeLocal(b.sc, b.r.st, a.Expression)
		b.handleDimFetchLValue(v, "assign_array", typ)
		return false
	case *expr.Variable:
		b.replaceVar(v, solver.ExprTypeLocal(b.sc, b.r.st, a.Expression), "assign", true)
	case *expr.List:
		b.handleAssignList(v.Items)
	case *expr.ShortList:
		b.handleAssignList(v.Items)
	case *expr.PropertyFetch:
		varNode, ok := v.Variable.(*expr.Variable)
		if !ok {
			break
		}

		id, ok := varNode.VarName.(*node.Identifier)
		if !ok {
			break
		}

		delete(b.unusedVars, id.Value)

		if id.Value != "this" {
			break
		}

		if b.r.st.CurrentClass == "" {
			break
		}

		propertyName, ok := v.Property.(*node.Identifier)
		if !ok {
			break
		}

		cls := b.r.getClass()

		p := cls.Properties[propertyName.Value]
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.sc, b.r.st, a.Expression, b.customTypes))
		cls.Properties[propertyName.Value] = p
	case *expr.StaticPropertyFetch:
		if b.r.st.CurrentClass == "" {
			break
		}

		className, ok := solver.GetClassName(b.r.st, v.Class)
		if !ok || className != b.r.st.CurrentClass {
			break
		}

		varNode, ok := v.Property.(*expr.Variable)
		if !ok {
			break
		}

		id, ok := varNode.VarName.(*node.Identifier)
		if !ok {
			break
		}

		cls := b.r.getClass()

		p := cls.Properties["$"+id.Value]
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.sc, b.r.st, a.Expression, b.customTypes))
		cls.Properties["$"+id.Value] = p
	default:
		a.Variable.Walk(b)
	}

	return false
}

func (b *BlockWalker) flushUnused() {
	if !meta.IsIndexingComplete() {
		return
	}

	visitedMap := make(map[node.Node]struct{})
	for name, nodes := range b.unusedVars {
		if name == "_" {
			// blank identifier is a way to tell linter (and PHPStorm) that result is explicitly unused
			continue
		}

		if _, ok := superGlobals[name]; ok {
			continue
		}

		for _, n := range nodes {
			if _, ok := visitedMap[n]; ok {
				continue
			}

			visitedMap[n] = struct{}{}
			b.r.Report(n, LevelUnused, "unused", `Unused variable %s (use $_ to ignore this inspection)`, name)
		}
	}
}

func (b *BlockWalker) handleVariableNode(n node.Node, typ *meta.TypesMap, what string) {
	if n == nil {
		return
	}

	vv, ok := n.(*expr.Variable)
	if !ok {
		return
	}

	b.addVar(vv, typ, what, true)
}

// GetChildrenVisitor is useless :)
func (b *BlockWalker) GetChildrenVisitor(key string) walker.Visitor {
	return b
}

// LeaveNode is called after all children have been visited.
func (b *BlockWalker) LeaveNode(w walker.Walkable) {
	for _, c := range b.custom {
		c.BeforeLeaveNode(w)
	}

	if b.exitFlags == 0 {
		switch w.(type) {
		case *stmt.Return:
			b.exitFlags |= FlagReturn
			b.containsExitFlags |= FlagReturn
		case *expr.Die, *expr.Exit:
			b.exitFlags |= FlagDie
			b.containsExitFlags |= FlagDie
		case *stmt.Throw:
			b.exitFlags |= FlagThrow
			b.containsExitFlags |= FlagThrow
		case *stmt.Continue:
			b.exitFlags |= FlagContinue
			b.containsExitFlags |= FlagContinue
		case *stmt.Break:
			b.exitFlags |= FlagBreak
			b.containsExitFlags |= FlagBreak
		}
	}

	for _, c := range b.custom {
		c.AfterLeaveNode(w)
	}
}

var fallthroughMarkerRegex = func() *regexp.Regexp {
	markers := []string{
		"fallthrough",
		"fall through",
		"falls through",
		"no break",
	}

	pattern := `(?:/\*|//)\s?(?:` + strings.Join(markers, `|`) + `)`
	return regexp.MustCompile(pattern)
}()

func (b *BlockWalker) caseHasFallthroughComment(n node.Node) bool {
	for _, comment := range b.r.comments[n] {
		str := comment.String()
		if fallthroughMarkerRegex.MatchString(str) {
			return true
		}
	}
	return false
}
