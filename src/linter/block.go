package linter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/z7zmey/php-parser/freefloating"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/expr/cast"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
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
	returnTypes *meta.TypesMap

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

	// shared state between all blocks
	unusedVars   map[string][]node.Node
	nonLocalVars map[string]struct{} // static, global and other vars that have complex control flow
}

func (b *BlockWalker) EnterChildNode(key string, w walker.Walkable) {}
func (b *BlockWalker) LeaveChildNode(key string, w walker.Walkable) {}
func (b *BlockWalker) EnterChildList(key string, w walker.Walkable) {}
func (b *BlockWalker) LeaveChildList(key string, w walker.Walkable) {}

// Scope returns block-level variable scope if it exists.
func (b *BlockWalker) Scope() *meta.Scope {
	return b.ctx.sc
}

// PrematureExitFlags returns information about whether or not all code branches have exit/return/throw/etc.
// You need to check what exactly you expect from a block to have (or not to have) by checking Flag* bits.
func (b *BlockWalker) PrematureExitFlags() int {
	return b.ctx.exitFlags
}

// RootState returns state that was stored in root context (if any) for use in custom hooks.
func (b *BlockWalker) RootState() map[string]interface{} {
	return b.r.State()
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

func (b *BlockWalker) checkRedundantCastArray(e node.Node) {
	if !meta.IsIndexingComplete() {
		return
	}
	typ := solver.ExprType(b.ctx.sc, b.r.st, e)
	if typ.Len() == 1 && typ.String() == "mixed[]" {
		b.r.Report(e, LevelDoNotReject, "redundantCast", "expression already has array type")
	}
}

func (b *BlockWalker) checkRedundantCast(e node.Node, dstType string) {
	if !meta.IsIndexingComplete() {
		return
	}
	typ := solver.ExprType(b.ctx.sc, b.r.st, e)
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
				b.parseComment(c)
			}
		}
	}

	switch s := w.(type) {
	case *binary.BitwiseAnd:
		b.handleBitwiseAnd(s)
	case *binary.BitwiseOr:
		b.handleBitwiseOr(s)

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
			b.addVar(ev, solver.ExprTypeLocalCustom(b.ctx.sc, b.r.st, v.Expr, b.ctx.customTypes), "static", true)
			b.addNonLocalVar(ev)
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
		isInstance := b.ctx.sc.IsInInstanceMethod()
		if isInstance {
			typ, _ = b.ctx.sc.GetVarNameType("this")
		}
		res = b.enterClosure(s, isInstance, typ)
	case *stmt.Return:
		b.handleReturn(s)
	case *stmt.Continue:
		b.handleContinue(s)
	case *binary.LogicalOr:
		res = b.handleLogicalOr(s)
	default:
		// b.d.debug(`  Statement: %T`, s)
	}

	for _, c := range b.custom {
		c.AfterEnterNode(w)
	}

	return res
}

func (b *BlockWalker) handleReturn(ret *stmt.Return) {
	if ret.Expr == nil {
		// Return without explicit return value.
		b.bareReturn = true
		return
	}
	b.returnsValue = true

	typ := solver.ExprTypeLocalCustom(b.ctx.sc, b.r.st, ret.Expr, b.ctx.customTypes)
	typ.Iterate(func(t string) {
		b.returnTypes = b.returnTypes.AppendString(t)
	})
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

func (b *BlockWalker) addNonLocalVar(v *expr.Variable) {
	name, ok := v.VarName.(*node.Identifier)
	if !ok {
		return
	}

	b.nonLocalVars[name.Value] = struct{}{}
}

// replaceVar must be used to track assignments to conrete var nodes if they are available
func (b *BlockWalker) replaceVar(v *expr.Variable, typ *meta.TypesMap, reason string, alwaysDefined bool) {
	b.ctx.sc.ReplaceVar(v, typ, reason, alwaysDefined)
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
	if !b.ctx.insideLoop {
		b.unusedVars[name.Value] = append(b.unusedVars[name.Value], v)
	}
}

// addVar must be used to track assignments to conrete var nodes if they are available
func (b *BlockWalker) addVar(v *expr.Variable, typ *meta.TypesMap, reason string, alwaysDefined bool) {
	b.ctx.sc.AddVar(v, typ, reason, alwaysDefined)
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
	if !b.ctx.insideLoop {
		b.unusedVars[name.Value] = append(b.unusedVars[name.Value], v)
	}
}

func (b *BlockWalker) parseComment(c freefloating.String) {
	if c.StringType != freefloating.CommentType {
		return
	}
	str := c.Value

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
		b.ctx.sc.AddVarFromPHPDoc(strings.TrimPrefix(varName, "$"), m, "@var")
	}
}

func (b *BlockWalker) handleUnset(s *stmt.Unset) bool {
	for _, v := range s.Vars {
		switch v := v.(type) {
		case *expr.Variable:
			if id, ok := v.VarName.(*node.Identifier); ok {
				delete(b.unusedVars, id.Value)
			}
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

	contexts := make([]*blockContext, 0, len(s.Catches)+1)

	// Assume that no code in try{} block has executed because exceptions can be thrown from anywhere.
	// So we handle catches and finally blocks first.
	for _, c := range s.Catches {
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

	ctx.sc.Iterate(func(varName string, typ *meta.TypesMap, alwaysDefined bool) {
		b.ctx.sc.AddVarName(varName, typ, "try var", alwaysDefined && othersExit)
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

	typ := solver.ExprType(b.ctx.sc, b.r.st, s.Variable)

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
		case *expr.Variable:
			if ref {
				b.addNonLocalVar(a)
				b.addVar(a, fn.Params[i].Typ, "call_with_ref", true /* TODO: variable may actually not be set by ref */)
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
	var fn meta.FuncInfo
	var fqName string

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
				fqName = nameStr
				fn, defined = meta.Info.GetFunction(fqName)
			} else {
				fqName = b.r.st.Namespace + `\` + nameStr
				fn, defined = meta.Info.GetFunction(fqName)
				if !defined && b.r.st.Namespace != "" {
					fqName = `\` + nameStr
					fn, defined = meta.Info.GetFunction(fqName)
				}
			}

		case *name.FullyQualified:
			fqName = meta.FullyQualifiedToString(nm)
			fn, defined = meta.Info.GetFunction(fqName)
		default:
			defined = false

			solver.ExprTypeCustom(b.ctx.sc, b.r.st, nm, b.ctx.customTypes).Iterate(func(typ string) {
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

	if fn.Doc.Deprecated {
		if fn.Doc.DeprecationNote != "" {
			b.r.Report(e.Function, LevelDoNotReject, "deprecated", "Call to deprecated function %s (%s)",
				meta.NameNodeToString(e.Function), fn.Doc.DeprecationNote)
		} else {
			b.r.Report(e.Function, LevelDoNotReject, "deprecated", "Call to deprecated function %s",
				meta.NameNodeToString(e.Function))
		}
	}

	e.Function.Walk(b)

	if fqName == `\compact` {
		b.handleCompactCallArgs(e.ArgumentList.Arguments)
	} else {
		b.handleCallArgs(e.Function, e.ArgumentList.Arguments, fn)
	}
	b.ctx.exitFlags |= fn.ExitFlags

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
			args = append(args, n.Items...)
		case *expr.ShortArray:
			args = append(args, n.Items...)
		case *expr.ArrayItem:
			args = append(args, n.Val)
		case *scalar.String:
			strs = append(strs, n)
		}
	}

	for _, s := range strs {
		id := node.NewIdentifier(unquote(s.Value))
		id.SetPosition(s.GetPosition())
		v := expr.NewVariable(id)
		v.SetPosition(s.GetPosition())
		b.handleVariable(v)
	}
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

	exprType := solver.ExprTypeCustom(b.ctx.sc, b.r.st, e.Variable, b.ctx.customTypes)

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
	} else {
		// Method is defined.

		if fn.Static && !magic {
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

	if foundMethod && !b.canAccess(implClass, fn.AccessLevel) {
		b.r.Report(e.Method, LevelError, "accessLevel", "Cannot access %s method %s->%s()", fn.AccessLevel, implClass, methodName)
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

	className, ok := solver.GetClassName(b.r.st, e.Class)
	if !ok {
		return true
	}

	fn, implClass, ok := solver.FindMethod(className, methodName)

	e.Class.Walk(b)
	e.Call.Walk(b)

	magic := haveMagicMethod(className, `__callStatic`)
	if !ok && !magic && !b.r.st.IsTrait {
		b.r.Report(e.Call, LevelError, "undefined", "Call to undefined method %s::%s()", className, methodName)
	} else {
		// Method is defined.

		// parent::f() is permitted.
		classNameNode, ok := e.Class.(*name.Name)
		parentCall := ok && meta.NameToString(classNameNode) == "parent"
		if !parentCall && !fn.Static && !magic {
			b.r.Report(e.Call, LevelWarning, "callStatic", "Calling instance method as static method")
		}
	}

	if ok && !b.canAccess(implClass, fn.AccessLevel) {
		b.r.Report(e.Call, LevelError, "accessLevel", "Cannot access %s method %s::%s()", fn.AccessLevel, implClass, methodName)
	}

	b.handleCallArgs(e.Call, e.ArgumentList.Arguments, fn)
	b.ctx.exitFlags |= fn.ExitFlags

	return false
}

func (b *BlockWalker) isThisInsideClosure(varNode node.Node) bool {
	if !b.ctx.sc.IsInClosure() {
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
	var implClass string
	var info meta.PropertyInfo

	typ := solver.ExprTypeCustom(b.ctx.sc, b.r.st, e.Variable, b.ctx.customTypes)
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
		varExpr.VarName.Walk(b)
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

		var key string
		var constKey bool

		switch k := item.Key.(type) {
		case *scalar.String:
			key = unquote(k.Value)
			constKey = true
		case *scalar.Lnumber:
			key = k.Value
			constKey = true
		}

		if !constKey {
			continue
		}

		if _, ok := keys[key]; ok {
			b.Report(item.Key, LevelWarning, "dupArrayKeys", "Duplicate array key '%s'", key)
		}

		keys[key] = struct{}{}
	}

	if haveImplicitKeys && haveKeys {
		b.Report(arr, LevelWarning, "mixedArrayKeys", "Mixing implicit and explicit array keys")
	}

	return false
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
		// If it's builtin constant, give a more precise report message.
		switch name := meta.NameNodeToString(e.Constant); strings.ToLower(name) {
		case "null", "true", "false":
			// TODO(quasilyte): should probably issue not "undefined" warning
			// here, but something else, like "constCase" or something.
			// Since it *was* "undefined" before, leave it as is for now,
			// only make error message more user-friendly.
			lcName := strings.ToLower(name)
			b.r.Report(e.Constant, LevelError, "undefined", "Use %s instead of %s", lcName, name)
		default:
			b.r.Report(e.Constant, LevelError, "undefined", "Undefined constant %s", name)
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

	if b.r.st.IsTrait {
		switch {
		case meta.NameNodeEquals(e.Class, "self"):
			// Don't try to resolve "self" inside trait context.
			return true
		case meta.NameNodeEquals(e.Class, "static"):
			// More or less identical to the "self" case.
			return true
		}
	}

	className, ok := solver.GetClassName(b.r.st, e.Class)
	if !ok {
		// perhaps something like 'new $class', cannot check this.
		return true
	}

	if _, ok := meta.Info.GetClass(className); !ok {
		b.r.Report(e.Class, LevelError, "undefined", "Class not found %s", className)
	}

	// Check implicitly invoked constructor method arguments count.
	ctor, _, ok := solver.FindMethod(className, "__construct")
	if !ok {
		return true
	}
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

	b.handleVariableNode(s.Key, nil, "foreach_key")
	if list, ok := s.Variable.(*expr.List); ok {
		for _, item := range list.Items {
			v, ok := item.(*expr.ArrayItem).Val.(*expr.Variable)
			if !ok {
				continue
			}
			b.handleVariableNode(v, nil, "foreach_value")
		}
	} else {
		b.handleVariableNode(s.Variable, nil, "foreach_value")
	}

	// expression is always executed and is executed in base context
	if s.Expr != nil {
		s.Expr.Walk(b)
		solver.ExprTypeLocalCustom(b.ctx.sc, b.r.st, s.Expr, b.ctx.customTypes).Iterate(func(typ string) {
			b.handleVariableNode(s.Variable, meta.NewTypesMap(meta.WrapElemOf(typ)), "foreach_value")
		})
	}

	// foreach body can do 0 cycles so we need a separate context for that
	if s.Stmt != nil {
		ctx := b.withNewContext(func() {
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

func (b *BlockWalker) enterClosure(fun *expr.Closure, haveThis bool, thisType *meta.TypesMap) bool {
	sc := meta.NewScope()
	sc.SetInClosure(true)

	if haveThis {
		sc.AddVarName("this", thisType, "closure inside instance method", true)
	} else {
		sc.AddVarName("this", meta.NewTypesMap("possibly_late_bound"), "possibly late bound $this", true)
	}

	doc := b.r.parsePHPDoc(fun.PhpDocComment, fun.Params)
	b.r.reportPhpdocErrors(fun, doc.errs)
	phpDocParamTypes := doc.types

	var closureUses []node.Node
	if fun.ClosureUse != nil {
		closureUses = fun.ClosureUse.Uses
	}
	for _, useExpr := range closureUses {
		var byRef bool
		var v *expr.Variable
		switch u := useExpr.(type) {
		case *expr.Reference:
			v = u.Variable.(*expr.Variable)
			byRef = true
		case *expr.Variable:
			v = u
		}

		varName := v.VarName.(*node.Identifier).Value

		if !b.ctx.sc.HaveVar(v) && !byRef {
			b.r.Report(v, LevelWarning, "undefined", "Undefined variable %s", varName)
		}

		typ, ok := b.ctx.sc.GetVarNameType(varName)
		if ok {
			sc.AddVarName(varName, typ, "use", true)
		}

		delete(b.unusedVars, varName)
	}

	params, _ := b.r.parseFuncArgs(fun.Params, phpDocParamTypes, sc)

	b.r.handleFuncStmts(params, closureUses, fun.Stmts, sc)
	b.r.addScope(fun, sc)

	return false
}

func (b *BlockWalker) maybeAddAllVars(sc *meta.Scope, reason string) {
	sc.Iterate(func(varName string, typ *meta.TypesMap, alwaysDefined bool) {
		b.ctx.sc.AddVarName(varName, typ, reason, false)
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

	varsToDelete []*expr.Variable
}

func (a *andWalker) EnterNode(w walker.Walkable) (res bool) {
	switch n := w.(type) {
	case *binary.BooleanAnd:
		return true

	case *expr.Isset:
		for _, v := range n.Variables {
			if v, ok := v.(*expr.Variable); ok {
				if a.b.ctx.sc.HaveVar(v) {
					continue
				}
				switch vn := v.VarName.(type) {
				case *node.Identifier:
					a.b.addVar(v, meta.NewTypesMap("isset_$"+vn.Value), "isset", true)
					a.varsToDelete = append(a.varsToDelete, v)
				case *expr.Variable:
					a.b.handleVariable(vn)
					name, ok := vn.VarName.(*node.Identifier)
					if !ok {
						continue
					}
					a.b.addVar(v, meta.NewTypesMap("isset_$$"+name.Value), "isset", true)
					a.varsToDelete = append(a.varsToDelete, v)
				}
			}
		}

	case *expr.InstanceOf:
		if className, ok := solver.GetClassName(a.b.r.st, n.Class); ok {
			if v, ok := n.Expr.(*expr.Variable); ok {
				a.b.ctx.sc.AddVar(v, meta.NewTypesMap(className), "instanceof", false)
			} else {
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
		v, ok := empty.Expr.(*expr.Variable)
		if !ok {
			break
		}
		if a.b.ctx.sc.HaveVar(v) {
			break
		}
		vn, ok := v.VarName.(*node.Identifier)
		if !ok {
			break
		}
		a.b.addVar(v, meta.NewTypesMap("isset_$"+vn.Value), "!empty", true)
		a.varsToDelete = append(a.varsToDelete, v)
	}

	w.Walk(a.b)
	return false
}

func (a *andWalker) GetChildrenVisitor(key string) walker.Visitor { return a }
func (a *andWalker) LeaveNode(w walker.Walkable)                  {}
func (a *andWalker) EnterChildNode(key string, w walker.Walkable) {}
func (a *andWalker) LeaveChildNode(key string, w walker.Walkable) {}
func (a *andWalker) EnterChildList(key string, w walker.Walkable) {}
func (a *andWalker) LeaveChildList(key string, w walker.Walkable) {}

func (b *BlockWalker) handleVariable(v *expr.Variable) bool {
	if !b.ctx.sc.HaveVar(v) {
		b.r.reportUndefinedVariable(v, b.ctx.sc.MaybeHaveVar(v))
		b.ctx.sc.AddVar(v, meta.NewTypesMap("undefined"), "undefined", true)
	} else if id, ok := v.VarName.(*node.Identifier); ok {
		delete(b.unusedVars, id.Value)
	}
	return false
}

func (b *BlockWalker) handleIf(s *stmt.If) bool {
	var varsToDelete []*expr.Variable
	// Remove all isset'ed variables after we're finished with this if statement.
	defer func() {
		for _, v := range varsToDelete {
			b.ctx.sc.DelVar(v, "isset/!empty")
		}
	}()
	walkCond := func(cond node.Node) {
		a := &andWalker{b: b}
		cond.Walk(a)
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

	varTypes := make(map[string]*meta.TypesMap, b.ctx.sc.Len())
	defCounts := make(map[string]int, b.ctx.sc.Len())

	for _, ctx := range contexts {
		if ctx.exitFlags != 0 {
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
		b.ctx.sc.AddVarName(nm, types, "all branches", defCounts[nm] == linksCount)
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

	for idx, c := range s.CaseList.Cases {
		var list []node.Node

		cond, list := b.getCaseStmts(c)
		if cond == nil {
			haveDefault = true
		} else {
			cond.Walk(b)
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

	varTypes := make(map[string]*meta.TypesMap, b.ctx.sc.Len())
	defCounts := make(map[string]int, b.ctx.sc.Len())

	for _, ctx := range contexts {
		b.propagateFlags(ctx)

		cleanFlags := ctx.exitFlags & (^breakFlags)
		if cleanFlags != 0 {
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
		b.ctx.sc.AddVarName(nm, types, "all cases", defCounts[nm] == linksCount)
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
	case *expr.Variable:
		b.addVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.st, a.Expression), "assign", true)
		b.addNonLocalVar(v)
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

func (b *BlockWalker) handleBitwiseAnd(s *binary.BitwiseAnd) {
	if b.isBool(s.Left) && b.isBool(s.Right) {
		b.ReportBitwiseOp(s, "&", "&&")
	}
}

func (b *BlockWalker) handleBitwiseOr(s *binary.BitwiseOr) {
	if b.isBool(s.Left) && b.isBool(s.Right) {
		b.ReportBitwiseOp(s, "|", "||")
	}
}

func (b *BlockWalker) ReportBitwiseOp(s node.Node, op string, rightOp string) {
	b.r.Report(s, LevelWarning, "bitwiseOps",
		"Used %s bitwise op over bool operands, perhaps %s is intended?", op, rightOp)
}

func (b *BlockWalker) handleAssign(a *assign.Assign) bool {
	a.Expression.Walk(b)

	switch v := a.Variable.(type) {
	case *expr.ArrayDimFetch:
		typ := solver.ExprTypeLocal(b.ctx.sc, b.r.st, a.Expression)
		b.handleDimFetchLValue(v, "assign_array", typ)
		return false
	case *expr.Variable:
		b.replaceVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.st, a.Expression), "assign", true)
	case *expr.List:
		b.handleAssignList(v.Items)
	case *expr.ShortList:
		b.handleAssignList(v.Items)
	case *expr.PropertyFetch:
		varNode, ok := v.Variable.(*expr.Variable)
		if !ok {
			v.Variable.Walk(b)
			v.Property.Walk(b)
			break
		}

		v.Property.Walk(b)

		id, ok := varNode.VarName.(*node.Identifier)
		if !ok {
			varNode.VarName.Walk(b)
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
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.ctx.sc, b.r.st, a.Expression, b.ctx.customTypes))
		cls.Properties[propertyName.Value] = p
	case *expr.StaticPropertyFetch:
		varNode, ok := v.Property.(*expr.Variable)
		if !ok {
			break
		}

		id, ok := varNode.VarName.(*node.Identifier)
		if !ok {
			varNode.VarName.Walk(b)
			break
		}

		if b.r.st.CurrentClass == "" {
			break
		}

		className, ok := solver.GetClassName(b.r.st, v.Class)
		if !ok || className != b.r.st.CurrentClass {
			break
		}

		cls := b.r.getClass()

		p := cls.Properties["$"+id.Value]
		p.Typ = p.Typ.Append(solver.ExprTypeLocalCustom(b.ctx.sc, b.r.st, a.Expression, b.ctx.customTypes))
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
			b.r.Report(n, LevelUnused, "unused", `Unused variable %s (use $_ to ignore this inspection)`, name)
		}
	}
}

func (b *BlockWalker) handleVariableNode(n node.Node, typ *meta.TypesMap, what string) {
	if n == nil {
		return
	}

	var vv *expr.Variable
	switch n := n.(type) {
	case *expr.Variable:
		vv = n
	case *expr.Reference:
		vv = n.Variable.(*expr.Variable)
	default:
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

func (b *BlockWalker) isBool(n node.Node) bool {
	return solver.ExprType(b.r.Scope(), b.r.st, n).Is("bool")
}
