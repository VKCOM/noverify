package linter

import (
	"fmt"
	"strings"

	"github.com/z7zmey/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
)

// loopKind describes current looping statement context.
type loopKind int

const (
	// loopNone is "there is no enclosing loop" context.
	loopNone loopKind = iota

	// loopFor is for all usual loops in PHP, like for/foreach loops.
	loopFor

	// loopSwitch is for switch statement, that is considered to
	// be a looping construction in PHP.
	loopSwitch
)

// To make "unused" linter happy.
const (
	_ = loopNone
	_ = varLocal
)

const (
	// FlagReturn shows whether or not block has "return"
	FlagReturn = 1 << iota
	FlagBreak
	FlagContinue
	FlagThrow
	FlagDie
)

type variableKind int

const (
	varLocal variableKind = iota
	varRef
	varGlobal
	varCondGlobal
	varStatic
)

// arrayKeyType is an universal PHP array key type.
// In PHP, all array keys are converted to either int or string,
// so we can always assume that `int|string` is good enough.
var arrayKeyType = meta.NewTypesMap("int|string").Immutable()

// blockWalker is used to process function/method contents.
type blockWalker struct {
	ctx *blockContext

	linter blockLinter

	// inferred return types if any
	returnTypes meta.TypesMap

	r *rootWalker

	custom []BlockChecker

	path irutil.NodePath

	ignoreFunctionBodies bool
	rootLevel            bool // analysing root-level code

	// state
	statements map[ir.Node]struct{}

	// whether a function has a return without explit expression.
	// Required to make a decision in void vs null type selection,
	// since "return" is the same as "return null".
	bareReturn bool
	// whether a function has a return with explicit expression.
	// When can't infer precise type, can use mixed.
	returnsValue bool
	// whether func_get_args() was called.
	callsFuncGetArgs bool

	// callsParentConstructor is set to true when parent::__construct() call
	// is found. This is needed for a root walker to report constructors
	// that do not call parent constructors.
	callsParentConstructor bool

	// shared state between all blocks
	unusedVars   map[string][]ir.Node
	unusedParams map[string]struct{}

	// static, global and other vars that have complex control flow.
	// Never contains varLocal elements.
	nonLocalVars map[string]variableKind

	inArrowFunction    bool
	parentBlockWalkers []*blockWalker // all parent block walkers if we handle nested arrow functions.
}

func newBlockWalker(r *rootWalker, sc *meta.Scope) *blockWalker {
	b := &blockWalker{
		r:            r,
		ctx:          &blockContext{sc: sc},
		unusedVars:   make(map[string][]ir.Node),
		nonLocalVars: make(map[string]variableKind),
		path:         irutil.NewNodePath(),
	}
	b.linter = blockLinter{walker: b}
	return b
}

func (b *blockWalker) isIndexingComplete() bool {
	return b.r.ctx.st.Info.IsIndexingComplete()
}

func (b *blockWalker) addStatement(n ir.Node) {
	if b.statements == nil {
		b.statements = make(map[ir.Node]struct{})
	}
	b.statements[n] = struct{}{}

	// A hack for assignment-used-as-expression checks to work
	e, ok := n.(*ir.ExpressionStmt)
	if !ok {
		return
	}

	assignment, ok := e.Expr.(*ir.Assign)
	if !ok {
		return
	}

	b.statements[assignment] = struct{}{}
}

func (b *blockWalker) reportDeadCode(n ir.Node) {
	if b.ctx.deadCodeReported {
		return
	}

	switch n.(type) {
	case *ir.BreakStmt, *ir.ReturnStmt, *ir.ExitExpr, *ir.ThrowStmt:
		// Allow to break code flow more than once.
		// This is useful in situation like this:
		//
		//    callSomeFuncThatExits(); exit;
		//
		// You can explicitly mark that function exits unconditionally for code clarity.
		return
	case *ir.FunctionStmt, *ir.ClassStmt, *ir.ConstListStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		// when we analyze root scope, function definions and other things are parsed even after exit, throw, etc
		if b.ignoreFunctionBodies {
			return
		}
	}

	b.ctx.deadCodeReported = true
	b.r.Report(n, LevelWarning, "deadCode", "Unreachable code")
}

func (b *blockWalker) handleComments(n ir.Node) {
	n.IterateTokens(func(t *token.Token) bool {
		b.handleToken(n, t)
		return true
	})
}

func (b *blockWalker) handleToken(n ir.Node, t *token.Token) {
	if t == nil {
		return
	}

	if t.ID != token.T_DOC_COMMENT && t.ID != token.T_COMMENT {
		return
	}
	str := string(t.Value)

	if !phpdoc.IsPHPDoc(str) {
		return
	}

	for _, p := range phpdoc.Parse(b.r.ctx.phpdocTypeParser, str) {
		p, ok := p.(*phpdoc.TypeVarCommentPart)
		if !ok || p.Name() != "var" {
			continue
		}

		types, warning := typesFromPHPDoc(&b.r.ctx, p.Type)
		if warning != "" {
			b.r.Report(n, LevelWarning, "phpdocType", "%s on line %d", warning, p.Line())
		}
		m := newTypesMap(&b.r.ctx, types)
		b.ctx.sc.AddVarFromPHPDoc(strings.TrimPrefix(p.Var, "$"), m, "@var")
	}
}

// EnterNode is called before walking to inner nodes.
func (b *blockWalker) EnterNode(n ir.Node) (res bool) {
	res = true

	for _, c := range b.custom {
		c.BeforeEnterNode(n)
	}

	b.path.Push(n)

	if b.ctx.exitFlags != 0 {
		b.reportDeadCode(n)
	}

	b.handleComments(n)

	switch s := n.(type) {
	case *ir.LogicalOrExpr:
		res = b.handleLogicalOr(s)
	case *ir.ArrayDimFetchExpr:
		b.checkArrayDimFetch(s)
	case *ir.GlobalStmt:
		b.handleAndCheckGlobalStmt(s)
		res = false
	case *ir.StaticStmt:
		for _, vv := range s.Vars {
			v := vv.(*ir.StaticVarStmt)
			ev := v.Variable
			typ := solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, v.Expr, b.ctx.customTypes)
			// Static vars can be assigned below and preserve the type of
			// the previously assigned value.
			typ.MarkAsImprecise()
			b.addVarName(v, ev.Name, typ, "static", meta.VarAlwaysDefined)
			b.addNonLocalVarName(ev.Name, varStatic)
			if v.Expr != nil {
				v.Expr.Walk(b)
			}
		}
		res = false
	case *ir.Root:
		for _, st := range s.Stmts {
			b.addStatement(st)
		}
	case *ir.StmtList:
		for _, st := range s.Stmts {
			b.addStatement(st)
		}
	// TODO: analyze control flow in try blocks separately and account for the fact that some functions or operations can
	// throw exceptions
	case *ir.TryStmt:
		res = b.handleTry(s)
	case *ir.Assign:
		// TODO: only accept first assignment, not all of them
		// e.g. if there is a condition like ($a = 10) || ($b = 5)
		// we must only accept $a = 10 as condition that is always executed
		res = b.handleAssign(s)
	case *ir.AssignReference:
		res = b.handleAssignReference(s)
	case *ir.AssignPlus:
		b.handleAssignOp(s)
	case *ir.AssignMinus:
		b.handleAssignOp(s)
	case *ir.AssignMul:
		b.handleAssignOp(s)
	case *ir.AssignDiv:
		b.handleAssignOp(s)
	case *ir.AssignConcat:
		b.handleAssignOp(s)
	case *ir.AssignShiftLeft:
		b.handleAssignOp(s)
	case *ir.AssignShiftRight:
		b.handleAssignOp(s)
	case *ir.AssignCoalesce:
		b.handleAssignOp(s)
	case *ir.ArrayExpr:
		res = b.handleArray(s)
	case *ir.ForeachStmt:
		res = b.handleForeach(s)
	case *ir.ForStmt:
		res = b.handleFor(s)
	case *ir.WhileStmt:
		res = b.handleWhile(s)
	case *ir.DoStmt:
		res = b.handleDo(s)
	case *ir.ElseIfStmt:
		b.handleElseIf(s)
	case *ir.IfStmt:
		// TODO: handle constant if expressions
		// TODO: maybe try to handle when variables are defined and used with the same condition
		res = b.handleIf(s)
	case *ir.SwitchStmt:
		res = b.handleSwitch(s)
	case *ir.TernaryExpr:
		res = b.handleTernary(s)
	case *ir.FunctionCallExpr:
		res = b.handleFunctionCall(s)
	case *ir.MethodCallExpr:
		res = b.handleMethodCall(s)
	case *ir.StaticCallExpr:
		res = b.handleStaticCall(s)
	case *ir.PropertyFetchExpr:
		res = b.handlePropertyFetch(s)
	case *ir.StaticPropertyFetchExpr:
		res = b.handleStaticPropertyFetch(s)
	case *ir.ClassConstFetchExpr:
		res = b.handleClassConstFetch(s)
	case *ir.UnsetStmt:
		res = b.handleUnset(s)
	case *ir.IssetExpr:
		res = b.handleIsset(s)
	case *ir.EmptyExpr:
		res = b.handleEmpty(s)
	case *ir.Var:
		res = b.handleVariable(s)
	case *ir.SimpleVar:
		res = b.handleVariable(s)
	case *ir.FunctionStmt:
		res = b.handleFunction(s)
	case *ir.ArrowFunctionExpr:
		res = b.handleArrowFunction(s)
	case *ir.AnonClassExpr:
		res = !b.ignoreFunctionBodies
	case *ir.ClassStmt:
		if b.ignoreFunctionBodies {
			res = false
		}
	case *ir.InterfaceStmt:
		if b.ignoreFunctionBodies {
			res = false
		}
	case *ir.TraitStmt:
		if b.ignoreFunctionBodies {
			res = false
		}
	case *ir.ClosureExpr:
		var typ meta.TypesMap
		isInstance := b.ctx.sc.IsInInstanceMethod()
		if isInstance {
			typ, _ = b.ctx.sc.GetVarNameType("this")
		}
		res = b.enterClosure(s, isInstance, typ, nil)

	case *ir.ReturnStmt:
		b.handleReturn(s)

	case *ir.CatchStmt:
		b.handleCatch(s)
		res = false
	}

	for _, c := range b.custom {
		c.AfterEnterNode(n)
	}

	if b.isIndexingComplete() {
		b.linter.enterNode(n)
	}
	if b.isIndexingComplete() {
		// Note: no need to check localRset for nil.
		kind := ir.GetNodeKind(n)
		if b.r.anyRset != nil {
			b.r.runRules(n, b.ctx.sc, b.r.anyRset.RulesByKind[kind])
		} else if !b.rootLevel && b.r.localRset != nil {
			b.r.runRules(n, b.ctx.sc, b.r.localRset.RulesByKind[kind])
		}
	}

	if !res {
		b.path.Pop()
	}
	return res
}

func (b *blockWalker) checkDupGlobal(s *ir.GlobalStmt) {
	vars := make(map[string]struct{}, len(s.Vars))
	for _, v := range s.Vars {
		v, ok := v.(*ir.SimpleVar)
		if !ok {
			continue
		}
		nm := v.Name

		// Check whether this var was already global'ed.
		// We use nonLocalVars for function-wide analysis and vars for local analysis.
		if _, ok := vars[nm]; ok {
			b.r.Report(v, LevelWarning, "dupGlobal", "global statement mentions $%s more than once", nm)
		} else {
			vars[nm] = struct{}{}
			if b.nonLocalVars[nm] == varGlobal {
				b.r.Report(v, LevelNotice, "dupGlobal", "$%s already global'ed above", nm)
			}
		}
	}
}

func (b *blockWalker) handleAndCheckGlobalStmt(s *ir.GlobalStmt) {
	if !b.rootLevel {
		b.checkDupGlobal(s)
	}

	for _, v := range s.Vars {
		nm := varToString(v)
		if nm == "" {
			continue
		}

		b.addVar(v, meta.NewTypesMap(meta.WrapGlobal(nm)), "global", meta.VarAlwaysDefined)
		if b.path.Conditional() {
			b.addNonLocalVar(v, varCondGlobal)
		} else {
			b.addNonLocalVar(v, varGlobal)
		}
	}
}

func (b *blockWalker) handleFunction(fun *ir.FunctionStmt) bool {
	if b.ignoreFunctionBodies {
		return false
	}

	return b.r.enterFunction(fun)
}

func (b *blockWalker) handleArrowFunction(fun *ir.ArrowFunctionExpr) bool {
	if b.ignoreFunctionBodies {
		return false
	}

	return b.enterArrowFunction(fun)
}

func (b *blockWalker) handleReturn(ret *ir.ReturnStmt) {
	if ret.Expr == nil {
		// Return without explicit return value.
		b.bareReturn = true
		return
	}
	b.returnsValue = true

	typ := solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, ret.Expr, b.ctx.customTypes)
	b.returnTypes = b.returnTypes.Append(typ)
}

func (b *blockWalker) handleLogicalOr(or *ir.LogicalOrExpr) bool {
	or.Left.Walk(b)

	// We're going to discard "or" RHS effects on the exit flags.
	exitFlags := b.ctx.exitFlags
	or.Right.Walk(b)
	b.ctx.exitFlags = exitFlags

	return false
}

func (b *blockWalker) addNonLocalVarName(nm string, kind variableKind) {
	b.nonLocalVars[nm] = kind
}

func (b *blockWalker) addNonLocalVar(v ir.Node, kind variableKind) {
	sv, ok := v.(*ir.SimpleVar)
	if !ok {
		return
	}
	b.addNonLocalVarName(sv.Name, kind)
}

// replaceVar must be used to track assignments to conrete var nodes if they are available
func (b *blockWalker) replaceVar(v ir.Node, typ meta.TypesMap, reason string, flags meta.VarFlags) {
	b.ctx.sc.ReplaceVar(v, typ, reason, flags)
	sv, ok := v.(*ir.SimpleVar)
	if !ok {
		return
	}

	b.trackVarName(v, sv.Name)
}

func (b *blockWalker) trackVarName(n ir.Node, nm string) {
	// Writes to non-local variables do count as usages
	if _, ok := b.nonLocalVars[nm]; ok {
		b.untrackVarName(nm)
		return
	}

	// Writes to variables that are done in a loop should not count as unused variables
	// because they can be read on the next iteration (ideally we should check for that too :))
	if !b.ctx.insideLoop {
		b.unusedVars[nm] = append(b.unusedVars[nm], n)
	}
}

func (b *blockWalker) untrackVarName(nm string) {
	delete(b.unusedVars, nm)
	delete(b.unusedParams, nm)
}

func (b *blockWalker) addVarName(n ir.Node, nm string, typ meta.TypesMap, reason string, flags meta.VarFlags) {
	b.ctx.sc.AddVarName(nm, typ, reason, flags)
	b.trackVarName(n, nm)
}

// addVar must be used to track assignments to conrete var nodes if they are available
func (b *blockWalker) addVar(v ir.Node, typ meta.TypesMap, reason string, flags meta.VarFlags) {
	b.ctx.sc.AddVar(v, typ, reason, flags)
	sv, ok := v.(*ir.SimpleVar)
	if !ok {
		return
	}
	b.trackVarName(v, sv.Name)
}

func (b *blockWalker) handleUnset(s *ir.UnsetStmt) bool {
	for _, v := range s.Vars {
		switch v := v.(type) {
		case *ir.SimpleVar:
			b.untrackVarName(v.Name)
			b.ctx.sc.DelVar(v, "unset")
		case *ir.Var:
			b.ctx.sc.DelVar(v, "unset")
		case *ir.ArrayDimFetchExpr:
			b.handleIssetDimFetch(v) // unset($a["something"]) does not unset $a itself, so no delVar here
		default:
			if v != nil {
				v.Walk(b)
			}
		}
	}

	return false
}

func (b *blockWalker) handleIsset(s *ir.IssetExpr) bool {
	for _, v := range s.Variables {
		switch v := v.(type) {
		case *ir.Var:
			// Do nothing.
		case *ir.SimpleVar:
			b.untrackVarName(v.Name)
		case *ir.ArrayDimFetchExpr:
			b.handleIssetDimFetch(v)
		default:
			if v != nil {
				v.Walk(b)
			}
		}
	}

	return false
}

func (b *blockWalker) handleEmpty(s *ir.EmptyExpr) bool {
	switch v := s.Expr.(type) {
	case *ir.Var:
		// Do nothing.
	case *ir.SimpleVar:
		b.untrackVarName(v.Name)
	case *ir.ArrayDimFetchExpr:
		b.handleIssetDimFetch(v)
	default:
		if v != nil {
			v.Walk(b)
		}
	}

	return false
}

// withNewContext runs a given function inside a new context.
// Upon function return, previous context is restored.
//
// While inside the callback (action), b.ctx is a new context.
//
// Returns the context that was assigned during callback execution (the new context),
// so it can be examined at the call site.
func (b *blockWalker) withNewContext(action func()) *blockContext {
	oldCtx := b.ctx
	newCtx := copyBlockContext(b.ctx)

	b.ctx = newCtx
	action()
	b.ctx = oldCtx

	return newCtx
}

func (b *blockWalker) handleTry(s *ir.TryStmt) bool {
	var linksCount int
	var finallyCtx *blockContext

	contexts := make([]*blockContext, 0, len(s.Catches)+1)

	// Assume that no code in try{} block has executed because exceptions can be thrown from anywhere.
	// So we handle catches and finally blocks first.
	for i := range s.Catches {
		c := s.Catches[i]
		ctx := b.withNewContext(func() {
			cc := c.(*ir.CatchStmt)
			for _, s := range cc.Stmts {
				b.addStatement(s)
			}
			cc.Walk(b)
		})
		contexts = append(contexts, ctx)

		if ctx.exitFlags == 0 {
			linksCount++
		}
	}

	if s.Finally != nil {
		finallyCtx = b.withNewContext(func() {
			cc := s.Finally.(*ir.FinallyStmt)
			for _, s := range cc.Stmts {
				b.addStatement(s)
			}
			s.Finally.Walk(b)
		})
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

		b.ctx.containsExitFlags |= ctx.containsExitFlags
	}

	ctx := b.withNewContext(func() {
		for _, s := range s.Stmts {
			b.addStatement(s)
			s.Walk(b)
		}
	})
	if ctx.exitFlags == 0 {
		linksCount++
	}

	contexts = append(contexts, ctx)

	varTypes := make(map[string]meta.TypesMap, b.ctx.sc.Len())
	defCounts := make(map[string]int, b.ctx.sc.Len())

	for _, ctx := range contexts {
		if ctx.exitFlags != 0 {
			continue
		}

		ctx.sc.Iterate(func(nm string, typ meta.TypesMap, flags meta.VarFlags) {
			varTypes[nm] = varTypes[nm].Append(typ)
			if flags.IsAlwaysDefined() {
				defCounts[nm]++
			}
		})
	}

	for nm, types := range varTypes {
		var flags meta.VarFlags
		flags.SetAlwaysDefined(defCounts[nm] == linksCount)
		b.ctx.sc.AddVarName(nm, types, "all branches try catch", flags)
	}

	if finallyCtx != nil {
		finallyCtx.sc.Iterate(func(nm string, typ meta.TypesMap, flags meta.VarFlags) {
			flags.SetAlwaysDefined(finallyCtx.exitFlags == 0)
			b.ctx.sc.AddVarName(nm, typ, "finally", flags)
		})
	}

	if othersExit && ctx.exitFlags != 0 {
		b.ctx.exitFlags |= prematureExitFlags
		b.ctx.exitFlags |= ctx.exitFlags
	}

	b.ctx.containsExitFlags |= ctx.containsExitFlags

	return false
}

func (b *blockWalker) handleCatch(s *ir.CatchStmt) {
	types := make([]meta.Type, 0, len(s.Types))
	for _, t := range s.Types {
		typ, ok := solver.GetClassName(b.r.ctx.st, t)
		if !ok {
			continue
		}
		types = append(types, meta.Type{Elem: typ})
	}
	m := meta.NewTypesMapFromTypes(types)

	b.handleVariableNode(s.Variable, m, "catch")

	for _, stmt := range s.Stmts {
		if stmt != nil {
			b.addStatement(stmt)
			stmt.Walk(b)
		}
	}
}

// We still need to analyze expressions in isset()/unset()/empty() statements
func (b *blockWalker) handleIssetDimFetch(e *ir.ArrayDimFetchExpr) {
	b.checkArrayDimFetch(e)

	switch v := e.Variable.(type) {
	case *ir.SimpleVar:
		b.untrackVarName(v.Name)
	case *ir.ArrayDimFetchExpr:
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

func (b *blockWalker) handleCallArgs(args []ir.Node, fn meta.FuncInfo) {
	for i, arg := range args {
		if i >= len(fn.Params) {
			arg.Walk(b)
			continue
		}

		ref := fn.Params[i].IsRef

		switch a := arg.(*ir.Argument).Expr.(type) {
		case *ir.Var, *ir.SimpleVar:
			if ref {
				b.addNonLocalVar(a, varRef)
				// TODO: variable may actually not be set by ref
				b.addVar(a, fn.Params[i].Typ, "call_with_ref", meta.VarAlwaysDefined)
				break
			}
			a.Walk(b)
		case *ir.ArrayDimFetchExpr:
			if ref {
				b.handleAndCheckDimFetchLValue(a, "call_with_ref", meta.MixedType)
				break
			}
			a.Walk(b)
		case *ir.ClosureExpr:
			var typ meta.TypesMap
			isInstance := b.ctx.sc.IsInInstanceMethod()
			if isInstance {
				typ, _ = b.ctx.sc.GetVarNameType("this")
			}

			// find the types for the arguments of the function that contains this closure
			var funcArgTypes []meta.TypesMap
			for _, arg := range args {
				tp := solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, arg.(*ir.Argument).Expr)
				funcArgTypes = append(funcArgTypes, tp)
			}

			closureSolver := &solver.ClosureCallerInfo{
				Name:     fn.Name,
				ArgTypes: funcArgTypes,
			}

			b.enterClosure(a, isInstance, typ, closureSolver)
		default:
			a.Walk(b)
		}
	}
}

func (b *blockWalker) handleFunctionCall(e *ir.FunctionCallExpr) bool {
	call := resolveFunctionCall(b.ctx.sc, b.r.ctx.st, b.ctx.customTypes, e)

	e.Function.Walk(b)

	switch call.funcName {
	case `\func_get_args`:
		b.callsFuncGetArgs = true
	case `\compact`:
		b.handleCompactCallArgs(e.Args)
	default:
		b.handleCallArgs(e.Args, call.info)
	}

	b.ctx.exitFlags |= call.info.ExitFlags

	return false
}

// handleCompactCallArgs treats strings anywhere in the argument list as uses
// of the variables named by those strings, which is how compact() behaves.
func (b *blockWalker) handleCompactCallArgs(args []ir.Node) {
	// Recursively flatten the argument list and extract strings
	var strs []*ir.String
	for len(args) > 0 {
		var head ir.Node
		head, args = args[0], args[1:]
		switch n := head.(type) {
		case *ir.Argument:
			args = append(args, n.Expr)
		case *ir.ArrayExpr:
			for _, item := range n.Items {
				args = append(args, item)
			}
		case *ir.ArrayItemExpr:
			args = append(args, n.Val)
		case *ir.String:
			strs = append(strs, n)
		}
	}

	for _, s := range strs {
		v := &ir.SimpleVar{
			Name:     s.Value,
			Position: ir.GetPosition(s),
		}
		b.handleVariable(v)
	}
}

func (b *blockWalker) handleMethodCall(e *ir.MethodCallExpr) bool {
	call := resolveMethodCall(b.ctx.sc, b.r.ctx.st, b.ctx.customTypes, e)

	e.Variable.Walk(b)
	e.Method.Walk(b)

	if !call.isMagic {
		b.handleCallArgs(e.Args, call.info)
	}
	b.ctx.exitFlags |= call.info.ExitFlags

	return false
}

func (b *blockWalker) handleStaticCall(e *ir.StaticCallExpr) bool {
	call := resolveStaticMethodCall(b.r.ctx.st, e)
	if !b.callsParentConstructor {
		b.callsParentConstructor = call.isCallsParentConstructor
	}

	e.Class.Walk(b)
	e.Call.Walk(b)

	b.handleCallArgs(e.Args, call.methodInfo.Info)
	b.ctx.exitFlags |= call.methodInfo.Info.ExitFlags

	return false
}

func (b *blockWalker) isThisInsideClosure(varNode ir.Node) bool {
	if !b.ctx.sc.IsInClosure() {
		return false
	}

	variable, ok := varNode.(*ir.SimpleVar)
	if !ok {
		return false
	}
	return variable.Name == `this`
}

func (b *blockWalker) handlePropertyFetch(e *ir.PropertyFetchExpr) bool {
	e.Variable.Walk(b)
	e.Property.Walk(b)
	return false
}

func (b *blockWalker) handleStaticPropertyFetch(e *ir.StaticPropertyFetchExpr) bool {
	e.Class.Walk(b)

	if propertyVarNode, propertyIsVarNode := e.Property.(*ir.Var); propertyIsVarNode {
		propertyVarNode.Expr.Walk(b)
	}

	return false
}

func (b *blockWalker) handleArray(arr *ir.ArrayExpr) bool {
	return b.handleArrayItems(arr.Items)
}

func (b *blockWalker) handleArrayItems(items []*ir.ArrayItemExpr) bool {
	for _, item := range items {
		if item.Val != nil {
			item.Val.Walk(b)
		}
		if item.Key != nil {
			item.Key.Walk(b)
		}
	}

	return false
}

func (b *blockWalker) handleClassConstFetch(e *ir.ClassConstFetchExpr) bool {
	e.Class.Walk(b)
	return false
}

func (b *blockWalker) handleForeach(s *ir.ForeachStmt) bool {
	// TODO: add reference semantics to foreach analyze as well

	// expression is always executed and is executed in base context
	if s.Expr != nil {
		s.Expr.Walk(b)
	}

	// foreach body can do 0 cycles so we need a separate context for that
	if s.Stmt != nil {
		ctx := b.withNewContext(func() {
			solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, s.Expr, b.ctx.customTypes).Iterate(func(typ string) {
				b.handleVariableNode(s.Variable, meta.NewTypesMap(meta.WrapElemOf(typ)), "foreach_value")
			})

			b.handleVariableNode(s.Key, arrayKeyType, "foreach_key")
			if list, ok := s.Variable.(*ir.ListExpr); ok {
				for _, item := range list.Items {
					b.handleVariableNode(item.Val, meta.TypesMap{}, "foreach_value")
				}
			} else {
				b.handleVariableNode(s.Variable, meta.TypesMap{}, "foreach_value")
			}

			b.ctx.innermostLoop = loopFor
			b.ctx.insideLoop = true
			if _, ok := s.Stmt.(*ir.StmtList); !ok {
				b.addStatement(s.Stmt)
			}
			s.Stmt.Walk(b)
		})

		b.maybeAddAllVars(ctx.sc, "foreach body")
		b.propagateFlags(ctx)
	}

	key, ok := s.Key.(*ir.SimpleVar)
	if !ok {
		return false
	}
	if b.r.config.IsDiscardVar(key.Name) {
		return false
	}

	_, ok = b.unusedVars[key.Name]
	if ok {
		variable, ok := s.Variable.(*ir.SimpleVar)
		if !ok {
			return false
		}

		b.untrackVarName(key.Name)

		b.r.Report(s.Key, LevelError, "unused", "foreach key $%s is unused, can simplify $%s => $%s to just $%s", key.Name, key.Name, variable.Name, variable.Name)
	}

	return false
}

func (b *blockWalker) handleFor(s *ir.ForStmt) bool {
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
		ctx := b.withNewContext(func() {
			b.ctx.innermostLoop = loopFor
			b.ctx.insideLoop = true
			s.Stmt.Walk(b)
		})

		b.maybeAddAllVars(ctx.sc, "while body")
		b.propagateFlags(ctx)
	}

	return false
}

func (b *blockWalker) enterArrowFunction(fun *ir.ArrowFunctionExpr) bool {
	sc := meta.NewScope()

	doc := b.r.parsePHPDoc(fun, fun.PhpDoc, fun.Params)
	b.r.reportPhpdocErrors(fun, doc.errs)
	phpDocParamTypes := doc.types

	params, _ := b.r.parseFuncArgs(fun.Params, phpDocParamTypes, sc, nil)
	b.r.handleArrowFuncExpr(params, fun.Expr, sc, b)

	return false
}

func (b *blockWalker) enterClosure(fun *ir.ClosureExpr, haveThis bool, thisType meta.TypesMap, closureSolver *solver.ClosureCallerInfo) bool {
	sc := meta.NewScope()
	sc.SetInClosure(true)

	if haveThis {
		sc.AddVarName("this", thisType, "closure inside instance method", meta.VarAlwaysDefined)
	} else {
		sc.AddVarName("this", meta.NewTypesMap("possibly_late_bound"), "possibly late bound $this", meta.VarAlwaysDefined)
	}

	doc := b.r.parsePHPDoc(fun, fun.PhpDoc, fun.Params)
	b.r.reportPhpdocErrors(fun, doc.errs)
	phpDocParamTypes := doc.types

	var closureUses []ir.Node
	if fun.ClosureUse != nil {
		closureUses = fun.ClosureUse.Uses
	}
	for _, useExpr := range closureUses {
		var byRef bool
		var v *ir.SimpleVar
		switch u := useExpr.(type) {
		case *ir.ReferenceExpr:
			v = u.Variable.(*ir.SimpleVar)
			byRef = true
		case *ir.SimpleVar:
			v = u
		default:
			continue
		}

		if !b.ctx.sc.HaveVar(v) && !byRef {
			b.r.Report(v, LevelWarning, "undefined", "Undefined variable %s", v.Name)
		}

		typ, ok := b.ctx.sc.GetVarNameType(v.Name)
		if ok {
			sc.AddVarName(v.Name, typ, "use", meta.VarAlwaysDefined)
		}

		b.untrackVarName(v.Name)
	}

	params, _ := b.r.parseFuncArgs(fun.Params, phpDocParamTypes, sc, closureSolver)

	b.r.handleFuncStmts(params, closureUses, fun.Stmts, sc)

	return false
}

func (b *blockWalker) maybeAddAllVars(sc *meta.Scope, reason string) {
	sc.Iterate(func(varName string, typ meta.TypesMap, flags meta.VarFlags) {
		flags &^= meta.VarAlwaysDefined
		b.ctx.sc.AddVarName(varName, typ, reason, flags)
	})
}

func (b *blockWalker) handleWhile(s *ir.WhileStmt) bool {
	if s.Cond != nil {
		s.Cond.Walk(b)
	}

	// while body can do 0 cycles so we need a separate context for that
	if s.Stmt != nil {
		ctx := b.withNewContext(func() {
			b.ctx.innermostLoop = loopFor
			b.ctx.insideLoop = true
			s.Stmt.Walk(b)
		})
		b.maybeAddAllVars(ctx.sc, "while body")
		b.propagateFlags(ctx)
	}

	return false
}

func (b *blockWalker) handleDo(s *ir.DoStmt) bool {
	if s.Stmt != nil {
		oldInnermostLoop := b.ctx.innermostLoop
		oldInsideLoop := b.ctx.insideLoop
		b.ctx.innermostLoop = loopFor
		b.ctx.insideLoop = true
		s.Stmt.Walk(b)
		b.ctx.innermostLoop = oldInnermostLoop
		b.ctx.insideLoop = oldInsideLoop
	}

	if s.Cond != nil {
		s.Cond.Walk(b)
	}

	return false
}

// propagateFlags is like propagateFlagsFromBranches, but for a simple single block case.
func (b *blockWalker) propagateFlags(other *blockContext) {
	b.ctx.containsExitFlags |= other.containsExitFlags
}

// Propagate premature exit flags from visited branches ("contexts").
func (b *blockWalker) propagateFlagsFromBranches(contexts []*blockContext, linksCount int) {
	allExit := false
	prematureExitFlags := 0

	for _, ctx := range contexts {
		b.ctx.containsExitFlags |= ctx.containsExitFlags
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
		b.ctx.exitFlags |= prematureExitFlags
	}
}

func (b *blockWalker) handleVariable(v ir.Node) bool {
	var varName string
	switch v := v.(type) {
	case *ir.Var:
		if vv, ok := v.Expr.(*ir.SimpleVar); ok {
			varName = vv.Name
			if b.inArrowFunction {
				for _, w := range b.parentBlockWalkers {
					w.untrackVarName(varName)
				}
			}

			b.untrackVarName(varName)
		}
	case *ir.SimpleVar:
		varName = v.Name
		if b.inArrowFunction {
			for _, w := range b.parentBlockWalkers {
				w.untrackVarName(varName)
			}
		}
		if b.r.config.IsDiscardVar(varName) && !isSuperGlobal(varName) {
			b.r.Report(v, LevelError, "discardVar", "Used var $%s that is supposed to be unused (rename variable if it's intended)", varName)
		}

		b.untrackVarName(varName)
	}

	have := b.ctx.sc.HaveVar(v)

	if !have && !b.inArrowFunction {
		b.r.reportUndefinedVariable(v, b.ctx.sc.MaybeHaveVar(v))
		b.ctx.sc.AddVar(v, meta.NewTypesMap("undefined"), "undefined", meta.VarAlwaysDefined)
	}

	// In case the required variable was not found in the current scope,
	// we need to look at all scopes above up to the scope in which the given
	// arrow function is declared or, in case it is a nested arrow function,
	// to the scope, which contains the parent arrow function.
	// This is necessary because arrow functions implicitly capture variables.
	if b.inArrowFunction && !have {
		var varNotFound bool
		var varMaybeNotDefined bool

		for _, bw := range b.parentBlockWalkers {
			s := bw.ctx.sc

			if !s.HaveVar(v) {
				if !varMaybeNotDefined {
					varMaybeNotDefined = s.MaybeHaveVar(v)
				}
				varNotFound = true
				continue
			}

			tp, _ := s.GetVarNameType(varName)

			// If a variable was found in one of the scopes,
			// we must add it to the current scope, that is, capture the variable.
			// Thus, by changing this variable inside the arrow function,
			// we will not change the variable that was captured,
			// as it should be according to the specification
			// (www.php.net/manual/en/functions.arrow.php).
			b.ctx.sc.AddVar(v, tp, "from_parent_scope", meta.VarAlwaysDefined)
			return false
		}

		if varNotFound {
			b.r.reportUndefinedVariable(v, varMaybeNotDefined)
			b.ctx.sc.AddVar(v, meta.NewTypesMap("undefined"), "undefined", meta.VarAlwaysDefined)
		}
	}

	return false
}

func (b *blockWalker) handleTernary(e *ir.TernaryExpr) bool {
	if e.IfTrue == nil {
		return true // Skip `$x ?: $y` expressions
	}

	b.withNewContext(func() {
		// No need to delete vars here as we run andWalker
		// only inside a new context.
		a := &andWalker{b: b}
		e.Condition.Walk(a)
		e.IfTrue.Walk(b)
	})
	e.IfFalse.Walk(b)
	return false
}

func (b *blockWalker) handleIf(s *ir.IfStmt) bool {
	var varsToDelete []ir.Node
	customMethods := len(b.ctx.customMethods)
	customFunctions := len(b.ctx.customFunctions)
	// Remove all isset'ed variables after we're finished with this if statement.
	defer func() {
		for _, v := range varsToDelete {
			b.ctx.sc.DelVar(v, "isset/!empty")
		}
		b.ctx.customMethods = b.ctx.customMethods[:customMethods]
		b.ctx.customFunctions = b.ctx.customFunctions[:customFunctions]
	}()
	walkCond := func(cond ir.Node) {
		a := &andWalker{b: b}
		cond.Walk(a)
		varsToDelete = append(varsToDelete, a.varsToDelete...)
	}

	// first condition is always executed, so run it in base context
	if s.Cond != nil {
		walkCond(s.Cond)
	}

	var contexts []*blockContext

	walk := func(n ir.Node) (links int) {
		// handle if (...) smth(); else other_thing(); // without braces
		switch n := n.(type) {
		case *ir.ElseStmt:
			b.addStatement(n.Stmt)
		case *ir.ElseIfStmt:
			b.addStatement(n.Stmt)
		default:
			b.addStatement(n)
		}

		ctx := b.withNewContext(func() {
			if elsif, ok := n.(*ir.ElseIfStmt); ok {
				walkCond(elsif.Cond)
				b.handleElseIf(elsif)
				elsif.Stmt.Walk(b)
			} else {
				n.Walk(b)
			}
		})

		contexts = append(contexts, ctx)

		if ctx.exitFlags != 0 {
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

	varTypes := make(map[string]meta.TypesMap, b.ctx.sc.Len())
	defCounts := make(map[string]int, b.ctx.sc.Len())

	for _, ctx := range contexts {
		if ctx.exitFlags != 0 {
			continue
		}

		ctx.sc.Iterate(func(nm string, typ meta.TypesMap, flags meta.VarFlags) {
			varTypes[nm] = varTypes[nm].Append(typ)
			if flags.IsAlwaysDefined() {
				defCounts[nm]++
			}
		})
	}

	for nm, types := range varTypes {
		var flags meta.VarFlags
		flags.SetAlwaysDefined(defCounts[nm] == linksCount)
		b.ctx.sc.AddVarName(nm, types, "all branches", flags)
	}

	return false
}

func (b *blockWalker) handleElseIf(s *ir.ElseIfStmt) {
	b.r.checkKeywordCase(s, "elseif")
}

func (b *blockWalker) iterateNextCases(cases []ir.Node, startIdx int) {
	for i := startIdx; i < len(cases); i++ {
		cond, list := getCaseStmts(cases[i])
		if cond != nil {
			cond.Walk(b)
		}

		for _, stmt := range list {
			if stmt != nil {
				b.addStatement(stmt)
				stmt.Walk(b)
				if b.ctx.exitFlags != 0 {
					return
				}
			}
		}
	}
}

func (b *blockWalker) handleSwitch(s *ir.SwitchStmt) bool {
	// first condition is always executed, so run it in base context
	if s.Cond != nil {
		s.Cond.Walk(b)
	}

	var contexts []*blockContext

	linksCount := 0
	haveDefault := false
	breakFlags := FlagBreak | FlagContinue

	for i := range s.Cases {
		idx := i
		c := s.Cases[i]
		var list []ir.Node

		cond, list := getCaseStmts(c)
		if cond == nil {
			haveDefault = true
			b.r.checkKeywordCase(c, "default")
		} else {
			cond.Walk(b)
			b.r.checkKeywordCase(c, "case")
		}

		// allow empty case body without "break;"
		// nothing new can be defined here so we just skip it
		if len(list) == 0 {
			continue
		}

		ctx := b.withNewContext(func() {
			b.ctx.innermostLoop = loopSwitch
			for _, s := range list {
				if s != nil {
					b.addStatement(s)
					s.Walk(b)
				}
			}

			// allow to omit "break;" in the final statement
			if idx != len(s.Cases)-1 && b.ctx.exitFlags == 0 {
				// allow the fallthrough if appropriate comment is present
				nextCase := s.Cases[idx+1]
				if !caseHasFallthroughComment(nextCase) {
					b.r.Report(c, LevelWarning, "caseBreak", "Add break or '// fallthrough' to the end of the case")
				}
			}

			if (b.ctx.exitFlags & (^breakFlags)) == 0 {
				linksCount++

				if b.ctx.exitFlags == 0 {
					b.iterateNextCases(s.Cases, idx+1)
				}
			}
		})

		contexts = append(contexts, ctx)
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
			b.ctx.containsExitFlags |= ctx.containsExitFlags
		}
	}

	if allExit {
		b.ctx.exitFlags |= prematureExitFlags
	}

	varTypes := make(map[string]meta.TypesMap, b.ctx.sc.Len())
	defCounts := make(map[string]int, b.ctx.sc.Len())

	for _, ctx := range contexts {
		b.propagateFlags(ctx)

		cleanFlags := ctx.exitFlags & (^breakFlags)
		if cleanFlags != 0 {
			continue
		}

		ctx.sc.Iterate(func(nm string, typ meta.TypesMap, flags meta.VarFlags) {
			varTypes[nm] = varTypes[nm].Append(typ)
			if flags.IsAlwaysDefined() {
				defCounts[nm]++
			}
		})
	}

	for nm, types := range varTypes {
		var flags meta.VarFlags
		flags.SetAlwaysDefined(defCounts[nm] == linksCount)
		b.ctx.sc.AddVarName(nm, types, "all cases", flags)
	}

	return false
}

// if $a was previously undefined,
// handle case when doing assignment like '$a[] = 4;'
// or call to function that accepts like exec("command", $a)
func (b *blockWalker) handleAndCheckDimFetchLValue(e *ir.ArrayDimFetchExpr, reason string, typ meta.TypesMap) {
	b.checkArrayDimFetch(e)

	switch v := e.Variable.(type) {
	case *ir.Var, *ir.SimpleVar:
		arrTyp := typ.Map(meta.WrapArrayOf)
		b.addVar(v, arrTyp, reason, meta.VarAlwaysDefined)
		b.handleVariable(v)
	case *ir.ArrayDimFetchExpr:
		b.handleAndCheckDimFetchLValue(v, reason, meta.MixedType)
	default:
		// probably not assignable?
		v.Walk(b)
	}

	if e.Dim != nil {
		e.Dim.Walk(b)
	}
}

func (b *blockWalker) checkArrayDimFetch(s *ir.ArrayDimFetchExpr) {
	if !b.isIndexingComplete() {
		return
	}

	typ := solver.ExprType(b.ctx.sc, b.r.ctx.st, s.Variable)

	var (
		maybeHaveClasses bool
		haveArrayAccess  bool
	)

	typ.Iterate(func(t string) {
		// FullyQualified class name will have "\" in the beginning
		if meta.IsClassType(t) {
			maybeHaveClasses = true

			if !haveArrayAccess && solver.Implements(b.r.ctx.st.Info, t, `\ArrayAccess`) {
				haveArrayAccess = true
			}
		}
	})

	if maybeHaveClasses && !haveArrayAccess {
		b.r.Report(s.Variable, LevelNotice, "arrayAccess", "Array access to non-array type %s", typ)
	}
}

// some day, perhaps, there will be some difference between handleAssignReference and handleAssign
func (b *blockWalker) handleAssignReference(a *ir.AssignReference) bool {
	switch v := a.Variable.(type) {
	case *ir.ArrayDimFetchExpr:
		b.handleAndCheckDimFetchLValue(v, "assign_array", meta.MixedType)
		a.Expression.Walk(b)
		return false
	case *ir.Var, *ir.SimpleVar:
		b.addVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expression), "assign", meta.VarAlwaysDefined)
		b.addNonLocalVar(v, varRef)
	case *ir.ListExpr:
		// TODO: figure out whether this case is reachable.
		for _, item := range v.Items {
			b.handleVariableNode(item.Val, meta.NewTypesMap("unknown_from_list"), "assign")
		}
	default:
		a.Variable.Walk(b)
	}

	a.Expression.Walk(b)
	return false
}

func (b *blockWalker) handleAssignShapeToList(items []*ir.ArrayItemExpr, info meta.ClassInfo) {
	for i, item := range items {
		var prop meta.PropertyInfo
		var ok bool

		if item.Key != nil {
			var key string
			switch keyNode := item.Key.(type) {
			case *ir.String:
				key = keyNode.Value
			case *ir.Lnumber:
				key = keyNode.Value
			case *ir.Dnumber:
				key = keyNode.Value
			}

			if key != "" {
				prop, ok = info.Properties[key]
			}
		} else {
			prop, ok = info.Properties[fmt.Sprint(i)]
		}

		var tp meta.TypesMap
		if !ok {
			tp = meta.NewTypesMap("unknown_from_list")
		} else {
			tp = prop.Typ
		}
		b.handleVariableNode(item.Val, tp, "list-assign")
	}
}

func (b *blockWalker) handleAssignList(list *ir.ListExpr, rhs ir.Node) {
	typ := solver.ExprType(b.ctx.sc, b.r.ctx.st, rhs)

	// TODO: test if we can prealloc elemTypes to const size hint like 2
	// and get stack allocation which will help to avoid the unwanted heap allocs.
	// Hint: only const (literal) size hints work for this.
	// Hint: check the compiler output to see whether elemTypes "escape" or not.
	//
	// We store meta.Type instead of string to avoid the need to do strings.Join
	// when we want to create a TypesMap.
	var elemTypes []meta.Type
	var shapeType string
	typ.Iterate(func(typ string) {
		switch {
		case meta.IsShapeType(typ):
			shapeType = typ
		case meta.IsArrayType(typ):
			elemType := strings.TrimSuffix(typ, "[]")
			elemTypes = append(elemTypes, meta.Type{Elem: elemType})
		}
	})

	// Try to handle it as a shape assignment.
	if shapeType != "" {
		class, ok := b.r.ctx.st.Info.GetClass(shapeType)
		if ok {
			b.handleAssignShapeToList(list.Items, class)
			return
		}
	}

	// Try to handle it as an array assignment.
	if len(elemTypes) != 0 {
		elemTypeMap := meta.NewTypesMapFromTypes(elemTypes).Immutable()
		for _, item := range list.Items {
			b.handleVariableNode(item.Val, elemTypeMap, "list-assign")
		}
		return
	}

	// Fallback: define vars with unknown types.
	//
	// TODO: shouldn't it be a mixed type? I would prefer a "mixed" type here
	// and "unknown from list" reason.
	for _, item := range list.Items {
		b.handleVariableNode(item.Val, meta.NewTypesMap("unknown_from_list"), "list-assign")
	}
}

func (b *blockWalker) paramClobberCheck(v *ir.SimpleVar) {
	if b.callsFuncGetArgs {
		return
	}
	if _, ok := b.unusedParams[v.Name]; ok && !b.path.Conditional() {
		b.r.Report(v, LevelWarning, "paramClobber", "$%s param re-assigned before being used", v.Name)
	}
}

func (b *blockWalker) handleAssign(a *ir.Assign) bool {
	a.Expression.Walk(b)

	switch v := a.Variable.(type) {
	case *ir.ArrayDimFetchExpr:
		typ := solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expression)
		b.handleAndCheckDimFetchLValue(v, "assign_array", typ)
		return false
	case *ir.SimpleVar:
		b.handleComments(v)
		b.paramClobberCheck(v)
		b.replaceVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expression), "assign", meta.VarAlwaysDefined)
	case *ir.Var:
		b.replaceVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expression), "assign", meta.VarAlwaysDefined)
	case *ir.ListExpr:
		if !b.isIndexingComplete() {
			return true
		}

		b.handleAssignList(v, a.Expression)
	case *ir.PropertyFetchExpr:
		v.Property.Walk(b)
		sv, ok := v.Variable.(*ir.SimpleVar)
		if !ok {
			v.Variable.Walk(b)
			break
		}

		b.untrackVarName(sv.Name)

		if sv.Name != "this" {
			break
		}

		if b.r.ctx.st.CurrentClass == "" {
			break
		}

		propertyName, ok := v.Property.(*ir.Identifier)
		if !ok {
			break
		}

		cls := b.r.getClass()

		p := cls.Properties[propertyName.Value]
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, a.Expression, b.ctx.customTypes))
		cls.Properties[propertyName.Value] = p
	case *ir.StaticPropertyFetchExpr:
		sv, ok := v.Property.(*ir.SimpleVar)
		if !ok {
			vv := v.Property.(*ir.Var)
			vv.Expr.Walk(b)
			break
		}

		if b.r.ctx.st.CurrentClass == "" {
			break
		}

		className, ok := solver.GetClassName(b.r.ctx.st, v.Class)
		if !ok || className != b.r.ctx.st.CurrentClass {
			break
		}

		cls := b.r.getClass()

		p := cls.Properties["$"+sv.Name]
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, a.Expression, b.ctx.customTypes))
		cls.Properties["$"+sv.Name] = p
	default:
		a.Variable.Walk(b)
	}

	return false
}

func (b *blockWalker) handleAssignOp(assign ir.Node) {
	var typ meta.TypesMap
	var v ir.Node

	switch assign := assign.(type) {
	case *ir.AssignPlus:
		e := &ir.PlusExpr{
			Left:  assign.Variable,
			Right: assign.Expression,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)
	case *ir.AssignMinus:
		e := &ir.MinusExpr{
			Left:  assign.Variable,
			Right: assign.Expression,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)
	case *ir.AssignMul:
		e := &ir.MulExpr{
			Left:  assign.Variable,
			Right: assign.Expression,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)
	case *ir.AssignDiv:
		e := &ir.DivExpr{
			Left:  assign.Variable,
			Right: assign.Expression,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)

	case *ir.AssignConcat:
		v = assign.Variable
		typ = meta.PreciseStringType
	case *ir.AssignShiftLeft:
		v = assign.Variable
		typ = meta.PreciseIntType
	case *ir.AssignShiftRight:
		v = assign.Variable
		typ = meta.PreciseIntType
	case *ir.AssignCoalesce:
		e := &ir.CoalesceExpr{
			Left:  assign.Variable,
			Right: assign.Expression,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)

	default:
		return
	}

	b.replaceVar(v, typ, "assign", meta.VarAlwaysDefined)
}

func (b *blockWalker) flushUnused() {
	if !b.isIndexingComplete() {
		return
	}

	visitedMap := make(map[ir.Node]struct{})
	for name, nodes := range b.unusedVars {
		if b.r.config.IsDiscardVar(name) {
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
			b.r.Report(n, LevelWarning, "unused", `Variable %s is unused (use $_ to ignore this inspection)`, name)
		}
	}
}

func (b *blockWalker) handleVariableNode(n ir.Node, typ meta.TypesMap, what string) {
	if n == nil {
		return
	}

	var vv ir.Node
	switch n := n.(type) {
	case *ir.Var, *ir.SimpleVar:
		vv = n
	case *ir.ReferenceExpr:
		vv = n.Variable
	default:
		return
	}

	b.addVar(vv, typ, what, meta.VarAlwaysDefined)
}

// LeaveNode is called after all children have been visited.
func (b *blockWalker) LeaveNode(w ir.Node) {
	for _, c := range b.custom {
		c.BeforeLeaveNode(w)
	}

	b.path.Pop()

	if b.ctx.exitFlags == 0 {
		b.updateExitFlags(w)
	}

	for _, c := range b.custom {
		c.AfterLeaveNode(w)
	}
}

func (b *blockWalker) updateExitFlags(n ir.Node) {
	switch n := n.(type) {
	case *ir.ReturnStmt:
		b.ctx.exitFlags |= FlagReturn
		b.ctx.containsExitFlags |= FlagReturn
	case *ir.ExitExpr:
		b.ctx.exitFlags |= FlagDie
		b.ctx.containsExitFlags |= FlagDie
	case *ir.ThrowStmt:
		b.ctx.exitFlags |= FlagThrow
		b.ctx.containsExitFlags |= FlagThrow
	case *ir.ContinueStmt:
		b.ctx.exitFlags |= FlagContinue
		b.ctx.containsExitFlags |= FlagContinue
	case *ir.BreakStmt:
		b.ctx.exitFlags |= FlagBreak
		b.ctx.containsExitFlags |= FlagBreak
	case *ir.ExpressionStmt:
		b.updateExitFlags(n.Expr)
	case *ir.FunctionCallExpr:
		if b.r.config.IgnoreTriggerError {
			return
		}
		nm, ok := n.Function.(*ir.Name)
		if !ok {
			return
		}
		// We can't use solver.GetFuncName here as PHP function names
		// lookup requires full symbol table information => we can't use
		// it during the indexing.
		funcName := strings.TrimPrefix(nm.Value, `\`)
		if (funcName != `trigger_error` && funcName != `user_error`) || len(n.Args) != 2 {
			return
		}
		errorLevel, ok := n.Arg(1).Expr.(*ir.ConstFetchExpr)
		// TODO: add meta.GetConstName() func and use it here.
		if ok && errorLevel.Constant.Value == `E_USER_ERROR` {
			b.ctx.exitFlags |= FlagDie
		}
	}
}

func (b *blockWalker) sideEffectFree(n ir.Node) bool {
	return solver.SideEffectFree(b.ctx.sc, b.r.ctx.st, b.ctx.customTypes, n)
}

func (b *blockWalker) exprType(n ir.Node) meta.TypesMap {
	return solver.ExprTypeCustom(b.ctx.sc, b.r.ctx.st, n, b.ctx.customTypes)
}
