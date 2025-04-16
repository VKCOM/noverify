package linter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/constfold"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/ir/phpcore"
	"github.com/VKCOM/noverify/src/linter/autogen"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/utils"
)

type blockLinter struct {
	walker   *blockWalker
	quickfix *QuickFixGenerator
}

func (b *blockLinter) enterNode(n ir.Node) {
	switch n := n.(type) {

	case *ir.Encapsed:
		b.checkStringInterpolationDeprecation(n)

	case *ir.Assign:
		b.checkAssign(n)

	case *ir.ArrayExpr:
		b.checkArray(n)

	case *ir.ClassStmt:
		b.checkClass(n)

	case *ir.TraitStmt:
		b.checkTrait(n)

	case *ir.FunctionCallExpr:
		b.checkFunctionCall(n)

	case *ir.ArrowFunctionExpr:
		phpDocParamTypes := b.walker.getParamsTypesFromPhpDoc(n.Doc)
		b.walker.CheckParamNullability(n.Params, phpDocParamTypes)

	case *ir.ClosureExpr:
		phpDocParamTypes := b.walker.getParamsTypesFromPhpDoc(n.Doc)
		b.walker.CheckParamNullability(n.Params, phpDocParamTypes)

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

	case *ir.UnaryPlusExpr:
		b.checkUnaryPlus(n)

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
		b.checkGettype(n)
	case *ir.NotEqualExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
		b.checkGettype(n)
	case *ir.IdenticalExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
		b.checkGettype(n)
	case *ir.NotIdenticalExpr:
		b.checkBinaryVoidType(n.Left, n.Right)
		b.checkBinaryDupArgsNoFloat(n, n.Left, n.Right)
		b.checkGettype(n)
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
		b.walker.r.checker.CheckKeywordCase(n, "clone")
	case *ir.ConstListStmt:
		b.walker.r.checker.CheckKeywordCase(n, "const")
	case *ir.GotoStmt:
		b.walker.r.checker.CheckKeywordCase(n, "goto")
	case *ir.ThrowStmt:
		b.walker.r.checker.CheckKeywordCase(n, "throw")
	case *ir.YieldExpr:
		b.walker.r.checker.CheckKeywordCase(n, "yield")
	case *ir.YieldFromExpr:
		b.walker.r.checker.CheckKeywordCase(n, "yield")
	case *ir.ImportExpr:
		b.walker.r.checker.CheckKeywordCase(n, n.Func)
	case *ir.BreakStmt:
		b.walker.r.checker.CheckKeywordCase(n, "break")
	case *ir.ReturnStmt:
		b.walker.r.checker.CheckKeywordCase(n, "return")
	case *ir.ElseStmt:
		b.walker.r.checker.CheckKeywordCase(n, "else")

	case *ir.ForeachStmt:
		b.checkForeach(n)
	case *ir.ForStmt:
		b.walker.r.checker.CheckKeywordCase(n, "for")
	case *ir.WhileStmt:
		b.walker.r.checker.CheckKeywordCase(n, "while")
		b.checkDangerousBoolCond(n.Cond)
	case *ir.DoStmt:
		b.walker.r.checker.CheckKeywordCase(n, "do")
		b.checkDangerousBoolCond(n.Cond)

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

func (b *blockLinter) checkStringInterpolationDeprecation(str *ir.Encapsed) {
	for _, item := range str.Parts {
		variable, ok := item.(*ir.SimpleVar)
		if ok {
			if variable.IdentifierTkn.Value[0] != '$' {
				b.report(str, LevelWarning, "stringInterpolationDeprecated", "use {$variable} instead ${variable}")
				break
			}
		}
	}
}

func (b *blockLinter) checkUnaryPlus(n *ir.UnaryPlusExpr) {
	val := constfold.Eval(b.classParseState(), n.Expr)
	if val.IsValid() {
		return
	}

	b.report(n, LevelWarning, "strangeCast", "Unary plus with non-constant expression, possible type cast, use an explicit cast to int or float instead of using the unary plus")
}

func (b *blockLinter) checkTrait(n *ir.TraitStmt) {
	for _, stmt := range n.Stmts {
		method, ok := stmt.(*ir.ClassMethodStmt)
		if ok {
			phpDocParamTypes := b.walker.getParamsTypesFromPhpDoc(method.Doc)
			b.walker.CheckParamNullability(method.Params, phpDocParamTypes)
		}
	}
}

func (b *blockLinter) checkClass(class *ir.ClassStmt) {
	const classMethod = 0
	const classOtherMember = 1

	var members = make([]int, 0, len(class.Stmts))
	for _, stmt := range class.Stmts {
		switch value := stmt.(type) {
		case *ir.ClassMethodStmt:
			members = append(members, classMethod)
			phpDocParamTypes := b.walker.getParamsTypesFromPhpDoc(value.Doc)
			b.walker.CheckParamNullability(value.Params, phpDocParamTypes)
		case *ir.PropertyListStmt:
			for _, element := range value.Doc.Parsed {
				if element.Name() == "deprecated" {
					b.report(stmt, LevelNotice, "deprecated", "Has deprecated field in class %s", class.ClassName.Value)
				}
			}
			members = append(members, classOtherMember)
		default:
			members = append(members, classOtherMember)
		}
	}

	var methodsBegin bool
	for index, member := range members {
		if member == classMethod {
			methodsBegin = true
		} else if methodsBegin {
			stmt := class.Stmts[index]
			memberType := ""
			memberName := ""
			switch stmt := stmt.(type) {
			case *ir.ClassConstListStmt:
				memberType = "Constant"
				memberName = stmt.Consts[0].(*ir.ConstantStmt).ConstantName.Value
			case *ir.PropertyListStmt:
				memberType = "Property"
				memberName = "$" + stmt.Properties[0].(*ir.PropertyStmt).Variable.Name
			default:
				continue
			}
			b.report(stmt, LevelError, "classMembersOrder", "%s %s must go before methods in the class %s", memberType, memberName, class.ClassName.Value)
		}
	}
}

func (b *blockLinter) checkForeach(n *ir.ForeachStmt) {
	b.walker.r.checker.CheckKeywordCase(n, "foreach")

	var vars []*ir.SimpleVar

	findAllVars := func(n ir.Node) bool {
		if vr, ok := n.(*ir.SimpleVar); ok {
			vars = append(vars, vr)
		}
		return true
	}

	if n.Variable != nil {
		irutil.Inspect(n.Variable, findAllVars)
	}
	if n.Key != nil {
		irutil.Inspect(n.Key, findAllVars)
	}

	fun, ok := b.walker.r.currentFunction()
	if !ok {
		return
	}

	for _, v := range vars {
		for _, param := range fun.Params {
			if v.Name == param.Name {
				b.report(v, LevelError, "varShadow", "Variable $%s shadow existing variable $%s from current function params", v.Name, param.Name)
			}
		}
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
		b.report(n, LevelNotice, "langDeprecated", "Use float cast instead of real")
	}
}

func (b *blockLinter) report(n ir.Node, level int, checkName, msg string, args ...interface{}) {
	b.walker.report(n, level, checkName, msg, args...)
}

func (b *blockLinter) checkNopStmt(n *ir.NopStmt) {
	switch b.walker.path.Parent().(type) {
	case *ir.DeclareStmt, *ir.IfStmt, *ir.ForStmt, *ir.ForeachStmt, *ir.WhileStmt, *ir.DoStmt:
		return
	}

	b.report(n, LevelNotice, "emptyStmt", "Semicolon (;) is not needed here, it can be safely removed")
}

func (b *blockLinter) checkCoalesceExpr(n *ir.CoalesceExpr) {
	lhsType := solver.ExprType(b.walker.ctx.sc, b.classParseState(), n.Left)
	if !lhsType.IsPrecise() {
		return
	}

	rhsVariableUndefined := false
	variable, ok := n.Left.(*ir.SimpleVar)
	if ok {
		have := b.walker.ctx.sc.HaveVar(variable)
		rhsVariableUndefined = !have
	}

	// If the variable is not defined, then ?? can be a test
	// for this, so we do not need to give this warning
	if !lhsType.Contains("null") && !rhsVariableUndefined {
		b.report(n.Right, LevelWarning, "deadCode", "%s is not nullable, right side of the expression is unreachable", irutil.FmtNode(n.Left))
	}
}

func (b *blockLinter) checkAssign(a *ir.Assign) {
	b.checkVoidType(a.Expr)

	var sign byte
	switch a.Expr.(type) {
	case *ir.UnaryPlusExpr:
		sign = '+'
	case *ir.UnaryMinusExpr:
		sign = '-'
	default:
		return
	}

	// Get sign token.
	signTkn := ir.GetFirstToken(a.Expr)

	// $a= 100;
	//   |
	//   - Free floating empty.
	//
	// $a = 100;
	//    |
	//    - Free floating contain space.
	containsSpaceBeforeAssign := len(a.EqualTkn.FreeFloating) != 0

	// $a =+ 100;
	//     |
	//     - Free floating empty.
	//
	// $a = +100;
	//      |
	//      - Free floating contain space.
	containsSpaceBeforeSign := len(signTkn.FreeFloating) != 0

	if !containsSpaceBeforeSign && containsSpaceBeforeAssign {
		b.report(a, LevelWarning, "reverseAssign", "Possible there should be '%c='", sign)
	}
}

func (b *blockLinter) checkTryStmt(s *ir.TryStmt) {
	if len(s.Catches) == 0 && s.Finally == nil {
		b.report(s, LevelError, "bareTry", "At least one catch or finally block must be present")
	}

	b.walker.r.checker.CheckKeywordCase(s, "try")

	for _, c := range s.Catches {
		b.walker.r.checker.CheckKeywordCase(c, "catch")
	}

	if s.Finally != nil {
		b.walker.r.checker.CheckKeywordCase(s.Finally, "finally")
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
				b.report(c.Types[0], LevelWarning, "dupCatch", "Duplicated catch on %s", class)
				add = false
				break
			}
			if solver.Extends(b.metaInfo(), class, otherClass) {
				b.report(c.Types[0], LevelWarning, "catchOrder", "Catch %s block will never run as it extends %s which is caught above", class, otherClass)
				add = false
				break
			}
			if solver.Implements(b.metaInfo(), class, otherClass) {
				b.report(c.Types[0], LevelWarning, "catchOrder", "Catch %s block will never run as it implements %s which is caught above", class, otherClass)
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
		b.report(n, LevelWarning, "dupSubExpr", "Duplicated operands value in %s expression", utils.BinaryOpString(n))
	}
}

// checkVoidType reports if node has void type
func (b *blockLinter) checkVoidType(n ir.Node) {
	if b.walker.exprType(n).Is("void") {
		b.report(n, LevelNotice, "voidResultUsed", "Void function result used")
	}
}

func (b *blockLinter) checkRedundantCastArray(e ir.Node) {
	typ := b.walker.exprType(e)
	if typ.Len() == 1 && typ.Is("mixed[]") {
		b.report(e, LevelNotice, "redundantCast", "Expression already has array type")
	}
}

func (b *blockLinter) checkRedundantCast(e ir.Node, dstType string) {
	typ := b.walker.exprType(e)
	if typ.Len() != 1 {
		return
	}
	typ.Iterate(func(x string) {
		if x == dstType {
			b.report(e, LevelNotice, "redundantCast", "Expression already has %s type", dstType)
		}
	})
}

func (b *blockLinter) checkNew(e *ir.NewExpr) {
	b.walker.r.checker.CheckKeywordCase(e, "new")

	if b.classParseState().IsTrait {
		switch {
		case utils.NameNodeEquals(e.Class, "self"):
			// Don't try to resolve "self" inside trait context.
			return
		case utils.NameNodeEquals(e.Class, "static"):
			// More or less identical to the "self" case.
			return
		}
	}

	var className string
	var args []ir.Node

	if anon, ok := e.Class.(*ir.AnonClassExpr); ok {
		className = autogen.GenerateAnonClassName(anon, b.walker.r.ctx.st.CurrentFile)
		className = b.classParseState().Namespace + className
		args = anon.Args
	} else {
		className, ok = solver.GetClassName(b.classParseState(), e.Class)
		if !ok {
			// perhaps something like 'new $class', cannot check this.
			return
		}
		args = e.Args
	}

	class, ok := b.metaInfo().GetClass(className)
	if !ok {
		class, ok = b.metaInfo().GetTrait(className)
		if ok {
			b.report(e.Class, LevelError, "invalidNew", "Cannot instantiate trait %s", class.Name)
		} else {
			b.report(e.Class, LevelError, "undefinedClass", "Class or interface named %s does not exist", className)
		}
	} else {
		if class.IsInterface() {
			b.report(e.Class, LevelError, "invalidNew", "Cannot instantiate interface %s", class.Name)
		}
		b.walker.r.checker.CheckNameCase(e.Class, className, class.Name)
	}

	// It's illegal to instantiate abstract class, but `static` can
	// resolve to something else due to the late static binding,
	// so it's the only exception to that rule.
	if class.IsAbstract() && !utils.NameNodeEquals(e.Class, "static") {
		b.report(e.Class, LevelError, "newAbstract", "Cannot instantiate abstract class %s", class.Name)
	}

	// Check implicitly invoked constructor method arguments count.
	m, ok := solver.FindMethod(b.metaInfo(), className, "__construct")
	if !ok {
		return
	}

	ctor := m.Info
	// If new expression is written without (), ArgumentList will be nil.
	// It's equivalent of 0 arguments constructor call.
	if ok && !enoughArgs(args, ctor) {
		b.report(e, LevelError, "argCount", "Too few arguments for %s constructor, expecting %d, saw %d", className, ctor.MinParamsCnt, len(args))
	}
}

func (b *blockLinter) checkStmtExpression(s *ir.ExpressionStmt) {
	report := false

	// All branches except default try to filter-out common
	// cases to reduce the number of type solving performed.
	if irutil.IsAssign(s.Expr) {
		assign, okCast := s.Expr.(*ir.Assign)
		if !okCast {
			return
		}
		if v, ok := assign.Variable.(*ir.StaticPropertyFetchExpr); ok {
			parseState := b.classParseState()
			left, ok := parseState.Info.GetVarType(v.Class)

			if ok {
				b.checkSafetyCall(s, left, "", "PropertyFetch")
			}
		}
		return
	}
	switch expr := s.Expr.(type) {
	case *ir.ImportExpr, *ir.ExitExpr:
		// Skip.
	case *ir.ArrayExpr, *ir.NewExpr:
		// Report these even if they are not pure.
		report = true
	case *ir.StaticCallExpr:
		if v, ok := expr.Class.(*ir.SimpleVar); ok {
			parseState := b.classParseState()
			left, ok := parseState.Info.GetVarType(v)

			if ok && left.Contains("null") {
				b.report(s, LevelWarning, "notNullSafetyStaticFunctionCall",
					"potential null dereference when accessing static call throw $%s", v.Name)
			}
		}
		report = b.isDiscardableExpr(s.Expr)
	default:
		report = b.isDiscardableExpr(s.Expr)
	}

	if report {
		b.report(s.Expr, LevelWarning, "discardExpr", "Expression evaluated but not used")
	}
}

func (b *blockLinter) isDiscardableExpr(expr ir.Node) bool {
	typ := b.walker.exprType(expr)
	if !typ.Is("void") {
		return b.walker.sideEffectFree(expr)
	}
	return false
}

func (b *blockLinter) checkConstFetch(e *ir.ConstFetchExpr) {
	_, _, defined := solver.GetConstant(b.classParseState(), e.Constant)

	if !defined {
		// If it's builtin constant, give a more precise report message.
		switch nm := utils.NameNodeToString(e.Constant); strings.ToLower(nm) {
		case "null", "true", "false":
			expected := strings.ToLower(nm)
			b.report(e.Constant, LevelError, "constCase", "Constant '%s' should be used in lower case as '%s'", nm, expected)
			b.addFixForBuiltInConstantCase(e.Constant, expected)
		default:
			b.report(e.Constant, LevelError, "undefinedConstant", "Undefined constant %s", nm)
		}
	}
}

func (b *blockLinter) addFixForBuiltInConstantCase(constant *ir.Name, expected string) {
	if !b.walker.r.config.ApplyQuickFixes {
		return
	}

	b.walker.r.addQuickFix("constCase", quickfix.TextEdit{
		StartPos:    constant.Position.StartPos,
		EndPos:      constant.Position.EndPos,
		Replacement: expected,
	})
}

func (b *blockLinter) checkTernary(e *ir.TernaryExpr) {
	if e.IfTrue == nil {
		return // Skip `$x ?: $y` expressions
	}

	_, nestedTernary := e.Condition.(*ir.TernaryExpr)
	if nestedTernary {
		b.report(e.Condition, LevelWarning, "nestedTernary", "Explicitly specify the order of operations for the ternary operator using parentheses")
	}

	// Check for `$cond ? $x : $x` which makes no sense.
	if irutil.NodeEqual(e.IfTrue, e.IfFalse) {
		b.report(e, LevelWarning, "dupBranchBody", "Branches for true and false have the same operands, ternary operator is meaningless")
	}
}

func (b *blockLinter) checkGlobalStmt(s *ir.GlobalStmt) {
	b.walker.r.checker.CheckKeywordCase(s, "global")

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

	for _, c := range s.Cases {
		caseNode, ok := c.(*ir.CaseStmt)
		if !ok {
			continue
		}

		// Probably the case:
		// case 1: case 2: case 3:
		if len(caseNode.Stmts) == 0 {
			continue
		}

		isDupBody := !nodeSet.Add(&ir.StmtList{Stmts: caseNode.Stmts})

		if isDupBody {
			msg := fmt.Sprintf("Branch 'case %s' in 'switch' is a duplicate, combine cases with the same body into one", irutil.FmtNode(caseNode.Cond))
			b.report(caseNode.Cond, LevelWarning, "dupBranchBody", "%s", msg)
		}
	}

	nodeSet.Reset()

	containsDefault := false

	for i, c := range s.Cases {
		defaultNode, ok := c.(*ir.DefaultStmt)
		if ok {
			containsDefault = true
			if i != 0 && i != len(s.Cases)-1 {
				b.report(defaultNode, LevelWarning, "switchDefault", "'default' should be first or last to improve readability")
			}
		}

		caseNode, ok := c.(*ir.CaseStmt)
		if !ok {
			continue
		}

		if !b.walker.sideEffectFree(caseNode.Cond) {
			continue
		}

		var constValue meta.ConstValue
		var isConstKey bool
		constFetchExpr, ok := caseNode.Cond.(*ir.ConstFetchExpr)
		if ok {
			constValue = constfold.Eval(b.classParseState(), constFetchExpr)
			if !constValue.IsValid() {
				continue
			}
			value := constValue.Value

			switch constValue.Type {
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
			isDupKey = !nodeSet.Add(caseNode.Cond)
		}

		if isDupKey {
			msg := fmt.Sprintf("Duplicated switch case for expression %s", irutil.FmtNode(caseNode.Cond))
			if isConstKey {
				dupKey := meta.GetConstValue(constValue)
				msg += " (value: " + dupKey + ")"
			}
			b.report(caseNode.Cond, LevelWarning, "dupCond", "%s", msg)
		}
	}

	if len(s.Cases) == 2 && containsDefault || len(s.Cases) == 1 {
		b.report(s, LevelWarning, "switchSimplify", "Switch can be rewritten into an 'if' statement to increase readability")
	}

	if len(s.Cases) == 0 {
		b.report(s, LevelWarning, "switchEmpty", "Switch has empty body")
	}

	if len(s.Cases) != 0 && !containsDefault {
		b.report(s, LevelWarning, "switchDefault", "Add 'default' branch to avoid unexpected unhandled condition values")
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
			b.report(s, LevelWarning, "dupBranchBody", "Duplicated if/else actions")
		}
	}
	b.checkDangerousBoolCond(s.Cond)

	b.checkIfStmtDupCond(s)
}

func (b *blockLinter) checkDangerousBoolCond(s ir.Node) {
	checkNodeDangerousBoolCond(s, b)
}

func checkNodeDangerousBoolCond(node ir.Node, b *blockLinter) {
	switch n := node.(type) {
	case *ir.ConstFetchExpr:
		if strings.EqualFold(n.Constant.Value, "true") || strings.EqualFold(n.Constant.Value, "false") {
			b.report(node, LevelNotice, "dangerousBoolCondition", "Potential dangerous bool value: you have constant bool value in condition")
		}
	case *ir.Lnumber:
		if n.Value == "0" || n.Value == "1" {
			b.report(node, LevelNotice, "dangerousBoolCondition", "Potential dangerous value: you have constant int value that interpreted as bool")
		}
	case *ir.BooleanOrExpr:
		checkNodeDangerousBoolCond(n.Left, b)
		checkNodeDangerousBoolCond(n.Right, b)
	case *ir.BooleanAndExpr:
		checkNodeDangerousBoolCond(n.Left, b)
		checkNodeDangerousBoolCond(n.Right, b)
	}
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
		b.checkDangerousBoolCond(elseif.Cond)
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
	b.walker.r.checker.CheckKeywordCase(c, "continue")
	if c.Expr == nil && b.walker.ctx.innermostLoop == loopSwitch {
		inLoop := irutil.InLoop(b.walker.path)
		msg := "Use 'break' instead of 'continue' in switch"
		if inLoop {
			msg = "Use 'break' instead of 'continue' or 'continue 2' to continue the loop around switch"
		}

		b.report(c, LevelError, "caseContinue", msg)
	}
}

func (b *blockLinter) checkArray(arr *ir.ArrayExpr) {
	if !arr.ShortSyntax {
		b.report(arr, LevelNotice, "arraySyntax", "Use the short form '[]' instead of the old 'array()'")
		b.walker.r.addQuickFix("arraySyntax", b.quickfix.Array(arr))
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

	b.walker.r.addQuickFix("trailingComma", quickfix.TextEdit{
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
		b.report(item, LevelNotice, "trailingComma", "Last element in a multi-line array should have a trailing comma")
		b.addFixForMultilineArrayTrailingComma(item)
	}
}

func (b *blockLinter) checkDeprecatedFunctionCall(e *ir.FunctionCallExpr, call *funcCallInfo) {
	if !call.info.Deprecated {
		return
	}

	if call.info.WithDeprecationNote() {
		b.report(e.Function, LevelNotice, "deprecated", "Call to deprecated function %s (%s)", utils.NameNodeToString(e.Function), call.info.DeprecationInfo)
		return
	}

	b.report(e.Function, LevelNotice, "deprecatedUntagged", "Call to deprecated function %s", utils.NameNodeToString(e.Function))
}

func (b *blockLinter) checkFunctionAvailability(e *ir.FunctionCallExpr, call *funcCallInfo) {
	if !call.isFound && !b.walker.ctx.customFunctionExists(e.Function) {
		b.report(e.Function, LevelError, "undefinedFunction", "Call to undefined function %s", utils.NameNodeToString(e.Function))
	}
}

func (b *blockLinter) checkCallArgs(fun ir.Node, args []ir.Node, fn meta.FuncInfo, callerClass string) {
	b.checkCallArgsCount(fun, args, fn, callerClass)
	b.checkArgsOrder(fun, args, fn)
}

func (b *blockLinter) checkArgsOrder(fun ir.Node, args []ir.Node, fn meta.FuncInfo) {
	if len(args) != 2 || len(fn.Params) < 2 {
		return
	}

	firstArg := args[0].(*ir.Argument)
	secondArg := args[1].(*ir.Argument)

	firstVar, ok := firstArg.Expr.(*ir.SimpleVar)
	if !ok {
		return
	}

	secondVar, ok := secondArg.Expr.(*ir.SimpleVar)
	if !ok {
		return
	}

	if firstVar.Name == fn.Params[1].Name && secondVar.Name == fn.Params[0].Name {
		b.report(fun, LevelWarning, "argsReverse", "Perhaps the order of the arguments is messed up, $%[1]s is passed to the $%[2]s parameter, and $%[2]s is passed to the $%[1]s parameter", firstVar.Name, secondVar.Name)
	}
}

func (b *blockLinter) checkCallArgsCount(fun ir.Node, args []ir.Node, fn meta.FuncInfo, callerClass string) {
	if fn.Name == `\mt_rand` {
		if len(args) != 0 && len(args) != 2 {
			b.report(fun, LevelWarning, "argCount", "mt_rand expects 0 or 2 args")
		}
		return
	}

	if fn.Name == `\compact` || fn.Name == `\func_get_args` {
		// there is no need to check the number of arguments for these functions.
		return
	}

	if !enoughArgs(args, fn) {
		name := strings.TrimPrefix(fn.Name, `\`)
		if callerClass != "" {
			name = fmt.Sprintf("%s::%s", strings.TrimPrefix(callerClass, `\`), name)
		} else if types.IsClosure(fn.Name) {
			name = autogen.TransformClosureToReadableName(fn.Name)
		}

		b.report(fun, LevelWarning, "argCount",
			"Too few arguments for %s, expecting %d, saw %d", name, fn.MinParamsCnt, len(args))
	}
}

func (b *blockLinter) checkFunctionCall(e *ir.FunctionCallExpr) {
	call := resolveFunctionCall(b.walker.ctx.sc, b.classParseState(), b.walker.ctx.customTypes, e)
	fqName := call.funcName

	var trimName = strings.TrimPrefix(fqName, `\`)
	var phpMasterFunc = phpcore.FuncAliases[trimName]
	if phpMasterFunc != nil {
		b.report(e, LevelWarning, "phpAliases", "Use %s instead of '%s'", phpMasterFunc.Value, trimName)
		b.walker.r.addQuickFix("phpAliases", b.quickfix.PhpAliasesReplace(e.Function.(*ir.Name), phpMasterFunc.Value))
	}

	if call.isClosure {
		b.walker.untrackVarName(trimName)
	} else {
		b.checkFunctionAvailability(e, &call)
		b.walker.r.checker.CheckNameCase(e.Function, call.funcName, call.info.Name)
	}

	if call.isFound {
		b.checkCallArgs(e.Function, e.Args, call.info, "")
		b.checkDeprecatedFunctionCall(e, &call)
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
		b.report(e, LevelNotice, "langDeprecated", "Use is_float function instead of is_real")
	case `\array_key_exists`:
		b.checkArrayKeyExistsCall(e)
	case `\random_int`:
		b.checkRandomIntCall(e)
	}
}

func (b *blockLinter) checkGettype(node ir.Node) {
	var left ir.Node
	var right ir.Node
	var isNegative = false

	switch node := node.(type) {
	case *ir.EqualExpr:
		left = node.Left
		right = node.Right
	case *ir.NotEqualExpr:
		left = node.Left
		right = node.Right
		isNegative = true
	case *ir.IdenticalExpr:
		left = node.Left
		right = node.Right
	case *ir.NotIdenticalExpr:
		left = node.Left
		right = node.Right
		isNegative = true
	default:
		return
	}

	call, ok := left.(*ir.FunctionCallExpr)
	if !ok {
		return
	}

	if len(call.Args) != 1 {
		return
	}

	name, ok := call.Function.(*ir.Name)
	if !ok {
		return
	}

	if name.Value != "gettype" {
		return
	}

	firstArg, ok := call.Args[0].(*ir.Argument)
	if !ok {
		return
	}

	nodeText := b.walker.r.nodeText(firstArg)

	str, ok := right.(*ir.String)
	if !ok {
		return
	}

	isFunctionName, ok := phpcore.TypeToIsFunction[str.Value]
	if !ok {
		return
	}

	b.report(node, LevelWarning, "getTypeMisUse", "use %s instead of '%s'", isFunctionName, b.walker.r.nodeText(node))
	b.walker.r.addQuickFix("getTypeMisUse", b.quickfix.GetType(node, isFunctionName, nodeText, isNegative))
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
	call := resolveMethodCall(b.walker.ctx.sc, parseState, b.walker.ctx.customTypes, e, b.walker.r.strictMixed)
	if !call.canAnalyze {
		return
	}

	b.checkMethodCallPackage(call.className, call.info, e)

	if !call.isMagic {
		b.checkCallArgs(e.Method, e.Args, call.info, call.methodCallerType.String())
	}

	if !call.isFound && !call.isMagic && !parseState.IsTrait && !b.walker.isThisInsideClosure(e.Variable) {
		needShowUndefinedMethod := !call.callerTypeIsMixed || b.walker.r.strictMixed

		// The method is undefined, but we permit calling it if `method_exists`
		// was called prior to that call.
		if !b.walker.ctx.customMethodExists(e.Variable, call.methodName) && needShowUndefinedMethod {
			b.report(e.Method, LevelError, "undefinedMethod", "Call to undefined method {%s}->%s()", call.methodCallerType, call.methodName)
		}
	} else if !call.isMagic && !parseState.IsTrait {
		// Method is defined.
		b.walker.r.checker.CheckNameCase(e.Method, call.methodName, call.info.Name)
		if call.info.IsStatic() {
			b.report(e.Method, LevelWarning, "callStatic", "Calling static method as instance method")
		}
	}

	switch caller := e.Variable.(type) {
	case *ir.FunctionCallExpr:
		var funcName string
		var ok bool

		switch fn := caller.Function.(type) {
		case *ir.SimpleVar:
			funcName, ok = solver.GetFuncName(parseState, &ir.Name{Value: fn.Name})

		case *ir.Name:
			funcName, ok = solver.GetFuncName(parseState, fn)
		}
		if ok {
			funInfo, found := parseState.Info.GetFunction(funcName)
			if found {
				funcType := funInfo.Typ
				b.checkSafetyCall(e, funcType, funInfo.Name, "FunctionCall")
			}
		}

	case *ir.SimpleVar:
		varType, ok := b.walker.ctx.sc.GetVarType(caller)
		if ok {
			b.checkSafetyCall(e, varType, caller.Name, "Variable")
		}
	}

	if call.info.Deprecated {
		deprecation := call.info.DeprecationInfo

		if deprecation.WithDeprecationNote() {
			b.report(e.Method, LevelNotice, "deprecated", "Call to deprecated method {%s}->%s() (%s)",
				call.methodCallerType, call.methodName, deprecation)
		} else {
			b.report(e.Method, LevelNotice, "deprecatedUntagged", "Call to deprecated method {%s}->%s()",
				call.methodCallerType, call.methodName)
		}
	}

	if call.isFound && !call.isMagic && !canAccess(parseState, call.className, call.info.AccessLevel) {
		b.report(e.Method, LevelError, "accessLevel", "Cannot access %s method %s->%s()", call.info.AccessLevel, call.className, call.methodName)
	}
}

func (b *blockLinter) checkSafetyCall(e ir.Node, typ types.Map, name string, suffix string) {
	if typ.Contains("null") {
		reportFullName := "notNullSafety" + suffix
		switch {
		case reportFullName == "notNullSafetyPropertyFetch":
			b.report(e, LevelWarning, "notNullSafety"+suffix,
				"potential attempt to access property through null")
			return
		case reportFullName == "notNullSafetyVariable" || reportFullName == "notNullSafetyFunctionCall":
			b.report(e, LevelWarning, reportFullName,
				"potential null dereference in %s when accessing method", name)
			return
		}
	}

	isSafetyCall := true
	typ.Iterate(func(typ string) {
		// TODO: here we can problem with mixed: $xs = [0, new Foo()]; $foo = $xs[0]; <== mixed. Need fix for array elem
		if types.IsScalar(typ) {
			isSafetyCall = false
		}
	})

	if !isSafetyCall {
		if name == "" {
			b.report(e, LevelWarning, "notSafeCall",
				"potentially not safe call when accessing property")
			return
		}
		b.report(e, LevelWarning, "notSafeCall",
			"potentially not safe call in %s when accessing method", name)
	}
}

func (b *blockLinter) checkStaticCall(e *ir.StaticCallExpr) {
	if utils.NameNodeToString(e.Class) == "parent" && b.classParseState().CurrentParentClass == "" {
		b.report(e, LevelError, "parentNotFound", "Cannot call method on parent as this class does not extend another")
		return
	}

	call := resolveStaticMethodCall(b.walker.ctx.sc, b.classParseState(), e)
	if !call.canAnalyze {
		return
	}

	b.checkMethodCallPackage(call.className, call.methodInfo.Info, e)

	b.checkClassSpecialNameCase(e, call.className)

	if !call.isMagic {
		b.checkCallArgs(e.Call, e.Args, call.methodInfo.Info, call.className)
	}

	if !call.isFound && !call.isMagic && !b.classParseState().IsTrait {
		b.report(e.Call, LevelError, "undefinedMethod", "Call to undefined method %s::%s()", call.className, call.methodName)
	} else if !call.isParentCall && !call.methodInfo.Info.IsStatic() && !call.isMagic && !b.classParseState().IsTrait {
		// Method is defined.
		// parent::f() is permitted.
		b.report(e.Call, LevelWarning, "callStatic", "Calling instance method as static method")
	}

	if call.methodInfo.Info.Deprecated {
		deprecation := call.methodInfo.Info.DeprecationInfo

		if deprecation.WithDeprecationNote() {
			b.report(e.Call, LevelNotice, "deprecated", "Call to deprecated static method %s::%s() (%s)",
				call.className, call.methodName, deprecation)
		} else {
			b.report(e.Call, LevelNotice, "deprecatedUntagged", "Call to deprecated static method %s::%s()",
				call.className, call.methodName)
		}
	}

	if call.isFound && !canAccess(b.classParseState(), call.methodInfo.ClassName, call.methodInfo.Info.AccessLevel) {
		b.report(e.Call, LevelError, "accessLevel", "Cannot access %s method %s::%s()", call.methodInfo.Info.AccessLevel, call.methodInfo.ClassName, call.methodName)
	}
}

func (b *blockLinter) checkMethodCallPackage(methodClassName string, methodInfo meta.FuncInfo, e ir.Node) {
	callClass, ok := b.classParseState().Info.GetClass(methodClassName)
	if !ok {
		return
	}
	if callClass.PackageInfo.Name == "" {
		return
	}
	if !callClass.PackageInfo.Internal && !methodInfo.Internal {
		return
	}

	currentClass, ok := b.walker.r.ctx.st.Info.GetClass(b.classParseState().CurrentClass)
	if !ok {
		return
	}

	if currentClass.PackageInfo.Name == callClass.PackageInfo.Name {
		return
	}

	b.report(e, LevelError, "packaging", "Call @internal method %s::%s outside package %s",
		methodClassName,
		methodInfo.Name,
		callClass.PackageInfo.Name,
	)
}

func (b *blockLinter) checkPropertyFetch(e *ir.PropertyFetchExpr) {
	globalMetaInfo := b.classParseState()

	fetch := resolvePropertyFetch(b.walker.ctx.sc, globalMetaInfo, b.walker.ctx.customTypes, e, b.walker.r.strictMixed)
	if !fetch.canAnalyze {
		return
	}

	needShowUndefinedProperty := !fetch.callerTypeIsMixed || b.walker.r.strictMixed

	if !fetch.isFound && !fetch.isMagic &&
		!globalMetaInfo.IsTrait &&
		!b.walker.isThisInsideClosure(e.Variable) &&
		needShowUndefinedProperty {
		b.report(e.Property, LevelError, "undefinedProperty", "Property {%s}->%s does not exist", fetch.propertyFetchType, fetch.propertyNode.Value)
	}

	if fetch.isFound && !fetch.isMagic && !canAccess(globalMetaInfo, fetch.className, fetch.info.AccessLevel) {
		b.report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s->%s", fetch.info.AccessLevel, fetch.className, fetch.propertyNode.Value)
	}

	left, ok := b.walker.ctx.sc.GetVarType(e.Variable)
	if ok {
		b.checkSafetyCall(e, left, "", "PropertyFetch")
	}
}

func (b *blockLinter) checkStaticPropertyFetch(e *ir.StaticPropertyFetchExpr) {
	globalMetaInfo := b.classParseState()
	fetch := resolveStaticPropertyFetch(globalMetaInfo, e)

	left, ok := globalMetaInfo.Info.GetVarType(e.Class)
	if ok && left.Contains("null") {
		b.report(e, LevelWarning, "notNullSafetyPropertyFetch",
			"attempt to access property that can be null")
	}

	if !fetch.canAnalyze {
		return
	}

	b.checkClassSpecialNameCase(e, fetch.className)

	if !fetch.isFound && !globalMetaInfo.IsTrait {
		b.report(e.Property, LevelError, "undefinedProperty", "Property %s::$%s does not exist", fetch.className, fetch.propertyName)
	}

	if fetch.isFound && !canAccess(globalMetaInfo, fetch.info.ClassName, fetch.info.Info.AccessLevel) {
		b.report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s::$%s", fetch.info.Info.AccessLevel, fetch.info.ClassName, fetch.propertyName)
	}
}

func (b *blockLinter) checkClassConstFetch(e *ir.ClassConstFetchExpr) {
	fetch := resolveClassConstFetch(b.classParseState(), e)
	if !fetch.canAnalyze {
		return
	}

	b.checkClassSpecialNameCase(e, fetch.className)

	if !utils.IsSpecialClassName(e.Class) {
		usedClassName, ok := solver.GetClassName(b.classParseState(), e.Class)
		if ok {
			b.walker.r.checker.CheckNameCase(e.Class, usedClassName, fetch.className)
		}
	}

	if !fetch.isFound && !b.classParseState().IsTrait {
		b.report(e.ConstantName, LevelError, "undefinedConstant", "Class constant %s::%s does not exist", fetch.className, fetch.constName)
	}

	if fetch.isFound && !canAccess(b.classParseState(), fetch.implClassName, fetch.info.AccessLevel) {
		b.report(e.ConstantName, LevelError, "accessLevel", "Cannot access %s constant %s::%s", fetch.info.AccessLevel, fetch.implClassName, fetch.constName)
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
			b.report(n, LevelNotice, "nameMismatch", "%s should be spelled as %s", strings.TrimPrefix(className, `\`), strings.TrimPrefix(name, `\`))
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
			phpDocParamTypes := b.walker.getParamsTypesFromPhpDoc(x.Doc)
			b.walker.CheckParamNullability(x.Params, phpDocParamTypes)
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
