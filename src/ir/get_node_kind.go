package ir

import (
	"fmt"
)

type NodeKind int

const (
	KindAssign NodeKind = iota
	KindAssignBitwiseAnd
	KindAssignBitwiseOr
	KindAssignBitwiseXor
	KindAssignCoalesce
	KindAssignConcat
	KindAssignDiv
	KindAssignMinus
	KindAssignMod
	KindAssignMul
	KindAssignPlus
	KindAssignPow
	KindAssignReference
	KindAssignShiftLeft
	KindAssignShiftRight
	KindBitwiseAndExpr
	KindBitwiseOrExpr
	KindBitwiseXorExpr
	KindBooleanAndExpr
	KindBooleanOrExpr
	KindCoalesceExpr
	KindConcatExpr
	KindDivExpr
	KindEqualExpr
	KindGreaterExpr
	KindGreaterOrEqualExpr
	KindIdenticalExpr
	KindLogicalAndExpr
	KindLogicalOrExpr
	KindLogicalXorExpr
	KindMinusExpr
	KindModExpr
	KindMulExpr
	KindNotEqualExpr
	KindNotIdenticalExpr
	KindPlusExpr
	KindPowExpr
	KindShiftLeftExpr
	KindShiftRightExpr
	KindSmallerExpr
	KindSmallerOrEqualExpr
	KindSpaceshipExpr
	KindArrayCastExpr
	KindBoolCastExpr
	KindDoubleCastExpr
	KindIntCastExpr
	KindObjectCastExpr
	KindStringCastExpr
	KindUnsetCastExpr
	KindArrayExpr
	KindArrayDimFetchExpr
	KindArrayItemExpr
	KindArrowFunctionExpr
	KindBitwiseNotExpr
	KindBooleanNotExpr
	KindClassConstFetchExpr
	KindCloneExpr
	KindClosureExpr
	KindClosureUseExpr
	KindConstFetchExpr
	KindEmptyExpr
	KindErrorSuppressExpr
	KindEvalExpr
	KindExitExpr
	KindFunctionCallExpr
	KindIncludeExpr
	KindIncludeOnceExpr
	KindInstanceOfExpr
	KindIssetExpr
	KindListExpr
	KindMethodCallExpr
	KindNewExpr
	KindParenExpr
	KindPostDecExpr
	KindPostIncExpr
	KindPreDecExpr
	KindPreIncExpr
	KindPrintExpr
	KindPropertyFetchExpr
	KindReferenceExpr
	KindRequireExpr
	KindRequireOnceExpr
	KindShellExecExpr
	KindStaticCallExpr
	KindStaticPropertyFetchExpr
	KindTernaryExpr
	KindUnaryMinusExpr
	KindUnaryPlusExpr
	KindYieldExpr
	KindYieldFromExpr
	KindFullyQualifiedName
	KindName
	KindNamePart
	KindRelativeName
	KindArgument
	KindArgumentList
	KindIdentifier
	KindNullable
	KindParameter
	KindRoot
	KindSimpleVar
	KindVar
	KindDnumber
	KindEncapsed
	KindEncapsedStringPart
	KindHeredoc
	KindLnumber
	KindMagicConstant
	KindString
	KindBreakStmt
	KindCaseStmt
	KindCaseListStmt
	KindCatchStmt
	KindClassStmt
	KindClassConstListStmt
	KindClassExtendsStmt
	KindClassImplementsStmt
	KindClassMethodStmt
	KindConstListStmt
	KindConstantStmt
	KindContinueStmt
	KindDeclareStmt
	KindDefaultStmt
	KindDoStmt
	KindEchoStmt
	KindElseStmt
	KindElseIfStmt
	KindExpressionStmt
	KindFinallyStmt
	KindForStmt
	KindForeachStmt
	KindFunctionStmt
	KindGlobalStmt
	KindGotoStmt
	KindGroupUseStmt
	KindHaltCompilerStmt
	KindIfStmt
	KindInlineHTMLStmt
	KindInterfaceStmt
	KindInterfaceExtendsStmt
	KindLabelStmt
	KindNamespaceStmt
	KindNopStmt
	KindPropertyStmt
	KindPropertyListStmt
	KindReturnStmt
	KindStaticStmt
	KindStaticVarStmt
	KindStmtList
	KindSwitchStmt
	KindThrowStmt
	KindTraitStmt
	KindTraitAdaptationListStmt
	KindTraitMethodRefStmt
	KindTraitUseStmt
	KindTraitUseAliasStmt
	KindTraitUsePrecedenceStmt
	KindTryStmt
	KindUnsetStmt
	KindUseStmt
	KindUseListStmt
	KindWhileStmt

	NumKinds
)

func GetNodeKind(n Node) NodeKind {
	switch n.(type) {
	case *Assign:
		return KindAssign
	case *AssignBitwiseAnd:
		return KindAssignBitwiseAnd
	case *AssignBitwiseOr:
		return KindAssignBitwiseOr
	case *AssignBitwiseXor:
		return KindAssignBitwiseXor
	case *AssignCoalesce:
		return KindAssignCoalesce
	case *AssignConcat:
		return KindAssignConcat
	case *AssignDiv:
		return KindAssignDiv
	case *AssignMinus:
		return KindAssignMinus
	case *AssignMod:
		return KindAssignMod
	case *AssignMul:
		return KindAssignMul
	case *AssignPlus:
		return KindAssignPlus
	case *AssignPow:
		return KindAssignPow
	case *AssignReference:
		return KindAssignReference
	case *AssignShiftLeft:
		return KindAssignShiftLeft
	case *AssignShiftRight:
		return KindAssignShiftRight
	case *BitwiseAndExpr:
		return KindBitwiseAndExpr
	case *BitwiseOrExpr:
		return KindBitwiseOrExpr
	case *BitwiseXorExpr:
		return KindBitwiseXorExpr
	case *BooleanAndExpr:
		return KindBooleanAndExpr
	case *BooleanOrExpr:
		return KindBooleanOrExpr
	case *CoalesceExpr:
		return KindCoalesceExpr
	case *ConcatExpr:
		return KindConcatExpr
	case *DivExpr:
		return KindDivExpr
	case *EqualExpr:
		return KindEqualExpr
	case *GreaterExpr:
		return KindGreaterExpr
	case *GreaterOrEqualExpr:
		return KindGreaterOrEqualExpr
	case *IdenticalExpr:
		return KindIdenticalExpr
	case *LogicalAndExpr:
		return KindLogicalAndExpr
	case *LogicalOrExpr:
		return KindLogicalOrExpr
	case *LogicalXorExpr:
		return KindLogicalXorExpr
	case *MinusExpr:
		return KindMinusExpr
	case *ModExpr:
		return KindModExpr
	case *MulExpr:
		return KindMulExpr
	case *NotEqualExpr:
		return KindNotEqualExpr
	case *NotIdenticalExpr:
		return KindNotIdenticalExpr
	case *PlusExpr:
		return KindPlusExpr
	case *PowExpr:
		return KindPowExpr
	case *ShiftLeftExpr:
		return KindShiftLeftExpr
	case *ShiftRightExpr:
		return KindShiftRightExpr
	case *SmallerExpr:
		return KindSmallerExpr
	case *SmallerOrEqualExpr:
		return KindSmallerOrEqualExpr
	case *SpaceshipExpr:
		return KindSpaceshipExpr
	case *ArrayCastExpr:
		return KindArrayCastExpr
	case *BoolCastExpr:
		return KindBoolCastExpr
	case *DoubleCastExpr:
		return KindDoubleCastExpr
	case *IntCastExpr:
		return KindIntCastExpr
	case *ObjectCastExpr:
		return KindObjectCastExpr
	case *StringCastExpr:
		return KindStringCastExpr
	case *UnsetCastExpr:
		return KindUnsetCastExpr
	case *ArrayExpr:
		return KindArrayExpr
	case *ArrayDimFetchExpr:
		return KindArrayDimFetchExpr
	case *ArrayItemExpr:
		return KindArrayItemExpr
	case *ArrowFunctionExpr:
		return KindArrowFunctionExpr
	case *BitwiseNotExpr:
		return KindBitwiseNotExpr
	case *BooleanNotExpr:
		return KindBooleanNotExpr
	case *ClassConstFetchExpr:
		return KindClassConstFetchExpr
	case *CloneExpr:
		return KindCloneExpr
	case *ClosureExpr:
		return KindClosureExpr
	case *ClosureUseExpr:
		return KindClosureUseExpr
	case *ConstFetchExpr:
		return KindConstFetchExpr
	case *EmptyExpr:
		return KindEmptyExpr
	case *ErrorSuppressExpr:
		return KindErrorSuppressExpr
	case *EvalExpr:
		return KindEvalExpr
	case *ExitExpr:
		return KindExitExpr
	case *FunctionCallExpr:
		return KindFunctionCallExpr
	case *IncludeExpr:
		return KindIncludeExpr
	case *IncludeOnceExpr:
		return KindIncludeOnceExpr
	case *InstanceOfExpr:
		return KindInstanceOfExpr
	case *IssetExpr:
		return KindIssetExpr
	case *ListExpr:
		return KindListExpr
	case *MethodCallExpr:
		return KindMethodCallExpr
	case *NewExpr:
		return KindNewExpr
	case *ParenExpr:
		return KindParenExpr
	case *PostDecExpr:
		return KindPostDecExpr
	case *PostIncExpr:
		return KindPostIncExpr
	case *PreDecExpr:
		return KindPreDecExpr
	case *PreIncExpr:
		return KindPreIncExpr
	case *PrintExpr:
		return KindPrintExpr
	case *PropertyFetchExpr:
		return KindPropertyFetchExpr
	case *ReferenceExpr:
		return KindReferenceExpr
	case *RequireExpr:
		return KindRequireExpr
	case *RequireOnceExpr:
		return KindRequireOnceExpr
	case *ShellExecExpr:
		return KindShellExecExpr
	case *StaticCallExpr:
		return KindStaticCallExpr
	case *StaticPropertyFetchExpr:
		return KindStaticPropertyFetchExpr
	case *TernaryExpr:
		return KindTernaryExpr
	case *UnaryMinusExpr:
		return KindUnaryMinusExpr
	case *UnaryPlusExpr:
		return KindUnaryPlusExpr
	case *YieldExpr:
		return KindYieldExpr
	case *YieldFromExpr:
		return KindYieldFromExpr
	case *FullyQualifiedName:
		return KindFullyQualifiedName
	case *Name:
		return KindName
	case *NamePart:
		return KindNamePart
	case *RelativeName:
		return KindRelativeName
	case *Argument:
		return KindArgument
	case *ArgumentList:
		return KindArgumentList
	case *Identifier:
		return KindIdentifier
	case *Nullable:
		return KindNullable
	case *Parameter:
		return KindParameter
	case *Root:
		return KindRoot
	case *SimpleVar:
		return KindSimpleVar
	case *Var:
		return KindVar
	case *Dnumber:
		return KindDnumber
	case *Encapsed:
		return KindEncapsed
	case *EncapsedStringPart:
		return KindEncapsedStringPart
	case *Heredoc:
		return KindHeredoc
	case *Lnumber:
		return KindLnumber
	case *MagicConstant:
		return KindMagicConstant
	case *String:
		return KindString
	case *BreakStmt:
		return KindBreakStmt
	case *CaseStmt:
		return KindCaseStmt
	case *CaseListStmt:
		return KindCaseListStmt
	case *CatchStmt:
		return KindCatchStmt
	case *ClassStmt:
		return KindClassStmt
	case *ClassConstListStmt:
		return KindClassConstListStmt
	case *ClassExtendsStmt:
		return KindClassExtendsStmt
	case *ClassImplementsStmt:
		return KindClassImplementsStmt
	case *ClassMethodStmt:
		return KindClassMethodStmt
	case *ConstListStmt:
		return KindConstListStmt
	case *ConstantStmt:
		return KindConstantStmt
	case *ContinueStmt:
		return KindContinueStmt
	case *DeclareStmt:
		return KindDeclareStmt
	case *DefaultStmt:
		return KindDefaultStmt
	case *DoStmt:
		return KindDoStmt
	case *EchoStmt:
		return KindEchoStmt
	case *ElseStmt:
		return KindElseStmt
	case *ElseIfStmt:
		return KindElseIfStmt
	case *ExpressionStmt:
		return KindExpressionStmt
	case *FinallyStmt:
		return KindFinallyStmt
	case *ForStmt:
		return KindForStmt
	case *ForeachStmt:
		return KindForeachStmt
	case *FunctionStmt:
		return KindFunctionStmt
	case *GlobalStmt:
		return KindGlobalStmt
	case *GotoStmt:
		return KindGotoStmt
	case *GroupUseStmt:
		return KindGroupUseStmt
	case *HaltCompilerStmt:
		return KindHaltCompilerStmt
	case *IfStmt:
		return KindIfStmt
	case *InlineHTMLStmt:
		return KindInlineHTMLStmt
	case *InterfaceStmt:
		return KindInterfaceStmt
	case *InterfaceExtendsStmt:
		return KindInterfaceExtendsStmt
	case *LabelStmt:
		return KindLabelStmt
	case *NamespaceStmt:
		return KindNamespaceStmt
	case *NopStmt:
		return KindNopStmt
	case *PropertyStmt:
		return KindPropertyStmt
	case *PropertyListStmt:
		return KindPropertyListStmt
	case *ReturnStmt:
		return KindReturnStmt
	case *StaticStmt:
		return KindStaticStmt
	case *StaticVarStmt:
		return KindStaticVarStmt
	case *StmtList:
		return KindStmtList
	case *SwitchStmt:
		return KindSwitchStmt
	case *ThrowStmt:
		return KindThrowStmt
	case *TraitStmt:
		return KindTraitStmt
	case *TraitAdaptationListStmt:
		return KindTraitAdaptationListStmt
	case *TraitMethodRefStmt:
		return KindTraitMethodRefStmt
	case *TraitUseStmt:
		return KindTraitUseStmt
	case *TraitUseAliasStmt:
		return KindTraitUseAliasStmt
	case *TraitUsePrecedenceStmt:
		return KindTraitUsePrecedenceStmt
	case *TryStmt:
		return KindTryStmt
	case *UnsetStmt:
		return KindUnsetStmt
	case *UseStmt:
		return KindUseStmt
	case *UseListStmt:
		return KindUseListStmt
	case *WhileStmt:
		return KindWhileStmt
	}

	panic(fmt.Sprintf("unhandled type %T", n))
}
