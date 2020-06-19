package linter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/astutil"
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/cast"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/rules"
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

const (
	// FlagReturn shows whether or not block has "return"
	FlagReturn = 1 << iota
	FlagBreak
	FlagContinue
	FlagThrow
	FlagDie
)

// BlockWalker is used to process function/method contents.
type BlockWalker struct {
	ctx *blockContext

	// inferred return types if any
	returnTypes meta.TypesMap

	r *RootWalker

	custom []BlockChecker

	ignoreFunctionBodies bool
	rootLevel            bool // analysing root-level code

	// state
	statements map[node.Node]struct{}

	// whether a function has a return without explit expression.
	// Required to make a decision in void vs null type selection,
	// since "return" is the same as "return null".
	bareReturn bool
	// whether a function has a return with explicit expression.
	// When can't infer precise type, can use mixed.
	returnsValue bool

	// callsParentConstructor is set to true when parent::__construct() call
	// is found. This is needed for a root walker to report constructors
	// that do not call parent constructors.
	callsParentConstructor bool

	// shared state between all blocks
	unusedVars   map[string][]node.Node
	nonLocalVars map[string]struct{} // static, global and other vars that have complex control flow
}

func (b *BlockWalker) addStatement(n node.Node) {
	if b.statements == nil {
		b.statements = make(map[node.Node]struct{})
	}
	b.statements[n] = struct{}{}

	// A hack for assignment-used-as-expression checks to work
	e, ok := n.(*stmt.Expression)
	if !ok {
		return
	}

	assignment, ok := e.Expr.(*assign.Assign)
	if !ok {
		return
	}

	b.statements[assignment] = struct{}{}
}

func (b *BlockWalker) reportDeadCode(n node.Node) {
	if b.ctx.deadCodeReported {
		return
	}

	switch n.(type) {
	case *stmt.Break, *stmt.Return, *expr.Exit, *stmt.Throw:
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

	b.ctx.deadCodeReported = true
	b.r.Report(n, LevelInformation, "deadCode", "Unreachable code")
}

func (b *BlockWalker) checkRedundantCastArray(e node.Node) {
	if !meta.IsIndexingComplete() {
		return
	}
	typ := solver.ExprType(b.ctx.sc, b.r.ctx.st, e)
	if typ.Len() == 1 && typ.Is("mixed[]") {
		b.r.Report(e, LevelDoNotReject, "redundantCast", "expression already has array type")
	}
}

func (b *BlockWalker) checkRedundantCast(e node.Node, dstType string) {
	if !meta.IsIndexingComplete() {
		return
	}
	typ := solver.ExprType(b.ctx.sc, b.r.ctx.st, e)
	if typ.Len() != 1 {
		return
	}
	typ.Iterate(func(x string) {
		if x == dstType {
			b.r.Report(e, LevelDoNotReject, "redundantCast",
				"expression already has %s type", dstType)
		}
	})
}

// checkVoidType reports if node has void type
func (b *BlockWalker) checkVoidType(n node.Node) {
	if b.isVoid(n) {
		b.r.Report(n, LevelDoNotReject, "voidResultUsed", "void function result used")
	}
}

func (b *BlockWalker) checkStrictCmp(n node.Node, left, right node.Node) {
	needsStrictCmp := func(n node.Node) bool {
		c, ok := n.(*expr.ConstFetch)
		if !ok {
			return false
		}
		nm, ok := c.Constant.(*name.Name)
		if !ok {
			return false
		}
		return meta.NameEquals(nm, "true") ||
			meta.NameEquals(nm, "false") ||
			meta.NameEquals(nm, "null")
	}

	var badNode node.Node
	switch {
	case needsStrictCmp(left):
		badNode = left
	case needsStrictCmp(right):
		badNode = right
	}
	if badNode != nil {
		suggest := "==="
		if _, ok := n.(*binary.NotEqual); ok {
			suggest = "!=="
		}
		b.r.Report(n, LevelWarning, "strictCmp", "non-strict comparison with %s (use %s)",
			astutil.FmtNode(badNode), suggest)
	}
}

func (b *BlockWalker) checkBinaryVoidType(left, right node.Node) {
	b.checkVoidType(left)
	b.checkVoidType(right)
}

func (b *BlockWalker) checkBinaryDupArgsNoFloat(n, left, right node.Node) {
	if b.exprType(left).Contains("float") || b.exprType(right).Contains("float") {
		return
	}
	b.checkBinaryDupArgs(n, left, right)
}

func (b *BlockWalker) checkBinaryDupArgs(n, left, right node.Node) {
	// Check for `$x <op> $y` where `<op>` is not a correct way to
	// handle identical operands.
	if !b.sideEffectFree(left) || !b.sideEffectFree(right) {
		return
	}
	if astutil.NodeEqual(left, right) {
		b.r.Report(n, LevelWarning, "dupSubExpr", "duplicated operands in %s expression", binaryOpString(n))
	}
}

// EnterNode is called before walking to inner nodes.
func (b *BlockWalker) EnterNode(w walker.Walkable) (res bool) {
	res = true

	for _, c := range b.custom {
		c.BeforeEnterNode(w)
	}

	n := w.(node.Node)

	if b.ctx.exitFlags != 0 {
		b.reportDeadCode(n)
	}

	if ffs := n.GetFreeFloating(); ffs != nil {
		for _, cs := range *ffs {
			for _, c := range cs {
				b.parseComment(n, c)
			}
		}
	}

	switch s := w.(type) {
	case *stmt.Expression:
		b.handleStmtExpression(s)

	case *binary.BitwiseAnd:
		b.handleBitwiseOp(s, s.Left, s.Right)
	case *binary.BitwiseOr:
		b.handleBitwiseOp(s, s.Left, s.Right)
	case *binary.BitwiseXor:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.LogicalAnd:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.BooleanAnd:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.LogicalOr:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
		res = b.handleLogicalOr(s)
	case *binary.BooleanOr:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.LogicalXor:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.Plus:
		b.checkBinaryVoidType(s.Left, s.Right)
	case *binary.Minus:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.Mul:
		b.checkBinaryVoidType(s.Left, s.Right)
	case *binary.Div:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.Mod:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.Pow:
		b.checkBinaryVoidType(s.Left, s.Right)
	case *binary.Equal:
		b.checkStrictCmp(s, s.Left, s.Right)
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.NotEqual:
		b.checkStrictCmp(s, s.Left, s.Right)
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.Identical:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.NotIdentical:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.Smaller:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.SmallerOrEqual:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.Greater:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgs(s, s.Left, s.Right)
	case *binary.GreaterOrEqual:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)
	case *binary.Spaceship:
		b.checkBinaryVoidType(s.Left, s.Right)
		b.checkBinaryDupArgsNoFloat(s, s.Left, s.Right)

	// end of binary functions

	case *cast.Double:
		b.checkRedundantCast(s.Expr, "float")
	case *cast.Int:
		b.checkRedundantCast(s.Expr, "int")
	case *cast.Bool:
		b.checkRedundantCast(s.Expr, "bool")
	case *cast.String:
		b.checkRedundantCast(s.Expr, "string")
	case *cast.Array:
		b.checkRedundantCastArray(s.Expr)
	case *stmt.Global:
		b.r.checkKeywordCase(s, "global")
		for _, v := range s.Vars {
			nm := varToString(v)
			if nm == "" {
				continue
			}
			if _, ok := superGlobals[nm]; ok {
				b.r.Report(v, LevelWarning, "redundantGlobal", "%s is superglobal", nm)
			}
			b.addVar(v, meta.NewTypesMap(meta.WrapGlobal(nm)), "global", meta.VarAlwaysDefined)
			b.addNonLocalVar(v)
		}
		res = false
	case *stmt.Static:
		for _, vv := range s.Vars {
			v := vv.(*stmt.StaticVar)
			ev := v.Variable
			typ := solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, v.Expr, b.ctx.customTypes)
			// Static vars can be assigned below and preserve the type of
			// the previously assigned value.
			typ.MarkAsImprecise()
			b.addVarName(v, ev.Name, typ, "static", meta.VarAlwaysDefined)
			b.addNonLocalVarName(ev.Name)
			if v.Expr != nil {
				v.Expr.Walk(b)
			}
		}
		res = false
	case *node.Root:
		for _, st := range s.Stmts {
			b.addStatement(st)
		}
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
	case *stmt.Foreach:
		res = b.handleForeach(s)
		b.r.checkKeywordCase(s, "foreach")
	case *stmt.For:
		res = b.handleFor(s)
		b.r.checkKeywordCase(s, "for")
	case *stmt.While:
		res = b.handleWhile(s)
		b.r.checkKeywordCase(s, "while")
	case *stmt.Do:
		res = b.handleDo(s)
		b.r.checkKeywordCase(s, "do")
	case *stmt.Else:
		b.r.checkKeywordCase(s, "else")
	case *stmt.ElseIf:
		if s.Merged {
			b.r.checkKeywordCase(s, "else")
			if ff := (*s.GetFreeFloating())[freefloating.Else]; len(ff) != 0 {
				rightmostPos := ff[len(ff)-1].Position
				b.r.checkKeywordCasePos(s, rightmostPos.EndPos, "if")
			}
		} else {
			b.r.checkKeywordCase(s, "elseif")
		}
	case *stmt.If:
		// TODO: handle constant if expressions
		// TODO: maybe try to handle when variables are defined and used with the same condition
		res = b.handleIf(s)
	case *stmt.Switch:
		res = b.handleSwitch(s)
	case *expr.Ternary:
		res = b.handleTernary(s)
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
		b.r.checkKeywordCase(s, "new")
	case *stmt.Unset:
		res = b.handleUnset(s)
	case *expr.Isset:
		res = b.handleIsset(s)
	case *expr.Empty:
		res = b.handleEmpty(s)
	case *node.Var:
		res = b.handleVariable(s)
	case *node.SimpleVar:
		res = b.handleVariable(s)
	case *expr.ArrayDimFetch:
		b.checkArrayDimFetch(s)
	case *stmt.Function:
		res = b.handleFunction(s)
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
		var typ meta.TypesMap
		isInstance := b.ctx.sc.IsInInstanceMethod()
		if isInstance {
			typ, _ = b.ctx.sc.GetVarNameType("this")
		}
		res = b.enterClosure(s, isInstance, typ)
	case *stmt.Return:
		b.handleReturn(s)
		b.r.checkKeywordCase(s, "return")
	case *stmt.Break:
		b.r.checkKeywordCase(s, "break")
	case *stmt.Continue:
		b.handleContinue(s)
		b.r.checkKeywordCase(s, "continue")
	case *expr.Clone:
		b.r.checkKeywordCase(s, "clone")
	case *stmt.ConstList:
		b.r.checkKeywordCase(s, "const")
	case *stmt.Goto:
		b.r.checkKeywordCase(s, "goto")
	case *stmt.Throw:
		b.r.checkKeywordCase(s, "throw")
	case *expr.Yield:
		b.r.checkKeywordCase(s, "yield")
	case *expr.YieldFrom:
		b.r.checkKeywordCase(s, "yield")
	case *expr.Include:
		b.r.checkKeywordCase(n, "include")
	case *expr.IncludeOnce:
		b.r.checkKeywordCase(n, "include_once")
	case *expr.Require:
		b.r.checkKeywordCase(n, "require")
	case *expr.RequireOnce:
		b.r.checkKeywordCase(n, "require_once")
	}

	for _, c := range b.custom {
		c.AfterEnterNode(w)
	}

	if meta.IsIndexingComplete() && b.r.anyRset != nil {
		// Note: no need to check localRset for nil.
		kind := rules.CategorizeNode(n)
		if kind != rules.KindNone {
			b.r.runRules(n, b.ctx.sc, b.r.anyRset.RulesByKind[kind])
			if !b.rootLevel {
				b.r.runRules(n, b.ctx.sc, b.r.localRset.RulesByKind[kind])
			}
		}
	}

	return res
}

func (b *BlockWalker) handleFunction(fun *stmt.Function) bool {
	if b.ignoreFunctionBodies {
		return false
	}

	return b.r.enterFunction(fun)
}

func (b *BlockWalker) handleReturn(ret *stmt.Return) {
	if ret.Expr == nil {
		// Return without explicit return value.
		b.bareReturn = true
		return
	}
	b.returnsValue = true

	typ := solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, ret.Expr, b.ctx.customTypes)
	b.returnTypes = b.returnTypes.Append(typ)
}

func (b *BlockWalker) handleLogicalOr(or *binary.LogicalOr) bool {
	or.Left.Walk(b)

	// We're going to discard "or" RHS effects on the exit flags.
	exitFlags := b.ctx.exitFlags
	or.Right.Walk(b)
	b.ctx.exitFlags = exitFlags

	return false
}

func (b *BlockWalker) handleContinue(s *stmt.Continue) {
	if s.Expr == nil && b.ctx.innermostLoop == loopSwitch {
		b.r.Report(s, LevelError, "caseContinue", "'continue' inside switch is 'break'")
	}
}

func (b *BlockWalker) addNonLocalVarName(nm string) {
	b.nonLocalVars[nm] = struct{}{}
}

func (b *BlockWalker) addNonLocalVar(v node.Node) {
	sv, ok := v.(*node.SimpleVar)
	if !ok {
		return
	}
	b.addNonLocalVarName(sv.Name)
}

// replaceVar must be used to track assignments to conrete var nodes if they are available
func (b *BlockWalker) replaceVar(v node.Node, typ meta.TypesMap, reason string, flags meta.VarFlags) {
	b.ctx.sc.ReplaceVar(v, typ, reason, flags)
	sv, ok := v.(*node.SimpleVar)
	if !ok {
		return
	}

	// Writes to non-local variables do count as usages
	if _, ok := b.nonLocalVars[sv.Name]; ok {
		delete(b.unusedVars, sv.Name)
		return
	}

	// Writes to variables that are done in a loop should not count as unused variables
	// because they can be read on the next iteration (ideally we should check for that too :))
	if !b.ctx.insideLoop {
		b.unusedVars[sv.Name] = append(b.unusedVars[sv.Name], sv)
	}
}

func (b *BlockWalker) trackVarName(n node.Node, nm string) {
	// Writes to non-local variables do count as usages
	if _, ok := b.nonLocalVars[nm]; ok {
		delete(b.unusedVars, nm)
		return
	}

	// Writes to variables that are done in a loop should not count as unused variables
	// because they can be read on the next iteration (ideally we should check for that too :))
	if !b.ctx.insideLoop {
		b.unusedVars[nm] = append(b.unusedVars[nm], n)
	}
}

func (b *BlockWalker) addVarName(n node.Node, nm string, typ meta.TypesMap, reason string, flags meta.VarFlags) {
	b.ctx.sc.AddVarName(nm, typ, reason, flags)
	b.trackVarName(n, nm)
}

// addVar must be used to track assignments to conrete var nodes if they are available
func (b *BlockWalker) addVar(v node.Node, typ meta.TypesMap, reason string, flags meta.VarFlags) {
	b.ctx.sc.AddVar(v, typ, reason, flags)
	sv, ok := v.(*node.SimpleVar)
	if !ok {
		return
	}
	b.trackVarName(v, sv.Name)
}

func (b *BlockWalker) parseComment(n node.Node, c freefloating.String) {
	if c.StringType != freefloating.CommentType {
		return
	}
	str := c.Value

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
			b.r.Report(n, LevelInformation, "phpdocType", "%s on line %d", warning, p.Line())
		}
		m := newTypesMap(&b.r.ctx, types)
		b.ctx.sc.AddVarFromPHPDoc(strings.TrimPrefix(p.Var, "$"), m, "@var")
	}
}

func (b *BlockWalker) handleUnset(s *stmt.Unset) bool {
	for _, v := range s.Vars {
		switch v := v.(type) {
		case *node.SimpleVar:
			delete(b.unusedVars, v.Name)
			b.ctx.sc.DelVar(v, "unset")
		case *node.Var:
			b.ctx.sc.DelVar(v, "unset")
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
		case *node.Var:
			// Do nothing.
		case *node.SimpleVar:
			delete(b.unusedVars, v.Name)
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
	case *node.Var:
		// Do nothing.
	case *node.SimpleVar:
		delete(b.unusedVars, v.Name)
	case *expr.ArrayDimFetch:
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
func (b *BlockWalker) withNewContext(action func()) *blockContext {
	oldCtx := b.ctx
	newCtx := copyBlockContext(b.ctx)

	b.ctx = newCtx
	action()
	b.ctx = oldCtx

	return newCtx
}

func (b *BlockWalker) handleTry(s *stmt.Try) bool {
	if len(s.Catches) == 0 && s.Finally == nil {
		b.r.Report(s, LevelError, "bareTry", "At least one catch or finally block must be present")
	}

	b.r.checkKeywordCase(s, "try")

	contexts := make([]*blockContext, 0, len(s.Catches)+1)

	// Assume that no code in try{} block has executed because exceptions can be thrown from anywhere.
	// So we handle catches and finally blocks first.
	for _, c := range s.Catches {
		b.r.checkKeywordCase(c, "catch")
		ctx := b.withNewContext(func() {
			b.r.addScope(c, b.ctx.sc)
			cc := c.(*stmt.Catch)
			for _, s := range cc.Stmts {
				b.addStatement(s)
			}
			b.handleCatch(cc)
		})
		contexts = append(contexts, ctx)
	}

	if s.Finally != nil {
		b.r.checkKeywordCase(s.Finally, "finally")
		b.withNewContext(func() {
			contexts = append(contexts, b.ctx)
			b.r.addScope(s.Finally, b.ctx.sc)
			cc := s.Finally.(*stmt.Finally)
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
			b.r.addScope(s, b.ctx.sc)
		}
	})

	ctx.sc.Iterate(func(varName string, typ meta.TypesMap, flags meta.VarFlags) {
		if !othersExit {
			flags &^= meta.VarAlwaysDefined
		}
		b.ctx.sc.AddVarName(varName, typ, "try var", flags)
	})

	if othersExit && ctx.exitFlags != 0 {
		b.ctx.exitFlags |= prematureExitFlags
		b.ctx.exitFlags |= ctx.exitFlags
	}

	b.ctx.containsExitFlags |= ctx.containsExitFlags

	return false
}

func (b *BlockWalker) handleCatch(s *stmt.Catch) bool {
	m := meta.NewEmptyTypesMap(len(s.Types))
	for _, t := range s.Types {
		typ, ok := solver.GetClassName(b.r.ctx.st, t)
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
	case *node.SimpleVar:
		delete(b.unusedVars, v.Name)
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

	typ := solver.ExprType(b.ctx.sc, b.r.ctx.st, s.Variable)

	var (
		maybeHaveClasses bool
		haveArrayAccess  bool
	)

	typ.Iterate(func(t string) {
		// FullyQualified class name will have "\" in the beginning
		if len(t) > 0 && t[0] == '\\' && !strings.HasSuffix(t, "[]") {
			if strings.HasPrefix(t, `\shape$`) {
				return
			}

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

func (b *BlockWalker) enoughArgs(args []node.Node, fn meta.FuncInfo) bool {
	if len(args) < fn.MinParamsCnt {
		// If the last argument is ...$arg, then assume it is an array with
		// sufficient values for the parameters
		if len(args) == 0 || !args[len(args)-1].(*node.Argument).Variadic {
			return false
		}
	}
	return true
}

func (b *BlockWalker) handleArgsCount(n node.Node, args []node.Node, fn meta.FuncInfo) {
	switch {
	case meta.NameNodeEquals(n, "mt_rand"):
		if len(args) != 0 && len(args) != 2 {
			b.r.Report(n, LevelWarning, "argCount", "mt_rand expects 0 or 2 args")
		}
		return
	}

	if !b.enoughArgs(args, fn) {
		b.r.Report(n, LevelWarning, "argCount", "Too few arguments for %s", meta.NameNodeToString(n))
	}
}

func (b *BlockWalker) handleCallArgs(n node.Node, args []node.Node, fn meta.FuncInfo) {
	b.handleArgsCount(n, args, fn)

	for i, arg := range args {
		if i >= len(fn.Params) {
			arg.Walk(b)
			continue
		}

		ref := fn.Params[i].IsRef

		switch a := arg.(*node.Argument).Expr.(type) {
		case *node.Var, *node.SimpleVar:
			if ref {
				b.addNonLocalVar(a)
				// TODO: variable may actually not be set by ref
				b.addVar(a, fn.Params[i].Typ, "call_with_ref", meta.VarAlwaysDefined)
				break
			}
			a.Walk(b)
		case *expr.ArrayDimFetch:
			if ref {
				b.handleDimFetchLValue(a, "call_with_ref", meta.MixedType)
				break
			}
			a.Walk(b)
		default:
			a.Walk(b)
		}
	}
}

func (b *BlockWalker) handleFunctionCall(e *expr.FunctionCall) bool {
	call := resolveFunctionCall(b.ctx.sc, b.r.ctx.st, b.ctx.customTypes, e)

	if meta.IsIndexingComplete() {
		if !call.canAnalyze {
			return true
		}

		if !call.defined && !b.ctx.customFunctionExists(e.Function) {
			b.r.Report(e.Function, LevelError, "undefined", "Call to undefined function %s", meta.NameNodeToString(e.Function))
		}
		b.r.checkNameCase(e.Function, call.fqName, call.info.Name)
	}

	if call.info.Doc.Deprecated {
		if call.info.Doc.DeprecationNote != "" {
			b.r.Report(e.Function, LevelDoNotReject, "deprecated", "Call to deprecated function %s (%s)",
				meta.NameNodeToString(e.Function), call.info.Doc.DeprecationNote)
		} else {
			b.r.Report(e.Function, LevelDoNotReject, "deprecated", "Call to deprecated function %s",
				meta.NameNodeToString(e.Function))
		}
	}

	e.Function.Walk(b)

	switch call.fqName {
	case `\preg_match`, `\preg_match_all`, `\preg_replace`, `\preg_split`:
		argNode := e.ArgumentList.Arguments[0]
		s, ok := argNode.(*node.Argument).Expr.(*scalar.String)
		if !ok {
			break
		}
		pat, ok := newRegexpPattern(s)
		if !ok {
			break
		}
		simplified := b.r.reSimplifier.simplifyRegexp(pat)
		if simplified != "" {
			b.r.Report(argNode, LevelDoNotReject, "regexpSimplify", "May re-write %s as '%s'",
				s.Value, simplified)
		}
		issues, err := b.r.reVet.CheckRegexp(pat)
		if err != nil {
			b.r.Report(argNode, LevelError, "regexpSyntax", "parse error: %v", err)
		}
		for _, issue := range issues {
			b.r.Report(argNode, LevelWarning, "regexpVet", "%s", issue)
		}
	}

	if call.fqName == `\compact` {
		b.handleCompactCallArgs(e.ArgumentList.Arguments)
	} else {
		b.handleCallArgs(e.Function, e.ArgumentList.Arguments, call.info)
	}
	b.ctx.exitFlags |= call.info.ExitFlags

	return false
}

// handleCompactCallArgs treats strings anywhere in the argument list as uses
// of the variables named by those strings, which is how compact() behaves.
func (b *BlockWalker) handleCompactCallArgs(args []node.Node) {
	// Recursively flatten the argument list and extract strings
	var strs []*scalar.String
	for len(args) > 0 {
		var head node.Node
		head, args = args[0], args[1:]
		switch n := head.(type) {
		case *node.Argument:
			args = append(args, n.Expr)
		case *expr.Array:
			for _, item := range n.Items {
				args = append(args, item)
			}
		case *expr.ArrayItem:
			args = append(args, n.Val)
		case *scalar.String:
			strs = append(strs, n)
		}
	}

	for _, s := range strs {
		v := node.NewSimpleVar(unquote(s.Value))
		v.SetPosition(s.GetPosition())
		b.handleVariable(v)
	}
}

// checks whether or not we can access to className::method/property/constant/etc from this context
func (b *BlockWalker) canAccess(className string, accessLevel meta.AccessLevel) bool {
	switch accessLevel {
	case meta.Private:
		return b.r.ctx.st.CurrentClass == className
	case meta.Protected:
		if b.r.ctx.st.CurrentClass == className {
			return true
		}

		// TODO: perhaps shpuld extract this common logic with visited map somewhere
		visited := make(map[string]struct{}, 8)
		parent := b.r.ctx.st.CurrentParentClass
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
		className   string
	)

	exprType := b.exprType(e.Variable)

	exprType.Find(func(typ string) bool {
		m, ok := solver.FindMethod(typ, methodName)
		fn = m.Info
		foundMethod = ok
		className = m.ClassName
		magic = haveMagicMethod(typ, `__call`)
		return foundMethod || magic
	})

	e.Variable.Walk(b)
	e.Method.Walk(b)

	if !foundMethod && !magic && !b.r.ctx.st.IsTrait && !b.isThisInsideClosure(e.Variable) {
		// The method is undefined but we permit calling it if `method_exists`
		// was called prior to that call.
		if !b.ctx.customMethodExists(e.Variable, methodName) {
			b.r.Report(e.Method, LevelError, "undefined", "Call to undefined method {%s}->%s()", exprType, methodName)
		}
	} else if !magic {
		// Method is defined.
		b.r.checkNameCase(e.Method, methodName, fn.Name)
		if fn.IsStatic() {
			b.r.Report(e.Method, LevelWarning, "callStatic", "Calling static method as instance method")
		}
	}

	if fn.Doc.Deprecated {
		if fn.Doc.DeprecationNote != "" {
			b.r.Report(e.Method, LevelDoNotReject, "deprecated", "Call to deprecated method {%s}->%s() (%s)",
				exprType, methodName, fn.Doc.DeprecationNote)
		} else {
			b.r.Report(e.Method, LevelDoNotReject, "deprecated", "Call to deprecated method {%s}->%s()",
				exprType, methodName)
		}
	}

	if foundMethod && !b.canAccess(className, fn.AccessLevel) {
		b.r.Report(e.Method, LevelError, "accessLevel", "Cannot access %s method %s->%s()", fn.AccessLevel, className, methodName)
	}

	b.handleCallArgs(e.Method, e.ArgumentList.Arguments, fn)
	b.ctx.exitFlags |= fn.ExitFlags

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

	classNameNode, ok := e.Class.(*name.Name)
	parentCall := ok && meta.NameEquals(classNameNode, "parent")
	if parentCall && methodName == "__construct" {
		b.callsParentConstructor = true
	}

	className, ok := solver.GetClassName(b.r.ctx.st, e.Class)
	if !ok {
		return true
	}

	m, ok := solver.FindMethod(className, methodName)
	fn := m.Info

	e.Class.Walk(b)
	e.Call.Walk(b)

	magic := haveMagicMethod(className, `__callStatic`)
	if !ok && !magic && !b.r.ctx.st.IsTrait {
		b.r.Report(e.Call, LevelError, "undefined", "Call to undefined method %s::%s()", className, methodName)
	} else {
		// Method is defined.

		// parent::f() is permitted.
		if !parentCall && !fn.IsStatic() && !magic {
			b.r.Report(e.Call, LevelWarning, "callStatic", "Calling instance method as static method")
		}
	}

	if ok && !b.canAccess(m.ClassName, fn.AccessLevel) {
		b.r.Report(e.Call, LevelError, "accessLevel", "Cannot access %s method %s::%s()", fn.AccessLevel, m.ClassName, methodName)
	}

	b.handleCallArgs(e.Call, e.ArgumentList.Arguments, fn)
	b.ctx.exitFlags |= fn.ExitFlags

	return false
}

func (b *BlockWalker) isThisInsideClosure(varNode node.Node) bool {
	if !b.ctx.sc.IsInClosure() {
		return false
	}

	variable, ok := varNode.(*node.SimpleVar)
	if !ok {
		return false
	}
	return variable.Name == `this`
}

func (b *BlockWalker) handlePropertyFetch(e *expr.PropertyFetch) bool {
	e.Variable.Walk(b)
	e.Property.Walk(b)

	if !meta.IsIndexingComplete() {
		return false
	}

	id, ok := e.Property.(*node.Identifier)
	if !ok {
		return false
	}

	found := false
	magic := false
	var className string
	var info meta.PropertyInfo

	typ := b.exprType(e.Variable)
	typ.Find(func(typ string) bool {
		p, ok := solver.FindProperty(typ, id.Value)
		info = p.Info
		className = p.ClassName
		found = ok
		magic = haveMagicMethod(typ, `__get`)
		return found || magic
	})

	if !found && !magic && !b.r.ctx.st.IsTrait && !b.isThisInsideClosure(e.Variable) {
		b.r.Report(e.Property, LevelError, "undefined", "Property {%s}->%s does not exist", typ, id.Value)
	}

	if found && !b.canAccess(className, info.AccessLevel) {
		b.r.Report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s->%s", info.AccessLevel, className, id.Value)
	}

	return false
}

func (b *BlockWalker) handleStaticPropertyFetch(e *expr.StaticPropertyFetch) bool {
	e.Class.Walk(b)

	if !meta.IsIndexingComplete() {
		return false
	}

	sv, ok := e.Property.(*node.SimpleVar)
	if !ok {
		vv := e.Property.(*node.Var)
		vv.Expr.Walk(b)
		return false
	}

	className, ok := solver.GetClassName(b.r.ctx.st, e.Class)
	if !ok {
		return false
	}

	p, ok := solver.FindProperty(className, "$"+sv.Name)
	if !ok && !b.r.ctx.st.IsTrait {
		b.r.Report(e.Property, LevelError, "undefined", "Property %s::$%s does not exist", className, sv.Name)
	}

	if ok && !b.canAccess(p.ClassName, p.Info.AccessLevel) {
		b.r.Report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s::$%s", p.Info.AccessLevel, p.ClassName, sv.Name)
	}

	return false
}

func (b *BlockWalker) handleArray(arr *expr.Array) bool {
	if !arr.ShortSyntax {
		b.r.Report(arr, LevelDoNotReject, "arraySyntax", "Use of old array syntax (use short form instead)")
	}
	return b.handleArrayItems(arr, arr.Items)
}

func (b *BlockWalker) handleArrayItems(arr node.Node, items []*expr.ArrayItem) bool {
	haveKeys := false
	haveImplicitKeys := false
	keys := make(map[string]struct{}, len(items))

	for _, item := range items {
		if item.Val == nil {
			continue
		}
		item.Val.Walk(b)

		if item.Key == nil {
			haveImplicitKeys = true
			continue
		}
		item.Key.Walk(b)

		haveKeys = true

		if !isValidArrayKey(item.Key) || !b.sideEffectFree(item.Key) {
			continue
		}

		var keyString string
		var warningKeyString string

		switch k := item.Key.(type) {
		case *scalar.String:
			keyString = unquote(k.Value)
			warningKeyString = k.Value
		case *scalar.Lnumber:
			keyString = numStringToDecimal(k.Value)
			warningKeyString = k.Value
		case *scalar.Dnumber:
			keyString = numStringToDecimal(k.Value)
			warningKeyString = k.Value
		case *expr.ConstFetch:
			if name, _, defined := solver.GetConstant(b.r.ctx.st, k.Constant); !defined {
				continue
			} else {
				keyString = name
				warningKeyString = astutil.FmtNode(k.Constant)
			}
		default:
			keyString = astutil.FmtNode(item.Key)
			warningKeyString = keyString
		}

		if _, ok := keys[keyString]; ok {
			b.r.Report(item.Key, LevelWarning, "dupArrayKeys", "Duplicate array key '%s'", warningKeyString)
		}

		keys[keyString] = struct{}{}
	}

	if haveImplicitKeys && haveKeys {
		b.r.Report(arr, LevelWarning, "mixedArrayKeys", "Mixing implicit and explicit array keys")
	}

	return false
}

func (b *BlockWalker) handleClassConstFetch(e *expr.ClassConstFetch) bool {
	if !meta.IsIndexingComplete() {
		return true
	}

	constName := e.ConstantName
	if constName.Value == `class` || constName.Value == `CLASS` {
		return false
	}

	className, ok := solver.GetClassName(b.r.ctx.st, e.Class)
	if !ok {
		return false
	}

	info, implClass, ok := solver.FindConstant(className, constName.Value)

	e.Class.Walk(b)

	if !ok && !b.r.ctx.st.IsTrait {
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

	_, _, defined := solver.GetConstant(b.r.ctx.st, e.Constant)

	if !defined {
		// If it's builtin constant, give a more precise report message.
		switch nm := meta.NameNodeToString(e.Constant); strings.ToLower(nm) {
		case "null", "true", "false":
			// TODO(quasilyte): should probably issue not "undefined" warning
			// here, but something else, like "constCase" or something.
			// Since it *was* "undefined" before, leave it as is for now,
			// only make error message more user-friendly.
			lcName := strings.ToLower(nm)
			b.r.Report(e.Constant, LevelError, "undefined", "Use %s instead of %s", lcName, nm)
		default:
			b.r.Report(e.Constant, LevelError, "undefined", "Undefined constant %s", nm)
		}
	}

	return true
}

func (b *BlockWalker) handleNew(e *expr.New) bool {
	// Can't handle `new class() ...` yet.
	if _, ok := e.Class.(*stmt.Class); ok {
		return false
	}

	if !meta.IsIndexingComplete() {
		return true
	}

	if b.r.ctx.st.IsTrait {
		switch {
		case meta.NameNodeEquals(e.Class, "self"):
			// Don't try to resolve "self" inside trait context.
			return true
		case meta.NameNodeEquals(e.Class, "static"):
			// More or less identical to the "self" case.
			return true
		}
	}

	className, ok := solver.GetClassName(b.r.ctx.st, e.Class)
	if !ok {
		// perhaps something like 'new $class', cannot check this.
		return true
	}

	class, ok := meta.Info.GetClass(className)
	if !ok {
		b.r.reportUndefinedType(e.Class, className)
	} else {
		b.r.checkNameCase(e.Class, className, class.Name)
	}

	// It's illegal to instantiate abstract class, but `static` can
	// resolve to something else due to the late static binding,
	// so it's the only exception to that rule.
	if class.IsAbstract() && !meta.NameNodeEquals(e.Class, "static") {
		b.r.Report(e.Class, LevelError, "newAbstract", "Cannot instantiate abstract class")
	}

	// Check implicitly invoked constructor method arguments count.
	m, ok := solver.FindMethod(className, "__construct")
	if !ok {
		return true
	}
	ctor := m.Info
	// If new expression is written without (), ArgumentList will be nil.
	// It's equivalent of 0 arguments constructor call.
	var args []node.Node
	if e.ArgumentList != nil {
		args = e.ArgumentList.Arguments
	}
	if ok && !b.enoughArgs(args, ctor) {
		b.r.Report(e, LevelError, "argCount", "Too few arguments for %s constructor", className)
	}

	return true
}

func (b *BlockWalker) handleForeach(s *stmt.Foreach) bool {
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

			b.handleVariableNode(s.Key, meta.TypesMap{}, "foreach_key")
			if list, ok := s.Variable.(*expr.List); ok {
				for _, item := range list.Items {
					b.handleVariableNode(item.Val, meta.TypesMap{}, "foreach_value")
				}
			} else {
				b.handleVariableNode(s.Variable, meta.TypesMap{}, "foreach_value")
			}

			b.ctx.innermostLoop = loopFor
			b.ctx.insideLoop = true
			if _, ok := s.Stmt.(*stmt.StmtList); !ok {
				b.addStatement(s.Stmt)
			}
			s.Stmt.Walk(b)
		})

		b.maybeAddAllVars(ctx.sc, "foreach body")
		b.propagateFlags(ctx)
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

func (b *BlockWalker) enterClosure(fun *expr.Closure, haveThis bool, thisType meta.TypesMap) bool {
	sc := meta.NewScope()
	sc.SetInClosure(true)

	if haveThis {
		sc.AddVarName("this", thisType, "closure inside instance method", meta.VarAlwaysDefined)
	} else {
		sc.AddVarName("this", meta.NewTypesMap("possibly_late_bound"), "possibly late bound $this", meta.VarAlwaysDefined)
	}

	doc := b.r.parsePHPDoc(fun, fun.PhpDocComment, fun.Params)
	b.r.reportPhpdocErrors(fun, doc.errs)
	phpDocParamTypes := doc.types

	var closureUses []node.Node
	if fun.ClosureUse != nil {
		closureUses = fun.ClosureUse.Uses
	}
	for _, useExpr := range closureUses {
		var byRef bool
		var v *node.SimpleVar
		switch u := useExpr.(type) {
		case *expr.Reference:
			v = u.Variable.(*node.SimpleVar)
			byRef = true
		case *node.SimpleVar:
			v = u
		}

		if !b.ctx.sc.HaveVar(v) && !byRef {
			b.r.Report(v, LevelWarning, "undefined", "Undefined variable %s", v.Name)
		}

		typ, ok := b.ctx.sc.GetVarNameType(v.Name)
		if ok {
			sc.AddVarName(v.Name, typ, "use", meta.VarAlwaysDefined)
		}

		delete(b.unusedVars, v.Name)
	}

	params, _ := b.r.parseFuncArgs(fun.Params, phpDocParamTypes, sc)

	b.r.handleFuncStmts(params, closureUses, fun.Stmts, sc)
	b.r.addScope(fun, sc)

	return false
}

func (b *BlockWalker) maybeAddAllVars(sc *meta.Scope, reason string) {
	sc.Iterate(func(varName string, typ meta.TypesMap, flags meta.VarFlags) {
		flags &^= meta.VarAlwaysDefined
		b.ctx.sc.AddVarName(varName, typ, reason, flags)
	})
}

func (b *BlockWalker) handleWhile(s *stmt.While) bool {
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

func (b *BlockWalker) handleDo(s *stmt.Do) bool {
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
func (b *BlockWalker) propagateFlags(other *blockContext) {
	b.ctx.containsExitFlags |= other.containsExitFlags
}

// Propagate premature exit flags from visited branches ("contexts").
func (b *BlockWalker) propagateFlagsFromBranches(contexts []*blockContext, linksCount int) {
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

// andWalker walks if conditions and adds isset/!empty/instanceof variables
// to the associated block walker.
//
// All variables defined by andWalker should be removed after
// if body is handled, this is why we collect varsToDelete.
type andWalker struct {
	b *BlockWalker

	varsToDelete []node.Node
}

func (a *andWalker) EnterNode(w walker.Walkable) (res bool) {
	switch n := w.(type) {
	case *expr.FunctionCall:
		args := n.ArgumentList.Arguments
		nm, ok := n.Function.(*name.Name)
		if !ok {
			break
		}
		switch {
		case len(args) == 2 && meta.NameEquals(nm, `method_exists`):
			obj := args[0].(*node.Argument).Expr
			methodName := args[1].(*node.Argument).Expr
			lit, ok := methodName.(*scalar.String)
			if ok {
				a.b.ctx.addCustomMethod(obj, unquote(lit.Value))
			}
		case len(args) == 1 && meta.NameEquals(nm, `function_exists`):
			functionName := args[0].(*node.Argument).Expr
			lit, ok := functionName.(*scalar.String)
			if ok {
				a.b.ctx.addCustomFunction(unquote(lit.Value))
			}
		}

	case *binary.BooleanAnd:
		return true

	case *expr.Isset:
		for _, v := range n.Variables {
			varNode := findVarNode(v)
			if varNode == nil {
				continue
			}
			if a.b.ctx.sc.HaveVar(varNode) {
				continue
			}

			switch v := varNode.(type) {
			case *node.SimpleVar:
				a.b.addVar(v, meta.NewTypesMap("isset_$"+v.Name), "isset", meta.VarAlwaysDefined)
				a.varsToDelete = append(a.varsToDelete, v)
			case *node.Var:
				a.b.handleVariable(v.Expr)
				vv, ok := v.Expr.(*node.SimpleVar)
				if !ok {
					continue
				}
				a.b.addVar(v, meta.NewTypesMap("isset_$$"+vv.Name), "isset", meta.VarAlwaysDefined)
				a.varsToDelete = append(a.varsToDelete, v)
			}
		}

	case *expr.InstanceOf:
		if className, ok := solver.GetClassName(a.b.r.ctx.st, n.Class); ok {
			switch v := n.Expr.(type) {
			case *node.Var, *node.SimpleVar:
				a.b.ctx.sc.AddVar(v, meta.NewTypesMap(className), "instanceof", 0)
			default:
				a.b.ctx.customTypes = append(a.b.ctx.customTypes, solver.CustomType{
					Node: n.Expr,
					Typ:  meta.NewTypesMap(className),
				})
			}
			// TODO: actually this needs to be present inside if body only
		}

	case *expr.BooleanNot:
		// TODO: consolidate with issets handling?
		// Probably could collect *expr.Variable instead of
		// isset and empty nodes and handle them in a single loop.

		// !empty($x) implies that isset($x) would return true.
		empty, ok := n.Expr.(*expr.Empty)
		if !ok {
			break
		}
		v, ok := empty.Expr.(*node.SimpleVar)
		if !ok {
			break
		}
		if a.b.ctx.sc.HaveVar(v) {
			break
		}
		a.b.addVar(v, meta.NewTypesMap("isset_$"+v.Name), "!empty", meta.VarAlwaysDefined)
		a.varsToDelete = append(a.varsToDelete, v)
	}

	w.Walk(a.b)
	return false
}

func (a *andWalker) LeaveNode(w walker.Walkable) {}

func (b *BlockWalker) handleVariable(v node.Node) bool {
	switch v := v.(type) {
	case *node.Var:
		if vv, ok := v.Expr.(*node.SimpleVar); ok {
			delete(b.unusedVars, vv.Name)
		}
	case *node.SimpleVar:
		delete(b.unusedVars, v.Name)
	}

	if !b.ctx.sc.HaveVar(v) {
		b.r.reportUndefinedVariable(v, b.ctx.sc.MaybeHaveVar(v))
		b.ctx.sc.AddVar(v, meta.NewTypesMap("undefined"), "undefined", meta.VarAlwaysDefined)
	}

	return false
}

func (b *BlockWalker) handleTernary(e *expr.Ternary) bool {
	if e.IfTrue == nil {
		return true // Skip `$x ?: $y` expressions
	}

	// Check for `$cond ? $x : $x` which makes no sense.
	if astutil.NodeEqual(e.IfTrue, e.IfFalse) {
		b.r.Report(e, LevelWarning, "dupBranchBody", "then/else operands are identical")
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

func (b *BlockWalker) handleIf(s *stmt.If) bool {
	// Check for `if ($cond) { $x } else { $x }`.
	// Leave more complex if chains to avoid false positives
	// until we get more examples of valid and invalid cases of
	// duplicated branches.
	if len(s.ElseIf) == 0 && s.Else != nil {
		x := s.Stmt
		y := s.Else.(*stmt.Else).Stmt
		if astutil.NodeEqual(x, y) {
			b.r.Report(s, LevelWarning, "dupBranchBody", "duplicated if/else actions")
		}
	}

	var varsToDelete []node.Node
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
	nodeSet := astutil.NewNodeSet()
	walkCond := func(cond node.Node) {
		a := &andWalker{b: b}
		cond.Walk(a)
		if b.sideEffectFree(cond) && !nodeSet.Add(cond) {
			b.r.Report(cond, LevelWarning, "dupCond", "duplicated condition in if-else chain")
		}
		varsToDelete = append(varsToDelete, a.varsToDelete...)
	}

	// first condition is always executed, so run it in base context
	if s.Cond != nil {
		walkCond(s.Cond)
	}

	var contexts []*blockContext

	walk := func(n node.Node) (links int) {
		// handle if (...) smth(); else other_thing(); // without braces
		if els, ok := n.(*stmt.Else); ok {
			b.addStatement(els.Stmt)
		} else if elsif, ok := n.(*stmt.ElseIf); ok {
			b.addStatement(elsif.Stmt)
		} else {
			b.addStatement(n)
		}

		ctx := b.withNewContext(func() {
			if elsif, ok := n.(*stmt.ElseIf); ok {
				walkCond(elsif.Cond)
			}
			n.Walk(b)
			b.r.addScope(n, b.ctx.sc)
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

func (b *BlockWalker) getCaseStmts(c node.Node) (cond node.Node, list []node.Node) {
	switch c := c.(type) {
	case *stmt.Case:
		cond = c.Cond
		list = c.Stmts
	case *stmt.Default:
		list = c.Stmts
	default:
		panic(fmt.Errorf("Unexpected type in switch statement: %T", c))
	}

	return cond, list
}

func (b *BlockWalker) iterateNextCases(cases []node.Node, startIdx int) {
	for i := startIdx; i < len(cases); i++ {
		cond, list := b.getCaseStmts(cases[i])
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

func (b *BlockWalker) handleSwitch(s *stmt.Switch) bool {
	// first condition is always executed, so run it in base context
	if s.Cond != nil {
		s.Cond.Walk(b)
	}

	var contexts []*blockContext

	linksCount := 0
	haveDefault := false
	breakFlags := FlagBreak | FlagContinue

	b.r.nodeSet.Reset()
	for i, c := range s.CaseList.Cases {
		c, ok := c.(*stmt.Case)
		if !ok {
			continue
		}
		if !b.sideEffectFree(c.Cond) {
			continue
		}
		if !b.r.nodeSet.Add(c.Cond) {
			b.r.Report(c.Cond, LevelWarning, "dupCond", "duplicated switch case #%d", i+1)
		}
	}

	for idx, c := range s.CaseList.Cases {
		var list []node.Node

		cond, list := b.getCaseStmts(c)
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
			if idx != len(s.CaseList.Cases)-1 && b.ctx.exitFlags == 0 {
				// allow the fallthrough if appropriate comment is present
				nextCase := s.CaseList.Cases[idx+1]
				if !b.caseHasFallthroughComment(nextCase) {
					b.r.Report(c, LevelInformation, "caseBreak", "Add break or '// fallthrough' to the end of the case")
				}
			}

			if (b.ctx.exitFlags & (^breakFlags)) == 0 {
				linksCount++

				if b.ctx.exitFlags == 0 {
					b.iterateNextCases(s.CaseList.Cases, idx+1)
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
func (b *BlockWalker) handleDimFetchLValue(e *expr.ArrayDimFetch, reason string, typ meta.TypesMap) {
	b.checkArrayDimFetch(e)

	switch v := e.Variable.(type) {
	case *node.Var, *node.SimpleVar:
		arrTyp := meta.NewEmptyTypesMap(typ.Len())
		typ.Iterate(func(t string) {
			arrTyp = arrTyp.AppendString(meta.WrapArrayOf(t))
		})
		b.addVar(v, arrTyp, reason, meta.VarAlwaysDefined)
		b.handleVariable(v)
	case *expr.ArrayDimFetch:
		b.handleDimFetchLValue(v, reason, meta.MixedType)
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
		b.handleDimFetchLValue(v, "assign_array", meta.MixedType)
		a.Expression.Walk(b)
		return false
	case *node.Var, *node.SimpleVar:
		b.addVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expression), "assign", meta.VarAlwaysDefined)
		b.addNonLocalVar(v)
	case *expr.List:
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

func (b *BlockWalker) handleAssignList(items []*expr.ArrayItem) {
	for _, item := range items {
		b.handleVariableNode(item.Val, meta.NewTypesMap("unknown_from_list"), "assign")
	}
}

func (b *BlockWalker) handleBitwiseOp(n, left, right node.Node) {
	// Note: we report `$x & $mask != $y` as a precedence issue even
	// if it can be caught with `typecheckOp` that checks both operand
	// types (bool is not a good operand for bitwise operation).
	//
	// Reporting `invalid types, expected number found bool` is
	// not that helpful, because the root of the problem is precedence.
	// Invalid types are a result of that.

	tok := "|"
	if _, ok := n.(*binary.BitwiseAnd); ok {
		tok = "&"
	}

	hasParens := func(n node.Node) bool {
		return findFreeFloatingToken(n, freefloating.Start, "(") ||
			findFreeFloatingToken(n, freefloating.End, ")")
	}
	checkArg := func(n, arg node.Node, tok string) {
		cmpTok := ""
		switch arg.(type) {
		case *binary.Equal:
			cmpTok = "=="
		case *binary.NotEqual:
			cmpTok = "!="
		case *binary.Identical:
			cmpTok = "==="
		case *binary.NotIdentical:
			cmpTok = "!=="
		}
		if cmpTok != "" && !hasParens(arg) {
			b.r.Report(n, LevelWarning, "precedence", "%s has higher precedence than %s", cmpTok, tok)
		}
	}

	b.checkBinaryDupArgs(n, left, right)
	b.checkBinaryVoidType(left, right)

	if b.exprType(left).Is("bool") && b.exprType(right).Is("bool") {
		b.r.Report(n, LevelWarning, "bitwiseOps",
			"Used %s bitwise op over bool operands, perhaps %s is intended?", tok, tok+tok)
		return
	}
	checkArg(n, left, tok)
	checkArg(n, right, tok)

}

func (b *BlockWalker) handleStmtExpression(s *stmt.Expression) {
	if !meta.IsIndexingComplete() {
		return
	}

	report := false

	// All branches except default try to filter-out common
	// cases to reduce the number of type solving performed.
	if astutil.IsAssign(s.Expr) {
		return
	}
	switch s.Expr.(type) {
	case *expr.Require, *expr.RequireOnce, *expr.Include, *expr.IncludeOnce, *expr.Exit:
		// Skip.
	case *expr.Array, *expr.New:
		// Report these even if they are not pure.
		report = true
	default:
		typ := b.exprType(s.Expr)
		if !typ.Is("void") {
			report = b.sideEffectFree(s.Expr)
		}
	}

	if report {
		ff := s.GetFreeFloating()
		if ff != nil {
			for _, tok := range (*ff)[freefloating.Expr] {
				if tok.StringType == freefloating.CommentType {
					return
				}
			}
		}

		b.r.Report(s.Expr, LevelWarning, "discardExpr", "expression evaluated but not used")
	}

}

func (b *BlockWalker) handleAssign(a *assign.Assign) bool {
	a.Expression.Walk(b)

	switch v := a.Variable.(type) {
	case *expr.ArrayDimFetch:
		typ := solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expression)
		b.handleDimFetchLValue(v, "assign_array", typ)
		return false
	case *node.Var, *node.SimpleVar:
		b.checkVoidType(a.Expression)
		b.replaceVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.ctx.st, a.Expression), "assign", meta.VarAlwaysDefined)
	case *expr.List:
		b.handleAssignList(v.Items)
	case *expr.PropertyFetch:
		v.Property.Walk(b)
		sv, ok := v.Variable.(*node.SimpleVar)
		if !ok {
			v.Variable.Walk(b)
			break
		}

		delete(b.unusedVars, sv.Name)

		if sv.Name != "this" {
			break
		}

		if b.r.ctx.st.CurrentClass == "" {
			break
		}

		propertyName, ok := v.Property.(*node.Identifier)
		if !ok {
			break
		}

		cls := b.r.getClass()

		p := cls.Properties[propertyName.Value]
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.ctx.sc, b.r.ctx.st, a.Expression, b.ctx.customTypes))
		cls.Properties[propertyName.Value] = p
	case *expr.StaticPropertyFetch:
		sv, ok := v.Property.(*node.SimpleVar)
		if !ok {
			vv := v.Property.(*node.Var)
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

func (b *BlockWalker) flushUnused() {
	if !meta.IsIndexingComplete() {
		return
	}

	visitedMap := make(map[node.Node]struct{})
	for name, nodes := range b.unusedVars {
		if IsDiscardVar(name) {
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
			b.r.Report(n, LevelUnused, "unused", `Variable %s is unused (use $_ to ignore this inspection)`, name)
		}
	}
}

func (b *BlockWalker) handleVariableNode(n node.Node, typ meta.TypesMap, what string) {
	if n == nil {
		return
	}

	var vv node.Node
	switch n := n.(type) {
	case *node.Var, *node.SimpleVar:
		vv = n
	case *expr.Reference:
		vv = n.Variable
	default:
		return
	}

	b.addVar(vv, typ, what, meta.VarAlwaysDefined)
}

// LeaveNode is called after all children have been visited.
func (b *BlockWalker) LeaveNode(w walker.Walkable) {
	for _, c := range b.custom {
		c.BeforeLeaveNode(w)
	}

	if b.ctx.exitFlags == 0 {
		switch w.(type) {
		case *stmt.Return:
			b.ctx.exitFlags |= FlagReturn
			b.ctx.containsExitFlags |= FlagReturn
		case *expr.Exit:
			b.ctx.exitFlags |= FlagDie
			b.ctx.containsExitFlags |= FlagDie
		case *stmt.Throw:
			b.ctx.exitFlags |= FlagThrow
			b.ctx.containsExitFlags |= FlagThrow
		case *stmt.Continue:
			b.ctx.exitFlags |= FlagContinue
			b.ctx.containsExitFlags |= FlagContinue
		case *stmt.Break:
			b.ctx.exitFlags |= FlagBreak
			b.ctx.containsExitFlags |= FlagBreak
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
	ffs := n.GetFreeFloating()
	if ffs == nil {
		return false
	}
	for _, cs := range *ffs {
		for _, c := range cs {
			if c.StringType == freefloating.CommentType {
				if fallthroughMarkerRegex.MatchString(c.Value) {
					return true
				}
			}
		}
	}
	return false
}

func (b *BlockWalker) sideEffectFree(n node.Node) bool {
	return sideEffectFree(b.ctx.sc, b.r.ctx.st, b.ctx.customTypes, n)
}

func (b *BlockWalker) isVoid(n node.Node) bool {
	return b.exprType(n).Is("void")
}

func (b *BlockWalker) exprType(n node.Node) meta.TypesMap {
	return solver.ExprTypeCustom(b.ctx.sc, b.r.ctx.st, n, b.ctx.customTypes)
}
