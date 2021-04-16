package linter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/constfold"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
)

type blockLinter struct {
	walker *blockWalker
}

func (b *blockLinter) enterNode(n ir.Node) {
	switch n := n.(type) {
	case *ir.Assign:
		b.checkAssign(n)

	case *ir.ArrayExpr:
		b.checkArray(n)

	case *ir.FunctionCallExpr:
		b.checkFunctionCall(n)

	case *ir.MethodCallExpr:
		b.checkMethodCall(n)

	case *ir.StaticCallExpr:
		b.checkStaticCall(n)

	case *ir.PropertyFetchExpr:
		b.checkPropertyFetch(n)

	case *ir.StaticPropertyFetchExpr:
		b.checkStaticPropertyFetch(n)

	case *ir.ClassConstFetchExpr:
		b.checkClassConstFetch(n)

	case *ir.NewExpr:
		b.checkNew(n)

	case *ir.ExpressionStmt:
		b.checkStmtExpression(n)

	case *ir.ConstFetchExpr:
		b.checkConstFetch(n)

	case *ir.TernaryExpr:
		b.checkTernary(n)

	case *ir.SwitchStmt:
		b.checkSwitch(n)

	case *ir.IfStmt:
		b.checkIfStmt(n)

	case *ir.GlobalStmt:
		b.checkGlobalStmt(n)

	case *ir.BitwiseAndExpr:
		b.checkBitwiseOp(n, n.Left, n.Right)
	case *ir.BitwiseOrExpr:
		b.checkBitwiseOp(n, n.Left, n.Right)
	case *ir.BitwiseXorExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.LogicalAndExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.BooleanAndExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.LogicalOrExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.BooleanOrExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.LogicalXorExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.PlusExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
	case *ir.MinusExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.MulExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
	case *ir.DivExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.ModExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.PowExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
	case *ir.EqualExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.NotEqualExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.IdenticalExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.NotIdenticalExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.SmallerExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.SmallerOrEqualExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.GreaterExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgs(n, n.Left, n.Right)
	case *ir.GreaterOrEqualExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.SpaceshipExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
	case *ir.CoalesceExpr:
		b.checkCoalesceExpr(n)
	case *ir.TypeCastExpr:
		b.checkTypeCaseExpr(n)

	case *ir.CloneExpr:
		b.walker.r.checkKeywordCase(n, "clone")
	case *ir.ConstListStmt:
		b.walker.r.checkKeywordCase(n, "const")
	case *ir.GotoStmt:
		b.walker.r.checkKeywordCase(n, "goto")
	case *ir.ThrowStmt:
		b.walker.r.checkKeywordCase(n, "throw")
	case *ir.YieldExpr:
		b.walker.r.checkKeywordCase(n, "yield")
	case *ir.YieldFromExpr:
		b.walker.r.checkKeywordCase(n, "yield")
	case *ir.ImportExpr:
		b.walker.r.checkKeywordCase(n, n.Func)
	case *ir.BreakStmt:
		b.walker.r.checkKeywordCase(n, "break")
	case *ir.ReturnStmt:
		b.walker.r.checkKeywordCase(n, "return")
	case *ir.ElseStmt:
		b.walker.r.checkKeywordCase(n, "else")

	case *ir.ForeachStmt:
		b.walker.r.checkKeywordCase(n, "foreach")
	case *ir.ForStmt:
		b.walker.r.checkKeywordCase(n, "for")
	case *ir.WhileStmt:
		b.walker.r.checkKeywordCase(n, "while")
	case *ir.DoStmt:
		b.walker.r.checkKeywordCase(n, "do")

	case *ir.ContinueStmt:
		b.checkContinueStmt(n)

	case *ir.Dnumber:
		b.checkIntOverflow(n)

	case *ir.TryStmt:
		b.checkTryStmt(n)

	case *ir.InterfaceStmt:
		b.checkInterfaceStmt(n)

	case *ir.NopStmt:
		b.checkNopStmt(n)

	case *ir.BadString:
		b.report(n, LevelError, "syntax", "%s", n.Error)
	}
}

func (b *blockLinter) checkTypeCaseExpr(n *ir.TypeCastExpr) {
	if n.Type == "array" {
		b.checkRedundantCastArray(n.Expr)
	} else {
		b.checkRedundantCast(n.Expr, n.Type)
	}

	// We cannot use the value directly, since for real it is equal to float,
	// so we have to use the token value.
	if bytes.EqualFold(n.CastTkn.Value, []byte("(real)")) {
		b.report(n, LevelNotice, "langDeprecated",
			"use float cast instead of real")
	}
}

func (b *blockLinter) report(n ir.Node, level int, checkName, msg string, args ...interface{}) {
	b.walker.r.Report(n, level, checkName, msg, args...)
}

func (b *blockLinter) checkNopStmt(n *ir.NopStmt) {
	switch b.walker.path.Parent().(type) {
	case *ir.DeclareStmt, *ir.IfStmt, *ir.ForStmt, *ir.ForeachStmt, *ir.WhileStmt, *ir.DoStmt:
		return
	}

	b.report(n, LevelNotice, "emptyStmt", "semicolon (;) is not needed here, it can be safely removed")
}

func (b *blockLinter) checkCoalesceExpr(n *ir.CoalesceExpr) {
	lhsType := solver.ExprType(b.walker.ctx.sc, b.classParseState(), n.Left)
	if !lhsType.IsPrecise() {
		return
	}

	if !lhsType.Contains("null") {
		b.report(n.Right, LevelWarning, "deadCode", "%s is not nullable, right side of the expression is unreachable", irutil.FmtNode(n.Left))
	}
}

func (b *blockLinter) checkAssign(a *ir.Assign) {
	b.checkVoidType(a.Expr)
}

func (b *blockLinter) checkTryStmt(s *ir.TryStmt) {
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

	if len(s.Catches) > 1 {
		b.checkCatchOrder(s)
	}
}

func (b *blockLinter) checkCatchOrder(s *ir.TryStmt) {
	// This code has O(n^2) complexity, but there are usually no more than 3-4 catch clauses in the code.
	// We could avoid some extra work if we would not add leaf types to the classes list,
	// but we don't have that information available.

	classes := make([]string, 0, len(s.Catches))

	for _, c := range s.Catches {
		c := c.(*ir.CatchStmt)
		if len(c.Types) > 1 {
			return // give up on A|B catch
		}

		class, ok := solver.GetClassName(b.classParseState(), c.Types[0])
		if !ok {
			continue
		}

		add := true
		for _, otherClass := range classes {
			if class == otherClass {
				b.report(c.Types[0], LevelWarning, "dupCatch", "duplicated catch on %s", class)
				add = false
				break
			}
			if solver.Extends(b.metaInfo(), class, otherClass) {
				b.report(c.Types[0], LevelWarning, "catchOrder", "catch %s block will never run as it extends %s which is caught above", class, otherClass)
				add = false
				break
			}
			if solver.Implements(b.metaInfo(), class, otherClass) {
				b.report(c.Types[0], LevelWarning, "catchOrder", "catch %s block will never run as it implements %s which is caught above", class, otherClass)
				add = false
				break
			}
		}
		if add {
			classes = append(classes, class)
		}
	}
}

func (b *blockLinter) checkBitwiseOp(n, left, right ir.Node) {
	b.checkBinaryDupArgs(n, left, right)
	b.checkBinaryVoidType(left, right)
}

func (b *blockLinter) checkBinaryVoidType(left, right ir.Node) {
	b.checkVoidType(left)
	b.checkVoidType(right)
}

func (b *blockLinter) checkBinaryDupArgsNoFloat(n, left, right ir.Node) {
	if b.walker.exprType(left).Contains("float") || b.walker.exprType(right).Contains("float") {
		return
	}
	b.checkBinaryDupArgs(n, left, right)
}

func (b *blockLinter) checkBinaryDupArgs(n, left, right ir.Node) {
	// Check for `$x <op> $y` where `<op>` is not a correct way to
	// handle identical operands.
	if !b.walker.sideEffectFree(left) || !b.walker.sideEffectFree(right) {
		return
	}
	if nodeEqual(b.classParseState(), left, right) {
		b.report(n, LevelWarning, "dupSubExpr", "duplicated operands value in %s expression", binaryOpString(n))
	}
}

// checkVoidType reports if node has void type
func (b *blockLinter) checkVoidType(n ir.Node) {
	if b.walker.exprType(n).Is("void") {
		b.report(n, LevelNotice, "voidResultUsed", "void function result used")
	}
}

func (b *blockLinter) checkRedundantCastArray(e ir.Node) {
	typ := b.walker.exprType(e)
	if typ.Len() == 1 && typ.Is("mixed[]") {
		b.report(e, LevelNotice, "redundantCast", "expression already has array type")
	}
}

func (b *blockLinter) checkRedundantCast(e ir.Node, dstType string) {
	typ := b.walker.exprType(e)
	if typ.Len() != 1 {
		return
	}
	typ.Iterate(func(x string) {
		if x == dstType {
			b.report(e, LevelNotice, "redundantCast",
				"expression already has %s type", dstType)
		}
	})
}

func (b *blockLinter) checkNew(e *ir.NewExpr) {
	b.walker.r.checkKeywordCase(e, "new")

	// Can't handle `new class() ...` yet.
	if _, ok := e.Class.(*ir.AnonClassExpr); ok {
		return
	}

	if b.classParseState().IsTrait {
		switch {
		case meta.NameNodeEquals(e.Class, "self"):
			// Don't try to resolve "self" inside trait context.
			return
		case meta.NameNodeEquals(e.Class, "static"):
			// More or less identical to the "self" case.
			return
		}
	}

	className, ok := solver.GetClassName(b.classParseState(), e.Class)
	if !ok {
		// perhaps something like 'new $class', cannot check this.
		return
	}

	class, ok := b.metaInfo().GetClass(className)
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
	m, ok := solver.FindMethod(b.metaInfo(), className, "__construct")
	if !ok {
		return
	}
	ctor := m.Info
	// If new expression is written without (), ArgumentList will be nil.
	// It's equivalent of 0 arguments constructor call.
	if ok && !enoughArgs(e.Args, ctor) {
		b.report(e, LevelError, "argCount", "Too few arguments for %s constructor", className)
	}
}

func (b *blockLinter) checkStmtExpression(s *ir.ExpressionStmt) {
	report := false

	// All branches except default try to filter-out common
	// cases to reduce the number of type solving performed.
	if irutil.IsAssign(s.Expr) {
		return
	}
	switch s.Expr.(type) {
	case *ir.ImportExpr, *ir.ExitExpr:
		// Skip.
	case *ir.ArrayExpr, *ir.NewExpr:
		// Report these even if they are not pure.
		report = true
	default:
		typ := b.walker.exprType(s.Expr)
		if !typ.Is("void") {
			report = b.walker.sideEffectFree(s.Expr)
		}
	}

	if report {
		b.report(s.Expr, LevelWarning, "discardExpr", "expression evaluated but not used")
	}
}

func (b *blockLinter) checkConstFetch(e *ir.ConstFetchExpr) {
	_, _, defined := solver.GetConstant(b.classParseState(), e.Constant)

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

func (b *blockLinter) checkTernary(e *ir.TernaryExpr) {
	if e.IfTrue == nil {
		return // Skip `$x ?: $y` expressions
	}

	_, nestedTernary := e.Condition.(*ir.TernaryExpr)
	if nestedTernary {
		b.report(e.Condition, LevelWarning, "nestedTernary", "in ternary operators, you must explicitly use parentheses to specify the order of operations")
	}

	// Check for `$cond ? $x : $x` which makes no sense.
	if irutil.NodeEqual(e.IfTrue, e.IfFalse) {
		b.report(e, LevelWarning, "dupBranchBody", "then/else operands are identical")
	}
}

func (b *blockLinter) checkGlobalStmt(s *ir.GlobalStmt) {
	b.walker.r.checkKeywordCase(s, "global")

	for _, v := range s.Vars {
		v, ok := v.(*ir.SimpleVar)
		if !ok {
			continue
		}
		if _, ok := superGlobals[v.Name]; ok {
			b.report(v, LevelWarning, "redundantGlobal", "%s is superglobal", v.Name)
		}
	}
}

func (b *blockLinter) checkSwitch(s *ir.SwitchStmt) {
	nodeSet := &b.walker.r.nodeSet
	nodeSet.Reset()
	wasAdded := false
	for i, c := range s.Cases {
		c, ok := c.(*ir.CaseStmt)
		if !ok {
			continue
		}
		if !b.walker.sideEffectFree(c.Cond) {
			continue
		}

		var v meta.ConstValue
		var isConstKey bool
		if k, ok := c.Cond.(*ir.ConstFetchExpr); ok {
			v = constfold.Eval(b.classParseState(), k)
			if !v.IsValid() {
				continue
			}
			value := v.Value

			switch v.Type {
			case meta.Float:
				val, ok := value.(float64)
				if !ok {
					continue
				}
				wasAdded = nodeSet.Add(&ir.Dnumber{Value: fmt.Sprint(val)})
			case meta.Integer:
				val, ok := value.(int64)
				if !ok {
					continue
				}
				wasAdded = nodeSet.Add(&ir.Lnumber{Value: fmt.Sprint(val)})
			case meta.String:
				val, ok := value.(string)
				if !ok {
					continue
				}
				wasAdded = nodeSet.Add(&ir.String{Value: fmt.Sprint(val)})
			case meta.Bool:
				val, ok := value.(bool)
				if !ok {
					continue
				}
				wasAdded = nodeSet.Add(&ir.Name{Value: fmt.Sprint(val)})
			default:
				continue
			}
			isConstKey = true
		}

		isDupKey := isConstKey && !wasAdded
		if !isDupKey {
			isDupKey = !nodeSet.Add(c.Cond)
		}

		if isDupKey {
			msg := fmt.Sprintf("duplicated switch case #%d", i+1)
			if isConstKey {
				dupKey := getConstValue(v)
				msg += " (value " + dupKey + ")"
			}
			b.report(c.Cond, LevelWarning, "dupCond", "%s", msg)
		}
	}
}

func (b *blockLinter) checkIfStmt(s *ir.IfStmt) {
	// Check for `if ($cond) { $x } else { $x }`.
	// Leave more complex if chains to avoid false positives
	// until we get more examples of valid and invalid cases of
	// duplicated branches.
	if len(s.ElseIf) == 0 && s.Else != nil {
		x := s.Stmt
		y := s.Else.(*ir.ElseStmt).Stmt
		if irutil.NodeEqual(x, y) {
			b.report(s, LevelWarning, "dupBranchBody", "duplicated if/else actions")
		}
	}

	b.checkIfStmtDupCond(s)
}

func (b *blockLinter) checkIfStmtDupCond(s *ir.IfStmt) {
	conditions := irutil.NewNodeSet()
	conditions.Add(s.Cond)
	for _, elseif := range s.ElseIf {
		elseif := elseif.(*ir.ElseIfStmt)
		if !b.walker.sideEffectFree(elseif.Cond) {
			continue
		}
		if !conditions.Add(elseif.Cond) {
			b.report(elseif.Cond, LevelWarning, "dupCond", "duplicated condition in if-else chain")
		}
	}
}

func (b *blockLinter) checkIntOverflow(num *ir.Dnumber) {
	// If value contains only [0-9], then it's probably the case
	// where lexer parsed int literal as Dnumber due to the overflow.
	for _, ch := range num.Value {
		if ch < '0' || ch > '9' {
			return
		}
	}
	b.report(num, LevelWarning, "intOverflow", "%s will be interpreted as float due to the overflow", num.Value)
}

func (b *blockLinter) checkContinueStmt(c *ir.ContinueStmt) {
	b.walker.r.checkKeywordCase(c, "continue")
	if c.Expr == nil && b.walker.ctx.innermostLoop == loopSwitch {
		b.report(c, LevelError, "caseContinue", "'continue' inside switch is 'break'")
	}
}

func (b *blockLinter) addFixForArray(arr *ir.ArrayExpr) {
	if !b.walker.r.config.ApplyQuickFixes {
		return
	}

	from := arr.Position.StartPos
	to := arr.Position.EndPos
	have := b.walker.r.file.Contents()[from:to]
	have = bytes.TrimPrefix(have, []byte("array("))
	have = bytes.TrimSuffix(have, []byte(")"))

	b.walker.r.ctx.fixes = append(b.walker.r.ctx.fixes, quickfix.TextEdit{
		StartPos:    arr.Position.StartPos,
		EndPos:      arr.Position.EndPos,
		Replacement: fmt.Sprintf("[%s]", string(have)),
	})
}

func (b *blockLinter) checkArray(arr *ir.ArrayExpr) {
	if !arr.ShortSyntax {
		b.report(arr, LevelNotice, "arraySyntax", "Use of old array syntax (use short form instead)")
		b.addFixForArray(arr)
	}

	multiline := false
	items := arr.Items
	haveKeys := false
	haveImplicitKeys := false
	keys := make(map[string]ir.Node, len(items))

	if arr.Position.EndLine != arr.Position.StartLine {
		multiline = true
	}

	for index, item := range items {
		if item.Val == nil {
			continue
		}

		if multiline && index == len(items)-1 {
			b.checkMultilineArrayTrailingComma(item)
		}

		if item.Key == nil {
			haveImplicitKeys = true
			continue
		}

		haveKeys = true

		var key string
		var constKey bool

		switch k := item.Key.(type) {
		case *ir.String:
			key = k.Value
			constKey = true
		case *ir.Lnumber:
			key = k.Value
			constKey = true
		case *ir.ConstFetchExpr:
			v := constfold.Eval(b.classParseState(), k)
			if !v.IsValid() {
				continue
			}

			value := v.Value
			switch v.Type {
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

			case meta.Bool:
				if value.(bool) {
					key = "1"
				} else {
					key = "0"
				}
			}

			constKey = true
		}

		if !constKey {
			continue
		}

		if n, ok := keys[key]; ok {
			origKey := b.walker.r.nodeText(n)
			dupKey := fmt.Sprintf("%#q", key)
			msg := fmt.Sprintf("Duplicate array key %s", origKey)
			if origKey != dupKey && origKey != key {
				msg += " (value " + dupKey + ")"
			}
			b.report(item.Key, LevelWarning, "dupArrayKeys", "%s", msg)
		}

		keys[key] = item.Key
	}

	if haveImplicitKeys && haveKeys {
		b.report(arr, LevelWarning, "mixedArrayKeys", "Mixing implicit and explicit array keys")
	}
}

func (b *blockLinter) addFixForMultilineArrayTrailingComma(item *ir.ArrayItemExpr) {
	if !b.walker.r.config.ApplyQuickFixes {
		return
	}

	from := item.Position.StartPos
	to := item.Position.EndPos
	have := b.walker.r.file.Contents()[from:to]

	b.walker.r.ctx.fixes = append(b.walker.r.ctx.fixes, quickfix.TextEdit{
		StartPos:    item.Position.StartPos,
		EndPos:      item.Position.EndPos,
		Replacement: string(have) + ",",
	})
}

func (b *blockLinter) checkMultilineArrayTrailingComma(item *ir.ArrayItemExpr) {
	from := item.Position.StartPos
	to := item.Position.EndPos
	src := b.walker.r.file.Contents()

	if to+1 >= len(src) {
		return
	}

	itemText := src[from : to+1]
	if itemText[len(itemText)-1] != ',' && itemText[len(itemText)-1] != ']' {
		b.report(item, LevelNotice, "trailingComma", "last element in a multi-line array must have a trailing comma")
		b.addFixForMultilineArrayTrailingComma(item)
	}
}

func (b *blockLinter) checkDeprecatedFunctionCall(e *ir.FunctionCallExpr, call *funcCallInfo) {
	if !call.info.Doc.Deprecated {
		return
	}

	if call.info.Doc.DeprecationNote != "" {
		b.report(e.Function, LevelNotice, "deprecated", "Call to deprecated function %s (%s)", meta.NameNodeToString(e.Function), call.info.Doc.DeprecationNote)
		return
	}

	b.report(e.Function, LevelNotice, "deprecated", "Call to deprecated function %s", meta.NameNodeToString(e.Function))
}

func (b *blockLinter) checkFunctionAvailability(e *ir.FunctionCallExpr, call *funcCallInfo) {
	if !call.isFound && !b.walker.ctx.customFunctionExists(e.Function) {
		b.report(e.Function, LevelError, "undefined", "Call to undefined function %s", meta.NameNodeToString(e.Function))
	}
}

func (b *blockLinter) checkCallArgs(n ir.Node, args []ir.Node, fn meta.FuncInfo) {
	b.checkCallArgsCount(n, args, fn)
}

func (b *blockLinter) checkCallArgsCount(n ir.Node, args []ir.Node, fn meta.FuncInfo) {
	if fn.Name == `\mt_rand` {
		if len(args) != 0 && len(args) != 2 {
			b.report(n, LevelWarning, "argCount", "mt_rand expects 0 or 2 args")
		}
		return
	}

	if fn.Name == `\compact` || fn.Name == `\func_get_args` {
		// there is no need to check the number of arguments for these functions.
		return
	}

	if !enoughArgs(args, fn) {
		b.report(n, LevelWarning, "argCount", "Too few arguments for %s", meta.NameNodeToString(n))
	}
}

func (b *blockLinter) checkFunctionCall(e *ir.FunctionCallExpr) {
	call := resolveFunctionCall(b.walker.ctx.sc, b.classParseState(), b.walker.ctx.customTypes, e)
	fqName := call.funcName

	if call.canAnalyze {
		b.checkCallArgs(e.Function, e.Args, call.info)
		b.checkDeprecatedFunctionCall(e, &call)
		b.checkFunctionAvailability(e, &call)
		b.walker.r.checkNameCase(e.Function, call.funcName, call.info.Name)
	}

	switch fqName {
	case `\strip_tags`:
		if len(e.Args) < 2 {
			break
		}
		b.checkStripTags(e)
	case `\preg_match`, `\preg_match_all`, `\preg_replace`, `\preg_split`:
		if len(e.Args) < 1 {
			break
		}
		b.checkRegexp(e, e.Arg(0))
	case `\sprintf`, `\printf`:
		if len(e.Args) < 1 {
			break
		}
		// TODO: handle fprintf as well?
		b.checkFormatString(e, e.Arg(0))
	case `\is_real`:
		b.report(e, LevelNotice, "langDeprecated", "use is_float function instead of is_real")
	case `\array_key_exists`:
		b.checkArrayKeyExistsCall(e)
	case `\random_int`:
		b.checkRandomIntCall(e)
	}
}

func (b *blockLinter) checkRandomIntCall(e *ir.FunctionCallExpr) {
	if len(e.Args) < 2 {
		return
	}

	arg1 := constfold.Eval(b.walker.r.ctx.st, e.Arg(0))
	if !arg1.IsValid() {
		return
	}

	arg2 := constfold.Eval(b.walker.r.ctx.st, e.Arg(1))
	if !arg2.IsValid() {
		return
	}

	min, ok := arg1.ToInt()
	if !ok {
		return
	}

	max, ok := arg2.ToInt()
	if !ok {
		return
	}

	if min > max {
		b.report(e, LevelNotice, "argsOrder", "possibly wrong order of arguments, min = %d, max = %d", min, max)
	}
}

func (b *blockLinter) checkArrayKeyExistsCall(e *ir.FunctionCallExpr) {
	if len(e.Args) < 2 {
		return
	}

	typ := solver.ExprType(b.walker.ctx.sc, b.walker.r.ctx.st, e.Arg(1).Expr)

	onlyObjects := !typ.Find(func(typ string) bool {
		return !types.IsClass(typ)
	})

	if onlyObjects {
		b.report(e, LevelWarning, "langDeprecated", "since PHP 7.4, using array_key_exists() with an object has been deprecated, use isset() or property_exists() instead")
	}
}

func (b *blockLinter) checkStripTags(e *ir.FunctionCallExpr) {
	reportArg := func(n ir.Node, format string, args ...interface{}) {
		message := fmt.Sprintf(format, args...)
		b.report(n, LevelWarning, "stripTags", "$allowed_tags argument: "+message)
	}

	normalizeTag := func(s string) string {
		s = strings.ReplaceAll(s, " ", "")
		if strings.HasSuffix(s, "/>") {
			s = strings.TrimSuffix(s, "/>") + ">"
		}
		s = strings.ToLower(s)
		return s
	}

	set := make(map[string]string)
	addTag := func(n ir.Node, tag string) {
		normalized := normalizeTag(tag)
		if prev := set[normalized]; prev != "" {
			if prev == tag {
				reportArg(n, "tag '%s' is duplicated", tag)
			} else {
				reportArg(n, "tag '%s' is duplicated, previously spelled as '%s'", tag, prev)
			}
		} else {
			set[normalized] = tag
		}
	}

	switch allowed := e.Arg(1).Expr.(type) {
	case *ir.ArrayExpr:
		for _, item := range allowed.Items {
			literal, ok := item.Val.(*ir.String)
			if !ok {
				continue
			}
			s := strings.TrimSpace(literal.Value)
			if strings.HasPrefix(s, "<") {
				reportArg(item.Val, "'<' and '>' are not needed for tags when using array argument")
			}
			addTag(literal, literal.Value)
		}
	case *ir.String:
		s := allowed.Value
		if strings.ContainsAny(s, `'"`) {
			reportArg(allowed, "using values/attrs is an error; they make matching always fail")
			break
		}
		for {
			s = strings.TrimLeft(s, " \n\r\t")
			end := strings.IndexByte(s, '>')
			if end == -1 {
				break
			}
			tag := s[:end+1]
			if strings.HasSuffix(tag, "/>") {
				fixed := strings.TrimSuffix(tag, "/>") + ">"
				reportArg(allowed, "'%s' should be written as '%s'", tag, fixed)
			}
			if strings.Contains(tag, " ") {
				reportArg(allowed, "tag '%s' should not contain spaces", tag)
			}
			addTag(allowed, tag)
			s = s[end+1:]
		}
	}
}

func (b *blockLinter) checkMethodCall(e *ir.MethodCallExpr) {
	parseState := b.classParseState()

	call := resolveMethodCall(b.walker.ctx.sc, parseState, b.walker.ctx.customTypes, e)
	if !call.canAnalyze {
		return
	}

	if !call.isMagic {
		b.checkCallArgs(e.Method, e.Args, call.info)
	}

	if !call.isFound && !call.isMagic && !parseState.IsTrait && !b.walker.isThisInsideClosure(e.Variable) {
		// The method is undefined but we permit calling it if `method_exists`
		// was called prior to that call.
		if !b.walker.ctx.customMethodExists(e.Variable, call.methodName) {
			b.report(e.Method, LevelError, "undefined", "Call to undefined method {%s}->%s()", call.methodCallerType, call.methodName)
		}
	} else if !call.isMagic && !parseState.IsTrait {
		// Method is defined.
		b.walker.r.checkNameCase(e.Method, call.methodName, call.info.Name)
		if call.info.IsStatic() {
			b.report(e.Method, LevelWarning, "callStatic", "Calling static method as instance method")
		}
	}

	if call.info.Doc.Deprecated {
		if call.info.Doc.DeprecationNote != "" {
			b.report(e.Method, LevelNotice, "deprecated", "Call to deprecated method {%s}->%s() (%s)",
				call.methodCallerType, call.methodName, call.info.Doc.DeprecationNote)
		} else {
			b.report(e.Method, LevelNotice, "deprecated", "Call to deprecated method {%s}->%s()",
				call.methodCallerType, call.methodName)
		}
	}

	if call.isFound && !call.isMagic && !canAccess(parseState, call.className, call.info.AccessLevel) {
		b.report(e.Method, LevelError, "accessLevel", "Cannot access %s method %s->%s()", call.info.AccessLevel, call.className, call.methodName)
	}
}

func (b *blockLinter) checkStaticCall(e *ir.StaticCallExpr) {
	call := resolveStaticMethodCall(b.classParseState(), e)
	if !call.canAnalyze {
		return
	}

	b.checkClassSpecialNameCase(e, call.className)

	if !call.isMagic {
		b.checkCallArgs(e.Call, e.Args, call.methodInfo.Info)
	}

	if !call.isFound && !call.isMagic && !b.classParseState().IsTrait {
		b.report(e.Call, LevelError, "undefined", "Call to undefined method %s::%s()", call.className, call.methodName)
	} else if !call.isParentCall && !call.methodInfo.Info.IsStatic() && !call.isMagic && !b.classParseState().IsTrait {
		// Method is defined.
		// parent::f() is permitted.
		b.report(e.Call, LevelWarning, "callStatic", "Calling instance method as static method")
	}

	if call.methodInfo.Info.Doc.Deprecated {
		if call.methodInfo.Info.Doc.DeprecationNote != "" {
			b.report(e.Call, LevelNotice, "deprecated", "Call to deprecated static method %s::%s() (%s)",
				call.className, call.methodName, call.methodInfo.Info.Doc.DeprecationNote)
		} else {
			b.report(e.Call, LevelNotice, "deprecated", "Call to deprecated static method %s::%s()",
				call.className, call.methodName)
		}
	}

	if call.isFound && !canAccess(b.classParseState(), call.methodInfo.ClassName, call.methodInfo.Info.AccessLevel) {
		b.report(e.Call, LevelError, "accessLevel", "Cannot access %s method %s::%s()", call.methodInfo.Info.AccessLevel, call.methodInfo.ClassName, call.methodName)
	}
}

func (b *blockLinter) checkPropertyFetch(e *ir.PropertyFetchExpr) {
	fetch := resolvePropertyFetch(b.walker.ctx.sc, b.classParseState(), b.walker.ctx.customTypes, e)
	if !fetch.canAnalyze {
		return
	}

	if !fetch.isFound && !fetch.isMagic && !b.classParseState().IsTrait && !b.walker.isThisInsideClosure(e.Variable) {
		b.report(e.Property, LevelError, "undefined", "Property {%s}->%s does not exist", fetch.propertyFetchType, fetch.propertyNode.Value)
	}

	if fetch.isFound && !fetch.isMagic && !canAccess(b.classParseState(), fetch.className, fetch.info.AccessLevel) {
		b.report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s->%s", fetch.info.AccessLevel, fetch.className, fetch.propertyNode.Value)
	}
}

func (b *blockLinter) checkStaticPropertyFetch(e *ir.StaticPropertyFetchExpr) {
	fetch := resolveStaticPropertyFetch(b.classParseState(), e)
	if !fetch.canAnalyze {
		return
	}

	b.checkClassSpecialNameCase(e, fetch.className)

	if !fetch.isFound && !b.classParseState().IsTrait {
		b.report(e.Property, LevelError, "undefined", "Property %s::$%s does not exist", fetch.className, fetch.propertyName)
	}

	if fetch.isFound && !canAccess(b.classParseState(), fetch.info.ClassName, fetch.info.Info.AccessLevel) {
		b.report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s::$%s", fetch.info.Info.AccessLevel, fetch.info.ClassName, fetch.propertyName)
	}
}

func (b *blockLinter) checkClassConstFetch(e *ir.ClassConstFetchExpr) {
	fetch := resolveClassConstFetch(b.classParseState(), e)
	if !fetch.canAnalyze {
		return
	}

	b.checkClassSpecialNameCase(e, fetch.className)

	if !fetch.isFound && !b.classParseState().IsTrait {
		b.walker.r.Report(e.ConstantName, LevelError, "undefined", "Class constant %s::%s does not exist", fetch.className, fetch.constName)
	}

	if fetch.isFound && !canAccess(b.classParseState(), fetch.implClassName, fetch.info.AccessLevel) {
		b.walker.r.Report(e.ConstantName, LevelError, "accessLevel", "Cannot access %s constant %s::%s", fetch.info.AccessLevel, fetch.implClassName, fetch.constName)
	}
}

func (b *blockLinter) checkClassSpecialNameCase(n ir.Node, className string) {
	// Since for resolving class names we use the solver.GetClassName function,
	// which resolves unknown classes as '\' + the passed class name, then for
	// misspelled special class names (self, static, parent) we get something
	// like '\SELF'. For correctly spelled ones, we get the specific class name.
	// Therefore, to catch this case, we compare the resolved class name with
	// '\' + the correct spelling of the special name, case insensitive.
	// If there is a match, it means that the name was originally spelled in the wrong case.

	names := []string{
		`\self`,
		`\static`,
		`\parent`,
	}

	for _, name := range names {
		if strings.EqualFold(className, name) {
			b.walker.r.Report(n, LevelNotice, "nameMismatch", "%s should be spelled as %s", strings.TrimPrefix(className, `\`), strings.TrimPrefix(name, `\`))
		}
	}
}

func (b *blockLinter) checkInterfaceStmt(iface *ir.InterfaceStmt) {
	for _, st := range iface.Stmts {
		switch x := st.(type) {
		case *ir.ClassMethodStmt:
			for _, modifier := range x.Modifiers {
				if strings.EqualFold(modifier.Value, "private") || strings.EqualFold(modifier.Value, "protected") {
					methodName := x.MethodName.Value
					b.report(x, LevelWarning, "nonPublicInterfaceMember", "'%s' can't be %s", methodName, modifier.Value)
				}
			}
		case *ir.ClassConstListStmt:
			for _, modifier := range x.Modifiers {
				if strings.EqualFold(modifier.Value, "private") || strings.EqualFold(modifier.Value, "protected") {
					for _, constant := range x.Consts {
						constantName := constant.(*ir.ConstantStmt).ConstantName.Value
						b.report(x, LevelWarning, "nonPublicInterfaceMember", "'%s' can't be %s", constantName, modifier.Value)
					}
				}
			}
		}
	}
}

func (b *blockLinter) checkRegexp(e *ir.FunctionCallExpr, arg *ir.Argument) {
	s, ok := arg.Expr.(*ir.String)
	if !ok {
		return
	}
	pat := s.Value
	simplified := b.walker.r.reSimplifier.simplifyRegexp(pat)
	if simplified != "" {
		rawPattern := b.walker.r.nodeText(s)
		b.report(arg, LevelNotice, "regexpSimplify", "May re-write %s as '%s'",
			rawPattern, simplified)
	}
	issues, err := b.walker.r.reVet.CheckRegexp(pat)
	if err != nil {
		b.report(arg, LevelError, "regexpSyntax", "parse error: %v", err)
	}
	for _, issue := range issues {
		b.report(arg, LevelWarning, "regexpVet", "%s", issue)
	}
}

func (b *blockLinter) checkFormatString(e *ir.FunctionCallExpr, arg *ir.Argument) {
	s, ok := arg.Expr.(*ir.String)
	if !ok {
		return
	}
	const argsLimit = 16
	if len(s.Value) > 255 || len(e.Args) > argsLimit {
		return
	}

	format, err := parseFormatString(s.Value)
	if err != nil {
		b.report(arg, LevelWarning, "printf", "%s", err.Error())
		return
	}

	// TODO: detect `% <char>` cases.
	// For example in, "Handler % tried to add additional_field %s but % could not be added!"
	// we have 2 bad formatting directives here, but only one is reported, `% t`, since
	// 't' is not a correct specifier (while '% c' is technically OK).
	//
	// TODO: test whether things like `%1%` make sense. We report all %% directive
	// usages that have any modifiers.

	usages := make([]uint8, argsLimit)
	for _, d := range format.directives {
		if d.specifier == '%' {
			hasModifiers := d.argNum != -1 || d.flags != "" || d.precision != -1 || d.width != -1
			if hasModifiers {
				b.report(arg, LevelWarning, "printf", "%%%% directive has modifiers")
			}
			continue
		}

		if d.argNum == -1 {
			continue
		}
		if d.argNum >= len(e.Args) {
			s := s.Value[d.begin:d.end]
			b.report(arg, LevelWarning, "printf", "%s directive refers to the args[%d] which is not provided", s, d.argNum)
			continue
		}
		if d.argNum < len(usages) {
			usages[d.argNum]++
		}

		arg := e.Arg(d.argNum)
		if d.specifier == 's' && b.walker.exprType(arg.Expr).IsArray() {
			b.report(arg, LevelWarning, "printf", "potential array to string conversion")
		}
	}

	for i := 1; i < len(e.Args); i++ {
		if usages[i] == 0 {
			b.report(e.Arg(i), LevelWarning, "printf", "argument is not referenced from the formatting string")
		}
	}
}

func (b *blockLinter) classParseState() *meta.ClassParseState {
	return b.walker.r.ctx.st
}

func (b *blockLinter) metaInfo() *meta.Info {
	return b.walker.r.ctx.st.Info
}
