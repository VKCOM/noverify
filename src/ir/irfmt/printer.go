package irfmt

import (
	"io"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
)

type PrettyPrinter struct {
	w           io.Writer
	indentStr   string
	indentDepth int
}

// NewPrettyPrinter -  Constructor for PrettyPrinter
func NewPrettyPrinter(w io.Writer, indentStr string) *PrettyPrinter {
	return &PrettyPrinter{
		w:           w,
		indentStr:   indentStr,
		indentDepth: 0,
	}
}

func (p *PrettyPrinter) Print(n ir.Node) {
	p.printNode(n)
}

func (p *PrettyPrinter) joinPrintIdents(glue string, items []*ir.Identifier) {
	for k, n := range items {
		if k > 0 {
			io.WriteString(p.w, glue)
		}

		p.Print(n)
	}
}

func (p *PrettyPrinter) joinPrintArrayItems(glue string, items []*ir.ArrayItemExpr) {
	for k, n := range items {
		if k > 0 {
			io.WriteString(p.w, glue)
		}

		p.Print(n)
	}
}

func (p *PrettyPrinter) joinPrint(glue string, nn []ir.Node) {
	for k, n := range nn {
		if k > 0 {
			io.WriteString(p.w, glue)
		}

		p.Print(n)
	}
}

func (p *PrettyPrinter) printNodes(nn []ir.Node) {
	p.indentDepth++
	l := len(nn) - 1
	for k, n := range nn {
		p.printIndent()
		p.Print(n)
		if k < l {
			io.WriteString(p.w, "\n")
		}
	}
	p.indentDepth--
}

func (p *PrettyPrinter) printIndent() {
	for i := 0; i < p.indentDepth; i++ {
		io.WriteString(p.w, p.indentStr)
	}
}

func (p *PrettyPrinter) printNode(n ir.Node) {
	switch n := n.(type) {

	// node

	case *ir.Root:
		p.printNodeRoot(n)
	case *ir.Identifier:
		p.printNodeIdentifier(n)
	case *ir.Parameter:
		p.printNodeParameter(n)
	case *ir.Nullable:
		p.printNodeNullable(n)
	case *ir.Argument:
		p.printNodeArgument(n)

		// name

	case *ir.NamePart:
		p.printNameNamePart(n)
	case *ir.Name:
		p.printNameName(n)
	case *ir.FullyQualifiedName:
		p.printNameFullyQualified(n)
	case *ir.RelativeName:
		p.printNameRelative(n)

		// scalar

	case *ir.Lnumber:
		p.printScalarLNumber(n)
	case *ir.Dnumber:
		p.printScalarDNumber(n)
	case *ir.String:
		p.printScalarString(n)
	case *ir.EncapsedStringPart:
		p.printScalarEncapsedStringPart(n)
	case *ir.Encapsed:
		p.printScalarEncapsed(n)
	case *ir.Heredoc:
		p.printScalarHeredoc(n)
	case *ir.MagicConstant:
		p.printScalarMagicConstant(n)

		// assign

	case *ir.Assign:
		p.printAssign(n)
	case *ir.AssignReference:
		p.printReference(n)
	case *ir.AssignBitwiseAnd:
		p.printAssignBitwiseAnd(n)
	case *ir.AssignBitwiseOr:
		p.printAssignBitwiseOr(n)
	case *ir.AssignBitwiseXor:
		p.printAssignBitwiseXor(n)
	case *ir.AssignConcat:
		p.printAssignConcat(n)
	case *ir.AssignDiv:
		p.printAssignDiv(n)
	case *ir.AssignMinus:
		p.printAssignMinus(n)
	case *ir.AssignMod:
		p.printAssignMod(n)
	case *ir.AssignMul:
		p.printAssignMul(n)
	case *ir.AssignPlus:
		p.printAssignPlus(n)
	case *ir.AssignPow:
		p.printAssignPow(n)
	case *ir.AssignShiftLeft:
		p.printAssignShiftLeft(n)
	case *ir.AssignShiftRight:
		p.printAssignShiftRight(n)

		// binary

	case *ir.BitwiseAndExpr:
		p.printBinaryBitwiseAnd(n)
	case *ir.BitwiseOrExpr:
		p.printBinaryBitwiseOr(n)
	case *ir.BitwiseXorExpr:
		p.printBinaryBitwiseXor(n)
	case *ir.BooleanAndExpr:
		p.printBinaryBooleanAnd(n)
	case *ir.BooleanOrExpr:
		p.printBinaryBooleanOr(n)
	case *ir.CoalesceExpr:
		p.printBinaryCoalesce(n)
	case *ir.ConcatExpr:
		p.printBinaryConcat(n)
	case *ir.DivExpr:
		p.printBinaryDiv(n)
	case *ir.EqualExpr:
		p.printBinaryEqual(n)
	case *ir.GreaterOrEqualExpr:
		p.printBinaryGreaterOrEqual(n)
	case *ir.GreaterExpr:
		p.printBinaryGreater(n)
	case *ir.IdenticalExpr:
		p.printBinaryIdentical(n)
	case *ir.LogicalAndExpr:
		p.printBinaryLogicalAnd(n)
	case *ir.LogicalOrExpr:
		p.printBinaryLogicalOr(n)
	case *ir.LogicalXorExpr:
		p.printBinaryLogicalXor(n)
	case *ir.MinusExpr:
		p.printBinaryMinus(n)
	case *ir.ModExpr:
		p.printBinaryMod(n)
	case *ir.MulExpr:
		p.printBinaryMul(n)
	case *ir.NotEqualExpr:
		p.printBinaryNotEqual(n)
	case *ir.NotIdenticalExpr:
		p.printBinaryNotIdentical(n)
	case *ir.PlusExpr:
		p.printBinaryPlus(n)
	case *ir.PowExpr:
		p.printBinaryPow(n)
	case *ir.ShiftLeftExpr:
		p.printBinaryShiftLeft(n)
	case *ir.ShiftRightExpr:
		p.printBinaryShiftRight(n)
	case *ir.SmallerOrEqualExpr:
		p.printBinarySmallerOrEqual(n)
	case *ir.SmallerExpr:
		p.printBinarySmaller(n)
	case *ir.SpaceshipExpr:
		p.printBinarySpaceship(n)

		// cast

	case *ir.ArrayCastExpr:
		p.printArray(n)
	case *ir.BoolCastExpr:
		p.printBool(n)
	case *ir.DoubleCastExpr:
		p.printDouble(n)
	case *ir.IntCastExpr:
		p.printInt(n)
	case *ir.ObjectCastExpr:
		p.printObject(n)
	case *ir.StringCastExpr:
		p.printString(n)
	case *ir.UnsetCastExpr:
		p.printUnset(n)

		// expr

	case *ir.ParenExpr:
		p.printExprParen(n)
	case *ir.ArrayDimFetchExpr:
		p.printExprArrayDimFetch(n)
	case *ir.ArrayItemExpr:
		p.printExprArrayItem(n)
	case *ir.ArrayExpr:
		p.printExprArray(n)
	case *ir.BitwiseNotExpr:
		p.printExprBitwiseNot(n)
	case *ir.BooleanNotExpr:
		p.printExprBooleanNot(n)
	case *ir.ClassConstFetchExpr:
		p.printExprClassConstFetch(n)
	case *ir.CloneExpr:
		p.printExprClone(n)
	case *ir.ClosureUseExpr:
		p.printExprClosureUse(n)
	case *ir.ClosureExpr:
		p.printExprClosure(n)
	case *ir.ConstFetchExpr:
		p.printExprConstFetch(n)
	case *ir.EmptyExpr:
		p.printExprEmpty(n)
	case *ir.ErrorSuppressExpr:
		p.printExprErrorSuppress(n)
	case *ir.EvalExpr:
		p.printExprEval(n)
	case *ir.ExitExpr:
		p.printExprExit(n)
	case *ir.FunctionCallExpr:
		p.printExprFunctionCall(n)
	case *ir.IncludeExpr:
		p.printExprInclude(n)
	case *ir.IncludeOnceExpr:
		p.printExprIncludeOnce(n)
	case *ir.InstanceOfExpr:
		p.printExprInstanceOf(n)
	case *ir.IssetExpr:
		p.printExprIsset(n)
	case *ir.ListExpr:
		p.printExprList(n)
	case *ir.MethodCallExpr:
		p.printExprMethodCall(n)
	case *ir.NewExpr:
		p.printExprNew(n)
	case *ir.PostDecExpr:
		p.printExprPostDec(n)
	case *ir.PostIncExpr:
		p.printExprPostInc(n)
	case *ir.PreDecExpr:
		p.printExprPreDec(n)
	case *ir.PreIncExpr:
		p.printExprPreInc(n)
	case *ir.PrintExpr:
		p.printExprPrint(n)
	case *ir.PropertyFetchExpr:
		p.printExprPropertyFetch(n)
	case *ir.ReferenceExpr:
		p.printExprReference(n)
	case *ir.RequireExpr:
		p.printExprRequire(n)
	case *ir.RequireOnceExpr:
		p.printExprRequireOnce(n)
	case *ir.ShellExecExpr:
		p.printExprShellExec(n)
	case *ir.StaticCallExpr:
		p.printExprStaticCall(n)
	case *ir.StaticPropertyFetchExpr:
		p.printExprStaticPropertyFetch(n)
	case *ir.TernaryExpr:
		p.printExprTernary(n)
	case *ir.UnaryMinusExpr:
		p.printExprUnaryMinus(n)
	case *ir.UnaryPlusExpr:
		p.printExprUnaryPlus(n)
	case *ir.Var:
		p.printExprVar(n)
	case *ir.SimpleVar:
		p.printExprSimpleVar(n)
	case *ir.YieldFromExpr:
		p.printExprYieldFrom(n)
	case *ir.YieldExpr:
		p.printExprYield(n)

		// stmt

	case *ir.BreakStmt:
		p.printStmtBreak(n)
	case *ir.CaseStmt:
		p.printStmtCase(n)
	case *ir.CatchStmt:
		p.printStmtCatch(n)
	case *ir.ClassMethodStmt:
		p.printStmtClassMethod(n)
	case *ir.ClassStmt:
		p.printStmtClass(n)
	case *ir.ClassConstListStmt:
		p.printStmtClassConstList(n)
	case *ir.ConstantStmt:
		p.printStmtConstant(n)
	case *ir.ContinueStmt:
		p.printStmtContinue(n)
	case *ir.DeclareStmt:
		p.printStmtDeclare(n)
	case *ir.DefaultStmt:
		p.printStmtDefault(n)
	case *ir.DoStmt:
		p.printStmtDo(n)
	case *ir.EchoStmt:
		p.printStmtEcho(n)
	case *ir.ElseIfStmt:
		p.printStmtElseif(n)
	case *ir.ElseStmt:
		p.printStmtElse(n)
	case *ir.ExpressionStmt:
		p.printStmtExpression(n)
	case *ir.FinallyStmt:
		p.printStmtFinally(n)
	case *ir.ForStmt:
		p.printStmtFor(n)
	case *ir.ForeachStmt:
		p.printStmtForeach(n)
	case *ir.FunctionStmt:
		p.printStmtFunction(n)
	case *ir.GlobalStmt:
		p.printStmtGlobal(n)
	case *ir.GotoStmt:
		p.printStmtGoto(n)
	case *ir.GroupUseStmt:
		p.printStmtGroupUse(n)
	case *ir.HaltCompilerStmt:
		p.printStmtHaltCompiler(n)
	case *ir.IfStmt:
		p.printStmtIf(n)
	case *ir.InlineHTMLStmt:
		p.printStmtInlineHTML(n)
	case *ir.InterfaceStmt:
		p.printStmtInterface(n)
	case *ir.LabelStmt:
		p.printStmtLabel(n)
	case *ir.NamespaceStmt:
		p.printStmtNamespace(n)
	case *ir.NopStmt:
		p.printStmtNop(n)
	case *ir.PropertyListStmt:
		p.printStmtPropertyList(n)
	case *ir.PropertyStmt:
		p.printStmtProperty(n)
	case *ir.ReturnStmt:
		p.printStmtReturn(n)
	case *ir.StaticVarStmt:
		p.printStmtStaticVar(n)
	case *ir.StaticStmt:
		p.printStmtStatic(n)
	case *ir.StmtList:
		p.printStmtStmtList(n)
	case *ir.SwitchStmt:
		p.printStmtSwitch(n)
	case *ir.ThrowStmt:
		p.printStmtThrow(n)
	case *ir.TraitMethodRefStmt:
		p.printStmtTraitMethodRef(n)
	case *ir.TraitUseAliasStmt:
		p.printStmtTraitUseAlias(n)
	case *ir.TraitUsePrecedenceStmt:
		p.printStmtTraitUsePrecedence(n)
	case *ir.TraitUseStmt:
		p.printStmtTraitUse(n)
	case *ir.TraitStmt:
		p.printStmtTrait(n)
	case *ir.TryStmt:
		p.printStmtTry(n)
	case *ir.UnsetStmt:
		p.printStmtUnset(n)
	case *ir.UseListStmt:
		p.printStmtUseList(n)
	case *ir.UseStmt:
		p.printStmtUse(n)
	case *ir.WhileStmt:
		p.printStmtWhile(n)
	}
}

// node

func (p *PrettyPrinter) printNodeRoot(n ir.Node) {
	v := n.(*ir.Root)

	stmts := v.Stmts
	if len(stmts) > 0 {
		firstStmt := stmts[0]
		stmts = stmts[1:]

		switch fs := firstStmt.(type) {
		case *ir.InlineHTMLStmt:
			io.WriteString(p.w, fs.Value)
			io.WriteString(p.w, "<?php\n")
		default:
			io.WriteString(p.w, "<?php\n")
			p.printIndent()
			p.Print(fs)
			io.WriteString(p.w, "\n")
		}
	}
	p.indentDepth--
	p.printNodes(stmts)
	io.WriteString(p.w, "\n")
}

func (p *PrettyPrinter) printNodeIdentifier(n ir.Node) {
	v := n.(*ir.Identifier).Value
	io.WriteString(p.w, v)
}

func (p *PrettyPrinter) printNodeParameter(n ir.Node) {
	nn := n.(*ir.Parameter)

	if nn.VariableType != nil {
		p.Print(nn.VariableType)
		io.WriteString(p.w, " ")
	}

	if nn.ByRef {
		io.WriteString(p.w, "&")
	}

	if nn.Variadic {
		io.WriteString(p.w, "...")
	}

	p.Print(nn.Variable)

	if nn.DefaultValue != nil {
		io.WriteString(p.w, " = ")
		p.Print(nn.DefaultValue)
	}
}

func (p *PrettyPrinter) printNodeNullable(n ir.Node) {
	nn := n.(*ir.Nullable)

	io.WriteString(p.w, "?")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printNodeArgument(n ir.Node) {
	nn := n.(*ir.Argument)

	if nn.IsReference {
		io.WriteString(p.w, "&")
	}

	if nn.Variadic {
		io.WriteString(p.w, "...")
	}

	p.Print(nn.Expr)
}

// name

func (p *PrettyPrinter) printNameNamePart(n ir.Node) {
	v := n.(*ir.NamePart).Value
	io.WriteString(p.w, v)
}

func (p *PrettyPrinter) printNameName(n ir.Node) {
	nn := n.(*ir.Name)

	for k, part := range nn.Parts {
		if k > 0 {
			io.WriteString(p.w, "\\")
		}

		p.Print(part)
	}
}

func (p *PrettyPrinter) printNameFullyQualified(n ir.Node) {
	nn := n.(*ir.FullyQualifiedName)

	for _, part := range nn.Parts {
		io.WriteString(p.w, "\\")
		p.Print(part)
	}
}

func (p *PrettyPrinter) printNameRelative(n ir.Node) {
	nn := n.(*ir.RelativeName)

	io.WriteString(p.w, "namespace")
	for _, part := range nn.Parts {
		io.WriteString(p.w, "\\")
		p.Print(part)
	}
}

// scalar

func (p *PrettyPrinter) printScalarLNumber(n ir.Node) {
	v := n.(*ir.Lnumber).Value
	io.WriteString(p.w, v)
}

func (p *PrettyPrinter) printScalarDNumber(n ir.Node) {
	v := n.(*ir.Dnumber).Value
	io.WriteString(p.w, v)
}

func (p *PrettyPrinter) printScalarString(n ir.Node) {
	v := n.(*ir.String).Value

	io.WriteString(p.w, v)
}

func (p *PrettyPrinter) printScalarEncapsedStringPart(n ir.Node) {
	v := n.(*ir.EncapsedStringPart).Value
	io.WriteString(p.w, v)
}

func (p *PrettyPrinter) printScalarEncapsed(n ir.Node) {
	nn := n.(*ir.Encapsed)
	io.WriteString(p.w, "\"")

	for _, part := range nn.Parts {
		switch part.(type) {
		case *ir.EncapsedStringPart:
			p.Print(part)
		default:
			io.WriteString(p.w, "{")
			p.Print(part)
			io.WriteString(p.w, "}")
		}
	}

	io.WriteString(p.w, "\"")
}

func (p *PrettyPrinter) printScalarHeredoc(n ir.Node) {
	nn := n.(*ir.Heredoc)

	io.WriteString(p.w, nn.Label)

	for _, part := range nn.Parts {
		switch part.(type) {
		case *ir.EncapsedStringPart:
			p.Print(part)
		default:
			io.WriteString(p.w, "{")
			p.Print(part)
			io.WriteString(p.w, "}")
		}
	}

	io.WriteString(p.w, strings.Trim(nn.Label, "<\"'\n"))
}

func (p *PrettyPrinter) printScalarMagicConstant(n ir.Node) {
	v := n.(*ir.MagicConstant).Value
	io.WriteString(p.w, v)
}

// Assign

func (p *PrettyPrinter) printAssign(n ir.Node) {
	nn := n.(*ir.Assign)
	p.Print(nn.Variable)
	io.WriteString(p.w, " = ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printReference(n ir.Node) {
	nn := n.(*ir.AssignReference)
	p.Print(nn.Variable)
	io.WriteString(p.w, " =& ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignBitwiseAnd(n ir.Node) {
	nn := n.(*ir.AssignBitwiseAnd)
	p.Print(nn.Variable)
	io.WriteString(p.w, " &= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignBitwiseOr(n ir.Node) {
	nn := n.(*ir.AssignBitwiseOr)
	p.Print(nn.Variable)
	io.WriteString(p.w, " |= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignBitwiseXor(n ir.Node) {
	nn := n.(*ir.AssignBitwiseXor)
	p.Print(nn.Variable)
	io.WriteString(p.w, " ^= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignConcat(n ir.Node) {
	nn := n.(*ir.AssignConcat)
	p.Print(nn.Variable)
	io.WriteString(p.w, " .= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignDiv(n ir.Node) {
	nn := n.(*ir.AssignDiv)
	p.Print(nn.Variable)
	io.WriteString(p.w, " /= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignMinus(n ir.Node) {
	nn := n.(*ir.AssignMinus)
	p.Print(nn.Variable)
	io.WriteString(p.w, " -= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignMod(n ir.Node) {
	nn := n.(*ir.AssignMod)
	p.Print(nn.Variable)
	io.WriteString(p.w, " %= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignMul(n ir.Node) {
	nn := n.(*ir.AssignMul)
	p.Print(nn.Variable)
	io.WriteString(p.w, " *= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignPlus(n ir.Node) {
	nn := n.(*ir.AssignPlus)
	p.Print(nn.Variable)
	io.WriteString(p.w, " += ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignPow(n ir.Node) {
	nn := n.(*ir.AssignPow)
	p.Print(nn.Variable)
	io.WriteString(p.w, " **= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignShiftLeft(n ir.Node) {
	nn := n.(*ir.AssignShiftLeft)
	p.Print(nn.Variable)
	io.WriteString(p.w, " <<= ")
	p.Print(nn.Expression)
}

func (p *PrettyPrinter) printAssignShiftRight(n ir.Node) {
	nn := n.(*ir.AssignShiftRight)
	p.Print(nn.Variable)
	io.WriteString(p.w, " >>= ")
	p.Print(nn.Expression)
}

// binary

func (p *PrettyPrinter) printBinaryBitwiseAnd(n ir.Node) {
	nn := n.(*ir.BitwiseAndExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " & ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryBitwiseOr(n ir.Node) {
	nn := n.(*ir.BitwiseOrExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " | ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryBitwiseXor(n ir.Node) {
	nn := n.(*ir.BitwiseXorExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " ^ ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryBooleanAnd(n ir.Node) {
	nn := n.(*ir.BooleanAndExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " && ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryBooleanOr(n ir.Node) {
	nn := n.(*ir.BooleanOrExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " || ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryCoalesce(n ir.Node) {
	nn := n.(*ir.CoalesceExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " ?? ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryConcat(n ir.Node) {
	nn := n.(*ir.ConcatExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " . ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryDiv(n ir.Node) {
	nn := n.(*ir.DivExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " / ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryEqual(n ir.Node) {
	nn := n.(*ir.EqualExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " == ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryGreaterOrEqual(n ir.Node) {
	nn := n.(*ir.GreaterOrEqualExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " >= ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryGreater(n ir.Node) {
	nn := n.(*ir.GreaterExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " > ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryIdentical(n ir.Node) {
	nn := n.(*ir.IdenticalExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " === ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryLogicalAnd(n ir.Node) {
	nn := n.(*ir.LogicalAndExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " and ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryLogicalOr(n ir.Node) {
	nn := n.(*ir.LogicalOrExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " or ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryLogicalXor(n ir.Node) {
	nn := n.(*ir.LogicalXorExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " xor ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryMinus(n ir.Node) {
	nn := n.(*ir.MinusExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " - ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryMod(n ir.Node) {
	nn := n.(*ir.ModExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " % ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryMul(n ir.Node) {
	nn := n.(*ir.MulExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " * ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryNotEqual(n ir.Node) {
	nn := n.(*ir.NotEqualExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " != ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryNotIdentical(n ir.Node) {
	nn := n.(*ir.NotIdenticalExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " !== ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryPlus(n ir.Node) {
	nn := n.(*ir.PlusExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " + ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryPow(n ir.Node) {
	nn := n.(*ir.PowExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " ** ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryShiftLeft(n ir.Node) {
	nn := n.(*ir.ShiftLeftExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " << ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinaryShiftRight(n ir.Node) {
	nn := n.(*ir.ShiftRightExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " >> ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinarySmallerOrEqual(n ir.Node) {
	nn := n.(*ir.SmallerOrEqualExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " <= ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinarySmaller(n ir.Node) {
	nn := n.(*ir.SmallerExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " < ")
	p.Print(nn.Right)
}

func (p *PrettyPrinter) printBinarySpaceship(n ir.Node) {
	nn := n.(*ir.SpaceshipExpr)

	p.Print(nn.Left)
	io.WriteString(p.w, " <=> ")
	p.Print(nn.Right)
}

// cast

func (p *PrettyPrinter) printArray(n ir.Node) {
	nn := n.(*ir.ArrayCastExpr)

	io.WriteString(p.w, "(array)")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printBool(n ir.Node) {
	nn := n.(*ir.BoolCastExpr)

	io.WriteString(p.w, "(bool)")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printDouble(n ir.Node) {
	nn := n.(*ir.DoubleCastExpr)

	io.WriteString(p.w, "(float)")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printInt(n ir.Node) {
	nn := n.(*ir.IntCastExpr)

	io.WriteString(p.w, "(int)")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printObject(n ir.Node) {
	nn := n.(*ir.ObjectCastExpr)

	io.WriteString(p.w, "(object)")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printString(n ir.Node) {
	nn := n.(*ir.StringCastExpr)

	io.WriteString(p.w, "(string)")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printUnset(n ir.Node) {
	nn := n.(*ir.UnsetCastExpr)

	io.WriteString(p.w, "(unset)")
	p.Print(nn.Expr)
}

// expr

func (p *PrettyPrinter) printExprParen(n *ir.ParenExpr) {
	io.WriteString(p.w, "(")
	p.Print(n.Expr)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprArrayDimFetch(n ir.Node) {
	nn := n.(*ir.ArrayDimFetchExpr)
	p.Print(nn.Variable)
	io.WriteString(p.w, "[")
	p.Print(nn.Dim)
	io.WriteString(p.w, "]")
}

func (p *PrettyPrinter) printExprArrayItem(n ir.Node) {
	nn := n.(*ir.ArrayItemExpr)

	if nn.Key != nil {
		p.Print(nn.Key)
		io.WriteString(p.w, " => ")
	}

	p.Print(nn.Val)
}

func (p *PrettyPrinter) printExprArray(n ir.Node) {
	nn := n.(*ir.ArrayExpr)

	if nn.ShortSyntax {
		io.WriteString(p.w, "[")
		p.joinPrintArrayItems(", ", nn.Items)
		io.WriteString(p.w, "]")
	} else {
		io.WriteString(p.w, "array(")
		p.joinPrintArrayItems(", ", nn.Items)
		io.WriteString(p.w, ")")
	}
}

func (p *PrettyPrinter) printExprBitwiseNot(n ir.Node) {
	nn := n.(*ir.BitwiseNotExpr)
	io.WriteString(p.w, "~")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprBooleanNot(n ir.Node) {
	nn := n.(*ir.BooleanNotExpr)
	io.WriteString(p.w, "!")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprClassConstFetch(n ir.Node) {
	nn := n.(*ir.ClassConstFetchExpr)

	p.Print(nn.Class)
	io.WriteString(p.w, "::")
	io.WriteString(p.w, nn.ConstantName.Value)
}

func (p *PrettyPrinter) printExprClone(n ir.Node) {
	nn := n.(*ir.CloneExpr)

	io.WriteString(p.w, "clone ")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprClosureUse(n ir.Node) {
	nn := n.(*ir.ClosureUseExpr)

	io.WriteString(p.w, "use (")
	p.joinPrint(", ", nn.Uses)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprClosure(n ir.Node) {
	nn := n.(*ir.ClosureExpr)

	if nn.Static {
		io.WriteString(p.w, "static ")
	}

	io.WriteString(p.w, "function ")

	if nn.ReturnsRef {
		io.WriteString(p.w, "&")
	}

	io.WriteString(p.w, "(")
	p.joinPrint(", ", nn.Params)
	io.WriteString(p.w, ")")

	if nn.ClosureUse != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.ClosureUse)
	}

	if nn.ReturnType != nil {
		io.WriteString(p.w, ": ")
		p.Print(nn.ReturnType)
	}

	io.WriteString(p.w, " {\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printExprConstFetch(n ir.Node) {
	nn := n.(*ir.ConstFetchExpr)

	p.Print(nn.Constant)
}

func (p *PrettyPrinter) printExprEmpty(n ir.Node) {
	nn := n.(*ir.EmptyExpr)

	io.WriteString(p.w, "empty(")
	p.Print(nn.Expr)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprErrorSuppress(n ir.Node) {
	nn := n.(*ir.ErrorSuppressExpr)

	io.WriteString(p.w, "@")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprEval(n ir.Node) {
	nn := n.(*ir.EvalExpr)

	io.WriteString(p.w, "eval(")
	p.Print(nn.Expr)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprExit(n ir.Node) {
	nn := n.(*ir.ExitExpr)

	if nn.Die {
		io.WriteString(p.w, "die(")
	} else {
		io.WriteString(p.w, "exit(")
	}
	p.Print(nn.Expr)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprFunctionCall(n ir.Node) {
	nn := n.(*ir.FunctionCallExpr)

	p.Print(nn.Function)
	io.WriteString(p.w, "(")
	p.joinPrint(", ", nn.ArgumentList.Arguments)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprInclude(n ir.Node) {
	nn := n.(*ir.IncludeExpr)

	io.WriteString(p.w, "include ")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprIncludeOnce(n ir.Node) {
	nn := n.(*ir.IncludeOnceExpr)

	io.WriteString(p.w, "include_once ")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprInstanceOf(n ir.Node) {
	nn := n.(*ir.InstanceOfExpr)

	p.Print(nn.Expr)
	io.WriteString(p.w, " instanceof ")
	p.Print(nn.Class)
}

func (p *PrettyPrinter) printExprIsset(n ir.Node) {
	nn := n.(*ir.IssetExpr)

	io.WriteString(p.w, "isset(")
	p.joinPrint(", ", nn.Variables)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprList(n ir.Node) {
	nn := n.(*ir.ListExpr)

	if nn.ShortSyntax {
		io.WriteString(p.w, "[")
		p.joinPrintArrayItems(", ", nn.Items)
		io.WriteString(p.w, "]")
	} else {
		io.WriteString(p.w, "list(")
		p.joinPrintArrayItems(", ", nn.Items)
		io.WriteString(p.w, ")")
	}
}

func (p *PrettyPrinter) printExprMethodCall(n ir.Node) {
	nn := n.(*ir.MethodCallExpr)

	p.Print(nn.Variable)
	io.WriteString(p.w, "->")
	p.Print(nn.Method)
	io.WriteString(p.w, "(")
	p.joinPrint(", ", nn.ArgumentList.Arguments)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprNew(n ir.Node) {
	nn := n.(*ir.NewExpr)

	io.WriteString(p.w, "new ")
	p.Print(nn.Class)

	if nn.ArgumentList != nil {
		io.WriteString(p.w, "(")
		p.joinPrint(", ", nn.ArgumentList.Arguments)
		io.WriteString(p.w, ")")
	}
}

func (p *PrettyPrinter) printExprPostDec(n ir.Node) {
	nn := n.(*ir.PostDecExpr)

	p.Print(nn.Variable)
	io.WriteString(p.w, "--")
}

func (p *PrettyPrinter) printExprPostInc(n ir.Node) {
	nn := n.(*ir.PostIncExpr)

	p.Print(nn.Variable)
	io.WriteString(p.w, "++")
}

func (p *PrettyPrinter) printExprPreDec(n ir.Node) {
	nn := n.(*ir.PreDecExpr)

	io.WriteString(p.w, "--")
	p.Print(nn.Variable)
}

func (p *PrettyPrinter) printExprPreInc(n ir.Node) {
	nn := n.(*ir.PreIncExpr)

	io.WriteString(p.w, "++")
	p.Print(nn.Variable)
}

func (p *PrettyPrinter) printExprPrint(n ir.Node) {
	nn := n.(*ir.PrintExpr)

	io.WriteString(p.w, "print(")
	p.Print(nn.Expr)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprPropertyFetch(n ir.Node) {
	nn := n.(*ir.PropertyFetchExpr)

	p.Print(nn.Variable)
	io.WriteString(p.w, "->")
	p.Print(nn.Property)
}

func (p *PrettyPrinter) printExprReference(n ir.Node) {
	nn := n.(*ir.ReferenceExpr)

	io.WriteString(p.w, "&")
	p.Print(nn.Variable)
}

func (p *PrettyPrinter) printExprRequire(n ir.Node) {
	nn := n.(*ir.RequireExpr)

	io.WriteString(p.w, "require ")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprRequireOnce(n ir.Node) {
	nn := n.(*ir.RequireOnceExpr)

	io.WriteString(p.w, "require_once ")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprShellExec(n ir.Node) {
	nn := n.(*ir.ShellExecExpr)

	io.WriteString(p.w, "`")
	for _, part := range nn.Parts {
		switch part.(type) {
		case *ir.EncapsedStringPart:
			p.Print(part)
		default:
			io.WriteString(p.w, "{")
			p.Print(part)
			io.WriteString(p.w, "}")
		}
	}
	io.WriteString(p.w, "`")
}

func (p *PrettyPrinter) printExprStaticCall(n ir.Node) {
	nn := n.(*ir.StaticCallExpr)

	p.Print(nn.Class)
	io.WriteString(p.w, "::")
	p.Print(nn.Call)
	io.WriteString(p.w, "(")
	p.joinPrint(", ", nn.ArgumentList.Arguments)
	io.WriteString(p.w, ")")
}

func (p *PrettyPrinter) printExprStaticPropertyFetch(n ir.Node) {
	nn := n.(*ir.StaticPropertyFetchExpr)

	p.Print(nn.Class)
	io.WriteString(p.w, "::")
	p.Print(nn.Property)
}

func (p *PrettyPrinter) printExprTernary(n ir.Node) {
	nn := n.(*ir.TernaryExpr)

	p.Print(nn.Condition)
	io.WriteString(p.w, " ?")

	if nn.IfTrue != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.IfTrue)
		io.WriteString(p.w, " ")
	}

	io.WriteString(p.w, ": ")
	p.Print(nn.IfFalse)
}

func (p *PrettyPrinter) printExprUnaryMinus(n ir.Node) {
	nn := n.(*ir.UnaryMinusExpr)

	io.WriteString(p.w, "-")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprUnaryPlus(n ir.Node) {
	nn := n.(*ir.UnaryPlusExpr)

	io.WriteString(p.w, "+")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprSimpleVar(nn *ir.SimpleVar) {
	io.WriteString(p.w, "$"+nn.Name)
}

func (p *PrettyPrinter) printExprVar(n ir.Node) {
	nn := n.(*ir.Var)
	io.WriteString(p.w, "$")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprYieldFrom(n ir.Node) {
	nn := n.(*ir.YieldFromExpr)

	io.WriteString(p.w, "yield from ")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printExprYield(n ir.Node) {
	nn := n.(*ir.YieldExpr)

	io.WriteString(p.w, "yield ")

	if nn.Key != nil {
		p.Print(nn.Key)
		io.WriteString(p.w, " => ")
	}

	p.Print(nn.Value)
}

// smtm

func (p *PrettyPrinter) printStmtBreak(n ir.Node) {
	nn := n.(*ir.BreakStmt)

	io.WriteString(p.w, "break")
	if nn.Expr != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.Expr)
	}

	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtCase(n ir.Node) {
	nn := n.(*ir.CaseStmt)

	io.WriteString(p.w, "case ")
	p.Print(nn.Cond)
	io.WriteString(p.w, ":")

	if len(nn.Stmts) > 0 {
		io.WriteString(p.w, "\n")
		p.printNodes(nn.Stmts)
	}
}

func (p *PrettyPrinter) printStmtCatch(n ir.Node) {
	nn := n.(*ir.CatchStmt)

	io.WriteString(p.w, "catch (")
	p.joinPrint(" | ", nn.Types)
	io.WriteString(p.w, " ")
	p.Print(nn.Variable)
	io.WriteString(p.w, ") {\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printStmtClassMethod(n ir.Node) {
	nn := n.(*ir.ClassMethodStmt)

	if nn.Modifiers != nil {
		p.joinPrintIdents(" ", nn.Modifiers)
		io.WriteString(p.w, " ")
	}
	io.WriteString(p.w, "function ")

	if nn.ReturnsRef {
		io.WriteString(p.w, "&")
	}

	p.Print(nn.MethodName)
	io.WriteString(p.w, "(")
	p.joinPrint(", ", nn.Params)
	io.WriteString(p.w, ")")

	if nn.ReturnType != nil {
		io.WriteString(p.w, ": ")
		p.Print(nn.ReturnType)
	}

	switch s := nn.Stmt.(type) {
	case *ir.StmtList:
		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "{\n")
		p.printNodes(s.Stmts)
		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "}")
	default:
		p.Print(s)
	}
}

func (p *PrettyPrinter) printStmtClass(n ir.Node) {
	nn := n.(*ir.ClassStmt)

	if nn.Modifiers != nil {
		p.joinPrintIdents(" ", nn.Modifiers)
		io.WriteString(p.w, " ")
	}
	io.WriteString(p.w, "class")

	if nn.ClassName != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.ClassName)
	}

	if nn.ArgumentList != nil {
		io.WriteString(p.w, "(")
		p.joinPrint(", ", nn.ArgumentList.Arguments)
		io.WriteString(p.w, ")")
	}

	if nn.Extends != nil {
		io.WriteString(p.w, " extends ")
		p.Print(nn.Extends.ClassName)
	}

	if nn.Implements != nil {
		io.WriteString(p.w, " implements ")
		p.joinPrint(", ", nn.Implements.InterfaceNames)
	}

	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "{\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printStmtClassConstList(n ir.Node) {
	nn := n.(*ir.ClassConstListStmt)

	if nn.Modifiers != nil {
		p.joinPrintIdents(" ", nn.Modifiers)
		io.WriteString(p.w, " ")
	}
	io.WriteString(p.w, "const ")

	p.joinPrint(", ", nn.Consts)

	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtConstant(n ir.Node) {
	nn := n.(*ir.ConstantStmt)

	p.Print(nn.ConstantName)
	io.WriteString(p.w, " = ")
	p.Print(nn.Expr)
}

func (p *PrettyPrinter) printStmtContinue(n ir.Node) {
	nn := n.(*ir.ContinueStmt)

	io.WriteString(p.w, "continue")
	if nn.Expr != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.Expr)
	}

	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtDeclare(n ir.Node) {
	nn := n.(*ir.DeclareStmt)

	io.WriteString(p.w, "declare(")
	p.joinPrint(", ", nn.Consts)
	io.WriteString(p.w, ")")

	switch s := nn.Stmt.(type) {
	case *ir.NopStmt:
		p.Print(s)
	case *ir.StmtList:
		io.WriteString(p.w, " ")
		p.Print(s)
	default:
		io.WriteString(p.w, "\n")
		p.indentDepth++
		p.printIndent()
		p.Print(s)
		p.indentDepth--
	}
}

func (p *PrettyPrinter) printStmtDefault(n ir.Node) {
	nn := n.(*ir.DefaultStmt)
	io.WriteString(p.w, "default:")

	if len(nn.Stmts) > 0 {
		io.WriteString(p.w, "\n")
		p.printNodes(nn.Stmts)
	}
}

func (p *PrettyPrinter) printStmtDo(n ir.Node) {
	nn := n.(*ir.DoStmt)
	io.WriteString(p.w, "do")

	switch s := nn.Stmt.(type) {
	case *ir.StmtList:
		io.WriteString(p.w, " ")
		p.Print(s)
		io.WriteString(p.w, " ")
	default:
		io.WriteString(p.w, "\n")
		p.indentDepth++
		p.printIndent()
		p.Print(s)
		p.indentDepth--
		io.WriteString(p.w, "\n")
		p.printIndent()
	}

	io.WriteString(p.w, "while (")
	p.Print(nn.Cond)
	io.WriteString(p.w, ");")
}

func (p *PrettyPrinter) printStmtEcho(n ir.Node) {
	nn := n.(*ir.EchoStmt)
	io.WriteString(p.w, "echo ")
	p.joinPrint(", ", nn.Exprs)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtElseif(n ir.Node) {
	nn := n.(*ir.ElseIfStmt)

	io.WriteString(p.w, "elseif (")
	p.Print(nn.Cond)

	if nn.AltSyntax {
		io.WriteString(p.w, ") :")

		if s := nn.Stmt.(*ir.StmtList).Stmts; len(s) > 0 {
			io.WriteString(p.w, "\n")
			p.printNodes(s)
		}
	} else {
		io.WriteString(p.w, ")")

		switch s := nn.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			io.WriteString(p.w, " ")
			p.Print(s)
		default:
			io.WriteString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtElse(n ir.Node) {
	nn := n.(*ir.ElseStmt)

	if nn.AltSyntax {
		io.WriteString(p.w, "else :")

		if s := nn.Stmt.(*ir.StmtList).Stmts; len(s) > 0 {
			io.WriteString(p.w, "\n")
			p.printNodes(s)
		}
	} else {
		io.WriteString(p.w, "else")

		switch s := nn.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			io.WriteString(p.w, " ")
			p.Print(s)
		default:
			io.WriteString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtExpression(n ir.Node) {
	nn := n.(*ir.ExpressionStmt)

	p.Print(nn.Expr)

	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtFinally(n ir.Node) {
	nn := n.(*ir.FinallyStmt)

	io.WriteString(p.w, "finally {\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printStmtFor(n ir.Node) {
	nn := n.(*ir.ForStmt)

	io.WriteString(p.w, "for (")
	p.joinPrint(", ", nn.Init)
	io.WriteString(p.w, "; ")
	p.joinPrint(", ", nn.Cond)
	io.WriteString(p.w, "; ")
	p.joinPrint(", ", nn.Loop)

	if nn.AltSyntax {
		io.WriteString(p.w, ") :\n")

		s := nn.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)
		io.WriteString(p.w, "\n")
		p.printIndent()

		io.WriteString(p.w, "endfor;")
	} else {
		io.WriteString(p.w, ")")

		switch s := nn.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			io.WriteString(p.w, " ")
			p.Print(s)
		default:
			io.WriteString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtForeach(n ir.Node) {
	nn := n.(*ir.ForeachStmt)

	io.WriteString(p.w, "foreach (")
	p.Print(nn.Expr)
	io.WriteString(p.w, " as ")

	if nn.Key != nil {
		p.Print(nn.Key)
		io.WriteString(p.w, " => ")
	}

	p.Print(nn.Variable)
	io.WriteString(p.w, ")")

	if nn.AltSyntax {
		io.WriteString(p.w, " :\n")

		s := nn.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)

		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "endforeach;")
	} else {
		switch s := nn.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			io.WriteString(p.w, " ")
			p.Print(s)
		default:
			io.WriteString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtFunction(n ir.Node) {
	nn := n.(*ir.FunctionStmt)

	io.WriteString(p.w, "function ")

	if nn.ReturnsRef {
		io.WriteString(p.w, "&")
	}

	p.Print(nn.FunctionName)

	io.WriteString(p.w, "(")
	p.joinPrint(", ", nn.Params)
	io.WriteString(p.w, ")")

	if nn.ReturnType != nil {
		io.WriteString(p.w, ": ")
		p.Print(nn.ReturnType)
	}

	io.WriteString(p.w, " {\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printStmtGlobal(n ir.Node) {
	nn := n.(*ir.GlobalStmt)

	io.WriteString(p.w, "global ")
	p.joinPrint(", ", nn.Vars)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtGoto(n ir.Node) {
	nn := n.(*ir.GotoStmt)

	io.WriteString(p.w, "goto ")
	p.Print(nn.Label)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtGroupUse(n ir.Node) {
	nn := n.(*ir.GroupUseStmt)

	io.WriteString(p.w, "use ")

	if nn.UseType != nil {
		p.Print(nn.UseType)
		io.WriteString(p.w, " ")
	}

	p.Print(nn.Prefix)
	io.WriteString(p.w, "\\{")
	p.joinPrint(", ", nn.UseList)
	io.WriteString(p.w, "};")
}

func (p *PrettyPrinter) printStmtHaltCompiler(n ir.Node) {
	io.WriteString(p.w, "__halt_compiler();")
}

func (p *PrettyPrinter) printStmtIf(n ir.Node) {
	nn := n.(*ir.IfStmt)

	io.WriteString(p.w, "if (")
	p.Print(nn.Cond)
	io.WriteString(p.w, ")")

	if nn.AltSyntax {
		io.WriteString(p.w, " :\n")

		s := nn.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)

		for _, elseif := range nn.ElseIf {
			io.WriteString(p.w, "\n")
			p.printIndent()
			p.Print(elseif)
		}

		if nn.Else != nil {
			io.WriteString(p.w, "\n")
			p.printIndent()
			p.Print(nn.Else)
		}

		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "endif;")
	} else {
		switch s := nn.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			io.WriteString(p.w, " ")
			p.Print(s)
		default:
			io.WriteString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}

		if nn.ElseIf != nil {
			io.WriteString(p.w, "\n")
			p.indentDepth--
			p.printNodes(nn.ElseIf)
			p.indentDepth++
		}

		if nn.Else != nil {
			io.WriteString(p.w, "\n")
			p.printIndent()
			p.Print(nn.Else)
		}
	}
}

func (p *PrettyPrinter) printStmtInlineHTML(n ir.Node) {
	nn := n.(*ir.InlineHTMLStmt)

	io.WriteString(p.w, "?>")
	io.WriteString(p.w, nn.Value)
	io.WriteString(p.w, "<?php")
}

func (p *PrettyPrinter) printStmtInterface(n ir.Node) {
	nn := n.(*ir.InterfaceStmt)

	io.WriteString(p.w, "interface")

	if nn.InterfaceName != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.InterfaceName)
	}

	if nn.Extends != nil {
		io.WriteString(p.w, " extends ")
		p.joinPrint(", ", nn.Extends.InterfaceNames)
	}

	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "{\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printStmtLabel(n ir.Node) {
	nn := n.(*ir.LabelStmt)

	p.Print(nn.LabelName)
	io.WriteString(p.w, ":")
}

func (p *PrettyPrinter) printStmtNamespace(n ir.Node) {
	nn := n.(*ir.NamespaceStmt)

	io.WriteString(p.w, "namespace")

	if nn.NamespaceName != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.NamespaceName)
	}

	if nn.Stmts != nil {
		io.WriteString(p.w, " {\n")
		p.printNodes(nn.Stmts)
		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "}")
	} else {
		io.WriteString(p.w, ";")
	}
}

func (p *PrettyPrinter) printStmtNop(n ir.Node) {
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtPropertyList(n ir.Node) {
	nn := n.(*ir.PropertyListStmt)

	p.joinPrintIdents(" ", nn.Modifiers)
	io.WriteString(p.w, " ")
	p.joinPrint(", ", nn.Properties)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtProperty(n ir.Node) {
	nn := n.(*ir.PropertyStmt)

	p.Print(nn.Variable)

	if nn.Expr != nil {
		io.WriteString(p.w, " = ")
		p.Print(nn.Expr)
	}
}

func (p *PrettyPrinter) printStmtReturn(n ir.Node) {
	nn := n.(*ir.ReturnStmt)

	io.WriteString(p.w, "return ")
	p.Print(nn.Expr)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtStaticVar(n ir.Node) {
	nn := n.(*ir.StaticVarStmt)
	p.Print(nn.Variable)

	if nn.Expr != nil {
		io.WriteString(p.w, " = ")
		p.Print(nn.Expr)
	}
}

func (p *PrettyPrinter) printStmtStatic(n ir.Node) {
	nn := n.(*ir.StaticStmt)

	io.WriteString(p.w, "static ")
	p.joinPrint(", ", nn.Vars)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtStmtList(n ir.Node) {
	nn := n.(*ir.StmtList)

	io.WriteString(p.w, "{\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printStmtSwitch(n ir.Node) {
	nn := n.(*ir.SwitchStmt)

	io.WriteString(p.w, "switch (")
	p.Print(nn.Cond)

	if nn.AltSyntax {
		io.WriteString(p.w, ") :\n")
		s := nn.CaseList.Cases
		p.printNodes(s)

		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "endswitch;")
	} else {
		io.WriteString(p.w, ")")

		io.WriteString(p.w, " {\n")
		p.printNodes(nn.CaseList.Cases)
		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "}")
	}
}

func (p *PrettyPrinter) printStmtThrow(n ir.Node) {
	nn := n.(*ir.ThrowStmt)

	io.WriteString(p.w, "throw ")
	p.Print(nn.Expr)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtTraitMethodRef(n ir.Node) {
	nn := n.(*ir.TraitMethodRefStmt)

	p.Print(nn.Trait)
	io.WriteString(p.w, "::")
	p.Print(nn.Method)
}

func (p *PrettyPrinter) printStmtTraitUseAlias(n ir.Node) {
	nn := n.(*ir.TraitUseAliasStmt)

	p.Print(nn.Ref)
	io.WriteString(p.w, " as")

	if nn.Modifier != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.Modifier)
	}

	if nn.Alias != nil {
		io.WriteString(p.w, " ")
		p.Print(nn.Alias)
	}

	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtTraitUsePrecedence(n ir.Node) {
	nn := n.(*ir.TraitUsePrecedenceStmt)

	p.Print(nn.Ref)
	io.WriteString(p.w, " insteadof ")
	p.joinPrint(", ", nn.Insteadof)

	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtTraitUse(n ir.Node) {
	nn := n.(*ir.TraitUseStmt)

	io.WriteString(p.w, "use ")
	p.joinPrint(", ", nn.Traits)

	if adaptationList, ok := nn.TraitAdaptationList.(*ir.TraitAdaptationListStmt); ok {
		adaptations := adaptationList.Adaptations
		io.WriteString(p.w, " {\n")
		p.printNodes(adaptations)
		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "}")
	} else {
		io.WriteString(p.w, ";")
	}
}

func (p *PrettyPrinter) printStmtTrait(n ir.Node) {
	nn := n.(*ir.TraitStmt)

	io.WriteString(p.w, "trait ")
	p.Print(nn.TraitName)

	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "{\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")
}

func (p *PrettyPrinter) printStmtTry(n ir.Node) {
	nn := n.(*ir.TryStmt)

	io.WriteString(p.w, "try {\n")
	p.printNodes(nn.Stmts)
	io.WriteString(p.w, "\n")
	p.printIndent()
	io.WriteString(p.w, "}")

	if nn.Catches != nil {
		io.WriteString(p.w, "\n")
		p.indentDepth--
		p.printNodes(nn.Catches)
		p.indentDepth++
	}

	if nn.Finally != nil {
		io.WriteString(p.w, "\n")
		p.printIndent()
		p.Print(nn.Finally)
	}
}

func (p *PrettyPrinter) printStmtUnset(n ir.Node) {
	nn := n.(*ir.UnsetStmt)

	io.WriteString(p.w, "unset(")
	p.joinPrint(", ", nn.Vars)
	io.WriteString(p.w, ");")
}

func (p *PrettyPrinter) printStmtUseList(n ir.Node) {
	nn := n.(*ir.UseListStmt)

	io.WriteString(p.w, "use ")

	if nn.UseType != nil {
		p.Print(nn.UseType)
		io.WriteString(p.w, " ")
	}

	p.joinPrint(", ", nn.Uses)
	io.WriteString(p.w, ";")
}

func (p *PrettyPrinter) printStmtUse(n ir.Node) {
	nn := n.(*ir.UseStmt)

	if nn.UseType != nil {
		p.Print(nn.UseType)
		io.WriteString(p.w, " ")
	}

	p.Print(nn.Use)

	if nn.Alias != nil {
		io.WriteString(p.w, " as ")
		p.Print(nn.Alias)
	}
}

func (p *PrettyPrinter) printStmtWhile(n ir.Node) {
	nn := n.(*ir.WhileStmt)

	io.WriteString(p.w, "while (")
	p.Print(nn.Cond)

	if nn.AltSyntax {
		io.WriteString(p.w, ") :\n")

		s := nn.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)

		io.WriteString(p.w, "\n")
		p.printIndent()
		io.WriteString(p.w, "endwhile;")
	} else {
		io.WriteString(p.w, ")")

		switch s := nn.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			io.WriteString(p.w, " ")
			p.Print(s)
		default:
			io.WriteString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}
