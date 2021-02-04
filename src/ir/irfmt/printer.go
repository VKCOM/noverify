package irfmt

import (
	"fmt"
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
			writeString(p.w, glue)
		}

		p.Print(n)
	}
}

func (p *PrettyPrinter) joinPrintArrayItems(glue string, items []*ir.ArrayItemExpr) {
	for k, n := range items {
		if k > 0 {
			writeString(p.w, glue)
		}

		p.Print(n)
	}
}

func (p *PrettyPrinter) joinPrint(glue string, nn []ir.Node) {
	for k, n := range nn {
		if k > 0 {
			writeString(p.w, glue)
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
			writeString(p.w, "\n")
		}
	}
	p.indentDepth--
}

func (p *PrettyPrinter) printIndent() {
	for i := 0; i < p.indentDepth; i++ {
		writeString(p.w, p.indentStr)
	}
}

func (p *PrettyPrinter) printNode(n ir.Node) {
	switch n := n.(type) {

	case *ir.AnonClassExpr:
		p.printExprAnonClass(n)

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

	case *ir.Name:
		p.printNameName(n)

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

	case *ir.TypeCastExpr:
		p.printTypeCastExpr(n)
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
	case *ir.ImportExpr:
		p.printExprImport(n)
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

func (p *PrettyPrinter) printClass(class ir.Class) {
	if class.Extends != nil {
		writeString(p.w, " extends ")
		p.Print(class.Extends.ClassName)
	}

	if class.Implements != nil {
		writeString(p.w, " implements ")
		p.joinPrint(", ", class.Implements.InterfaceNames)
	}

	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "{\n")
	p.printNodes(class.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printExprAnonClass(n *ir.AnonClassExpr) {
	writeString(p.w, "class")

	if len(n.Args) != 0 {
		writeString(p.w, "(")
		p.joinPrint(", ", n.Args)
		writeString(p.w, ")")
	}

	p.printClass(n.Class)
}

// node

func (p *PrettyPrinter) printNodeRoot(n *ir.Root) {
	stmts := n.Stmts
	if len(stmts) > 0 {
		firstStmt := stmts[0]
		stmts = stmts[1:]

		switch fs := firstStmt.(type) {
		case *ir.InlineHTMLStmt:
			writeString(p.w, fs.Value)
			writeString(p.w, "<?php\n")
		default:
			writeString(p.w, "<?php\n")
			p.printIndent()
			p.Print(fs)
			writeString(p.w, "\n")
		}
	}
	p.indentDepth--
	p.printNodes(stmts)
	writeString(p.w, "\n")
}

func (p *PrettyPrinter) printNodeIdentifier(n *ir.Identifier) {
	writeString(p.w, n.Value)
}

func (p *PrettyPrinter) printNodeParameter(n *ir.Parameter) {
	if n.VariableType != nil {
		p.Print(n.VariableType)
		writeString(p.w, " ")
	}

	if n.ByRef {
		writeString(p.w, "&")
	}

	if n.Variadic {
		writeString(p.w, "...")
	}

	p.Print(n.Variable)

	if n.DefaultValue != nil {
		writeString(p.w, " = ")
		p.Print(n.DefaultValue)
	}
}

func (p *PrettyPrinter) printNodeNullable(n *ir.Nullable) {
	writeString(p.w, "?")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printNodeArgument(n *ir.Argument) {
	if n.IsReference {
		writeString(p.w, "&")
	}

	if n.Variadic {
		writeString(p.w, "...")
	}

	p.Print(n.Expr)
}

// name

func (p *PrettyPrinter) printNameName(n *ir.Name) {
	writeString(p.w, n.Value)
}

// scalar

func (p *PrettyPrinter) printScalarLNumber(n *ir.Lnumber) {
	writeString(p.w, n.Value)
}

func (p *PrettyPrinter) printScalarDNumber(n *ir.Dnumber) {
	writeString(p.w, n.Value)
}

func (p *PrettyPrinter) printScalarString(n *ir.String) {
	var s string
	if n.DoubleQuotes {
		s = `"` + n.Value + `"`
	} else {
		s = "'" + n.Value + "'"
	}
	writeString(p.w, s)
}

func (p *PrettyPrinter) printScalarEncapsedStringPart(n *ir.EncapsedStringPart) {
	writeString(p.w, n.Value)
}

func (p *PrettyPrinter) printScalarEncapsed(n *ir.Encapsed) {
	writeString(p.w, "\"")

	for _, part := range n.Parts {
		switch part.(type) {
		case *ir.EncapsedStringPart:
			p.Print(part)
		default:
			writeString(p.w, "{")
			p.Print(part)
			writeString(p.w, "}")
		}
	}

	writeString(p.w, "\"")
}

func (p *PrettyPrinter) printScalarHeredoc(n *ir.Heredoc) {
	writeString(p.w, n.Label)

	for _, part := range n.Parts {
		switch part.(type) {
		case *ir.EncapsedStringPart:
			p.Print(part)
		default:
			writeString(p.w, "{")
			p.Print(part)
			writeString(p.w, "}")
		}
	}

	writeString(p.w, strings.Trim(n.Label, "<\"'\n"))
}

func (p *PrettyPrinter) printScalarMagicConstant(n *ir.MagicConstant) {
	writeString(p.w, n.Value)
}

// Assign

func (p *PrettyPrinter) printAssign(n *ir.Assign) {
	p.Print(n.Variable)
	writeString(p.w, " = ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printReference(n *ir.AssignReference) {
	p.Print(n.Variable)
	writeString(p.w, " =& ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignBitwiseAnd(n *ir.AssignBitwiseAnd) {
	p.Print(n.Variable)
	writeString(p.w, " &= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignBitwiseOr(n *ir.AssignBitwiseOr) {
	p.Print(n.Variable)
	writeString(p.w, " |= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignBitwiseXor(n *ir.AssignBitwiseXor) {
	p.Print(n.Variable)
	writeString(p.w, " ^= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignConcat(n *ir.AssignConcat) {
	p.Print(n.Variable)
	writeString(p.w, " .= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignDiv(n *ir.AssignDiv) {
	p.Print(n.Variable)
	writeString(p.w, " /= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignMinus(n *ir.AssignMinus) {
	p.Print(n.Variable)
	writeString(p.w, " -= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignMod(n *ir.AssignMod) {
	p.Print(n.Variable)
	writeString(p.w, " %= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignMul(n *ir.AssignMul) {
	p.Print(n.Variable)
	writeString(p.w, " *= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignPlus(n *ir.AssignPlus) {
	p.Print(n.Variable)
	writeString(p.w, " += ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignPow(n *ir.AssignPow) {
	p.Print(n.Variable)
	writeString(p.w, " **= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignShiftLeft(n *ir.AssignShiftLeft) {
	p.Print(n.Variable)
	writeString(p.w, " <<= ")
	p.Print(n.Expression)
}

func (p *PrettyPrinter) printAssignShiftRight(n *ir.AssignShiftRight) {
	p.Print(n.Variable)
	writeString(p.w, " >>= ")
	p.Print(n.Expression)
}

// binary

func (p *PrettyPrinter) printBinaryBitwiseAnd(n *ir.BitwiseAndExpr) {
	p.Print(n.Left)
	writeString(p.w, " & ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryBitwiseOr(n *ir.BitwiseOrExpr) {
	p.Print(n.Left)
	writeString(p.w, " | ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryBitwiseXor(n *ir.BitwiseXorExpr) {
	p.Print(n.Left)
	writeString(p.w, " ^ ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryBooleanAnd(n *ir.BooleanAndExpr) {
	p.Print(n.Left)
	writeString(p.w, " && ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryBooleanOr(n *ir.BooleanOrExpr) {
	p.Print(n.Left)
	writeString(p.w, " || ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryCoalesce(n *ir.CoalesceExpr) {
	p.Print(n.Left)
	writeString(p.w, " ?? ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryConcat(n *ir.ConcatExpr) {
	p.Print(n.Left)
	writeString(p.w, " . ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryDiv(n *ir.DivExpr) {
	p.Print(n.Left)
	writeString(p.w, " / ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryEqual(n *ir.EqualExpr) {
	p.Print(n.Left)
	writeString(p.w, " == ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryGreaterOrEqual(n *ir.GreaterOrEqualExpr) {
	p.Print(n.Left)
	writeString(p.w, " >= ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryGreater(n *ir.GreaterExpr) {
	p.Print(n.Left)
	writeString(p.w, " > ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryIdentical(n *ir.IdenticalExpr) {
	p.Print(n.Left)
	writeString(p.w, " === ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryLogicalAnd(n *ir.LogicalAndExpr) {
	p.Print(n.Left)
	writeString(p.w, " and ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryLogicalOr(n *ir.LogicalOrExpr) {
	p.Print(n.Left)
	writeString(p.w, " or ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryLogicalXor(n *ir.LogicalXorExpr) {
	p.Print(n.Left)
	writeString(p.w, " xor ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryMinus(n *ir.MinusExpr) {
	p.Print(n.Left)
	writeString(p.w, " - ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryMod(n *ir.ModExpr) {
	p.Print(n.Left)
	writeString(p.w, " % ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryMul(n *ir.MulExpr) {
	p.Print(n.Left)
	writeString(p.w, " * ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryNotEqual(n *ir.NotEqualExpr) {
	p.Print(n.Left)
	writeString(p.w, " != ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryNotIdentical(n *ir.NotIdenticalExpr) {
	p.Print(n.Left)
	writeString(p.w, " !== ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryPlus(n *ir.PlusExpr) {
	p.Print(n.Left)
	writeString(p.w, " + ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryPow(n *ir.PowExpr) {
	p.Print(n.Left)
	writeString(p.w, " ** ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryShiftLeft(n *ir.ShiftLeftExpr) {
	p.Print(n.Left)
	writeString(p.w, " << ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinaryShiftRight(n *ir.ShiftRightExpr) {
	p.Print(n.Left)
	writeString(p.w, " >> ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinarySmallerOrEqual(n *ir.SmallerOrEqualExpr) {
	p.Print(n.Left)
	writeString(p.w, " <= ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinarySmaller(n *ir.SmallerExpr) {
	p.Print(n.Left)
	writeString(p.w, " < ")
	p.Print(n.Right)
}

func (p *PrettyPrinter) printBinarySpaceship(n *ir.SpaceshipExpr) {
	p.Print(n.Left)
	writeString(p.w, " <=> ")
	p.Print(n.Right)
}

// cast

func (p *PrettyPrinter) printTypeCastExpr(n *ir.TypeCastExpr) {
	fmt.Fprintf(p.w, "(%s)", n.Type)
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printUnset(n *ir.UnsetCastExpr) {
	writeString(p.w, "(unset)")
	p.Print(n.Expr)
}

// expr

func (p *PrettyPrinter) printExprParen(n *ir.ParenExpr) {
	writeString(p.w, "(")
	p.Print(n.Expr)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprArrayDimFetch(n *ir.ArrayDimFetchExpr) {
	p.Print(n.Variable)
	openBrace, closeBrace := "[", "]"
	if n.CurlyBrace {
		openBrace, closeBrace = "{", "}"
	}
	writeString(p.w, openBrace)
	p.Print(n.Dim)
	writeString(p.w, closeBrace)
}

func (p *PrettyPrinter) printExprArrayItem(n *ir.ArrayItemExpr) {
	if n.Unpack {
		writeString(p.w, "...")
	}

	if n.Key != nil {
		p.Print(n.Key)
		writeString(p.w, " => ")
	}

	p.Print(n.Val)
}

func (p *PrettyPrinter) printExprArray(n *ir.ArrayExpr) {
	if n.ShortSyntax {
		writeString(p.w, "[")
		p.joinPrintArrayItems(", ", n.Items)
		writeString(p.w, "]")
	} else {
		writeString(p.w, "array(")
		p.joinPrintArrayItems(", ", n.Items)
		writeString(p.w, ")")
	}
}

func (p *PrettyPrinter) printExprBitwiseNot(n *ir.BitwiseNotExpr) {
	writeString(p.w, "~")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprBooleanNot(n *ir.BooleanNotExpr) {
	writeString(p.w, "!")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprClassConstFetch(n *ir.ClassConstFetchExpr) {
	p.Print(n.Class)
	writeString(p.w, "::")
	writeString(p.w, n.ConstantName.Value)
}

func (p *PrettyPrinter) printExprClone(n *ir.CloneExpr) {
	writeString(p.w, "clone ")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprClosureUse(n *ir.ClosureUseExpr) {
	writeString(p.w, "use (")
	p.joinPrint(", ", n.Uses)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprClosure(n *ir.ClosureExpr) {
	if n.Static {
		writeString(p.w, "static ")
	}

	writeString(p.w, "function ")

	if n.ReturnsRef {
		writeString(p.w, "&")
	}

	writeString(p.w, "(")
	p.joinPrint(", ", n.Params)
	writeString(p.w, ")")

	if n.ClosureUse != nil {
		writeString(p.w, " ")
		p.Print(n.ClosureUse)
	}

	if n.ReturnType != nil {
		writeString(p.w, ": ")
		p.Print(n.ReturnType)
	}

	writeString(p.w, " {\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printExprConstFetch(n *ir.ConstFetchExpr) {
	p.Print(n.Constant)
}

func (p *PrettyPrinter) printExprEmpty(n *ir.EmptyExpr) {
	writeString(p.w, "empty(")
	p.Print(n.Expr)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprErrorSuppress(n *ir.ErrorSuppressExpr) {
	writeString(p.w, "@")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprEval(n *ir.EvalExpr) {
	writeString(p.w, "eval(")
	p.Print(n.Expr)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprExit(n *ir.ExitExpr) {
	if n.Die {
		writeString(p.w, "die(")
	} else {
		writeString(p.w, "exit(")
	}
	p.Print(n.Expr)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprFunctionCall(n *ir.FunctionCallExpr) {
	p.Print(n.Function)
	writeString(p.w, "(")
	p.joinPrint(", ", n.Args)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprInstanceOf(n *ir.InstanceOfExpr) {
	p.Print(n.Expr)
	writeString(p.w, " instanceof ")
	p.Print(n.Class)
}

func (p *PrettyPrinter) printExprIsset(n *ir.IssetExpr) {
	writeString(p.w, "isset(")
	p.joinPrint(", ", n.Variables)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprList(n *ir.ListExpr) {
	if n.ShortSyntax {
		writeString(p.w, "[")
		p.joinPrintArrayItems(", ", n.Items)
		writeString(p.w, "]")
	} else {
		writeString(p.w, "list(")
		p.joinPrintArrayItems(", ", n.Items)
		writeString(p.w, ")")
	}
}

func (p *PrettyPrinter) printExprMethodCall(n *ir.MethodCallExpr) {
	p.Print(n.Variable)
	writeString(p.w, "->")
	p.Print(n.Method)
	writeString(p.w, "(")
	p.joinPrint(", ", n.Args)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprNew(n *ir.NewExpr) {
	writeString(p.w, "new ")
	p.Print(n.Class)

	if n.Args != nil {
		writeString(p.w, "(")
		p.joinPrint(", ", n.Args)
		writeString(p.w, ")")
	}
}

func (p *PrettyPrinter) printExprPostDec(n *ir.PostDecExpr) {
	p.Print(n.Variable)
	writeString(p.w, "--")
}

func (p *PrettyPrinter) printExprPostInc(n *ir.PostIncExpr) {
	p.Print(n.Variable)
	writeString(p.w, "++")
}

func (p *PrettyPrinter) printExprPreDec(n *ir.PreDecExpr) {
	writeString(p.w, "--")
	p.Print(n.Variable)
}

func (p *PrettyPrinter) printExprPreInc(n *ir.PreIncExpr) {
	writeString(p.w, "++")
	p.Print(n.Variable)
}

func (p *PrettyPrinter) printExprPrint(n *ir.PrintExpr) {
	writeString(p.w, "print(")
	p.Print(n.Expr)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprPropertyFetch(n *ir.PropertyFetchExpr) {
	p.Print(n.Variable)
	writeString(p.w, "->")
	p.Print(n.Property)
}

func (p *PrettyPrinter) printExprReference(n *ir.ReferenceExpr) {
	writeString(p.w, "&")
	p.Print(n.Variable)
}

func (p *PrettyPrinter) printExprImport(n *ir.ImportExpr) {
	writeString(p.w, n.Func+" ")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprShellExec(n *ir.ShellExecExpr) {
	writeString(p.w, "`")
	for _, part := range n.Parts {
		switch part.(type) {
		case *ir.EncapsedStringPart:
			p.Print(part)
		default:
			writeString(p.w, "{")
			p.Print(part)
			writeString(p.w, "}")
		}
	}
	writeString(p.w, "`")
}

func (p *PrettyPrinter) printExprStaticCall(n *ir.StaticCallExpr) {
	p.Print(n.Class)
	writeString(p.w, "::")
	p.Print(n.Call)
	writeString(p.w, "(")
	p.joinPrint(", ", n.Args)
	writeString(p.w, ")")
}

func (p *PrettyPrinter) printExprStaticPropertyFetch(n *ir.StaticPropertyFetchExpr) {
	p.Print(n.Class)
	writeString(p.w, "::")
	p.Print(n.Property)
}

func (p *PrettyPrinter) printExprTernary(n *ir.TernaryExpr) {
	p.Print(n.Condition)
	writeString(p.w, " ?")

	if n.IfTrue != nil {
		writeString(p.w, " ")
		p.Print(n.IfTrue)
		writeString(p.w, " ")
	}

	writeString(p.w, ": ")
	p.Print(n.IfFalse)
}

func (p *PrettyPrinter) printExprUnaryMinus(n *ir.UnaryMinusExpr) {
	writeString(p.w, "-")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprUnaryPlus(n *ir.UnaryPlusExpr) {
	writeString(p.w, "+")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprSimpleVar(n *ir.SimpleVar) {
	writeString(p.w, "$"+n.Name)
}

func (p *PrettyPrinter) printExprVar(n *ir.Var) {
	writeString(p.w, "$")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprYieldFrom(n *ir.YieldFromExpr) {
	writeString(p.w, "yield from ")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printExprYield(n *ir.YieldExpr) {
	writeString(p.w, "yield ")

	if n.Key != nil {
		p.Print(n.Key)
		writeString(p.w, " => ")
	}

	p.Print(n.Value)
}

// smtm

func (p *PrettyPrinter) printStmtBreak(n *ir.BreakStmt) {
	writeString(p.w, "break")
	if n.Expr != nil {
		writeString(p.w, " ")
		p.Print(n.Expr)
	}

	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtCase(n *ir.CaseStmt) {
	writeString(p.w, "case ")
	p.Print(n.Cond)
	writeString(p.w, ":")

	if len(n.Stmts) > 0 {
		writeString(p.w, "\n")
		p.printNodes(n.Stmts)
	}
}

func (p *PrettyPrinter) printStmtCatch(n *ir.CatchStmt) {
	writeString(p.w, "catch (")
	p.joinPrint(" | ", n.Types)
	writeString(p.w, " ")
	p.Print(n.Variable)
	writeString(p.w, ") {\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printStmtClassMethod(n *ir.ClassMethodStmt) {
	if n.Modifiers != nil {
		p.joinPrintIdents(" ", n.Modifiers)
		writeString(p.w, " ")
	}
	writeString(p.w, "function ")

	if n.ReturnsRef {
		writeString(p.w, "&")
	}

	p.Print(n.MethodName)
	writeString(p.w, "(")
	p.joinPrint(", ", n.Params)
	writeString(p.w, ")")

	if n.ReturnType != nil {
		writeString(p.w, ": ")
		p.Print(n.ReturnType)
	}

	switch s := n.Stmt.(type) {
	case *ir.StmtList:
		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "{\n")
		p.printNodes(s.Stmts)
		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "}")
	default:
		p.Print(s)
	}
}

func (p *PrettyPrinter) printStmtClass(n *ir.ClassStmt) {
	if n.Modifiers != nil {
		p.joinPrintIdents(" ", n.Modifiers)
		writeString(p.w, " ")
	}
	writeString(p.w, "class")

	writeString(p.w, " ")
	p.Print(n.ClassName)

	p.printClass(n.Class)
}

func (p *PrettyPrinter) printStmtClassConstList(n *ir.ClassConstListStmt) {
	if n.Modifiers != nil {
		p.joinPrintIdents(" ", n.Modifiers)
		writeString(p.w, " ")
	}
	writeString(p.w, "const ")

	p.joinPrint(", ", n.Consts)

	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtConstant(n *ir.ConstantStmt) {
	p.Print(n.ConstantName)
	writeString(p.w, " = ")
	p.Print(n.Expr)
}

func (p *PrettyPrinter) printStmtContinue(n *ir.ContinueStmt) {
	writeString(p.w, "continue")
	if n.Expr != nil {
		writeString(p.w, " ")
		p.Print(n.Expr)
	}

	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtDeclare(n *ir.DeclareStmt) {
	writeString(p.w, "declare(")
	p.joinPrint(", ", n.Consts)
	writeString(p.w, ")")

	switch s := n.Stmt.(type) {
	case *ir.NopStmt:
		p.Print(s)
	case *ir.StmtList:
		writeString(p.w, " ")
		p.Print(s)
	default:
		writeString(p.w, "\n")
		p.indentDepth++
		p.printIndent()
		p.Print(s)
		p.indentDepth--
	}
}

func (p *PrettyPrinter) printStmtDefault(n *ir.DefaultStmt) {
	writeString(p.w, "default:")

	if len(n.Stmts) > 0 {
		writeString(p.w, "\n")
		p.printNodes(n.Stmts)
	}
}

func (p *PrettyPrinter) printStmtDo(n *ir.DoStmt) {
	writeString(p.w, "do")

	switch s := n.Stmt.(type) {
	case *ir.StmtList:
		writeString(p.w, " ")
		p.Print(s)
		writeString(p.w, " ")
	default:
		writeString(p.w, "\n")
		p.indentDepth++
		p.printIndent()
		p.Print(s)
		p.indentDepth--
		writeString(p.w, "\n")
		p.printIndent()
	}

	writeString(p.w, "while (")
	p.Print(n.Cond)
	writeString(p.w, ");")
}

func (p *PrettyPrinter) printStmtEcho(n *ir.EchoStmt) {
	writeString(p.w, "echo ")
	p.joinPrint(", ", n.Exprs)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtElseif(n *ir.ElseIfStmt) {
	writeString(p.w, "elseif (")
	p.Print(n.Cond)

	if n.AltSyntax {
		writeString(p.w, ") :")

		if s := n.Stmt.(*ir.StmtList).Stmts; len(s) > 0 {
			writeString(p.w, "\n")
			p.printNodes(s)
		}
	} else {
		writeString(p.w, ")")

		switch s := n.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			writeString(p.w, " ")
			p.Print(s)
		default:
			writeString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtElse(n *ir.ElseStmt) {
	if n.AltSyntax {
		writeString(p.w, "else :")

		if s := n.Stmt.(*ir.StmtList).Stmts; len(s) > 0 {
			writeString(p.w, "\n")
			p.printNodes(s)
		}
	} else {
		writeString(p.w, "else")

		switch s := n.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			writeString(p.w, " ")
			p.Print(s)
		default:
			writeString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtExpression(n *ir.ExpressionStmt) {
	p.Print(n.Expr)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtFinally(n *ir.FinallyStmt) {
	writeString(p.w, "finally {\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printStmtFor(n *ir.ForStmt) {
	writeString(p.w, "for (")
	p.joinPrint(", ", n.Init)
	writeString(p.w, "; ")
	p.joinPrint(", ", n.Cond)
	writeString(p.w, "; ")
	p.joinPrint(", ", n.Loop)

	if n.AltSyntax {
		writeString(p.w, ") :\n")

		s := n.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)
		writeString(p.w, "\n")
		p.printIndent()

		writeString(p.w, "endfor;")
	} else {
		writeString(p.w, ")")

		switch s := n.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			writeString(p.w, " ")
			p.Print(s)
		default:
			writeString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtForeach(n *ir.ForeachStmt) {
	writeString(p.w, "foreach (")
	p.Print(n.Expr)
	writeString(p.w, " as ")

	if n.Key != nil {
		p.Print(n.Key)
		writeString(p.w, " => ")
	}

	p.Print(n.Variable)
	writeString(p.w, ")")

	if n.AltSyntax {
		writeString(p.w, " :\n")

		s := n.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)

		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "endforeach;")
	} else {
		switch s := n.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			writeString(p.w, " ")
			p.Print(s)
		default:
			writeString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func (p *PrettyPrinter) printStmtFunction(n *ir.FunctionStmt) {
	writeString(p.w, "function ")

	if n.ReturnsRef {
		writeString(p.w, "&")
	}

	p.Print(n.FunctionName)

	writeString(p.w, "(")
	p.joinPrint(", ", n.Params)
	writeString(p.w, ")")

	if n.ReturnType != nil {
		writeString(p.w, ": ")
		p.Print(n.ReturnType)
	}

	writeString(p.w, " {\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printStmtGlobal(n *ir.GlobalStmt) {
	writeString(p.w, "global ")
	p.joinPrint(", ", n.Vars)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtGoto(n *ir.GotoStmt) {
	writeString(p.w, "goto ")
	p.Print(n.Label)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtGroupUse(n *ir.GroupUseStmt) {
	writeString(p.w, "use ")

	if n.UseType != nil {
		p.Print(n.UseType)
		writeString(p.w, " ")
	}

	p.Print(n.Prefix)
	writeString(p.w, "\\{")
	p.joinPrint(", ", n.UseList)
	writeString(p.w, "};")
}

func (p *PrettyPrinter) printStmtHaltCompiler(n ir.Node) {
	writeString(p.w, "__halt_compiler();")
}

func (p *PrettyPrinter) printStmtIf(n *ir.IfStmt) {
	writeString(p.w, "if (")
	p.Print(n.Cond)
	writeString(p.w, ")")

	if n.AltSyntax {
		writeString(p.w, " :\n")

		s := n.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)

		for _, elseif := range n.ElseIf {
			writeString(p.w, "\n")
			p.printIndent()
			p.Print(elseif)
		}

		if n.Else != nil {
			writeString(p.w, "\n")
			p.printIndent()
			p.Print(n.Else)
		}

		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "endif;")
	} else {
		switch s := n.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			writeString(p.w, " ")
			p.Print(s)
		default:
			writeString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}

		if n.ElseIf != nil {
			writeString(p.w, "\n")
			p.indentDepth--
			p.printNodes(n.ElseIf)
			p.indentDepth++
		}

		if n.Else != nil {
			writeString(p.w, "\n")
			p.printIndent()
			p.Print(n.Else)
		}
	}
}

func (p *PrettyPrinter) printStmtInlineHTML(n *ir.InlineHTMLStmt) {
	writeString(p.w, "?>")
	writeString(p.w, n.Value)
	writeString(p.w, "<?php")
}

func (p *PrettyPrinter) printStmtInterface(n *ir.InterfaceStmt) {
	writeString(p.w, "interface")

	if n.InterfaceName != nil {
		writeString(p.w, " ")
		p.Print(n.InterfaceName)
	}

	if n.Extends != nil {
		writeString(p.w, " extends ")
		p.joinPrint(", ", n.Extends.InterfaceNames)
	}

	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "{\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printStmtLabel(n *ir.LabelStmt) {
	p.Print(n.LabelName)
	writeString(p.w, ":")
}

func (p *PrettyPrinter) printStmtNamespace(n *ir.NamespaceStmt) {
	writeString(p.w, "namespace")

	if n.NamespaceName != nil {
		writeString(p.w, " ")
		p.Print(n.NamespaceName)
	}

	if n.Stmts != nil {
		writeString(p.w, " {\n")
		p.printNodes(n.Stmts)
		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "}")
	} else {
		writeString(p.w, ";")
	}
}

func (p *PrettyPrinter) printStmtNop(n *ir.NopStmt) {
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtPropertyList(n *ir.PropertyListStmt) {
	p.joinPrintIdents(" ", n.Modifiers)
	writeString(p.w, " ")
	p.joinPrint(", ", n.Properties)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtProperty(n *ir.PropertyStmt) {
	p.Print(n.Variable)

	if n.Expr != nil {
		writeString(p.w, " = ")
		p.Print(n.Expr)
	}
}

func (p *PrettyPrinter) printStmtReturn(n *ir.ReturnStmt) {
	writeString(p.w, "return ")
	p.Print(n.Expr)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtStaticVar(n *ir.StaticVarStmt) {
	p.Print(n.Variable)

	if n.Expr != nil {
		writeString(p.w, " = ")
		p.Print(n.Expr)
	}
}

func (p *PrettyPrinter) printStmtStatic(n *ir.StaticStmt) {
	writeString(p.w, "static ")
	p.joinPrint(", ", n.Vars)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtStmtList(n *ir.StmtList) {
	writeString(p.w, "{\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printStmtSwitch(n *ir.SwitchStmt) {
	writeString(p.w, "switch (")
	p.Print(n.Cond)

	if n.AltSyntax {
		writeString(p.w, ") :\n")
		p.printNodes(n.Cases)

		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "endswitch;")
	} else {
		writeString(p.w, ")")

		writeString(p.w, " {\n")
		p.printNodes(n.Cases)
		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "}")
	}
}

func (p *PrettyPrinter) printStmtThrow(n *ir.ThrowStmt) {
	writeString(p.w, "throw ")
	p.Print(n.Expr)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtTraitMethodRef(n *ir.TraitMethodRefStmt) {
	p.Print(n.Trait)
	writeString(p.w, "::")
	p.Print(n.Method)
}

func (p *PrettyPrinter) printStmtTraitUseAlias(n *ir.TraitUseAliasStmt) {
	p.Print(n.Ref)
	writeString(p.w, " as")

	if n.Modifier != nil {
		writeString(p.w, " ")
		p.Print(n.Modifier)
	}

	if n.Alias != nil {
		writeString(p.w, " ")
		p.Print(n.Alias)
	}

	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtTraitUsePrecedence(n *ir.TraitUsePrecedenceStmt) {
	p.Print(n.Ref)
	writeString(p.w, " insteadof ")
	p.joinPrint(", ", n.Insteadof)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtTraitUse(n *ir.TraitUseStmt) {
	writeString(p.w, "use ")
	p.joinPrint(", ", n.Traits)

	if adaptationList, ok := n.TraitAdaptationList.(*ir.TraitAdaptationListStmt); ok {
		adaptations := adaptationList.Adaptations
		writeString(p.w, " {\n")
		p.printNodes(adaptations)
		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "}")
	} else {
		writeString(p.w, ";")
	}
}

func (p *PrettyPrinter) printStmtTrait(n *ir.TraitStmt) {
	writeString(p.w, "trait ")
	p.Print(n.TraitName)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "{\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")
}

func (p *PrettyPrinter) printStmtTry(n *ir.TryStmt) {
	writeString(p.w, "try {\n")
	p.printNodes(n.Stmts)
	writeString(p.w, "\n")
	p.printIndent()
	writeString(p.w, "}")

	if n.Catches != nil {
		writeString(p.w, "\n")
		p.indentDepth--
		p.printNodes(n.Catches)
		p.indentDepth++
	}

	if n.Finally != nil {
		writeString(p.w, "\n")
		p.printIndent()
		p.Print(n.Finally)
	}
}

func (p *PrettyPrinter) printStmtUnset(n *ir.UnsetStmt) {
	writeString(p.w, "unset(")
	p.joinPrint(", ", n.Vars)
	writeString(p.w, ");")
}

func (p *PrettyPrinter) printStmtUseList(n *ir.UseListStmt) {
	writeString(p.w, "use ")

	if n.UseType != nil {
		p.Print(n.UseType)
		writeString(p.w, " ")
	}

	p.joinPrint(", ", n.Uses)
	writeString(p.w, ";")
}

func (p *PrettyPrinter) printStmtUse(n *ir.UseStmt) {
	if n.UseType != nil {
		p.Print(n.UseType)
		writeString(p.w, " ")
	}

	p.Print(n.Use)

	if n.Alias != nil {
		writeString(p.w, " as ")
		p.Print(n.Alias)
	}
}

func (p *PrettyPrinter) printStmtWhile(n *ir.WhileStmt) {
	writeString(p.w, "while (")
	p.Print(n.Cond)

	if n.AltSyntax {
		writeString(p.w, ") :\n")

		s := n.Stmt.(*ir.StmtList)
		p.printNodes(s.Stmts)

		writeString(p.w, "\n")
		p.printIndent()
		writeString(p.w, "endwhile;")
	} else {
		writeString(p.w, ")")

		switch s := n.Stmt.(type) {
		case *ir.NopStmt:
			p.Print(s)
		case *ir.StmtList:
			writeString(p.w, " ")
			p.Print(s)
		default:
			writeString(p.w, "\n")
			p.indentDepth++
			p.printIndent()
			p.Print(s)
			p.indentDepth--
		}
	}
}

func writeString(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}
