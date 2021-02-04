package ir

import (
	"github.com/z7zmey/php-parser/pkg/position"
	"github.com/z7zmey/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/phpdoc"
)

// TODO: make Alt and AltSyntax field names consistent.

// Assign is a `$Variable = $Expression` expression.
type Assign struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignBitwiseAnd is a `$Variable &= $Expression` expression.
type AssignBitwiseAnd struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignBitwiseOr is a `$Variable |= $Expression` expression.
type AssignBitwiseOr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignBitwiseXor is a `$Variable ^= $Expression` expression.
type AssignBitwiseXor struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignCoalesce is a `$Variable ??= $Expression` expression.
type AssignCoalesce struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignConcat is a `$Variable .= $Expression` expression.
type AssignConcat struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignDiv is a `$Variable /= $Expression` expression.
type AssignDiv struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignMinus is a `$Variable -= $Expression` expression.
type AssignMinus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignMod is a `$Variable %= $Expression` expression.
type AssignMod struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignMul is a `$Variable *= $Expression` expression.
type AssignMul struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignPlus is a `$Variable += $Expression` expression.
type AssignPlus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignPow is a `$Variable **= $Expression` expression.
type AssignPow struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignReference is a `$Variable &= $Expression` expression.
type AssignReference struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignShiftLeft is a `$Variable <<= $Expression` expression.
type AssignShiftLeft struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AssignShiftRight is a `$Variable >>= $Expression` expression.
type AssignShiftRight struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *token.Token
	Expression   Node
}

// AnonClassExpr is an anonymous class expression.
// $Args may contain constructor call arguments `new class ($Args...) {}`.
type AnonClassExpr struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	ClassTkn             *token.Token
	OpenParenthesisTkn   *token.Token
	ArgsFreeFloating     freefloating.Collection
	Args                 []Node
	SeparatorTkns        []*token.Token
	CloseParenthesisTkn  *token.Token
	OpenCurlyBracketTkn  *token.Token
	CloseCurlyBracketTkn *token.Token
	Class
}

// BitwiseAndExpr is a `$Left & $Right` expression.
type BitwiseAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// BitwiseOrExpr is a `$Left | $Right` expression.
type BitwiseOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// BitwiseXorExpr is a `$Left ^ $Right` expression.
type BitwiseXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// BooleanAndExpr is a `$Left && $Right` expression.
type BooleanAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// BooleanOrExpr is a `$Left || $Right` expression.
type BooleanOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// CoalesceExpr is a `$Left ?? $Right` expression.
type CoalesceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// ConcatExpr is a `$Left . $Right` expression.
type ConcatExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// DivExpr is a `$Left / $Right` expression.
type DivExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// EqualExpr is a `$Left == $Right` expression.
type EqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// GreaterExpr is a `$Left > $Right` expression.
type GreaterExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// GreaterOrEqualExpr is a `$Left >= $Right` expression.
type GreaterOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// IdenticalExpr is a `$Left === $Right` expression.
type IdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// LogicalAndExpr is a `$Left and $Right` expression.
type LogicalAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// LogicalOrExpr is a `$Left or $Right` expression.
type LogicalOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// LogicalXorExpr is a `$Left xor $Right` expression.
type LogicalXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// MinusExpr is a `$Left - $Right` expression.
type MinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// ModExpr is a `$Left % $Right` expression.
type ModExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// MulExpr is a `$Left * $Right` expression.
type MulExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// NotEqualExpr is a `$Left != $Right` expression.
type NotEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// NotIdenticalExpr is a `$Left !== $Right` expression.
type NotIdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// PlusExpr is a `$Left + $Right` expression.
type PlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// PowExpr is a `$Left ** $Right` expression.
type PowExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// ShiftLeftExpr is a `$Left << $Right` expression.
type ShiftLeftExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// ShiftRightExpr is a `$Left >> $Right` expression.
type ShiftRightExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// SmallerExpr is a `$Left < $Right` expression.
type SmallerExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// SmallerOrEqualExpr is a `$Left <= $Right` expression.
type SmallerOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// SpaceshipExpr is a `$Left <=> $Right` expression.
type SpaceshipExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *token.Token
	Right        Node
}

// TypeCastExpr is a `($Type)$Expr` expression.
type TypeCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	CastTkn      *token.Token
	Type         string // "array" "bool" "int" "float" "object" "string"
	Expr         Node
}

// UnsetCastExpr is a `(unset)$Expr` expression.
type UnsetCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	CastTkn      *token.Token
	Expr         Node
}

// ArrayExpr is a `array($Items...)` expression.
// If $ShortSyntax is true, it's `[$Items...]`.
type ArrayExpr struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	ArrayTkn        *token.Token
	OpenBracketTkn  *token.Token
	Items           []*ArrayItemExpr
	SeparatorTkns   []*token.Token
	CloseBracketTkn *token.Token
	ShortSyntax     bool
}

// ArrayDimFetchExpr is a `$Variable[$Dim]` expression.
// If $CurlyBrace is true, it's `$Variable{$Dim}`
type ArrayDimFetchExpr struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	Variable        Node
	OpenBracketTkn  *token.Token
	Dim             Node
	CloseBracketTkn *token.Token
	CurlyBrace      bool
}

// ArrayItemExpr is a `$Key => $Val` expression.
// If $Unpack is true, it's `...$Val` ($Key is nil).
//
// TODO: make unpack a separate node?
type ArrayItemExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	EllipsisTkn    *token.Token
	Key            Node
	DoubleArrowTkn *token.Token
	AmpersandTkn   *token.Token
	Val            Node
	Unpack         bool
}

// ArrowFunctionExpr is a `fn($Params...): $ReturnType => $Expr` expression.
// If $ReturnsRef is true, it's `fn&($Params...): $ReturnType => $Expr`.
// If $Static is true, it's `static fn($Params...): $ReturnType => $Expr`.
// $ReturnType is optional, without it we have `fn($Params...) => $Expr` syntax.
type ArrowFunctionExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	StaticTkn           *token.Token
	FnTkn               *token.Token
	AmpersandTkn        *token.Token
	OpenParenthesisTkn  *token.Token
	Params              []Node
	SeparatorTkns       []*token.Token
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	ReturnType          Node
	DoubleArrowTkn      *token.Token
	Expr                Node
	ReturnsRef          bool
	Static              bool

	Doc
}

// BitwiseNotExpr is a `~$Expr` expression.
type BitwiseNotExpr struct {
	FreeFloating freefloating.Collection
	TildaTkn     *token.Token
	Position     *position.Position
	Expr         Node
}

// BooleanNotExpr is a `!$Expr` expression.
type BooleanNotExpr struct {
	FreeFloating   freefloating.Collection
	ExclamationTkn *token.Token
	Position       *position.Position
	Expr           Node
}

// ClassConstFetchExpr is a `$Class::$ConstantName` expression.
type ClassConstFetchExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	Class          Node
	DoubleColonTkn *token.Token
	ConstantName   *Identifier
}

// CloneExpr is a `clone $Expr` expression.
type CloneExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	CloneTkn     *token.Token
	Expr         Node
}

// ClosureExpr is a `function($Params...) use ($ClosureUse) : $ReturnType { $Stmts... }` expression.
// If $ReturnsRef is true, it's `function&($Params...) use ($ClosureUse) : $ReturnType { $Stmts... }`.
// If $Static is true, it's `static function($Params...) use ($ClosureUse) : $ReturnType { $Stmts... }`.
// $ReturnType is optional, without it we have `function($Params...) use ($ClosureUse) { $Stmts... }` syntax.
// $ClosureUse is optional, without it we have `function($Params...) : $ReturnType { $Stmts... }` syntax.
type ClosureExpr struct {
	FreeFloating           freefloating.Collection
	Position               *position.Position
	StaticTkn              *token.Token
	FunctionTkn            *token.Token
	AmpersandTkn           *token.Token
	OpenParenthesisTkn     *token.Token
	Params                 []Node
	SeparatorTkns          []*token.Token
	CloseParenthesisTkn    *token.Token
	UseTkn                 *token.Token
	UseOpenParenthesisTkn  *token.Token
	ClosureUse             *ClosureUseExpr
	UseSeparatorTkns       []*token.Token
	UseCloseParenthesisTkn *token.Token
	ColonTkn               *token.Token
	ReturnType             Node
	OpenCurlyBracketTkn    *token.Token
	Stmts                  []Node
	CloseCurlyBracketTkn   *token.Token
	ReturnsRef             bool
	Static                 bool
	PhpDocComment          string
	PhpDoc                 []phpdoc.CommentPart
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
	FreeFloating        freefloating.Collection
	Position            *position.Position
	EmptyTkn            *token.Token
	OpenParenthesisTkn  *token.Token
	Expr                Node
	CloseParenthesisTkn *token.Token
}

// ErrorSuppressExpr is a `@$Expr` expression.
type ErrorSuppressExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	AtTkn        *token.Token
	Expr         Node
}

// EvalExpr is a `eval($Expr)` expression.
type EvalExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	EvalTkn             *token.Token
	OpenParenthesisTkn  *token.Token
	Expr                Node
	CloseParenthesisTkn *token.Token
}

// ExitExpr is a `exit($Expr)` expression.
// If $Die is true, it's `die($Expr)`.
type ExitExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ExitTkn             *token.Token
	OpenParenthesisTkn  *token.Token
	Expr                Node
	CloseParenthesisTkn *token.Token
	Die                 bool
}

// FunctionCallExpr is a `$Function($Args...)` expression.
type FunctionCallExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	Function            Node
	OpenParenthesisTkn  *token.Token
	ArgsFreeFloating    freefloating.Collection
	Args                []Node
	SeparatorTkns       []*token.Token
	CloseParenthesisTkn *token.Token
}

// ImportExpr is a `$Func $Expr` expression.
// It could be `include $Expr`, `require $Expr` and so on.
type ImportExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ImportTkn    *token.Token
	Func         string // "include" "include_once" "require" "require_once"
	Expr         Node
}

// InstanceOfExpr is a `$Expr instanceof $Class` expression.
type InstanceOfExpr struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	Expr          Node
	InstanceOfTkn *token.Token
	Class         Node
}

// IssetExpr is a `isset($Variables...)` expression.
type IssetExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	IssetTkn            *token.Token
	OpenParenthesisTkn  *token.Token
	Variables           []Node
	SeparatorTkns       []*token.Token
	CloseParenthesisTkn *token.Token
}

// ListExpr is a `list($Items...)` expression.
// Note that it may appear not only in assignments as LHS, but
// also in foreach value expressions.
// If $ShortSyntax is true, it's `[$Items]`.
type ListExpr struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	ListTkn         *token.Token
	OpenBracketTkn  *token.Token
	Items           []*ArrayItemExpr
	SeparatorTkns   []*token.Token
	CloseBracketTkn *token.Token
	ShortSyntax     bool
}

// MethodCallExpr is a `$Variable->$Method($Args...)` expression.
type MethodCallExpr struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	Variable             Node
	ObjectOperatorTkn    *token.Token
	OpenCurlyBracketTkn  *token.Token
	Method               Node
	CloseCurlyBracketTkn *token.Token
	OpenParenthesisTkn   *token.Token
	ArgsFreeFloating     freefloating.Collection
	Args                 []Node
	SeparatorTkns        []*token.Token
	CloseParenthesisTkn  *token.Token
}

// NewExpr is a `new $Class($Args...)` expression.
// If $Args is nil, it's `new $Class`.
type NewExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	NewTkn              *token.Token
	Class               Node
	OpenParenthesisTkn  *token.Token
	ArgsFreeFloating    freefloating.Collection
	Args                []Node
	SeparatorTkns       []*token.Token
	CloseParenthesisTkn *token.Token
}

// ParenExpr is a `($Expr)` expression.
type ParenExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	OpenParenthesisTkn  *token.Token
	Expr                Node
	CloseParenthesisTkn *token.Token
}

// PostDecExpr is a `$Variable--` expression.
type PostDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	DecTkn       *token.Token
}

// PostIncExpr is a `$Variable++` expression.
type PostIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	IncTkn       *token.Token
}

// PreDecExpr is a `--$Variable` expression.
type PreDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	DecTkn       *token.Token
	Variable     Node
}

// PreIncExpr is a `++$Variable` expression.
type PreIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	IncTkn       *token.Token
	Variable     Node
}

// PrintExpr is a `print $Expr` expression.
type PrintExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	PrintTkn     *token.Token
	Expr         Node
}

// PropertyFetchExpr is a `$Variable->$Property` expression.
type PropertyFetchExpr struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	Variable             Node
	ObjectOperatorTkn    *token.Token
	OpenCurlyBracketTkn  *token.Token
	Property             Node
	CloseCurlyBracketTkn *token.Token
}

// ReferenceExpr is a `&$Variable` expression.
type ReferenceExpr struct {
	FreeFloating freefloating.Collection
	AmpersandTkn *token.Token
	Position     *position.Position
	Variable     Node
}

// ShellExecExpr is a ``-quoted string.
type ShellExecExpr struct {
	FreeFloating     freefloating.Collection
	Position         *position.Position
	OpenBacktickTkn  *token.Token
	Parts            []Node
	CloseBacktickTkn *token.Token
}

// StaticCallExpr is a `$Class::$Call($Args...)` expression.
type StaticCallExpr struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	Class                Node
	DoubleColonTkn       *token.Token
	OpenCurlyBracketTkn  *token.Token
	Call                 Node
	CloseCurlyBracketTkn *token.Token
	OpenParenthesisTkn   *token.Token
	ArgsFreeFloating     freefloating.Collection
	Args                 []Node
	SeparatorTkns        []*token.Token
	CloseParenthesisTkn  *token.Token
}

// StaticPropertyFetchExpr is a `$Class::$Property` expression.
type StaticPropertyFetchExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	Class          Node
	DoubleColonTkn *token.Token
	Property       Node
}

// TernaryExpr is a `$Condition ? $IfTrue : $IfFalse` expression.
// If $IfTrue is nil, it's `$Condition ?: $IfFalse`.
type TernaryExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Condition    Node
	QuestionTkn  *token.Token
	IfTrue       Node
	ColonTkn     *token.Token
	IfFalse      Node
}

// UnaryMinusExpr is a `-$Expr` expression.
type UnaryMinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	MinusTkn     *token.Token
	Expr         Node
}

// UnaryPlusExpr is a `+$Expr` expression.
type UnaryPlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	PlusTkn      *token.Token
	Expr         Node
}

// YieldExpr is a `yield $Key => $Value` expression.
// If $Key is nil, it's `yield $Value`.
type YieldExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	YieldTkn       *token.Token
	Key            Node
	DoubleArrowTkn *token.Token
	Value          Node
}

// YieldFromExpr is a `yield from $Expr` expression.
type YieldFromExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	YieldFromTkn *token.Token
	Expr         Node
}

// Name is either a FQN, local name or a name that may need a further resolving.
// Use Name methods to interpret the $Value correctly.
type Name struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	NameTkn      *token.Token
	Value        string
}

// Argument is a wrapper node for func/method arguments.
// If $Variadic is true, it's `...$Expr`.
// If $IsReference is true, it's `&$Expr`.
type Argument struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	VariadicTkn  *token.Token
	AmpersandTkn *token.Token
	Expr         Node
	Variadic     bool
	IsReference  bool
}

// Identifier is like a name, but it's always resolved to itself.
// Identifier always consists of a single part.
type Identifier struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	IdentifierTkn *token.Token
	Value         string
}

// Nullable is a `?$Expr` expression.
type Nullable struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	QuestionTkn  *token.Token
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
	VariableType Node
	AmpersandTkn *token.Token
	VariadicTkn  *token.Token
	Variable     *SimpleVar
	EqualTkn     *token.Token
	DefaultValue Node
	ByRef        bool
	Variadic     bool
}

// Root is a node that wraps all file statements.
type Root struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
	EndTkn       *token.Token
}

// SimpleVar is a normal PHP variable like `$foo` or `$bar`.
type SimpleVar struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	DollarTkn     *token.Token
	IdentifierTkn *token.Token
	Name          string
}

// Var is variable variable expression like `$$foo` or `${"foo"}`.
type Var struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	DollarTkn            *token.Token
	OpenCurlyBracketTkn  *token.Token
	Expr                 Node
	CloseCurlyBracketTkn *token.Token
}

// Dnumber is a floating point literal.
type Dnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	NumberTkn    *token.Token
	Value        string
}

// Encapsed is a string literal with interpolated parts.
type Encapsed struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	OpenQuoteTkn  *token.Token
	Parts         []Node
	CloseQuoteTkn *token.Token
}

// EncapsedStringPart is a part of the Encapsed literal.
type EncapsedStringPart struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	EncapsedStrTkn *token.Token
	Value          string
}

// Heredoc is special PHP literal.
// Note that it may be a nowdoc, depending on the label.
type Heredoc struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	Label           string
	OpenHeredocTkn  *token.Token
	Parts           []Node
	CloseHeredocTkn *token.Token
}

// Lnumber is an integer literal.
type Lnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	NumberTkn    *token.Token
	Value        string
}

// MagicConstant is a special PHP constant like __FILE__ or __CLASS__.
// TODO: do we really need a separate node for these constants?
type MagicConstant struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	MagicConstTkn *token.Token
	Value         string
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
	MinusTkn     *token.Token
	StringTkn    *token.Token
	Value        string
	DoubleQuotes bool
}

// BadString is a string that we couldn't interpret correctly.
// The $Value contains uninterpreted (raw) string bytes.
// $Error contains the reason why this string is "bad".
//
// TODO: Maybe make String + Error field
type BadString struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	MinusTkn     *token.Token
	StringTkn    *token.Token
	Value        string
	DoubleQuotes bool
	Error        string
}

// BreakStmt is a `break $Expr` statement.
type BreakStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	BreakTkn     *token.Token
	Expr         Node
	SemiColonTkn *token.Token
}

// CaseStmt is a `case $Cond: $Stmts...` statement.
type CaseStmt struct {
	FreeFloating     freefloating.Collection
	Position         *position.Position
	CaseTkn          *token.Token
	Cond             Node
	CaseSeparatorTkn *token.Token
	Stmts            []Node
}

// CatchStmt is a `catch ($Types... $Variable) { $Stmts... }` statement.
// Note that $Types are |-separated, like in `T1 | T2`.
type CatchStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	CatchTkn             *token.Token
	OpenParenthesisTkn   *token.Token
	Types                []Node
	SeparatorTkns        []*token.Token
	Variable             *SimpleVar
	CloseParenthesisTkn  *token.Token
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token
}

// ClassStmt is a named class declaration.
// $Modifiers consist of identifiers like `final` and `abstract`.
type ClassStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	Modifiers            []*Identifier
	ClassTkn             *token.Token
	ClassName            *Identifier
	OpenCurlyBracketTkn  *token.Token
	CloseCurlyBracketTkn *token.Token
	Class
}

// ClassConstListStmt is a `$Modifiers... const $Consts...` statement.
// $Modifiers may specify the constant access level.
// Every element in $Consts is a *ConstantStmt.
type ClassConstListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	Modifiers     []*Identifier
	ConstTkn      *token.Token
	Consts        []Node
	SeparatorTkns []*token.Token
	SemiColonTkn  *token.Token

	Doc
}

// ClassExtendsStmt is a `extends $ClassName` statement.
type ClassExtendsStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ExtendsTkn   *token.Token
	ClassName    *Name
}

// ClassImplementsStmt is a `implements $InterfaceNames...` statement.
// TODO: shouldn't every InterfaceName be a *Name?
type ClassImplementsStmt struct {
	FreeFloating            freefloating.Collection
	Position                *position.Position
	ImplementsTkn           *token.Token
	ImplementsSeparatorTkns []*token.Token
	InterfaceNames          []Node
}

// ClassMethodStmt is a class method declaration.
type ClassMethodStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	Modifiers           []*Identifier
	FunctionTkn         *token.Token
	AmpersandTkn        *token.Token
	MethodName          *Identifier
	OpenParenthesisTkn  *token.Token
	Params              []Node
	SeparatorTkns       []*token.Token
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	ReturnType          Node
	Stmt                Node
	ReturnsRef          bool

	Doc
}

// ConstListStmt is a `const $Consts` statement.
// Every element in $Consts is a *ConstantStmt.
type ConstListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ConstTkn      *token.Token
	Consts        []Node
	SeparatorTkns []*token.Token
	SemiColonTkn  *token.Token
}

// ConstantStmt is a `$ConstantName = $Expr` statement.
// It's a part of the *ConstListStmt, *ClassConstListStmt and *DeclareStmt.
type ConstantStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ConstantName *Identifier
	EqualTkn     *token.Token
	Expr         Node
}

// ContinueStmt is a `continue $Expe` statement.
type ContinueStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ContinueTkn  *token.Token
	Expr         Node
	SemiColonTkn *token.Token
}

// DeclareStmt is a `declare ($Consts...) $Stmt` statement.
// $Stmt can be an empty statement, like in `declare ($Consts...);`,
// but it can also be a block like in `declare ($Consts...) {}`.
// If $Alt is true, the block will begin with `:` and end with `enddeclare`.
// Every element in $Consts is a *ConstantStmt.
type DeclareStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	DeclareTkn          *token.Token
	OpenParenthesisTkn  *token.Token
	Consts              []Node
	SeparatorTkns       []*token.Token
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	Stmt                Node
	EndDeclareTkn       *token.Token
	SemiColonTkn        *token.Token
	Alt                 bool
}

// DefaultStmt is a `default: $Stmts...` statement.
type DefaultStmt struct {
	FreeFloating     freefloating.Collection
	Position         *position.Position
	DefaultTkn       *token.Token
	CaseSeparatorTkn *token.Token
	Stmts            []Node
}

// DoStmt is a `do $Stmt while ($Cond)` statement.
type DoStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	DoTkn               *token.Token
	Stmt                Node
	WhileTkn            *token.Token
	OpenParenthesisTkn  *token.Token
	Cond                Node
	CloseParenthesisTkn *token.Token
	SemiColonTkn        *token.Token
}

// EchoStmt is a `echo $Exprs...` statement.
type EchoStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	EchoTkn       *token.Token
	Exprs         []Node
	SeparatorTkns []*token.Token
	SemiColonTkn  *token.Token
}

// ElseStmt is a `else $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:`.
type ElseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ElseTkn      *token.Token
	ColonTkn     *token.Token
	Stmt         Node
	AltSyntax    bool
}

// ElseIfStmt is a `elseif ($Cond) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endif`.
// $Merged tells whether this elseif is a merged `else if` statement.
type ElseIfStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ElseIfTkn           *token.Token
	ElseTkn             *token.Token
	IfTkn               *token.Token
	OpenParenthesisTkn  *token.Token
	Cond                Node
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	Stmt                Node
	AltSyntax           bool
	Merged              bool
}

// ExpressionStmt is an expression $Expr that is evaluated for side-effects only.
// When expression is used in a place where statement is expected, it
// becomes an ExpressionStmt.
type ExpressionStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
	SemiColonTkn *token.Token
}

// FinallyStmt is a `finally { $Stmts... }` statement.
type FinallyStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	FinallyTkn           *token.Token
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token
}

// ForStmt is a `for ($Init; $Cond; $Loop) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endfor`.
type ForStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ForTkn              *token.Token
	OpenParenthesisTkn  *token.Token
	Init                []Node
	InitSeparatorTkns   []*token.Token
	InitSemiColonTkn    *token.Token
	Cond                []Node
	CondSeparatorTkns   []*token.Token
	CondSemiColonTkn    *token.Token
	Loop                []Node
	LoopSeparatorTkns   []*token.Token
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	Stmt                Node
	EndForTkn           *token.Token
	SemiColonTkn        *token.Token
	AltSyntax           bool
}

// ForeachStmt is a `foreach ($Expr as $Key => $Variable) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endforeach`.
type ForeachStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ForeachTkn          *token.Token
	OpenParenthesisTkn  *token.Token
	Expr                Node
	AsTkn               *token.Token
	Key                 Node
	DoubleArrowTkn      *token.Token
	AmpersandTkn        *token.Token
	Variable            Node
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	Stmt                Node
	EndForeachTkn       *token.Token
	SemiColonTkn        *token.Token
	AltSyntax           bool
}

// FunctionStmt is a named function declaration.
type FunctionStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	FunctionTkn          *token.Token
	AmpersandTkn         *token.Token
	FunctionName         *Identifier
	OpenParenthesisTkn   *token.Token
	Params               []Node
	SeparatorTkns        []*token.Token
	CloseParenthesisTkn  *token.Token
	ColonTkn             *token.Token
	ReturnType           Node
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token
	ReturnsRef           bool
	PhpDocComment        string
	PhpDoc               []phpdoc.CommentPart
}

// GlobalStmt is a `global $Vars` statement.
type GlobalStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	GlobalTkn     *token.Token
	Vars          []Node
	SeparatorTkns []*token.Token
	SemiColonTkn  *token.Token
}

// GotoStmt is a `goto $Label` statement.
type GotoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	GotoTkn      *token.Token
	Label        *Identifier
	SemiColonTkn *token.Token
}

// GroupUseStmt is a `use $UseType $Prefix\{ $UseList }` statement.
// $UseType is a "function" or "const".
// TODO: change $UseType type to *Identifier?
type GroupUseStmt struct {
	FreeFloating          freefloating.Collection
	Position              *position.Position
	UseTkn                *token.Token
	UseType               *Identifier
	LeadingNsSeparatorTkn *token.Token
	Prefix                *Name
	NsSeparatorTkn        *token.Token
	OpenCurlyBracketTkn   *token.Token
	UseList               []Node
	SeparatorTkns         []*token.Token
	CloseCurlyBracketTkn  *token.Token
	SemiColonTkn          *token.Token
}

// HaltCompilerStmt is a `__halt_compiler()` statement.
type HaltCompilerStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	HaltCompilerTkn     *token.Token
	OpenParenthesisTkn  *token.Token
	CloseParenthesisTkn *token.Token
	SemiColonTkn        *token.Token
}

// IfStmt is a `if ($Cond) $Stmt` statement.
// $ElseIf contains an entire elseif chain, if any.
// $Else may contain an else part of the statement.
type IfStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	IfTkn               *token.Token
	OpenParenthesisTkn  *token.Token
	Cond                Node
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	Stmt                Node
	ElseIf              []Node
	Else                Node
	EndIfTkn            *token.Token
	SemiColonTkn        *token.Token
	ElseTkn             *token.Token
	AltSyntax           bool
}

// InlineHTMLStmt is a part of the script that will not be interpreted
// as a PHP script. In other words, it's everything outside of
// the <? and ?> tags.
type InlineHTMLStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	InlineHTMLTkn *token.Token
	Value         string
}

// InterfaceStmt is an interface declaration.
type InterfaceStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	InterfaceTkn         *token.Token
	InterfaceName        *Identifier
	ExtendsTkn           *token.Token
	Extends              *InterfaceExtendsStmt
	ExtendsSeparatorTkns []*token.Token
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token

	Doc
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
	ColonTkn     *token.Token
}

// NamespaceStmt is a `namespace $NamespaceName` statement.
// If $Stmts is not nil, it's `namespace $NamespaceName { $Stmts... }`.
type NamespaceStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	NsTkn                *token.Token
	NamespaceName        *Name
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token
	SemiColonTkn         *token.Token
}

// NopStmt is a `;` statement.
// It's also known as "empty statement".
type NopStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	SemiColonTkn *token.Token
}

// PropertyStmt is a `$Variable = $Expr` statement.
// It's a part of the *PropertyListStmt.
type PropertyStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     *SimpleVar
	EqualTkn     *token.Token
	Expr         Node

	Doc
}

// PropertyListStmt is a `$Modifiers $Type $Properties` statement.
// Every element in $Properties is a *PropertyStmt.
type PropertyListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	Modifiers     []*Identifier
	Type          Node
	Properties    []Node
	SeparatorTkns []*token.Token
	SemiColonTkn  *token.Token

	Doc
}

// ReturnStmt is a `return $Expr` statement.
type ReturnStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ReturnTkn    *token.Token
	Expr         Node
	SemiColonTkn *token.Token
}

// StaticStmt is a `static $Vars...` statement.
// Every element in $Vars is a *StaticVarStmt.
type StaticStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	StaticTkn     *token.Token
	Vars          []Node
	SeparatorTkns []*token.Token
	SemiColonTkn  *token.Token
}

// StaticVarStmt is a `$Variable = $Expr`.
// It's a part of the *StaticStmt.
type StaticVarStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     *SimpleVar
	EqualTkn     *token.Token
	Expr         Node
}

// StmtList is a `{ $Stmts... }` statement.
// It's also known as "block statement".
type StmtList struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token
}

// SwitchStmt is a `switch ($Cond) $CaseList` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endswitch`.
type SwitchStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	SwitchTkn            *token.Token
	OpenParenthesisTkn   *token.Token
	Cond                 Node
	CloseParenthesisTkn  *token.Token
	ColonTkn             *token.Token
	OpenCurlyBracketTkn  *token.Token
	CaseSeparatorTkn     *token.Token
	Cases                []Node
	CloseCurlyBracketTkn *token.Token
	EndSwitchTkn         *token.Token
	SemiColonTkn         *token.Token
	AltSyntax            bool
}

// ThrowStmt is a `throw $Expr` statement.
type ThrowStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ThrowTkn     *token.Token
	Expr         Node
	SemiColonTkn *token.Token
}

// TraitStmt is a trait declaration.
type TraitStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	TraitTkn             *token.Token
	TraitName            *Identifier
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token

	Doc
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
	FreeFloating         freefloating.Collection
	Position             *position.Position
	UseTkn               *token.Token
	Traits               []Node
	SeparatorTkns        []*token.Token
	OpenCurlyBracketTkn  *token.Token
	TraitAdaptationList  Node
	CloseCurlyBracketTkn *token.Token
	SemiColonTkn         *token.Token
}

type TraitUseAliasStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	DoubleColonTkn *token.Token
	Ref            Node
	AsTkn          *token.Token
	Modifier       Node
	Alias          *Identifier
	SemiColonTkn   *token.Token
}

type TraitUsePrecedenceStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	DoubleColonTkn *token.Token
	Ref            Node
	InsteadofTkn   *token.Token
	Insteadof      []Node
	SeparatorTkns  []*token.Token
	SemiColonTkn   *token.Token
}

// TryStmt is a `try { $Stmts... } $Catches` statement.
// $Finally only presents if `finally {...}` block exists.
type TryStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	TryTkn               *token.Token
	OpenCurlyBracketTkn  *token.Token
	Stmts                []Node
	CloseCurlyBracketTkn *token.Token
	Catches              []Node
	Finally              Node
}

// UnsetStmt is a `unset($Vars...)` statement.
type UnsetStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	UnsetTkn            *token.Token
	OpenParenthesisTkn  *token.Token
	Vars                []Node
	SeparatorTkns       []*token.Token
	CloseParenthesisTkn *token.Token
	SemiColonTkn        *token.Token
}

type UseStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	UseType        *Identifier
	NsSeparatorTkn *token.Token
	Use            *Name
	AsTkn          *token.Token
	Alias          *Identifier
}

type UseListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	UseTkn        *token.Token
	UseType       *Identifier
	Uses          []Node
	SeparatorTkns []*token.Token
	SemiColonTkn  *token.Token
}

// WhileStmt is a `while ($Cond) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endwhile`.
type WhileStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	WhileTkn            *token.Token
	OpenParenthesisTkn  *token.Token
	Cond                Node
	CloseParenthesisTkn *token.Token
	ColonTkn            *token.Token
	Stmt                Node
	EndWhileTkn         *token.Token
	SemiColonTkn        *token.Token
	AltSyntax           bool
}
