package linter

import (
	"fmt"
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
	"github.com/VKCOM/noverify/src/solver"
)

type blockLinter struct {
	walker *BlockWalker
}

func (b *blockLinter) enterNode(n node.Node) {
	switch n := n.(type) {
	case *assign.Assign:
		b.checkAssign(n)

	case *expr.Array:
		b.checkArray(n)

	case *expr.FunctionCall:
		b.checkFunctionCall(n)

	case *expr.New:
		b.checkNew(n)

	case *stmt.Expression:
		b.checkStmtExpression(n)

	case *expr.ConstFetch:
		b.checkConstFetch(n)

	case *expr.Ternary:
		b.checkTernary(n)

	case *stmt.Switch:
		b.checkSwitch(n)

	case *stmt.If:
		b.checkIfStmt(n)

	case *stmt.Global:
		b.checkGlobalStmt(n)

	case *binary.BitwiseAnd:
		b.checkBitwiseOp(n, n.Left, n.Right)
	case *binary.BitwiseOr:
		b.checkBitwiseOp(n, n.Left, n.Right)
	case *binary.BitwiseXor:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.LogicalAnd:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.BooleanAnd:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.LogicalOr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.BooleanOr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.LogicalXor:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.Plus:
		b.checkBinaryVoidType(n.Left, n.Right)
	case *binary.Minus:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.Mul:
		b.checkBinaryVoidType(n.Left, n.Right)
	case *binary.Div:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.Mod:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.Pow:
		b.checkBinaryVoidType(n.Left, n.Right)
	case *binary.Equal:
		b.checkStrictCmp(n, n.Left, n.Right)
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.NotEqual:
		b.checkStrictCmp(n, n.Left, n.Right)
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.Identical:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.NotIdentical:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.Smaller:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.SmallerOrEqual:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.Greater:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *binary.GreaterOrEqual:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *binary.Spaceship:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)

	case *cast.Double:
		b.checkRedundantCast(n.Expr, "float")
	case *cast.Int:
		b.checkRedundantCast(n.Expr, "int")
	case *cast.Bool:
		b.checkRedundantCast(n.Expr, "bool")
	case *cast.String:
		b.checkRedundantCast(n.Expr, "string")
	case *cast.Array:
		b.checkRedundantCastArray(n.Expr)

	case *expr.Clone:
		b.walker.r.checkKeywordCase(n, "clone")
	case *stmt.ConstList:
		b.walker.r.checkKeywordCase(n, "const")
	case *stmt.Goto:
		b.walker.r.checkKeywordCase(n, "goto")
	case *stmt.Throw:
		b.walker.r.checkKeywordCase(n, "throw")
	case *expr.Yield:
		b.walker.r.checkKeywordCase(n, "yield")
	case *expr.YieldFrom:
		b.walker.r.checkKeywordCase(n, "yield")
	case *expr.Include:
		b.walker.r.checkKeywordCase(n, "include")
	case *expr.IncludeOnce:
		b.walker.r.checkKeywordCase(n, "include_once")
	case *expr.Require:
		b.walker.r.checkKeywordCase(n, "require")
	case *expr.RequireOnce:
		b.walker.r.checkKeywordCase(n, "require_once")
	case *stmt.Break:
		b.walker.r.checkKeywordCase(n, "break")
	case *stmt.Return:
		b.walker.r.checkKeywordCase(n, "return")
	case *stmt.Else:
		b.walker.r.checkKeywordCase(n, "else")

	case *stmt.Foreach:
		b.walker.r.checkKeywordCase(n, "foreach")
	case *stmt.For:
		b.walker.r.checkKeywordCase(n, "for")
	case *stmt.While:
		b.walker.r.checkKeywordCase(n, "while")
	case *stmt.Do:
		b.walker.r.checkKeywordCase(n, "do")

	case *stmt.Continue:
		b.checkContinueStmt(n)

	case *scalar.Dnumber:
		b.checkIntOverflow(n)

	case *stmt.Try:
		b.checkTryStmt(n)

	case *stmt.Interface:
		b.checkInterfaceStmt(n)
	}
}

func (b *blockLinter) report(n node.Node, level int, checkName, msg string, args ...interface{}) {
	b.walker.r.Report(n, level, checkName, msg, args...)
}

func (b *blockLinter) checkAssign(a *assign.Assign) {
	b.checkVoidType(a.Expression)
}

func (b *blockLinter) checkTryStmt(s *stmt.Try) {
	if len(s.Catches) == 0 && s.Finally == nil {
		b.report(s, LevelError, "bareTry", "At least one catch or finally block must be present")
	}

	b.walker.r.checkKeywordCase(s, "try")

	for _, c := range s.Catches {
		b.walker.r.checkKeywordCase(c, "catch")
	}

	if s.Finally != nil {
		b.walker.r.checkKeywordCase(s.Finally, "finally")
	}
}

func (b *blockLinter) checkBitwiseOp(n, left, right node.Node) {
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
			b.report(n, LevelWarning, "precedence", "%s has higher precedence than %s", cmpTok, tok)
		}
	}

	b.checkBinaryDupArgs(n, left, right)
	b.checkBinaryVoidType(left, right)

	if b.walker.exprType(left).Is("bool") && b.walker.exprType(right).Is("bool") {
		b.report(n, LevelWarning, "bitwiseOps",
			"Used %s bitwise op over bool operands, perhaps %s is intended?", tok, tok+tok)
		return
	}
	checkArg(n, left, tok)
	checkArg(n, right, tok)
}

func (b *blockLinter) checkBinaryVoidType(left, right node.Node) {
	b.checkVoidType(left)
	b.checkVoidType(right)
}

func (b *blockLinter) checkBinaryDupArgsNoFloat(n, left, right node.Node) {
	if b.walker.exprType(left).Contains("float") || b.walker.exprType(right).Contains("float") {
		return
	}
	b.checkBinaryDupArgs(n, left, right)
}

func (b *blockLinter) checkBinaryDupArgs(n, left, right node.Node) {
	// Check for `$x <op> $y` where `<op>` is not a correct way to
	// handle identical operands.
	if !b.walker.sideEffectFree(left) || !b.walker.sideEffectFree(right) {
		return
	}
	if nodeEqual(b.walker.r.ctx.st, left, right) {
		b.report(n, LevelWarning, "dupSubExpr", "duplicated operands value in %s expression", binaryOpString(n))
	}
}

func (b *blockLinter) checkStrictCmp(n node.Node, left, right node.Node) {
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
		b.report(n, LevelWarning, "strictCmp", "non-strict comparison with %s (use %s)",
			astutil.FmtNode(badNode), suggest)
	}
}

// checkVoidType reports if node has void type
func (b *blockLinter) checkVoidType(n node.Node) {
	if b.walker.exprType(n).Is("void") {
		b.report(n, LevelDoNotReject, "voidResultUsed", "void function result used")
	}
}

func (b *blockLinter) checkRedundantCastArray(e node.Node) {
	typ := b.walker.exprType(e)
	if typ.Len() == 1 && typ.Is("mixed[]") {
		b.report(e, LevelDoNotReject, "redundantCast", "expression already has array type")
	}
}

func (b *blockLinter) checkRedundantCast(e node.Node, dstType string) {
	typ := b.walker.exprType(e)
	if typ.Len() != 1 {
		return
	}
	typ.Iterate(func(x string) {
		if x == dstType {
			b.report(e, LevelDoNotReject, "redundantCast",
				"expression already has %s type", dstType)
		}
	})
}

func (b *blockLinter) checkNew(e *expr.New) {
	b.walker.r.checkKeywordCase(e, "new")

	// Can't handle `new class() ...` yet.
	if _, ok := e.Class.(*stmt.Class); ok {
		return
	}

	if b.walker.r.ctx.st.IsTrait {
		switch {
		case meta.NameNodeEquals(e.Class, "self"):
			// Don't try to resolve "self" inside trait context.
			return
		case meta.NameNodeEquals(e.Class, "static"):
			// More or less identical to the "self" case.
			return
		}
	}

	className, ok := solver.GetClassName(b.walker.r.ctx.st, e.Class)
	if !ok {
		// perhaps something like 'new $class', cannot check this.
		return
	}

	class, ok := meta.Info.GetClass(className)
	if !ok {
		b.walker.r.reportUndefinedType(e.Class, className)
	} else {
		b.walker.r.checkNameCase(e.Class, className, class.Name)
	}

	// It's illegal to instantiate abstract class, but `static` can
	// resolve to something else due to the late static binding,
	// so it's the only exception to that rule.
	if class.IsAbstract() && !meta.NameNodeEquals(e.Class, "static") {
		b.report(e.Class, LevelError, "newAbstract", "Cannot instantiate abstract class")
	}

	// Check implicitly invoked constructor method arguments count.
	m, ok := solver.FindMethod(className, "__construct")
	if !ok {
		return
	}
	ctor := m.Info
	// If new expression is written without (), ArgumentList will be nil.
	// It's equivalent of 0 arguments constructor call.
	var args []node.Node
	if e.ArgumentList != nil {
		args = e.ArgumentList.Arguments
	}
	if ok && !enoughArgs(args, ctor) {
		b.report(e, LevelError, "argCount", "Too few arguments for %s constructor", className)
	}
}

func (b *blockLinter) checkStmtExpression(s *stmt.Expression) {
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
		typ := b.walker.exprType(s.Expr)
		if !typ.Is("void") {
			report = b.walker.sideEffectFree(s.Expr)
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

		b.report(s.Expr, LevelWarning, "discardExpr", "expression evaluated but not used")
	}
}

func (b *blockLinter) checkConstFetch(e *expr.ConstFetch) {
	_, _, defined := solver.GetConstant(b.walker.r.ctx.st, e.Constant)

	if !defined {
		// If it's builtin constant, give a more precise report message.
		switch nm := meta.NameNodeToString(e.Constant); strings.ToLower(nm) {
		case "null", "true", "false":
			// TODO(quasilyte): should probably issue not "undefined" warning
			// here, but something else, like "constCase" or something.
			// Since it *was* "undefined" before, leave it as is for now,
			// only make error message more user-friendly.
			lcName := strings.ToLower(nm)
			b.report(e.Constant, LevelError, "undefined", "Use %s instead of %s", lcName, nm)
		default:
			b.report(e.Constant, LevelError, "undefined", "Undefined constant %s", nm)
		}
	}
}

func (b *blockLinter) checkTernary(e *expr.Ternary) {
	if e.IfTrue == nil {
		return // Skip `$x ?: $y` expressions
	}

	// Check for `$cond ? $x : $x` which makes no sense.
	if astutil.NodeEqual(e.IfTrue, e.IfFalse) {
		b.report(e, LevelWarning, "dupBranchBody", "then/else operands are identical")
	}
}

func (b *blockLinter) checkGlobalStmt(s *stmt.Global) {
	b.walker.r.checkKeywordCase(s, "global")

	for _, v := range s.Vars {
		v, ok := v.(*node.SimpleVar)
		if !ok {
			continue
		}
		if _, ok := superGlobals[v.Name]; ok {
			b.report(v, LevelWarning, "redundantGlobal", "%s is superglobal", v.Name)
		}
	}
}

func (b *blockLinter) checkSwitch(s *stmt.Switch) {
	nodeSet := &b.walker.r.nodeSet
	nodeSet.Reset()
	for i, c := range s.CaseList.Cases {
		c, ok := c.(*stmt.Case)
		if !ok {
			continue
		}
		if !b.walker.sideEffectFree(c.Cond) {
			continue
		}
		if !nodeSet.Add(c.Cond) {
			b.report(c.Cond, LevelWarning, "dupCond", "duplicated switch case #%d", i+1)
		}
	}
}

func (b *blockLinter) checkIfStmt(s *stmt.If) {
	// Check for `if ($cond) { $x } else { $x }`.
	// Leave more complex if chains to avoid false positives
	// until we get more examples of valid and invalid cases of
	// duplicated branches.
	if len(s.ElseIf) == 0 && s.Else != nil {
		x := s.Stmt
		y := s.Else.(*stmt.Else).Stmt
		if astutil.NodeEqual(x, y) {
			b.report(s, LevelWarning, "dupBranchBody", "duplicated if/else actions")
		}
	}

	b.checkIfStmtDupCond(s)
}

func (b *blockLinter) checkIfStmtDupCond(s *stmt.If) {
	conditions := astutil.NewNodeSet()
	conditions.Add(s.Cond)
	for _, elseif := range s.ElseIf {
		elseif := elseif.(*stmt.ElseIf)
		if !b.walker.sideEffectFree(elseif.Cond) {
			continue
		}
		if !conditions.Add(elseif.Cond) {
			b.report(elseif.Cond, LevelWarning, "dupCond", "duplicated condition in if-else chain")
		}
	}
}

func (b *blockLinter) checkIntOverflow(num *scalar.Dnumber) {
	// If value contains only [0-9], then it's probably the case
	// where lexer parsed int literal as Dnumber due to the overflow.
	for _, ch := range num.Value {
		if ch < '0' || ch > '9' {
			return
		}
	}
	b.report(num, LevelWarning, "intOverflow", "%s will be interpreted as float due to the overflow", num.Value)
}

func (b *blockLinter) checkContinueStmt(c *stmt.Continue) {
	b.walker.r.checkKeywordCase(c, "continue")
	if c.Expr == nil && b.walker.ctx.innermostLoop == loopSwitch {
		b.report(c, LevelError, "caseContinue", "'continue' inside switch is 'break'")
	}
}

func (b *blockLinter) checkArray(arr *expr.Array) {
	if !arr.ShortSyntax {
		b.report(arr, LevelDoNotReject, "arraySyntax", "Use of old array syntax (use short form instead)")
	}

	items := arr.Items
	haveKeys := false
	haveImplicitKeys := false
	keys := make(map[string]struct{}, len(items))

	for _, item := range items {
		if item.Val == nil {
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
			key = unquote(k.Value)
			constKey = true
		case *scalar.Lnumber:
			key = k.Value
			constKey = true
		case *expr.ConstFetch:
			_, info, ok := solver.GetConstant(b.walker.r.ctx.st, k.Constant)

			if !ok {
				continue
			}
			if info.Value.Type == meta.Undefined {
				continue
			}

			value := info.Value.Value

			switch info.Value.Type {
			case meta.Float:
				val, ok := value.(float64)
				if !ok {
					continue
				}
				value := int64(val)
				key = fmt.Sprint(value)

			case meta.Integer:
				key = fmt.Sprint(value)

			case meta.String:
				key = value.(string)
			}

			constKey = true
		}

		if !constKey {
			continue
		}

		if _, ok := keys[key]; ok {
			b.report(item.Key, LevelWarning, "dupArrayKeys", "Duplicate array key '%s'", key)
		}

		keys[key] = struct{}{}
	}

	if haveImplicitKeys && haveKeys {
		b.report(arr, LevelWarning, "mixedArrayKeys", "Mixing implicit and explicit array keys")
	}
}

func (b *blockLinter) checkFunctionCall(e *expr.FunctionCall) {
	fqName, ok := solver.GetFuncName(b.walker.r.ctx.st, e.Function)
	if !ok {
		return
	}

	switch fqName {
	case `\preg_match`, `\preg_match_all`, `\preg_replace`, `\preg_split`:
		if len(e.ArgumentList.Arguments) < 1 {
			break
		}
		b.checkRegexp(e, e.ArgumentList.Arguments[0].(*node.Argument))
	}
}

func (b *blockLinter) checkInterfaceStmt(iface *stmt.Interface) {
	for _, st := range iface.Stmts {
		switch x := st.(type) {
		case *stmt.ClassMethod:
			for _, modifier := range x.Modifiers {
				if strings.EqualFold(modifier.Value, "private") || strings.EqualFold(modifier.Value, "protected") {
					methodName := x.MethodName.Value
					b.report(x, LevelWarning, "nonPublicInterfaceMember", "'%s' can't be %s", methodName, modifier.Value)
				}
			}
		case *stmt.ClassConstList:
			for _, modifier := range x.Modifiers {
				if strings.EqualFold(modifier.Value, "private") || strings.EqualFold(modifier.Value, "protected") {
					for _, constant := range x.Consts {
						constantName := constant.(*stmt.Constant).ConstantName.Value
						b.report(x, LevelWarning, "nonPublicInterfaceMember", "'%s' can't be %s", constantName, modifier.Value)
					}
				}
			}
		}
	}
}

func (b *blockLinter) checkRegexp(e *expr.FunctionCall, arg *node.Argument) {
	s, ok := arg.Expr.(*scalar.String)
	if !ok {
		return
	}
	pat, ok := newRegexpPattern(s)
	if !ok {
		return
	}
	simplified := b.walker.r.reSimplifier.simplifyRegexp(pat)
	if simplified != "" {
		b.report(arg, LevelDoNotReject, "regexpSimplify", "May re-write %s as '%s'",
			s.Value, simplified)
	}
	issues, err := b.walker.r.reVet.CheckRegexp(pat)
	if err != nil {
		b.report(arg, LevelError, "regexpSyntax", "parse error: %v", err)
	}
	for _, issue := range issues {
		b.report(arg, LevelWarning, "regexpVet", "%s", issue)
	}
}
