package ir

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

type Visitor interface {
	EnterNode(Node) bool
	LeaveNode(Node)
}

type Node interface {
	Walk(Visitor)
	GetFreeFloating() *freefloating.Collection
}

type Assign struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignBitwiseAnd struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignBitwiseOr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignBitwiseXor struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignCoalesce struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignConcat struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignDiv struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignMinus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignMod struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignMul struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignPlus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignPow struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignReference struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignShiftLeft struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type AssignShiftRight struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

type BitwiseAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type BitwiseOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type BitwiseXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type BooleanAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type BooleanOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type CoalesceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type ConcatExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type DivExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type EqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type GreaterExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type GreaterOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type IdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type LogicalAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type LogicalOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type LogicalXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type MinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type ModExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type MulExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type NotEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type NotIdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type PlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type PowExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type ShiftLeftExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type ShiftRightExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type SmallerExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type SmallerOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type SpaceshipExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

type ArrayCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type BoolCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type DoubleCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type IntCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type ObjectCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type StringCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type UnsetCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type ArrayExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Items        []*ArrayItemExpr
	ShortSyntax  bool
}

type ArrayDimFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Dim          Node
}

type ArrayItemExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Key          Node
	Val          Node
	Unpack       bool
}

type ArrowFunctionExpr struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	Static        bool
	PhpDocComment string
	Params        []Node
	ReturnType    Node
	Expr          Node
}

type BitwiseNotExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type BooleanNotExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type ClassConstFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node
	ConstantName *Identifier
}

type CloneExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type ClosureExpr struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	Static        bool
	PhpDocComment string
	Params        []Node
	ClosureUse    *ClosureUseExpr
	ReturnType    Node
	Stmts         []Node
}

type ClosureUseExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Uses         []Node
}

type ConstFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Constant     Node
}

type EmptyExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type ErrorSuppressExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type EvalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type ExitExpr struct {
	FreeFloating freefloating.Collection
	Die          bool
	Position     *position.Position
	Expr         Node
}

type FunctionCallExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Function     Node
	ArgumentList *ArgumentList
}

type IncludeExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type IncludeOnceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type InstanceOfExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
	Class        Node
}

type IssetExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variables    []Node
}

type ListExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Items        []*ArrayItemExpr
	ShortSyntax  bool
}

type MethodCallExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Method       Node
	ArgumentList *ArgumentList
}

type NewExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node
	ArgumentList *ArgumentList
}

type ParenExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type PostDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

type PostIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

type PreDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

type PreIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

type PrintExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type PropertyFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Property     Node
}

type ReferenceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

type RequireExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type RequireOnceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type ShellExecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []Node
}

type StaticCallExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node
	Call         Node
	ArgumentList *ArgumentList
}

type StaticPropertyFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node
	Property     Node
}

type TernaryExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Condition    Node
	IfTrue       Node
	IfFalse      Node
}

type UnaryMinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type UnaryPlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type YieldExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Key          Node
	Value        Node
}

type YieldFromExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type FullyQualifiedName struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []Node
}

type Name struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []Node
}

type NamePart struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type RelativeName struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []Node
}

type Argument struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variadic     bool
	IsReference  bool
	Expr         Node
}

type ArgumentList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Arguments    []Node
}

type Identifier struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type Nullable struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type Parameter struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ByRef        bool
	Variadic     bool
	VariableType Node
	Variable     *SimpleVar
	DefaultValue Node
}

type Root struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

type SimpleVar struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Name         string
}

type Var struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type Dnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type Encapsed struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []Node
}

type EncapsedStringPart struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type Heredoc struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Label        string
	Parts        []Node
}

type Lnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type MagicConstant struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type String struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type BreakStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type CaseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmts        []Node
}

type CaseListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cases        []Node
}

type CatchStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Types        []Node
	Variable     *SimpleVar
	Stmts        []Node
}

type ClassStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	ClassName     *Identifier
	Modifiers     []*Identifier
	ArgumentList  *ArgumentList
	Extends       *ClassExtendsStmt
	Implements    *ClassImplementsStmt
	Stmts         []Node
}

type ClassConstListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Modifiers    []*Identifier
	Consts       []Node
}

type ClassExtendsStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ClassName    Node
}

type ClassImplementsStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	InterfaceNames []Node
}

type ClassMethodStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	PhpDocComment string
	MethodName    *Identifier
	Modifiers     []*Identifier
	Params        []Node
	ReturnType    Node
	Stmt          Node
}

type ConstListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Consts       []Node
}

type ConstantStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	ConstantName  *Identifier
	Expr          Node
}

type ContinueStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type DeclareStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Consts       []Node
	Stmt         Node
	Alt          bool
}

type DefaultStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

type DoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmt         Node
	Cond         Node
}

type EchoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Exprs        []Node
}

type ElseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmt         Node
	AltSyntax    bool
}

type ElseIfStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmt         Node
	AltSyntax    bool
	Merged       bool
}

type ExpressionStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type FinallyStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

type ForStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Init         []Node
	Cond         []Node
	Loop         []Node
	Stmt         Node
	AltSyntax    bool
}

type ForeachStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
	Key          Node
	Variable     Node
	Stmt         Node
	AltSyntax    bool
}

type FunctionStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	PhpDocComment string
	FunctionName  *Identifier
	Params        []Node
	ReturnType    Node
	Stmts         []Node
}

type GlobalStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Vars         []Node
}

type GotoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Label        *Identifier
}

type GroupUseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      Node
	Prefix       Node
	UseList      []Node
}

type HaltCompilerStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
}

type IfStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmt         Node
	ElseIf       []Node
	Else         Node
	AltSyntax    bool
}

type InlineHTMLStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

type InterfaceStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	InterfaceName *Identifier
	Extends       *InterfaceExtendsStmt
	Stmts         []Node
}

type InterfaceExtendsStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	InterfaceNames []Node
}

type LabelStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	LabelName    *Identifier
}

type NamespaceStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	NamespaceName Node
	Stmts         []Node
}

type NopStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
}

type PropertyStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	Variable      *SimpleVar
	Expr          Node
}

type PropertyListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Modifiers    []*Identifier
	Type         Node
	Properties   []Node
}

type ReturnStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type StaticStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Vars         []Node
}

type StaticVarStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     *SimpleVar
	Expr         Node
}

type StmtList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

type SwitchStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	CaseList     *CaseListStmt
	AltSyntax    bool
}

type ThrowStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

type TraitStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	TraitName     *Identifier
	Stmts         []Node
}

type TraitAdaptationListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Adaptations  []Node
}

type TraitMethodRefStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Trait        Node
	Method       *Identifier
}

type TraitUseStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	Traits              []Node
	TraitAdaptationList Node
}

type TraitUseAliasStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Ref          Node
	Modifier     Node
	Alias        *Identifier
}

type TraitUsePrecedenceStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Ref          Node
	Insteadof    []Node
}

type TryStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
	Catches      []Node
	Finally      Node
}

type UnsetStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Vars         []Node
}

type UseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      *Identifier
	Use          Node
	Alias        *Identifier
}

type UseListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      Node
	Uses         []Node
}

type WhileStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmt         Node
	AltSyntax    bool
}
