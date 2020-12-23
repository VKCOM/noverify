package ir

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/phpdoc"
)

// TODO: make Alt and AltSyntax field names consistent.

// Assign is a `$Variable = $Expression` expression.
type Assign struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignBitwiseAnd is a `$Variable &= $Expression` expression.
type AssignBitwiseAnd struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignBitwiseOr is a `$Variable |= $Expression` expression.
type AssignBitwiseOr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignBitwiseXor is a `$Variable ^= $Expression` expression.
type AssignBitwiseXor struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignCoalesce is a `$Variable ??= $Expression` expression.
type AssignCoalesce struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignConcat is a `$Variable .= $Expression` expression.
type AssignConcat struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignDiv is a `$Variable /= $Expression` expression.
type AssignDiv struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignMinus is a `$Variable -= $Expression` expression.
type AssignMinus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignMod is a `$Variable %= $Expression` expression.
type AssignMod struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignMul is a `$Variable *= $Expression` expression.
type AssignMul struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignPlus is a `$Variable += $Expression` expression.
type AssignPlus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignPow is a `$Variable **= $Expression` expression.
type AssignPow struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignReference is a `$Variable &= $Expression` expression.
type AssignReference struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignShiftLeft is a `$Variable <<= $Expression` expression.
type AssignShiftLeft struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AssignShiftRight is a `$Variable >>= $Expression` expression.
type AssignShiftRight struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Expression   Node
}

// AnonClassExpr is an anonymous class expression.
// $Args may contain constructor call arguments `new class ($Args...) {}`.
type AnonClassExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class

	ArgsFreeFloating freefloating.Collection
	Args             []Node
}

// BitwiseAndExpr is a `$Left & $Right` expression.
type BitwiseAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// BitwiseOrExpr is a `$Left | $Right` expression.
type BitwiseOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// BitwiseXorExpr is a `$Left ^ $Right` expression.
type BitwiseXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// BooleanAndExpr is a `$Left && $Right` expression.
type BooleanAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// BooleanOrExpr is a `$Left || $Right` expression.
type BooleanOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// CoalesceExpr is a `$Left ?? $Right` expression.
type CoalesceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// ConcatExpr is a `$Left . $Right` expression.
type ConcatExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// DivExpr is a `$Left / $Right` expression.
type DivExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// EqualExpr is a `$Left == $Right` expression.
type EqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// GreaterExpr is a `$Left > $Right` expression.
type GreaterExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// GreaterOrEqualExpr is a `$Left >= $Right` expression.
type GreaterOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// IdenticalExpr is a `$Left === $Right` expression.
type IdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// LogicalAndExpr is a `$Left and $Right` expression.
type LogicalAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// LogicalOrExpr is a `$Left or $Right` expression.
type LogicalOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// LogicalXorExpr is a `$Left xor $Right` expression.
type LogicalXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// MinusExpr is a `$Left - $Right` expression.
type MinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// ModExpr is a `$Left % $Right` expression.
type ModExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// MulExpr is a `$Left * $Right` expression.
type MulExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// NotEqualExpr is a `$Left != $Right` expression.
type NotEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// NotIdenticalExpr is a `$Left !== $Right` expression.
type NotIdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// PlusExpr is a `$Left + $Right` expression.
type PlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// PowExpr is a `$Left ** $Right` expression.
type PowExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// ShiftLeftExpr is a `$Left << $Right` expression.
type ShiftLeftExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// ShiftRightExpr is a `$Left >> $Right` expression.
type ShiftRightExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// SmallerExpr is a `$Left < $Right` expression.
type SmallerExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// SmallerOrEqualExpr is a `$Left <= $Right` expression.
type SmallerOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// SpaceshipExpr is a `$Left <=> $Right` expression.
type SpaceshipExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	Right        Node
}

// TypeCastExpr is a `($Type)$Expr` expression.
type TypeCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Type         string // "array" "bool" "int" "float" "object" "string"
	Expr         Node
}

// UnsetCastExpr is a `(unset)$Expr` expression.
type UnsetCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// ArrayExpr is a `array($Items...)` expression.
// If $ShortSyntax is true, it's `[$Items...]`.
type ArrayExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Items        []*ArrayItemExpr
	ShortSyntax  bool
}

// ArrayDimFetchExpr is a `$Variable[$Dim]` expression.
type ArrayDimFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Dim          Node
}

// ArrayItemExpr is a `$Key => $Val` expression.
// If $Unpack is true, it's `...$Val` ($Key is nil).
//
// TODO: make unpack a separate node?
type ArrayItemExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Key          Node
	Val          Node
	Unpack       bool
}

// ArrowFunctionExpr is a `fn($Params...): $ReturnType => $Expr` expression.
// If $ReturnsRef is true, it's `fn&($Params...): $ReturnType => $Expr`.
// If $Static is true, it's `static fn($Params...): $ReturnType => $Expr`.
// $ReturnType is optional, without it we have `fn($Params...) => $Expr` syntax.
type ArrowFunctionExpr struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	Static        bool
	PhpDocComment string
	PhpDoc        []phpdoc.CommentPart
	Params        []Node
	ReturnType    Node
	Expr          Node
}

// BitwiseNotExpr is a `~$Expr` expression.
type BitwiseNotExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// BooleanNotExpr is a `!$Expr` expression.
type BooleanNotExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// ClassConstFetchExpr is a `$Class::$ConstantName` expression.
type ClassConstFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node
	ConstantName *Identifier
}

// CloneExpr is a `clone $Expr` expression.
type CloneExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// ClosureExpr is a `function($Params...) use ($ClosureUse) : $ReturnType { $Stmts... }` expression.
// If $ReturnsRef is true, it's `function&($Params...) use ($ClosureUse) : $ReturnType { $Stmts... }`.
// If $Static is true, it's `static function($Params...) use ($ClosureUse) : $ReturnType { $Stmts... }`.
// $ReturnType is optional, without it we have `function($Params...) use ($ClosureUse) { $Stmts... }` syntax.
// $ClosureUse is optional, without it we have `function($Params...) : $ReturnType { $Stmts... }` syntax.
type ClosureExpr struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	Static        bool
	PhpDocComment string
	PhpDoc        []phpdoc.CommentPart
	Params        []Node
	ClosureUse    *ClosureUseExpr
	ReturnType    Node
	Stmts         []Node
}

// ClosureUseExpr is a `use ($Uses...)` expression.
// TODO: it's not a expression really.
type ClosureUseExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Uses         []Node
}

// ConstFetchExpr is a `$Constant` expression.
type ConstFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Constant     *Name
}

// EmptyExpr is a `empty($Expr)` expression.
type EmptyExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// ErrorSuppressExpr is a `@$Expr` expression.
type ErrorSuppressExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// EvalExpr is a `eval($Expr)` expression.
type EvalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// ExitExpr is a `exit($Expr)` expression.
// If $Die is true, it's `die($Expr)`.
type ExitExpr struct {
	FreeFloating freefloating.Collection
	Die          bool
	Position     *position.Position
	Expr         Node
}

// FunctionCallExpr is a `$Function($Args...)` expression.
type FunctionCallExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Function     Node

	ArgsFreeFloating freefloating.Collection
	Args             []Node
}

// ImportExpr is a `$Func $Expr` expression.
// It could be `include $Expr`, `require $Expr` and so on.
type ImportExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Func         string // "include" "include_once" "require" "require_once"
	Expr         Node
}

// InstanceOfExpr is a `$Expr instanceof $Class` expression.
type InstanceOfExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
	Class        Node
}

// IssetExpr is a `isset($Variables...)` expression.
type IssetExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variables    []Node
}

// ListExpr is a `list($Items...)` expression.
// Note that it may appear not only in assignments as LHS, but
// also in foreach value expressions.
type ListExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Items        []*ArrayItemExpr
	ShortSyntax  bool
}

// MethodCallExpr is a `$Variable->$Method($Args...)` expression.
type MethodCallExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Method       Node

	ArgsFreeFloating freefloating.Collection
	Args             []Node
}

// NewExpr is a `new $Class($Args...)` expression.
// If $Args is nil, it's `new $Class`.
type NewExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node

	ArgsFreeFloating freefloating.Collection
	Args             []Node
}

// ParenExpr is a `($Expr)` expression.
type ParenExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// PostDecExpr is a `$Variable--` expression.
type PostDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

// PostIncExpr is a `$Variable++` expression.
type PostIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

// PreDecExpr is a `--$Variable` expression.
type PreDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

// PreIncExpr is a `++$Variable` expression.
type PreIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

// PrintExpr is a `print $Expr` expression.
type PrintExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// PropertyFetchExpr is a `$Variable->$Property` expression.
type PropertyFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	Property     Node
}

// ReferenceExpr is a `&$Variable` expression.
type ReferenceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
}

// ShellExecExpr is a ``-quoted string.
type ShellExecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []Node
}

// StaticCallExpr is a `$Class::$Call($Args...)` expression.
type StaticCallExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node
	Call         Node

	ArgsFreeFloating freefloating.Collection
	Args             []Node
}

// StaticPropertyFetchExpr is a `$Class::$Property` expression.
type StaticPropertyFetchExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        Node
	Property     Node
}

// TernaryExpr is a `$Condition ? $IfTrue : $IfFalse` expression.
// If $IfTrue is nil, it's `$Condition ?: $IfFalse`.
type TernaryExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Condition    Node
	IfTrue       Node
	IfFalse      Node
}

// UnaryMinusExpr is a `-$Expr` expression.
type UnaryMinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// UnaryPlusExpr is a `+$Expr` expression.
type UnaryPlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// YieldExpr is a `yield $Key => $Value` expression.
// If $Key is nil, it's `yield $Value`.
type YieldExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Key          Node
	Value        Node
}

// YieldFromExpr is a `yield from $Expr` expression.
type YieldFromExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// Name is either a FQN, local name or a name that may need a further resolving.
// Use Name methods to interpret the $Value correctly.
type Name struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// Argument is a wrapper node for func/method arguments.
// If $Variadic is true, it's `...$Expr`.
// If $IsReference is true, it's `&$Expr`.
type Argument struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variadic     bool
	IsReference  bool
	Expr         Node
}

// Identifier is like a name, but it's always resolved to itself.
// Identifier always consists of a single part.
type Identifier struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// Nullable is a `?$Expr` expression.
type Nullable struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// Parameter is a function param declaration.
// Possible syntaxes:
// $VariableType $Variable = $DefaultValue
// $VariableType $Variable
// $Variable
// If $ByRef is true, it's `&$Variable`.
// If $Variadic is true, it's `...$Variable`.
type Parameter struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ByRef        bool
	Variadic     bool
	VariableType Node
	Variable     *SimpleVar
	DefaultValue Node
}

// Root is a node that wraps all file statements.
type Root struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

// SimpleVar is a normal PHP variable like `$foo` or `$bar`.
type SimpleVar struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Name         string
}

// Var is variable variable expression like `$$foo` or `${"foo"}`.
type Var struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// Dnumber is a floating point literal.
type Dnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// Encapsed is a string literal with interpolated parts.
type Encapsed struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []Node
}

// EncapsedStringPart is a part of the Encapsed literal.
type EncapsedStringPart struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// Heredoc is special PHP literal.
// Note that it may be a nowdoc, depending on the label.
type Heredoc struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Label        string
	Parts        []Node
}

// Lnumber is an integer literal.
type Lnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// MagicConstant is a special PHP constant like __FILE__ or __CLASS__.
// TODO: do we really need a separate node for these constants?
type MagicConstant struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// String is a simple (no interpolation) string literal.
//
// The $Value contains interpreted string bytes, if you need a raw
// string value, use positions and fetch relevant the source bytes.
//
// $DoubleQuotes tell whether originally this string literal was ""-quoted.
type String struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
	DoubleQuotes bool
}

// BadString is a string that we couldn't interpret correctly.
// The $Value contains uninterpreted (raw) string bytes.
// $Error contains the reason why this string is "bad".
type BadString struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
	DoubleQuotes bool
	Error        string
}

// BreakStmt is a `break $Expr` statement.
type BreakStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// CaseStmt is a `case $Cond: $Stmts...` statement.
type CaseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmts        []Node
}

// CaseListStmt is a wrapper node that contains all switch statement cases.
// TODO: can we get rid of it?
type CaseListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cases        []Node
}

// CatchStmt is a `catch ($Types... $Variable) { $Stmts... }` statement.
// Note that $Types are |-separated, like in `T1 | T2`.
type CatchStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Types        []Node
	Variable     *SimpleVar
	Stmts        []Node
}

// ClassStmt is a named class declaration.
// $Modifiers consist of identifiers like `final` and `abstract`.
type ClassStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ClassName    *Identifier
	Modifiers    []*Identifier
	Class
}

// ClassConstListStmt is a `$Modifiers... const $Consts...` statement.
// $Modifiers may specify the constant access level.
// Every element in $Consts is a *ConstantStmt.
type ClassConstListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Modifiers    []*Identifier
	Consts       []Node
}

// ClassExtendsStmt is a `extends $ClassName` statement.
type ClassExtendsStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ClassName    *Name
}

// ClassImplementsStmt is a `implements $InterfaceNames...` statement.
// TODO: shouldn't every InterfaceName be a *Name?
type ClassImplementsStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	InterfaceNames []Node
}

// ClassMethodStmt is a class method declaration.
type ClassMethodStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	PhpDocComment string
	PhpDoc        []phpdoc.CommentPart
	MethodName    *Identifier
	Modifiers     []*Identifier
	Params        []Node
	ReturnType    Node
	Stmt          Node
}

// ConstListStmt is a `const $Consts` statement.
// Every element in $Consts is a *ConstantStmt.
type ConstListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Consts       []Node
}

// ConstantStmt is a `$ConstantName = $Expr` statement.
// It's a part of the *ConstListStmt, *ClassConstListStmt and *DeclareStmt.
type ConstantStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	ConstantName  *Identifier
	Expr          Node
}

// ContinueStmt is a `continue $Expe` statement.
type ContinueStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// DeclareStmt is a `declare ($Consts...) $Stmt` statement.
// $Stmt can be an empty statement, like in `declare ($Consts...);`,
// but it can also be a block like in `declare ($Consts...) {}`.
// If $Alt is true, the block will begin with `:` and end with `enddeclare`.
// Every element in $Consts is a *ConstantStmt.
type DeclareStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Consts       []Node
	Stmt         Node
	Alt          bool
}

// DefaultStmt is a `default: $Stmts...` statement.
type DefaultStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

// DoStmt is a `do $Stmt while ($Cond)` statement.
type DoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmt         Node
	Cond         Node
}

// EchoStmt is a `echo $Exprs...` statement.
type EchoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Exprs        []Node
}

// ElseStmt is a `else $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:`.
type ElseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmt         Node
	AltSyntax    bool
}

// ElseIfStmt is a `elseif ($Cond) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endif`.
// $Merged tells whether this elseif is a merged `else if` statement.
type ElseIfStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmt         Node
	AltSyntax    bool
	Merged       bool
}

// ExpressionStmt is an expression $Expr that is evaluated for side-effects only.
// When expression is used in a place where statement is expected, it
// becomes an ExpressionStmt.
type ExpressionStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// FinallyStmt is a `finally { $Stmts... }` statement.
type FinallyStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

// ForStmt is a `for ($Init; $Cond; $Loop) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endfor`.
type ForStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Init         []Node
	Cond         []Node
	Loop         []Node
	Stmt         Node
	AltSyntax    bool
}

// ForeachStmt is a `foreach ($Expr as $Key => $Variable) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endforeach`.
type ForeachStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
	Key          Node
	Variable     Node
	Stmt         Node
	AltSyntax    bool
}

// FunctionStmt is a named function declaration.
type FunctionStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	PhpDocComment string
	PhpDoc        []phpdoc.CommentPart
	FunctionName  *Identifier
	Params        []Node
	ReturnType    Node
	Stmts         []Node
}

// GlobalStmt is a `global $Vars` statement.
type GlobalStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Vars         []Node
}

// GotoStmt is a `goto $Label` statement.
type GotoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Label        *Identifier
}

// GroupUseStmt is a `use $UseType $Prefix\{ $UseList }` statement.
// $UseType is a "function" or "const".
// TODO: change $UseType type to *Identifier?
type GroupUseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      *Identifier
	Prefix       *Name
	UseList      []Node
}

// HaltCompilerStmt is a `__halt_compiler()` statement.
type HaltCompilerStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
}

// IfStmt is a `if ($Cond) $Stmt` statement.
// $ElseIf contains an entire elseif chain, if any.
// $Else may contain an else part of the statement.
type IfStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmt         Node
	ElseIf       []Node
	Else         Node
	AltSyntax    bool
}

// InlineHTMLStmt is a part of the script that will not be interpreted
// as a PHP script. In other words, it's everything outside of
// the <? and ?> tags.
type InlineHTMLStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// InterfaceStmt is an interface declaration.
type InterfaceStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	InterfaceName *Identifier
	Extends       *InterfaceExtendsStmt
	Stmts         []Node
}

// InterfaceExtendsStmt is a `extends $InterfaceNames...` statement.
// TODO: do we need this wrapper node?
// TODO: InterfaceNames could be a []*Name.
type InterfaceExtendsStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	InterfaceNames []Node
}

// LabelStmt is a `$LabelName:` statement.
type LabelStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	LabelName    *Identifier
}

// NamespaceStmt is a `namespace $NamespaceName` statement.
// If $Stmts is not nil, it's `namespace $NamespaceName { $Stmts... }`.
type NamespaceStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	NamespaceName *Name
	Stmts         []Node
}

// NopStmt is a `;` statement.
// It's also known as "empty statement".
type NopStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
}

// PropertyStmt is a `$Variable = $Expr` statement.
// It's a part of the *PropertyListStmt.
type PropertyStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	PhpDoc        []phpdoc.CommentPart
	Variable      *SimpleVar
	Expr          Node
}

// PropertyListStmt is a `$Modifiers $Type $Properties` statement.
// Every element in $Properties is a *PropertyStmt.
type PropertyListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Modifiers    []*Identifier
	Type         Node
	Properties   []Node
}

// ReturnStmt is a `return $Expr` statement.
type ReturnStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// StaticStmt is a `static $Vars...` statement.
// Every element in $Vars is a *StaticVarStmt.
type StaticStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Vars         []Node
}

// StaticVarStmt is a `$Variable = $Expr`.
// It's a part of the *StaticStmt.
type StaticVarStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     *SimpleVar
	Expr         Node
}

// StmtList is a `{ $Stmts... }` statement.
// It's also known as "block statement".
type StmtList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

// SwitchStmt is a `switch ($Cond) $CaseList` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endswitch`.
type SwitchStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	CaseList     *CaseListStmt
	AltSyntax    bool
}

// ThrowStmt is a `throw $Expr` statement.
type ThrowStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// TraitStmt is a trait declaration.
type TraitStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	TraitName     *Identifier
	Stmts         []Node
}

// TraitAdaptationListStmt is a block inside a *TraitUseStmt.
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

// TryStmt is a `try { $Stmts... } $Catches` statement.
// $Finally only presents if `finally {...}` block exists.
type TryStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
	Catches      []Node
	Finally      Node
}

// UnsetStmt is a `unset($Vars...)` statement.
type UnsetStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Vars         []Node
}

type UseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      *Identifier
	Use          *Name
	Alias        *Identifier
}

type UseListStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      *Identifier
	Uses         []Node
}

// WhileStmt is a `while ($Cond) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endwhile`.
type WhileStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         Node
	Stmt         Node
	AltSyntax    bool
}
