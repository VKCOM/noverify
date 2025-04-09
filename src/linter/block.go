package linter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/VKCOM/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/phpdoctypes"
	"github.com/VKCOM/noverify/src/utils"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/linter/autogen"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
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
var arrayKeyType = types.NewMap("int|string").Immutable()

// blockWalker is used to process function/method contents.
type blockWalker struct {
	ctx *blockContext

	linter blockLinter

	// inferred return types if any
	returnTypes types.Map

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
	b.linter = blockLinter{walker: b, quickfix: r.checker.quickfix}
	return b
}

func (b *blockWalker) report(n ir.Node, level int, checkName, msg string, args ...interface{}) {
	if b.isSuppressed(n, checkName) {
		return
	}

	b.r.Report(n, level, checkName, msg, args...)
}

func (b *blockWalker) isSuppressed(n ir.Node, checkName string) bool {
	if containLinterSuppress(n, checkName) {
		return true
	}

	// We go up the tree in search of a comment that disables this checker.
	for i := 0; b.path.NthParent(i) != nil; i++ {
		parent := b.path.NthParent(i)
		if containLinterSuppress(parent, checkName) {
			return true
		}
	}

	return false
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

	if b.containsDisableInspection(n, "PhpUnreachableStatementInspection") {
		b.ctx.deadCodeReported = true
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
	b.report(n, LevelWarning, "deadCode", "Unreachable code")
}

func containLinterSuppress(n ir.Node, needInspection string) bool {
	if n == nil {
		return false
	}

	phpdocTypeParser := phpdoc.NewTypeParser()
	firstTkn := ir.GetFirstToken(n)
	if firstTkn == nil {
		return false
	}

	for _, tkn := range firstTkn.FreeFloating {
		if !phpdoc.IsPHPDocToken(tkn) {
			continue
		}

		if !bytes.Contains(tkn.Value, []byte("@noverify-suppress")) {
			continue
		}

		parsed := phpdoc.Parse(phpdocTypeParser, string(tkn.Value))
		for _, p := range parsed.Parsed {
			part, ok := p.(*phpdoc.RawCommentPart)
			if !ok {
				continue
			}

			if part.Name() == "noverify-suppress" {
				inspection := part.Params[0]

				if inspection == "all" || inspection == needInspection {
					return true
				}
			}
		}
	}

	return false
}

func (b *blockWalker) containsDisableInspection(n ir.Node, needInspection string) bool {
	firstTkn := ir.GetFirstToken(n)
	if firstTkn == nil {
		return false
	}

	for _, tkn := range firstTkn.FreeFloating {
		if !phpdoc.IsPHPDocToken(tkn) {
			continue
		}

		if !bytes.Contains(tkn.Value, []byte("@noinspection")) {
			continue
		}

		parsed := phpdoc.Parse(b.r.ctx.phpdocTypeParser, string(tkn.Value))
		for _, p := range parsed.Parsed {
			part, ok := p.(*phpdoc.RawCommentPart)
			if !ok {
				continue
			}

			if part.Name() == "noinspection" {
				inspection := part.Params[0]

				if inspection == needInspection {
					return true
				}
			}
		}
	}

	return false
}

func (b *blockWalker) handleComments(n ir.Node) {
	switch node := n.(type) {
	case *ir.ArrayDimFetchExpr:
		n = node.Variable
	default:
		n = node
	}

	n.IterateTokens(func(t *token.Token) bool {
		b.handleCommentToken(n, t)
		return true
	})
}

func (b *blockWalker) handleCommentToken(n ir.Node, t *token.Token) {
	if !phpdoc.IsPHPDocToken(t) {
		return
	}

	doc := phpdoc.Parse(b.r.ctx.phpdocTypeParser, string(t.Value))

	if phpdoc.IsSuspicious(t.Value) {
		b.r.ReportPHPDoc(PHPDocLine(n, 1),
			LevelWarning, "invalidDocblock",
			"Multiline PHPDoc comment should start with /**, not /*",
		)
	}

	for _, p := range doc.Parsed {
		part, ok := p.(*phpdoc.TypeVarCommentPart)
		if !ok || p.Name() != "var" {
			continue
		}

		converted := phpdoctypes.ToRealType(b.r.ctx.typeNormalizer.ClassFQNProvider(), b.r.config.KPHP, part.Type)
		moveShapesToContext(&b.r.ctx, converted.Shapes)
		b.r.handleClosuresFromDoc(converted.Closures)

		if converted.Warning != "" {
			b.r.ReportPHPDoc(
				PHPDocLineField(n, part.Line(), 1),
				LevelNotice, "invalidDocblockType", converted.Warning,
			)
		}

		simpleVar, isSimpleVar := n.(*ir.SimpleVar)
		if isSimpleVar && part.Var == "" {
			part.Var = simpleVar.Name
		}

		typesMap := types.NewMapWithNormalization(b.r.ctx.typeNormalizer, converted.Types)
		b.ctx.sc.AddVarFromPHPDoc(strings.TrimPrefix(part.Var, "$"), typesMap, "@var")

		b.r.checker.checkUndefinedClassesInPHPDoc(n, typesMap, part)
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
		s.Walk(b.r)
		res = false
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
		var typ types.Map
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
			b.report(v, LevelWarning, "dupGlobal", "Global statement mentions $%s more than once", nm)
		} else {
			vars[nm] = struct{}{}
			if b.nonLocalVars[nm] == varGlobal {
				b.report(v, LevelNotice, "dupGlobal", "$%s already global'ed above", nm)
			}
		}
	}
}

func (b *blockWalker) handleAndCheckGlobalStmt(s *ir.GlobalStmt) {
	if !b.rootLevel {
		b.checkDupGlobal(s)
	}

	for _, v := range s.Vars {
		nm := utils.VarToString(v)
		if nm == "" {
			continue
		}

		b.addVar(v, types.NewMap(types.WrapGlobal(nm)), "global", meta.VarAlwaysDefined)
		if b.path.Conditional() {
			b.addNonLocalVar(v, varCondGlobal)
		} else {
			b.addNonLocalVar(v, varGlobal)
		}
	}
}

func (b *blockWalker) CheckParamNullability(params []ir.Node) {
	for _, param := range params {
		if p, ok := param.(*ir.Parameter); ok {
			var paramType ir.Node
			paramType, paramOk := p.VariableType.(*ir.Name)
			if !paramOk {
				paramIdentifier, paramIdentifierOk := p.VariableType.(*ir.Identifier)
				if !paramIdentifierOk {
					continue
				}
				paramType = paramIdentifier
			}

			paramName, ok := paramType.(*ir.Name)
			if ok {
				if paramName.Value == "mixed" {
					continue
				}
			}

			defValue, defValueOk := p.DefaultValue.(*ir.ConstFetchExpr)
			if !defValueOk {
				continue
			}

			if defValue.Constant.Value != "null" {
				continue
			}

			b.linter.report(paramType, LevelWarning, "notExplicitNullableParam", "parameter with null default value should be explicitly nullable")
			b.r.addQuickFix("notExplicitNullableParam", b.linter.quickfix.notExplicitNullableParam(paramType))
		}
	}
}

func (b *blockWalker) handleFunction(fun *ir.FunctionStmt) bool {
	if b.ignoreFunctionBodies {
		b.CheckParamNullability(fun.Params)
		return false
	}

	if b.r.metaInfo().IsIndexingComplete() {
		return b.r.checker.CheckFunction(fun)
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
func (b *blockWalker) replaceVar(v ir.Node, typ types.Map, reason string, flags meta.VarFlags) {
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

func (b *blockWalker) untrackVarNameImpl(nm string) {
	delete(b.unusedVars, nm)
	delete(b.unusedParams, nm)
}

func (b *blockWalker) untrackVarName(nm string) {
	if b.inArrowFunction {
		b.untrackVarNameImpl(nm)
		for _, w := range b.parentBlockWalkers {
			w.untrackVarNameImpl(nm)
		}
		return
	}
	b.untrackVarNameImpl(nm)
}

func (b *blockWalker) addVarName(n ir.Node, nm string, typ types.Map, reason string, flags meta.VarFlags) {
	b.ctx.sc.AddVarName(nm, typ, reason, flags)
	b.trackVarName(n, nm)
}

// addVar must be used to track assignments to conrete var nodes if they are available
func (b *blockWalker) addVar(v ir.Node, typ types.Map, reason string, flags meta.VarFlags) {
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

func (b *blockWalker) withSpecificContext(ctx *blockContext, action func()) {
	oldCtx := b.ctx
	newCtx := ctx

	b.ctx = newCtx
	action()
	b.ctx = oldCtx
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

	tryCtx := b.withNewContext(func() {
		for _, s := range s.Stmts {
			b.addStatement(s)
			s.Walk(b)
		}
	})
	if tryCtx.exitFlags == 0 {
		linksCount++
	}

	b.checkUnreachableForFinallyReturn(s, tryCtx, finallyCtx, contexts)

	contexts = append(contexts, tryCtx)

	varTypes := make(map[string]types.Map, b.ctx.sc.Len())
	defCounts := make(map[string]int, b.ctx.sc.Len())

	for _, ctx := range contexts {
		if ctx.exitFlags != 0 {
			continue
		}

		ctx.sc.Iterate(func(nm string, typ types.Map, flags meta.VarFlags) {
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
		finallyCtx.sc.Iterate(func(nm string, typ types.Map, flags meta.VarFlags) {
			flags.SetAlwaysDefined(finallyCtx.exitFlags == 0)
			b.ctx.sc.AddVarName(nm, typ, "finally", flags)
		})
	}

	if othersExit && tryCtx.exitFlags != 0 {
		b.ctx.exitFlags |= prematureExitFlags
		b.ctx.exitFlags |= tryCtx.exitFlags
	}

	b.ctx.containsExitFlags |= tryCtx.containsExitFlags

	return false
}

func (b *blockWalker) checkUnreachableForFinallyReturn(tryStmt *ir.TryStmt, tryCtx *blockContext, finallyCtx *blockContext, catchContexts []*blockContext) {
	if finallyCtx == nil {
		return
	}

	var exitPoints []exitPoint

	// If try block contains some other return/die statements.
	containsOtherNoThrowExitPoints := tryCtx.containsExitFlags&FlagReturn != 0 ||
		tryCtx.containsExitFlags&FlagDie != 0

	if containsOtherNoThrowExitPoints {
		exitFlagsWithoutThrow := tryCtx.containsExitFlags ^ FlagThrow
		points := b.findExitPointsByFlags(&ir.StmtList{Stmts: tryStmt.Stmts}, exitFlagsWithoutThrow)
		exitPoints = append(exitPoints, points...)
	}

	var catchContainsDie bool
	var catchWithDieIndex int

	for i, context := range catchContexts {
		containsDie := context.exitFlags&FlagDie != 0
		exitFlagsWithoutDie := context.exitFlags ^ FlagDie
		containsOtherNoDieExitPoints := exitFlagsWithoutDie != 0

		if containsOtherNoDieExitPoints {
			points := b.findExitPointsByFlags(tryStmt.Catches[i], exitFlagsWithoutDie)
			exitPoints = append(exitPoints, points...)
		}

		if containsDie {
			catchContainsDie = true
			catchWithDieIndex = i + 1
		}
	}

	if catchContainsDie {
		b.report(tryStmt.Finally, LevelError, "deadCode", "Block finally is unreachable (because catch block %d contains a exit/die)", catchWithDieIndex)

		// If there is an error when the finally block is unreachable,
		// then errors due to return in finally are skipped.
		return
	}

	if finallyCtx.exitFlags == FlagReturn {
		var finallyReturnPos position.Position
		finallyReturns := b.findExitPointsByFlags(tryStmt.Finally, FlagReturn)
		if len(finallyReturns) > 0 {
			finallyReturnPos = *ir.GetPosition(finallyReturns[0].n)
		}

		for _, point := range exitPoints {
			b.report(point.n, LevelError, "deadCode", "%s is unreachable (because finally block contains a return on line %d)", point.kind, finallyReturnPos.StartLine)
		}
	}
}

type exitPoint struct {
	n    ir.Node
	kind string
}

func (b *blockWalker) findExitPointsByFlags(where ir.Node, exitFlags int) (points []exitPoint) {
	irutil.Inspect(where, func(n ir.Node) bool {
		if exitFlags&FlagReturn != 0 {
			ret, ok := n.(*ir.ReturnStmt)
			if ok {
				points = append(points, exitPoint{
					n:    ret,
					kind: "return",
				})
			}
		}

		if exitFlags&FlagThrow != 0 {
			thr, ok := n.(*ir.ThrowStmt)
			if ok {
				points = append(points, exitPoint{
					n:    thr,
					kind: "throw",
				})
			}
		}

		if exitFlags&FlagDie != 0 {
			exit, ok := n.(*ir.ExitExpr)
			if ok {
				typ := "exit"
				if exit.Die {
					typ = "die"
				}

				points = append(points, exitPoint{
					n:    exit,
					kind: typ,
				})
			}
		}
		return true
	})

	return points
}

func (b *blockWalker) handleCatch(s *ir.CatchStmt) {
	typeList := make([]types.Type, 0, len(s.Types))
	for _, t := range s.Types {
		typ, ok := solver.GetClassName(b.r.ctx.st, t)
		if !ok {
			continue
		}
		typeList = append(typeList, types.Type{Elem: typ})
	}
	m := types.NewMapFromTypes(typeList)

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

func nullSafetyRealParamForCheck(fn meta.FuncInfo, paramIndex int, haveVariadic bool) meta.FuncParam {
	if haveVariadic && paramIndex >= len(fn.Params)-1 {
		return fn.Params[len(fn.Params)-1]
	}
	return fn.Params[paramIndex]
}

func formatSlashesFuncName(fn meta.FuncInfo) string {
	return strings.TrimPrefix(fn.Name, "\\")
}

func (b *blockWalker) checkNotSafetyCallArgsF(args []ir.Node, fn meta.FuncInfo) {
	if fn.Params == nil || fn.Name == "" {
		return
	}
	haveVariadic := fn.Flags&meta.FuncVariadic != 0

	for i, arg := range args {
		if arg == nil {
			continue
		}

		// If there are more arguments than declared and function is not variadic, ignore extra arguments.
		if !haveVariadic && i > len(fn.Params)-1 {
			return
		}

		switch a := arg.(*ir.Argument).Expr.(type) {
		case *ir.SimpleVar:
			b.checkSimpleVarSafety(arg, fn, i, a, haveVariadic)
		case *ir.ConstFetchExpr:
			b.checkConstFetchSafety(arg, fn, i, a, haveVariadic)
		case *ir.ArrayDimFetchExpr:
			b.checkArrayDimFetchSafety(arg, fn, i, a, haveVariadic)
		case *ir.ListExpr:
			b.checkListExprSafety(arg, fn, i, a, haveVariadic)
		case *ir.PropertyFetchExpr:
			b.checkUnifiedPropertyFetchNotSafety(a, fn, i, haveVariadic)
		case *ir.StaticCallExpr:
			b.checkStaticCallSafety(arg, fn, i, a, haveVariadic)
		case *ir.StaticPropertyFetchExpr:
			b.checkUnifiedPropertyFetchNotSafety(a, fn, i, haveVariadic)
		case *ir.FunctionCallExpr:
			b.checkFunctionCallSafety(arg, fn, i, a, haveVariadic)
		}
	}
}

func (b *blockWalker) checkFunctionCallSafety(arg ir.Node, fn meta.FuncInfo, paramIndex int, funcCall *ir.FunctionCallExpr, haveVariadic bool) {
	var funcName string

	var isClearF bool
	var callType types.Map

	switch f := funcCall.Function.(type) {
	case *ir.Name:
		funcName = f.Value
		isClearF = true
	case *ir.SimpleVar:
		funcName = f.Name
		varInfo, found := b.ctx.sc.GetVar(f)

		if !found {
			return
		}
		callType = varInfo.Type // nolint:ineffassign,staticcheck
	default:
		return
	}

	funcInfo, ok := b.linter.metaInfo().GetFunction("\\" + funcName)
	if !ok && isClearF {
		return
	} else {
		callType = funcInfo.Typ
	}

	param := nullSafetyRealParamForCheck(fn, paramIndex, haveVariadic)
	paramType := param.Typ
	if haveVariadic && paramIndex >= len(fn.Params)-1 {
		// For variadic parameter check, if type is mixed then skip.
		if types.IsTypeMixed(paramType) {
			return
		}
	}
	paramAllowsNull := types.IsTypeNullable(paramType)
	varIsNullable := types.IsTypeNullable(callType)
	if varIsNullable && !paramAllowsNull {
		b.report(arg, LevelWarning, "notNullSafetyFunctionArgumentFunctionCall",
			"not null safety call in function %s signature of param %s when calling function %s",
			formatSlashesFuncName(fn), param.Name, funcInfo.Name)
	}

	if paramType.Empty() || varIsNullable {
		return
	}

	if !b.isTypeCompatible(callType, param.Typ) {
		b.report(arg, LevelWarning, "notSafetyCall",
			"potential not safety call in function %s signature of param %s when calling function %s",
			formatSlashesFuncName(fn), param.Name, funcInfo.Name)
	}
}

func (b *blockWalker) checkStaticCallSafety(arg ir.Node, fn meta.FuncInfo, paramIndex int, staticCallF *ir.StaticCallExpr, haveVariadic bool) {
	funcName, ok := staticCallF.Call.(*ir.Identifier)
	if !ok {
		return
	}

	var className string
	switch classNameNode := staticCallF.Class.(type) {
	case *ir.SimpleVar:
		varInfo, found := b.ctx.sc.GetVar(classNameNode)
		if !found {
			return
		}
		className = varInfo.Type.String()
	case *ir.Name:
		className = "\\" + classNameNode.Value
	}

	classInfo, ok := b.r.ctx.st.Info.GetClass(className)
	if !ok {
		return
	}

	funcInfo, ok := classInfo.Methods.Get(funcName.Value)
	if !ok {
		return
	}

	param := nullSafetyRealParamForCheck(fn, paramIndex, haveVariadic)
	paramType := param.Typ
	if haveVariadic && paramIndex >= len(fn.Params)-1 {
		// For variadic parameter check, if type is mixed then skip.
		if types.IsTypeMixed(paramType) {
			return
		}
	}
	funcType := funcInfo.Typ
	paramAllowsNull := types.IsTypeNullable(paramType)
	varIsNullable := types.IsTypeNullable(funcType)
	if varIsNullable && !paramAllowsNull {
		b.report(arg, LevelWarning, "notNullSafetyFunctionArgumentStaticFunctionCall",
			"not null safety call in function %s signature of param %s when calling static function %s",
			formatSlashesFuncName(fn), param.Name, funcInfo.Name)
	}

	if paramType.Empty() || varIsNullable {
		return
	}

	if !b.isTypeCompatible(funcType, param.Typ) {
		b.report(arg, LevelWarning, "notSafetyCall",
			"potential not safety static call in function %s signature of param %s",
			formatSlashesFuncName(fn), param.Name)
	}
}

func (b *blockWalker) checkSimpleVarSafety(arg ir.Node, fn meta.FuncInfo, paramIndex int, variable *ir.SimpleVar, haveVariadic bool) {
	varInfo, ok := b.ctx.sc.GetVar(variable)
	if !ok {
		return
	}

	param := nullSafetyRealParamForCheck(fn, paramIndex, haveVariadic)
	paramType := param.Typ
	if haveVariadic && paramIndex >= len(fn.Params)-1 {
		// For variadic parameter check, if type is mixed then skip.
		if types.IsTypeMixed(paramType) {
			return
		}
	}
	paramAllowsNull := types.IsTypeNullable(paramType)
	varType := varInfo.Type
	varIsNullable := types.IsTypeNullable(varType)
	if varIsNullable && !paramAllowsNull {
		b.report(arg, LevelWarning, "notNullSafetyFunctionArgumentVariable",
			"not null safety call in function %s signature of param %s",
			formatSlashesFuncName(fn), param.Name)
	}

	if paramType.Empty() || varIsNullable {
		return
	}

	if !b.isTypeCompatible(varType, paramType) {
		b.report(arg, LevelWarning, "notSafetyCall",
			"potential not safety call in function %s signature of param %s",
			formatSlashesFuncName(fn), param.Name)
	}
}

func (b *blockWalker) isTypeCompatible(varType types.Map, paramType types.Map) bool {
	if paramType.Empty() {
		return true
	}

	var forcedVarType = types.NewMapFromMap(solver.ResolveTypes(b.r.metaInfo(), "", varType, solver.ResolverMap{}))

	// Attempt to merge union types if one is a subclass/implementation of the other
	if forcedVarType.Len() > 1 {
		metaInfo := b.r.metaInfo()
		forcedVarType = solver.MergeUnionTypes(metaInfo, forcedVarType)
	}

	if forcedVarType.Len() > paramType.Len() {
		if paramType.Contains(types.WrapArrayOf("mixed")) || paramType.Contains("mixed") {
			return true
		}
		return false
	}

	isVarBoolean := forcedVarType.IsBoolean()
	isClass := forcedVarType.IsClass()
	varClassName := forcedVarType.String()

	for _, param := range paramType.Keys() {
		// boolean case
		if isVarBoolean && (param == "bool" || param == "boolean") {
			return true
		}

		if paramType.Contains(types.WrapArrayOf("mixed")) || paramType.Contains("mixed") {
			return true
		}

		// exact match
		if forcedVarType.Contains(param) {
			return true
		}

		// class check
		if isClass {
			if param == "object" || strings.Contains(param, varClassName) {
				return true
			}
			if !types.IsScalar(param) {
				metaInfo := b.r.metaInfo()
				if solver.Implements(metaInfo, varClassName, param) {
					return true
				} else if solver.ImplementsAbstract(metaInfo, varClassName, param) {
					return true
				}
			}
		}
	}

	forcedParamType := types.NewMapFromMap(solver.ResolveTypes(b.r.metaInfo(), "", paramType, solver.ResolverMap{}))

	// TODO: This is bullshit because we have no good type inferring for arrays: bool[1] will be bool[]! !not bool!
	if strings.Contains(forcedParamType.String(), "[") {
		idx := strings.Index(forcedParamType.String(), "[")
		arrayType := forcedParamType.String()[:idx] //nolint:gocritic
		return forcedParamType.Contains(arrayType)
	} else if strings.Contains(varType.String(), "[") {
		idx := strings.Index(varType.String(), "[")
		arrayType := varType.String()[:idx] //nolint:gocritic
		return forcedParamType.Contains(arrayType)
	}
	return !forcedParamType.Intersect(forcedVarType).Empty()
}

func (b *blockWalker) checkConstFetchSafety(arg ir.Node, fn meta.FuncInfo, paramIndex int, constExpr *ir.ConstFetchExpr, haveVariadic bool) {
	constVal := constExpr.Constant.Value
	isNull := constVal == "null"

	param := nullSafetyRealParamForCheck(fn, paramIndex, haveVariadic)
	paramType := param.Typ
	if haveVariadic && paramIndex >= len(fn.Params)-1 {
		if types.IsTypeMixed(paramType) {
			return
		}
	}
	paramAllowsNull := types.IsTypeNullable(paramType)
	if isNull {
		if !paramAllowsNull {
			b.report(arg, LevelWarning, "notNullSafetyFunctionArgumentConstFetch",
				"null passed to non-nullable parameter %s in function %s",
				param.Name, formatSlashesFuncName(fn))
		}
	} else {
		isBool := constVal == "true" || constVal == "false"
		if isBool {
			typ := types.NewMap(constVal)
			if !b.isTypeCompatible(typ, paramType) {
				b.report(arg, LevelWarning, "notSafetyCall",
					"potential not safety access in parameter %s of function %s",
					param.Name, formatSlashesFuncName(fn))
			}
		}
	}
}

func (b *blockWalker) checkArrayDimFetchSafety(arg ir.Node, fn meta.FuncInfo, paramIndex int, arrayExpr *ir.ArrayDimFetchExpr, haveVariadic bool) {
	baseVar, ok := arrayExpr.Variable.(*ir.SimpleVar)
	if !ok {
		return
	}

	varInfo, found := b.ctx.sc.GetVar(baseVar)
	if !found {
		return
	}

	param := nullSafetyRealParamForCheck(fn, paramIndex, haveVariadic)
	if haveVariadic && paramIndex >= len(fn.Params)-1 {
		if types.IsTypeMixed(param.Typ) {
			return
		}
	}
	paramAllowsNull := types.IsTypeNullable(param.Typ)
	varType := varInfo.Type
	varIsNullable := types.IsTypeNullable(varType)
	if varIsNullable && !paramAllowsNull {
		b.report(arg, LevelWarning, "notNullSafetyFunctionArgumentArrayDimFetch",
			"potential null array access in parameter %s of function %s",
			param.Name, formatSlashesFuncName(fn))
	}

	if param.Typ.Empty() || varIsNullable {
		return
	}

	if !b.isTypeCompatible(varType, param.Typ) {
		b.report(arg, LevelWarning, "notSafetyCall",
			"potential not safety array access in parameter %s of function %s",
			param.Name, formatSlashesFuncName(fn))
	}
}

func (b *blockWalker) checkListExprSafety(arg ir.Node, fn meta.FuncInfo, paramIndex int, listExpr *ir.ListExpr, haveVariadic bool) {
	for _, item := range listExpr.Items {
		if item == nil {
			continue
		}

		if item.Key != nil {
			b.checkNotSafetyCallArgsF([]ir.Node{item.Key}, fn)
		}

		if item.Val != nil {
			param := nullSafetyRealParamForCheck(fn, paramIndex, haveVariadic)
			paramType := param.Typ
			if simpleVar, ok := item.Val.(*ir.SimpleVar); ok {
				varInfo, found := b.ctx.sc.GetVar(simpleVar)
				if found {
					varType := varInfo.Type
					varIsNullable := types.IsTypeNullable(varType)
					if varIsNullable && !types.IsTypeNullable(paramType) {
						{
							b.report(arg, LevelWarning, "notNullSafetyFunctionArgumentList",
								"potential null value in list assignment for param %s in function %s",
								param.Name, formatSlashesFuncName(fn))
						}
					}
					if paramType.Empty() || varIsNullable {
						return
					}

					if !b.isTypeCompatible(varType, paramType) {
						b.report(arg, LevelWarning, "notSafetyCall",
							"potential not safety list assignment for param %s in function %s",
							param.Name, formatSlashesFuncName(fn))
					}
				}
			}
			b.checkNotSafetyCallArgsF([]ir.Node{item.Val}, fn)
		}
	}
}

// checkingPropertyFetchSafetyCondition verifies the safety of the final property fetch
// (the rightmost node in the chain) by checking null-safety and type compatibility
func (b *blockWalker) checkingPropertyFetchSafetyCondition(
	expr ir.Node,
	propType types.Map,
	prpName string,
	fn meta.FuncInfo,
	paramIndex int,
	haveVariadic bool,
) {
	isPrpNullable := types.IsTypeNullable(propType)
	param := nullSafetyRealParamForCheck(fn, paramIndex, haveVariadic)
	paramType := param.Typ

	if haveVariadic && paramIndex >= len(fn.Params)-1 {
		if types.IsTypeMixed(paramType) {
			return
		}
		paramAllowsNull := types.IsTypeNullable(paramType)
		if isPrpNullable && !paramAllowsNull {
			b.report(expr, LevelWarning, "notNullSafetyFunctionArgumentPropertyFetch",
				"potential null dereference when accessing property '%s'", prpName)
		}
		return
	}

	paramAllowsNull := types.IsTypeNullable(paramType)
	if isPrpNullable && !paramAllowsNull {
		b.report(expr, LevelWarning, "notNullSafetyFunctionArgumentPropertyFetch",
			"potential null dereference when accessing property '%s'", prpName)
	}

	if paramType.Empty() || isPrpNullable {
		return
	}

	if !b.isTypeCompatible(propType, paramType) {
		b.report(expr, LevelWarning, "notSafetyCall",
			"potential not safety accessing property '%s'", prpName)
	}
}

// collectUnifiedPropertyFetchChain collects the entire chain of property fetches (instance or static)
// and returns a slice of nodes in the order: [rightmost property fetch, ..., base node (e.g. SimpleVar)]
func (b *blockWalker) collectUnifiedPropertyFetchChain(expr ir.Node) []ir.Node {
	var chain []ir.Node
	switch e := expr.(type) {
	case *ir.PropertyFetchExpr:
		cur := e
		for {
			chain = append(chain, cur)
			if nested, ok := cur.Variable.(*ir.PropertyFetchExpr); ok {
				cur = nested
			} else {
				chain = append(chain, cur.Variable)
				break
			}
		}
	case *ir.StaticPropertyFetchExpr:
		cur := e
		for {
			chain = append(chain, cur)
			if nested, ok := cur.Class.(*ir.StaticPropertyFetchExpr); ok {
				cur = nested
			} else {
				chain = append(chain, cur.Class)
				break
			}
		}
	}
	return chain
}

// checkUnifiedPropertyFetchNotSafety combines checks for instance and static property access
// For the final (rightmost) node (the one being substituted), we check null and type safety
// and all intermediate nodes (except the base one) must be classes
func (b *blockWalker) checkUnifiedPropertyFetchNotSafety(expr ir.Node, fn meta.FuncInfo, paramIndex int, haveVariadic bool) {
	chain := b.collectUnifiedPropertyFetchChain(expr)
	if len(chain) == 0 {
		return
	}

	globalMetaInfo := b.linter.classParseState()

	var finalPropType types.Map
	var finalPropName string
	switch node := chain[0].(type) {
	case *ir.PropertyFetchExpr:
		propInfo := resolvePropertyFetch(b.ctx.sc, globalMetaInfo, b.ctx.customTypes, node, b.r.strictMixed)
		if !propInfo.isFound {
			return
		}
		ip, ok := node.Property.(*ir.Identifier)
		if !ok {
			return
		}
		finalPropType = propInfo.info.Typ
		finalPropName = ip.Value
	case *ir.StaticPropertyFetchExpr:
		variable, isVar := node.Property.(*ir.SimpleVar)
		class, isClass := node.Class.(*ir.SimpleVar)

		if isVar && isClass {
			if !isClass {
				return
			}
			classTyp, ok := b.ctx.sc.GetVarType(class)
			if !ok {
				return
			}
			if classTyp.Contains("null") {
				classTyp.Erase("null")
			}

			property, found := solver.FindProperty(b.r.ctx.st.Info, classTyp.String(), "$"+variable.Name)
			if !found {
				return
			}

			finalPropType = property.Info.Typ
			finalPropName = variable.Name
		} else {
			propInfo := resolveStaticPropertyFetch(globalMetaInfo, node)
			if !propInfo.isFound {
				return
			}

			finalPropType = propInfo.info.Info.Typ
			finalPropName = propInfo.propertyName
		}
	default:
		return
	}

	b.checkingPropertyFetchSafetyCondition(expr, finalPropType, finalPropName, fn, paramIndex, haveVariadic)

	for i := 1; i < len(chain); i++ {
		switch node := chain[i].(type) {
		case *ir.PropertyFetchExpr:
			propInfo := resolvePropertyFetch(b.ctx.sc, globalMetaInfo, b.ctx.customTypes, node, b.r.strictMixed)
			if !propInfo.isFound {
				return
			}
			propType := propInfo.info.Typ
			if types.IsTypeNullable(propType) {
				b.report(node, LevelWarning, "notSafetyCall",
					"potential null dereference when accessing property '%s'", propInfo.propertyNode.Value)
				return
			}

			propType.Iterate(func(typ string) {
				if types.IsTrivial(typ) {
					b.report(node, LevelWarning, "notSafetyCall",
						"potential not safety accessing property '%s': intermediary node is not a class", propInfo.propertyNode.Value)
					return
				}
			})
		case *ir.SimpleVar:
			varType, ok := b.ctx.sc.GetVarType(node)
			if !ok {
				return
			}
			varType = solver.MergeUnionTypes(b.r.metaInfo(), varType)
			if types.IsTypeNullable(varType) {
				b.report(node, LevelWarning, "notSafetyCall",
					"potential null dereference when accessing variable '%s'", node.Name)
				return
			}
			varType.Iterate(func(typ string) {
				if types.IsTrivial(typ) {
					b.report(node, LevelWarning, "notSafetyCall",
						"potential not safety accessing variable '%s': intermediary node is not a class", node.Name)
					return
				}
			})
		}
	}
}

func (b *blockWalker) handleCallArgs(args []ir.Node, fn meta.FuncInfo) {
	b.checkNotSafetyCallArgsF(args, fn)

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
				b.handleAndCheckDimFetchLValue(a, "call_with_ref", types.MixedType)
				break
			}
			a.Walk(b)
		case *ir.ClosureExpr:
			var typ types.Map
			isInstance := b.ctx.sc.IsInInstanceMethod()
			if isInstance {
				typ, _ = b.ctx.sc.GetVarNameType("this")
			}

			// find the types for the arguments of the function that contains this closure
			var funcArgTypes []types.Map
			for _, arg := range args {
				tp := solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, arg.(*ir.Argument).Expr)
				funcArgTypes = append(funcArgTypes, tp)
			}

			closureSolver := &solver.ClosureCallerInfo{
				Name:     fn.Name,
				ArgTypes: funcArgTypes,
			}

			b.CheckParamNullability(a.Params)
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
	call := resolveMethodCall(b.ctx.sc, b.r.ctx.st, b.ctx.customTypes, e, b.r.strictMixed)

	e.Variable.Walk(b)
	e.Method.Walk(b)

	if !call.isMagic {
		b.handleCallArgs(e.Args, call.info)
	}
	b.ctx.exitFlags |= call.info.ExitFlags

	return false
}

func (b *blockWalker) handleStaticCall(e *ir.StaticCallExpr) bool {
	call := resolveStaticMethodCall(b.ctx.sc, b.r.ctx.st, e)
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
				b.handleVariableNode(s.Variable, types.NewMap(types.WrapElemOf(typ)), "foreach_value")
			})

			b.handleVariableNode(s.Key, arrayKeyType, "foreach_key")
			if list, ok := s.Variable.(*ir.ListExpr); ok {
				for _, item := range list.Items {
					b.handleVariableNode(item.Val, types.Map{}, "foreach_value")
				}
			} else {
				b.handleVariableNode(s.Variable, types.Map{}, "foreach_value")
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

		b.report(s.Key, LevelWarning, "unused", "Foreach key $%s is unused, can simplify $%s => $%s to just $%s", key.Name, key.Name, variable.Name, variable.Name)
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

	// Indexing stage.
	doc := phpdoctypes.Parse(fun.Doc, fun.Params, b.r.ctx.typeNormalizer)
	moveShapesToContext(&b.r.ctx, doc.Shapes)
	b.r.handleClosuresFromDoc(doc.Closures)

	funcParams := b.r.parseFuncParams(fun.Params, doc.ParamTypes, sc, nil)
	b.r.handleArrowFuncExpr(funcParams.params, fun.Expr, sc, b)

	name := &ir.Identifier{Value: "arrow function"}
	b.r.checker.CheckFuncParams(name, fun.Params, funcParams, doc.ParamTypes)

	// Check stage.
	errors := b.r.checker.CheckPHPDoc(fun, fun.Doc, fun.Params)
	b.r.reportPHPDocErrors(errors)

	return false
}

func (b *blockWalker) enterClosure(fun *ir.ClosureExpr, haveThis bool, thisType types.Map, closureSolver *solver.ClosureCallerInfo) bool {
	sc := meta.NewScope()
	sc.SetInClosure(true)

	if haveThis {
		sc.AddVarName("this", thisType, "closure inside instance method", meta.VarAlwaysDefined)
	} else {
		sc.AddVarName("this", types.NewMap("possibly_late_bound"), "possibly late bound $this", meta.VarAlwaysDefined)
	}

	// Indexing stage.
	doc := phpdoctypes.Parse(fun.Doc, fun.Params, b.r.ctx.typeNormalizer)
	moveShapesToContext(&b.r.ctx, doc.Shapes)
	b.r.handleClosuresFromDoc(doc.Closures)

	// Check stage.
	errors := b.r.checker.CheckPHPDoc(fun, fun.Doc, fun.Params)
	b.r.reportPHPDocErrors(errors)

	var hintReturnType types.Map
	if typ, ok := b.r.parseTypeHintNode(fun.ReturnType); ok {
		hintReturnType = typ
	}
	b.r.checker.CheckTypeHintNode(fun.ReturnType, "closure return type")

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
			b.report(v, LevelWarning, "undefinedVariable", "Cannot find referenced variable $%s", v.Name)
		}

		typ, ok := b.ctx.sc.GetVarNameType(v.Name)
		if ok {
			sc.AddVarName(v.Name, typ, "use", meta.VarAlwaysDefined)
		}

		b.untrackVarName(v.Name)
	}

	params := b.r.parseFuncParams(fun.Params, doc.ParamTypes, sc, closureSolver)

	funcInfo := b.r.handleFuncStmts(params.params, closureUses, fun.Stmts, sc)
	phpDocReturnTypes := doc.ReturnType
	actualReturnTypes := funcInfo.returnTypes
	exitFlags := funcInfo.prematureExitFlags

	returnTypes := functionReturnType(phpDocReturnTypes, hintReturnType, actualReturnTypes)

	name := autogen.GenerateClosureName(fun, b.r.ctx.st.CurrentFunction, b.r.ctx.st.CurrentFile)

	b.r.checker.CheckFuncParams(&ir.Identifier{Value: name}, fun.Params, params, doc.ParamTypes)

	if b.r.meta.Functions.H == nil {
		b.r.meta.Functions = meta.NewFunctionsMap()
	}

	var funcFlags meta.FuncFlags

	if params.isVariadic {
		funcFlags |= meta.FuncVariadic
	}

	b.r.meta.Functions.Set(name, meta.FuncInfo{
		Params:          params.params,
		Name:            name,
		Pos:             b.r.getElementPos(fun),
		Typ:             returnTypes.Immutable(),
		MinParamsCnt:    params.minParamsCount,
		Flags:           funcFlags,
		ExitFlags:       exitFlags,
		DeprecationInfo: doc.Deprecation,
	})

	return false
}

func (b *blockWalker) maybeAddAllVars(sc *meta.Scope, reason string) {
	sc.Iterate(func(varName string, typ types.Map, flags meta.VarFlags) {
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
			b.report(v, LevelError, "discardVar", "Used var $%s that is supposed to be unused (rename variable if it's intended or respecify --unused-var-regex flag)", varName)
		}

		b.untrackVarName(varName)
	}

	have := b.ctx.sc.HaveVar(v)

	if !have && !b.inArrowFunction {
		b.r.reportUndefinedVariable(v, b.ctx.sc.MaybeHaveVar(v), b.path)
		b.ctx.sc.AddVar(v, types.NewMap("undefined"), "undefined", meta.VarAlwaysDefined)
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
			b.r.reportUndefinedVariable(v, varMaybeNotDefined, b.path)
			b.ctx.sc.AddVar(v, types.NewMap("undefined"), "undefined", meta.VarAlwaysDefined)
		}
	}

	return false
}

func (b *blockWalker) handleTernary(e *ir.TernaryExpr) bool {
	if e.IfTrue == nil {
		return true // Skip `$x ?: $y` expressions
	}

	contexts := make([]*blockContext, 0, 2)

	initialContext := b.ctx
	falseContext := copyBlockContext(initialContext)
	trueContext := copyBlockContext(initialContext)
	contexts = append(contexts, trueContext, falseContext)

	b.withSpecificContext(trueContext, func() {
		a := &andWalker{
			b:              b,
			initialContext: initialContext,
			trueContext:    trueContext,
			falseContext:   falseContext,
		}
		e.Condition.Walk(a)
	})

	b.withSpecificContext(trueContext, func() {
		e.IfTrue.Walk(b)
	})
	b.replaceAllImplicitVars(trueContext, initialContext)

	b.withSpecificContext(falseContext, func() {
		e.IfFalse.Walk(b)
	})
	b.replaceAllImplicitVars(trueContext, initialContext)

	b.extractVariablesTo(b.ctx, contexts, 2)

	return false
}

// extractVariablesTo extracts all variables from contexts and creates
// new ones in the passed context, taking into account the number of
// places where the variable is defined, if the number is less than
// requiredCountToAlwaysDefined, then the variable will be designated
// as "not always defined".
func (b *blockWalker) extractVariablesTo(targetContext *blockContext, contexts []*blockContext, requiredCountToAlwaysDefined int) {
	varTypes := make(map[string]types.Map, targetContext.sc.Len())
	defCounts := make(map[string]int, targetContext.sc.Len())

	for _, ctx := range contexts {
		if ctx.exitFlags != 0 {
			continue
		}

		ctx.sc.Iterate(func(nm string, typ types.Map, flags meta.VarFlags) {
			varTypes[nm] = varTypes[nm].Append(typ)
			if flags.IsAlwaysDefined() {
				defCounts[nm]++
			}
		})
	}

	for nm, typeMap := range varTypes {
		var flags meta.VarFlags
		if defCounts[nm] == requiredCountToAlwaysDefined {
			flags = meta.VarAlwaysDefined
		}

		targetContext.sc.AddVarName(nm, typeMap, "all branches", flags)
	}
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

	var linksCount int
	var contexts []*blockContext

	onlyInstanceof := true
	// Add all new variables from the condition to the current scope.
	irutil.Inspect(s.Cond, func(n ir.Node) bool {
		switch n := n.(type) {
		case *ir.BooleanAndExpr:
		case *ir.BooleanOrExpr:
		case *ir.BooleanNotExpr:
			return true
		case *ir.InstanceOfExpr:
			return false

		case *ir.Assign:
			b.handleAssign(n)
			return false
		default:
			onlyInstanceof = false
		}

		return true
	})

	// initialContext is the context of the block in which the if-else is located.
	initialContext := b.ctx
	// First, we need to traverse the main condition.
	//   trueContext  will store the state of the variables in the **if** block.
	//   falseContext will store the state of the variables in the **else** block.
	falseContext := copyBlockContext(initialContext)
	trueContext := copyBlockContext(initialContext)
	b.withSpecificContext(trueContext, func() {
		a := &andWalker{
			b:              b,
			initialContext: initialContext,
			trueContext:    trueContext,
			falseContext:   falseContext,
		}

		s.Cond.Walk(a)
		varsToDelete = append(varsToDelete, a.varsToDelete...)
	})
	contexts = append(contexts, trueContext)

	// New reference variables can be created in the condition,
	// which should be moved outside the if block.
	moveNonLocalVariables(trueContext, b.ctx, b.nonLocalVars)

	if s.Stmt != nil {
		// We process the if body with a context if the condition is **true**.
		b.withSpecificContext(trueContext, func() {
			s.Stmt.Walk(b)
		})

		if trueContext.exitFlags == 0 {
			linksCount++
		}

		// Case:
		// if (!$a instanceof Boo) {
		//   return
		// }
		//
		// $a has Boo type
		if trueContext.exitFlags != 0 && onlyInstanceof && len(s.ElseIf) == 0 && s.Else == nil {
			b.ctx = falseContext
		}

		b.replaceAllImplicitVars(trueContext, initialContext)
	} else {
		linksCount++
	}

	for _, n := range s.ElseIf {
		if elseif, ok := n.(*ir.ElseIfStmt); ok {
			elseifTrueContext := copyBlockContext(falseContext)
			elseifFalseContext := copyBlockContext(falseContext)

			b.withSpecificContext(elseifTrueContext, func() {
				a := &andWalker{
					b:              b,
					initialContext: initialContext,
					trueContext:    elseifTrueContext,
					falseContext:   elseifFalseContext,
				}
				elseif.Cond.Walk(a)
				varsToDelete = append(varsToDelete, a.varsToDelete...)
			})

			// New reference variables can be created in the condition,
			// which should be moved outside the if block.
			moveNonLocalVariables(elseifTrueContext, b.ctx, b.nonLocalVars)

			// Handle if (...) smth(); else other_thing(); // without braces.
			switch n := n.(type) {
			case *ir.ElseStmt:
				b.addStatement(n.Stmt)
			case *ir.ElseIfStmt:
				b.addStatement(n.Stmt)
			default:
				b.addStatement(n)
			}

			b.withSpecificContext(elseifTrueContext, func() {
				b.handleElseIf(elseif)
				elseif.Stmt.Walk(b)
			})

			b.replaceAllImplicitVars(elseifTrueContext, initialContext)

			contexts = append(contexts, elseifTrueContext)
			if elseifTrueContext.exitFlags == 0 {
				linksCount++
			}

			// Every elseif changes variables in else.
			falseContext = elseifFalseContext
		} else {
			n.Walk(b)
		}
	}

	if s.Else != nil {
		// We process the else body with a context if the condition is **false**.
		b.withSpecificContext(falseContext, func() {
			s.Else.Walk(b)
		})

		if falseContext.exitFlags == 0 {
			linksCount++
		}

		b.replaceAllImplicitVars(falseContext, initialContext)

		contexts = append(contexts, falseContext)
	} else {
		linksCount++
	}

	b.propagateFlagsFromBranches(contexts, linksCount)
	b.extractVariablesTo(b.ctx, contexts, linksCount)

	return false
}

// replaceAllImplicitVars replaces any implicit variables that were added by
// instanceof, isset, etc. with their original versions from initialContext.
func (b *blockWalker) replaceAllImplicitVars(targetContext *blockContext, initialContext *blockContext) {
	targetContext.sc.Iterate(func(name string, typ types.Map, flags meta.VarFlags) {
		if !flags.IsImplicit() {
			return
		}

		oldVar, ok := initialContext.sc.GetVarName(name)
		if !ok {
			return
		}

		targetContext.sc.ReplaceVarName(name, oldVar.Type, "fallback", oldVar.Flags)
	})
}

func moveNonLocalVariables(trueContext, toContext *blockContext, nonLocalVars map[string]variableKind) {
	nonLocalToDelete := make([]string, 0, len(nonLocalVars))
	for name, kind := range nonLocalVars {
		if kind != varRef {
			continue
		}

		typesMap, ok := trueContext.sc.GetVarNameType(name)
		if !ok {
			continue
		}
		toContext.sc.AddVarName(name, typesMap, "ref", meta.VarAlwaysDefined)
		nonLocalToDelete = append(nonLocalToDelete, name)
	}
	for _, name := range nonLocalToDelete {
		delete(nonLocalVars, name)
	}
}

func (b *blockWalker) handleElseIf(s *ir.ElseIfStmt) {
	b.r.checker.CheckKeywordCase(s, "elseif")
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
			b.r.checker.CheckKeywordCase(c, "default")
		} else {
			cond.Walk(b)
			b.r.checker.CheckKeywordCase(c, "case")
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
					b.report(c, LevelWarning, "caseBreak", "Add break or '// fallthrough' to the end of the case")
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

	varTypes := make(map[string]types.Map, b.ctx.sc.Len())
	defCounts := make(map[string]int, b.ctx.sc.Len())

	for _, ctx := range contexts {
		b.propagateFlags(ctx)

		cleanFlags := ctx.exitFlags & (^breakFlags)
		if cleanFlags != 0 {
			continue
		}

		ctx.sc.Iterate(func(nm string, typ types.Map, flags meta.VarFlags) {
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
func (b *blockWalker) handleAndCheckDimFetchLValue(e *ir.ArrayDimFetchExpr, reason string, typ types.Map) {
	b.checkArrayDimFetch(e)

	switch v := e.Variable.(type) {
	case *ir.Var, *ir.SimpleVar:
		arrayOfType := typ.Map(types.WrapArrayOf)

		varType, ok := b.ctx.sc.GetVarType(v)
		// If the variable contains the type of an empty array, then it is
		// necessary to replace this type with a new, more precise one.
		if ok && varType.Len() == 1 && varType.Contains("empty_array") {
			b.replaceVar(v, arrayOfType, reason, meta.VarAlwaysDefined)
			sv, ok := v.(*ir.SimpleVar)
			if !ok {
				return
			}
			b.untrackVarName(sv.Name)
		} else {
			b.addVar(v, arrayOfType, reason, meta.VarAlwaysDefined)
			b.handleVariable(v)
		}
	case *ir.ArrayDimFetchExpr:
		arrayOfType := typ.Map(types.WrapArrayOf)
		b.handleAndCheckDimFetchLValue(v, reason, arrayOfType)
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
		if types.IsClass(t) {
			maybeHaveClasses = true

			if !haveArrayAccess && solver.Implements(b.r.ctx.st.Info, t, `\ArrayAccess`) {
				haveArrayAccess = true
			}
		}
	})

	if maybeHaveClasses && !haveArrayAccess {
		b.report(s.Variable, LevelNotice, "arrayAccess", "Array access to non-array type %s", typ)
	}
}

// some day, perhaps, there will be some difference between handleAssignReference and handleAssign
func (b *blockWalker) handleAssignReference(a *ir.AssignReference) bool {
	switch v := a.Variable.(type) {
	case *ir.ArrayDimFetchExpr:
		typ := solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expr)
		b.handleAndCheckDimFetchLValue(v, "assign_array", typ)
		a.Expr.Walk(b)
		return false
	case *ir.Var, *ir.SimpleVar:
		b.addVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expr), "assign", meta.VarAlwaysDefined)
		b.addNonLocalVar(v, varRef)
	case *ir.ListExpr:
		// TODO: figure out whether this case is reachable.
		for _, item := range v.Items {
			b.handleVariableNode(item.Val, types.NewMap("unknown_from_list"), "assign")
		}
	default:
		a.Variable.Walk(b)
	}

	a.Expr.Walk(b)
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

		var tp types.Map
		if !ok {
			tp = types.NewMap("unknown_from_list")
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
	// We store types.Type instead of string to avoid the need to do strings.Join
	// when we want to create a TypesMap.
	var elemTypes []types.Type
	var shapeType string
	typ.Iterate(func(typ string) {
		switch {
		case types.IsShape(typ):
			shapeType = typ
		case types.IsArray(typ):
			elemType := types.ArrayType(typ)
			elemTypes = append(elemTypes, types.Type{Elem: elemType})
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
		elemTypeMap := types.NewMapFromTypes(elemTypes).Immutable()
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
		b.handleVariableNode(item.Val, types.NewMap("unknown_from_list"), "list-assign")
	}
}

func (b *blockWalker) paramClobberCheck(v *ir.SimpleVar) {
	if b.callsFuncGetArgs {
		return
	}
	if _, ok := b.unusedParams[v.Name]; ok && !b.path.Conditional() {
		b.report(v, LevelWarning, "paramClobber", "Param $%s re-assigned before being used", v.Name)
	}
}

func (b *blockWalker) handleAssign(a *ir.Assign) bool {
	b.handleComments(a.Variable)

	a.Expr.Walk(b)

	switch v := a.Variable.(type) {
	case *ir.ArrayDimFetchExpr:
		typ := solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expr)
		b.handleAndCheckDimFetchLValue(v, "assign_array", typ)
		return false
	case *ir.SimpleVar:
		b.paramClobberCheck(v)
		b.replaceVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expr), "assign", meta.VarAlwaysDefined)
	case *ir.Var:
		b.replaceVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expr), "assign", meta.VarAlwaysDefined)
	case *ir.ListExpr:
		if !b.isIndexingComplete() {
			return true
		}

		b.handleAssignList(v, a.Expr)
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
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, a.Expr, b.ctx.customTypes))
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
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, a.Expr, b.ctx.customTypes))
		cls.Properties["$"+sv.Name] = p
	default:
		a.Variable.Walk(b)
	}

	return false
}

func (b *blockWalker) handleAssignOp(assign ir.Node) {
	var typ types.Map
	var v ir.Node

	switch assign := assign.(type) {
	case *ir.AssignPlus:
		e := &ir.PlusExpr{
			Left:  assign.Variable,
			Right: assign.Expr,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)
	case *ir.AssignMinus:
		e := &ir.MinusExpr{
			Left:  assign.Variable,
			Right: assign.Expr,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)
	case *ir.AssignMul:
		e := &ir.MulExpr{
			Left:  assign.Variable,
			Right: assign.Expr,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)
	case *ir.AssignDiv:
		e := &ir.DivExpr{
			Left:  assign.Variable,
			Right: assign.Expr,
		}
		v = assign.Variable
		typ = solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, e)

	case *ir.AssignConcat:
		v = assign.Variable
		typ = types.PreciseStringType
	case *ir.AssignShiftLeft:
		v = assign.Variable
		typ = types.PreciseIntType
	case *ir.AssignShiftRight:
		v = assign.Variable
		typ = types.PreciseIntType
	case *ir.AssignCoalesce:
		e := &ir.CoalesceExpr{
			Left:  assign.Variable,
			Right: assign.Expr,
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
			b.report(n, LevelWarning, "unused", `Variable $%s is unused (use $_ to ignore this inspection or specify --unused-var-regex flag)`, name)
		}
	}
}

func (b *blockWalker) handleVariableNode(n ir.Node, typ types.Map, what string) {
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

func (b *blockWalker) exprType(n ir.Node) types.Map {
	return solver.ExprTypeCustom(b.ctx.sc, b.r.ctx.st, n, b.ctx.customTypes)
}
