package ir

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/phpdoc"
)

// Token is a stub while switching to a new version of the parser.
// Replace later to token.Token from php-parser
type Token struct{}

// TODO: make Alt and AltSyntax field names consistent.

// Assign is a `$Variable = $Expression` expression.
type Assign struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignBitwiseAnd is a `$Variable &= $Expression` expression.
type AssignBitwiseAnd struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignBitwiseOr is a `$Variable |= $Expression` expression.
type AssignBitwiseOr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignBitwiseXor is a `$Variable ^= $Expression` expression.
type AssignBitwiseXor struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignCoalesce is a `$Variable ??= $Expression` expression.
type AssignCoalesce struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignConcat is a `$Variable .= $Expression` expression.
type AssignConcat struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignDiv is a `$Variable /= $Expression` expression.
type AssignDiv struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignMinus is a `$Variable -= $Expression` expression.
type AssignMinus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignMod is a `$Variable %= $Expression` expression.
type AssignMod struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignMul is a `$Variable *= $Expression` expression.
type AssignMul struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignPlus is a `$Variable += $Expression` expression.
type AssignPlus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignPow is a `$Variable **= $Expression` expression.
type AssignPow struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignReference is a `$Variable &= $Expression` expression.
type AssignReference struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignShiftLeft is a `$Variable <<= $Expression` expression.
type AssignShiftLeft struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AssignShiftRight is a `$Variable >>= $Expression` expression.
type AssignShiftRight struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	EqualTkn     *Token
	Expression   Node
}

// AnonClassExpr is an anonymous class expression.
// $Args may contain constructor call arguments `new class ($Args...) {}`.
type AnonClassExpr struct {
	FreeFloating            freefloating.Collection
	Position                *position.Position
	ClassTkn                *Token
	OpenParenthesisTkn      *Token
	ArgsFreeFloating        freefloating.Collection
	Args                    []Node
	SeparatorTkns           []*Token
	CloseParenthesisTkn     *Token
	ExtendsTkn              *Token
	ImplementsTkn           *Token
	ImplementsSeparatorTkns []*Token
	OpenCurlyBracketTkn     *Token
	CloseCurlyBracketTkn    *Token
	Class
}

// BitwiseAndExpr is a `$Left & $Right` expression.
type BitwiseAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// BitwiseOrExpr is a `$Left | $Right` expression.
type BitwiseOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// BitwiseXorExpr is a `$Left ^ $Right` expression.
type BitwiseXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// BooleanAndExpr is a `$Left && $Right` expression.
type BooleanAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// BooleanOrExpr is a `$Left || $Right` expression.
type BooleanOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// CoalesceExpr is a `$Left ?? $Right` expression.
type CoalesceExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// ConcatExpr is a `$Left . $Right` expression.
type ConcatExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// DivExpr is a `$Left / $Right` expression.
type DivExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// EqualExpr is a `$Left == $Right` expression.
type EqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// GreaterExpr is a `$Left > $Right` expression.
type GreaterExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// GreaterOrEqualExpr is a `$Left >= $Right` expression.
type GreaterOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// IdenticalExpr is a `$Left === $Right` expression.
type IdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// LogicalAndExpr is a `$Left and $Right` expression.
type LogicalAndExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// LogicalOrExpr is a `$Left or $Right` expression.
type LogicalOrExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// LogicalXorExpr is a `$Left xor $Right` expression.
type LogicalXorExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// MinusExpr is a `$Left - $Right` expression.
type MinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// ModExpr is a `$Left % $Right` expression.
type ModExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// MulExpr is a `$Left * $Right` expression.
type MulExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// NotEqualExpr is a `$Left != $Right` expression.
type NotEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// NotIdenticalExpr is a `$Left !== $Right` expression.
type NotIdenticalExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// PlusExpr is a `$Left + $Right` expression.
type PlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// PowExpr is a `$Left ** $Right` expression.
type PowExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// ShiftLeftExpr is a `$Left << $Right` expression.
type ShiftLeftExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// ShiftRightExpr is a `$Left >> $Right` expression.
type ShiftRightExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// SmallerExpr is a `$Left < $Right` expression.
type SmallerExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// SmallerOrEqualExpr is a `$Left <= $Right` expression.
type SmallerOrEqualExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// SpaceshipExpr is a `$Left <=> $Right` expression.
type SpaceshipExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         Node
	OpTkn        *Token
	Right        Node
}

// TypeCastExpr is a `($Type)$Expr` expression.
type TypeCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	CastTkn      *Token
	Type         string // "array" "bool" "int" "float" "object" "string"
	Expr         Node
}

// UnsetCastExpr is a `(unset)$Expr` expression.
type UnsetCastExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	CastTkn      *Token
	Expr         Node
}

// ArrayExpr is a `array($Items...)` expression.
// If $ShortSyntax is true, it's `[$Items...]`.
type ArrayExpr struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	ArrayTkn        *Token
	OpenBracketTkn  *Token
	Items           []*ArrayItemExpr
	SeparatorTkns   []*Token
	CloseBracketTkn *Token
	ShortSyntax     bool
}

// ArrayDimFetchExpr is a `$Variable[$Dim]` expression.
// If $CurlyBrace is true, it's `$Variable{$Dim}`
type ArrayDimFetchExpr struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	Variable        Node
	OpenBracketTkn  *Token
	Dim             Node
	CloseBracketTkn *Token
	CurlyBrace      bool
}

// ArrayItemExpr is a `$Key => $Val` expression.
// If $Unpack is true, it's `...$Val` ($Key is nil).
//
// TODO: make unpack a separate node?
type ArrayItemExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	EllipsisTkn    *Token
	Key            Node
	DoubleArrowTkn *Token
	AmpersandTkn   *Token
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
	StaticTkn           *Token
	FnTkn               *Token
	AmpersandTkn        *Token
	OpenParenthesisTkn  *Token
	Params              []Node
	SeparatorTkns       []*Token
	CloseParenthesisTkn *Token
	ColonTkn            *Token
	ReturnType          Node
	DoubleArrowTkn      *Token
	Expr                Node
	ReturnsRef          bool
	Static              bool
	PhpDocComment       string
	PhpDoc              []phpdoc.CommentPart
}

// BitwiseNotExpr is a `~$Expr` expression.
type BitwiseNotExpr struct {
	FreeFloating freefloating.Collection
	TildaTkn     *Token
	Position     *position.Position
	Expr         Node
}

// BooleanNotExpr is a `!$Expr` expression.
type BooleanNotExpr struct {
	FreeFloating   freefloating.Collection
	ExclamationTkn *Token
	Position       *position.Position
	Expr           Node
}

// ClassConstFetchExpr is a `$Class::$ConstantName` expression.
type ClassConstFetchExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	Class          Node
	DoubleColonTkn *Token
	ConstantName   *Identifier
}

// CloneExpr is a `clone $Expr` expression.
type CloneExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	CloneTkn     *Token
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
	StaticTkn              *Token
	FunctionTkn            *Token
	AmpersandTkn           *Token
	OpenParenthesisTkn     *Token
	Params                 []Node
	SeparatorTkns          []*Token
	CloseParenthesisTkn    *Token
	UseTkn                 *Token
	UseOpenParenthesisTkn  *Token
	ClosureUse             *ClosureUseExpr
	UseSeparatorTkns       []*Token
	UseCloseParenthesisTkn *Token
	ColonTkn               *Token
	ReturnType             Node
	OpenCurlyBracketTkn    *Token
	Stmts                  []Node
	CloseCurlyBracketTkn   *Token
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
	EmptyTkn            *Token
	OpenParenthesisTkn  *Token
	Expr                Node
	CloseParenthesisTkn *Token
}

// ErrorSuppressExpr is a `@$Expr` expression.
type ErrorSuppressExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	AtTkn        *Token
	Expr         Node
}

// EvalExpr is a `eval($Expr)` expression.
type EvalExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	EvalTkn             *Token
	OpenParenthesisTkn  *Token
	Expr                Node
	CloseParenthesisTkn *Token
}

// ExitExpr is a `exit($Expr)` expression.
// If $Die is true, it's `die($Expr)`.
type ExitExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ExitTkn             *Token
	OpenParenthesisTkn  *Token
	Expr                Node
	CloseParenthesisTkn *Token
	Die                 bool
}

// FunctionCallExpr is a `$Function($Args...)` expression.
type FunctionCallExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	Function            Node
	OpenParenthesisTkn  *Token
	ArgsFreeFloating    freefloating.Collection
	Args                []Node
	SeparatorTkns       []*Token
	CloseParenthesisTkn *Token
}

// ImportExpr is a `$Func $Expr` expression.
// It could be `include $Expr`, `require $Expr` and so on.
type ImportExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ImportTkn    *Token
	Func         string // "include" "include_once" "require" "require_once"
	Expr         Node
}

// InstanceOfExpr is a `$Expr instanceof $Class` expression.
type InstanceOfExpr struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	Expr          Node
	InstanceOfTkn *Token
	Class         Node
}

// IssetExpr is a `isset($Variables...)` expression.
type IssetExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	IssetTkn            *Token
	OpenParenthesisTkn  *Token
	Variables           []Node
	SeparatorTkns       []*Token
	CloseParenthesisTkn *Token
}

// ListExpr is a `list($Items...)` expression.
// Note that it may appear not only in assignments as LHS, but
// also in foreach value expressions.
// If $ShortSyntax is true, it's `[$Items]`.
type ListExpr struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	ListTkn         *Token
	OpenBracketTkn  *Token
	Items           []*ArrayItemExpr
	SeparatorTkns   []*Token
	CloseBracketTkn *Token
	ShortSyntax     bool
}

// MethodCallExpr is a `$Variable->$Method($Args...)` expression.
type MethodCallExpr struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	Variable             Node
	ObjectOperatorTkn    *Token
	OpenCurlyBracketTkn  *Token
	Method               Node
	CloseCurlyBracketTkn *Token
	OpenParenthesisTkn   *Token
	ArgsFreeFloating     freefloating.Collection
	Args                 []Node
	SeparatorTkns        []*Token
	CloseParenthesisTkn  *Token
}

// NewExpr is a `new $Class($Args...)` expression.
// If $Args is nil, it's `new $Class`.
type NewExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	NewTkn              *Token
	Class               Node
	OpenParenthesisTkn  *Token
	ArgsFreeFloating    freefloating.Collection
	Args                []Node
	SeparatorTkns       []*Token
	CloseParenthesisTkn *Token
}

// ParenExpr is a `($Expr)` expression.
type ParenExpr struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	OpenParenthesisTkn  *Token
	Expr                Node
	CloseParenthesisTkn *Token
}

// PostDecExpr is a `$Variable--` expression.
type PostDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	DecTkn       *Token
}

// PostIncExpr is a `$Variable++` expression.
type PostIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     Node
	IncTkn       *Token
}

// PreDecExpr is a `--$Variable` expression.
type PreDecExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	DecTkn       *Token
	Variable     Node
}

// PreIncExpr is a `++$Variable` expression.
type PreIncExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	IncTkn       *Token
	Variable     Node
}

// PrintExpr is a `print $Expr` expression.
type PrintExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	PrintTkn     *Token
	Expr         Node
}

// PropertyFetchExpr is a `$Variable->$Property` expression.
type PropertyFetchExpr struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	Variable             Node
	ObjectOperatorTkn    *Token
	OpenCurlyBracketTkn  *Token
	Property             Node
	CloseCurlyBracketTkn *Token
}

// ReferenceExpr is a `&$Variable` expression.
type ReferenceExpr struct {
	FreeFloating freefloating.Collection
	AmpersandTkn *Token
	Position     *position.Position
	Variable     Node
}

// ShellExecExpr is a ``-quoted string.
type ShellExecExpr struct {
	FreeFloating     freefloating.Collection
	Position         *position.Position
	OpenBacktickTkn  *Token
	Parts            []Node
	CloseBacktickTkn *Token
}

// StaticCallExpr is a `$Class::$Call($Args...)` expression.
type StaticCallExpr struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	Class                Node
	DoubleColonTkn       *Token
	OpenCurlyBracketTkn  *Token
	Call                 Node
	CloseCurlyBracketTkn *Token
	OpenParenthesisTkn   *Token
	ArgsFreeFloating     freefloating.Collection
	Args                 []Node
	SeparatorTkns        []*Token
	CloseParenthesisTkn  *Token
}

// StaticPropertyFetchExpr is a `$Class::$Property` expression.
type StaticPropertyFetchExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	Class          Node
	DoubleColonTkn *Token
	Property       Node
}

// TernaryExpr is a `$Condition ? $IfTrue : $IfFalse` expression.
// If $IfTrue is nil, it's `$Condition ?: $IfFalse`.
type TernaryExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Condition    Node
	QuestionTkn  *Token
	IfTrue       Node
	ColonTkn     *Token
	IfFalse      Node
}

// UnaryMinusExpr is a `-$Expr` expression.
type UnaryMinusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	MinusTkn     *Token
	Expr         Node
}

// UnaryPlusExpr is a `+$Expr` expression.
type UnaryPlusExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	PlusTkn      *Token
	Expr         Node
}

// YieldExpr is a `yield $Key => $Value` expression.
// If $Key is nil, it's `yield $Value`.
type YieldExpr struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	YieldTkn       *Token
	Key            Node
	DoubleArrowTkn *Token
	Value          Node
}

// YieldFromExpr is a `yield from $Expr` expression.
type YieldFromExpr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	YieldFromTkn *Token
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
	VariadicTkn  *Token
	AmpersandTkn *Token
	Expr         Node
	Variadic     bool
	IsReference  bool
}

// Identifier is like a name, but it's always resolved to itself.
// Identifier always consists of a single part.
type Identifier struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	IdentifierTkn *Token
	Value         string
}

// Nullable is a `?$Expr` expression.
type Nullable struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	QuestionTkn  *Token
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
	AmpersandTkn *Token
	VariadicTkn  *Token
	Variable     *SimpleVar
	EqualTkn     *Token
	DefaultValue Node
	ByRef        bool
	Variadic     bool
}

// Root is a node that wraps all file statements.
type Root struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
	EndTkn       *Token
}

// SimpleVar is a normal PHP variable like `$foo` or `$bar`.
type SimpleVar struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	DollarTkn    *Token
	Name         string
}

// Var is variable variable expression like `$$foo` or `${"foo"}`.
type Var struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	DollarTkn            *Token
	OpenCurlyBracketTkn  *Token
	Expr                 Node
	CloseCurlyBracketTkn *Token
}

// Dnumber is a floating point literal.
type Dnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	NumberTkn    *Token
	Value        string
}

// Encapsed is a string literal with interpolated parts.
type Encapsed struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	OpenQuoteTkn  *Token
	Parts         []Node
	CloseQuoteTkn *Token
}

// EncapsedStringPart is a part of the Encapsed literal.
type EncapsedStringPart struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	EncapsedStrTkn *Token
	Value          string
}

// Heredoc is special PHP literal.
// Note that it may be a nowdoc, depending on the label.
type Heredoc struct {
	FreeFloating    freefloating.Collection
	Position        *position.Position
	Label           string
	OpenHeredocTkn  *Token
	Parts           []Node
	CloseHeredocTkn *Token
}

// Lnumber is an integer literal.
type Lnumber struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	NumberTkn    *Token
	Value        string
}

// MagicConstant is a special PHP constant like __FILE__ or __CLASS__.
// TODO: do we really need a separate node for these constants?
type MagicConstant struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	MagicConstTkn *Token
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
	MinusTkn     *Token
	StringTkn    *Token
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
	MinusTkn     *Token
	StringTkn    *Token
	Value        string
	DoubleQuotes bool
	Error        string
}

// BreakStmt is a `break $Expr` statement.
type BreakStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	BreakTkn     *Token
	Expr         Node
	SemiColonTkn *Token
}

// CaseStmt is a `case $Cond: $Stmts...` statement.
type CaseStmt struct {
	FreeFloating     freefloating.Collection
	Position         *position.Position
	CaseTkn          *Token
	Cond             Node
	CaseSeparatorTkn *Token
	Stmts            []Node
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
	FreeFloating         freefloating.Collection
	Position             *position.Position
	CatchTkn             *Token
	OpenParenthesisTkn   *Token
	Types                []Node
	SeparatorTkns        []*Token
	Variable             *SimpleVar
	CloseParenthesisTkn  *Token
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
}

// ClassStmt is a named class declaration.
// $Modifiers consist of identifiers like `final` and `abstract`.
type ClassStmt struct {
	FreeFloating            freefloating.Collection
	Position                *position.Position
	Modifiers               []*Identifier
	ClassTkn                *Token
	ClassName               *Identifier
	ExtendsTkn              *Token
	ImplementsTkn           *Token
	ImplementsSeparatorTkns []*Token
	OpenCurlyBracketTkn     *Token
	CloseCurlyBracketTkn    *Token
	Class
}

// ClassConstListStmt is a `$Modifiers... const $Consts...` statement.
// $Modifiers may specify the constant access level.
// Every element in $Consts is a *ConstantStmt.
type ClassConstListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	Modifiers     []*Identifier
	ConstTkn      *Token
	Consts        []Node
	SeparatorTkns []*Token
	SemiColonTkn  *Token
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
	FreeFloating        freefloating.Collection
	Position            *position.Position
	Modifiers           []*Identifier
	FunctionTkn         *Token
	AmpersandTkn        *Token
	MethodName          *Identifier
	OpenParenthesisTkn  *Token
	Params              []Node
	SeparatorTkns       []*Token
	CloseParenthesisTkn *Token
	ColonTkn            *Token
	ReturnType          Node
	Stmt                Node
	ReturnsRef          bool
	PhpDocComment       string
	PhpDoc              []phpdoc.CommentPart
}

// ConstListStmt is a `const $Consts` statement.
// Every element in $Consts is a *ConstantStmt.
type ConstListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ConstTkn      *Token
	Consts        []Node
	SeparatorTkns []*Token
	SemiColonTkn  *Token
}

// ConstantStmt is a `$ConstantName = $Expr` statement.
// It's a part of the *ConstListStmt, *ClassConstListStmt and *DeclareStmt.
type ConstantStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ConstantName  *Identifier
	EqualTkn      *Token
	Expr          Node
	PhpDocComment string
}

// ContinueStmt is a `continue $Expe` statement.
type ContinueStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ContinueTkn  *Token
	Expr         Node
	SemiColonTkn *Token
}

// DeclareStmt is a `declare ($Consts...) $Stmt` statement.
// $Stmt can be an empty statement, like in `declare ($Consts...);`,
// but it can also be a block like in `declare ($Consts...) {}`.
// If $Alt is true, the block will begin with `:` and end with `enddeclare`.
// Every element in $Consts is a *ConstantStmt.
type DeclareStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	DeclareTkn          *Token
	OpenParenthesisTkn  *Token
	Consts              []Node
	SeparatorTkns       []*Token
	CloseParenthesisTkn *Token
	ColonTkn            *Token
	Stmt                Node
	EndDeclareTkn       *Token
	SemiColonTkn        *Token
	Alt                 bool
}

// DefaultStmt is a `default: $Stmts...` statement.
type DefaultStmt struct {
	FreeFloating     freefloating.Collection
	Position         *position.Position
	DefaultTkn       *Token
	CaseSeparatorTkn *Token
	Stmts            []Node
}

// DoStmt is a `do $Stmt while ($Cond)` statement.
type DoStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	DoTkn               *Token
	Stmt                Node
	WhileTkn            *Token
	OpenParenthesisTkn  *Token
	Cond                Node
	CloseParenthesisTkn *Token
	SemiColonTkn        *Token
}

// EchoStmt is a `echo $Exprs...` statement.
type EchoStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	EchoTkn       *Token
	Exprs         []Node
	SeparatorTkns []*Token
	SemiColonTkn  *Token
}

// ElseStmt is a `else $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:`.
type ElseStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ElseTkn      *Token
	ColonTkn     *Token
	Stmt         Node
	AltSyntax    bool
}

// ElseIfStmt is a `elseif ($Cond) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endif`.
// $Merged tells whether this elseif is a merged `else if` statement.
type ElseIfStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ElseIfTkn           *Token
	OpenParenthesisTkn  *Token
	Cond                Node
	CloseParenthesisTkn *Token
	ColonTkn            *Token
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
	SemiColonTkn *Token
}

// FinallyStmt is a `finally { $Stmts... }` statement.
type FinallyStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	FinallyTkn           *Token
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
}

// ForStmt is a `for ($Init; $Cond; $Loop) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endfor`.
type ForStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ForTkn              *Token
	OpenParenthesisTkn  *Token
	Init                []Node
	InitSeparatorTkns   []*Token
	InitSemiColonTkn    *Token
	Cond                []Node
	CondSeparatorTkns   []*Token
	CondSemiColonTkn    *Token
	Loop                []Node
	LoopSeparatorTkns   []*Token
	CloseParenthesisTkn *Token
	ColonTkn            *Token
	Stmt                Node
	EndForTkn           *Token
	SemiColonTkn        *Token
	AltSyntax           bool
}

// ForeachStmt is a `foreach ($Expr as $Key => $Variable) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endforeach`.
type ForeachStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	ForeachTkn          *Token
	OpenParenthesisTkn  *Token
	Expr                Node
	AsTkn               *Token
	Key                 Node
	DoubleArrowTkn      *Token
	AmpersandTkn        *Token
	Variable            Node
	CloseParenthesisTkn *Token
	ColonTkn            *Token
	Stmt                Node
	EndForeachTkn       *Token
	SemiColonTkn        *Token
	AltSyntax           bool
}

// FunctionStmt is a named function declaration.
type FunctionStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	FunctionTkn          *Token
	AmpersandTkn         *Token
	FunctionName         *Identifier
	OpenParenthesisTkn   *Token
	Params               []Node
	SeparatorTkns        []*Token
	CloseParenthesisTkn  *Token
	ColonTkn             *Token
	ReturnType           Node
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
	ReturnsRef           bool
	PhpDocComment        string
	PhpDoc               []phpdoc.CommentPart
}

// GlobalStmt is a `global $Vars` statement.
type GlobalStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	GlobalTkn     *Token
	Vars          []Node
	SeparatorTkns []*Token
	SemiColonTkn  *Token
}

// GotoStmt is a `goto $Label` statement.
type GotoStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	GotoTkn      *Token
	Label        *Identifier
	SemiColonTkn *Token
}

// GroupUseStmt is a `use $UseType $Prefix\{ $UseList }` statement.
// $UseType is a "function" or "const".
// TODO: change $UseType type to *Identifier?
type GroupUseStmt struct {
	FreeFloating          freefloating.Collection
	Position              *position.Position
	UseTkn                *Token
	UseType               *Identifier
	LeadingNsSeparatorTkn *Token
	Prefix                *Name
	NsSeparatorTkn        *Token
	OpenCurlyBracketTkn   *Token
	UseList               []Node
	SeparatorTkns         []*Token
	CloseCurlyBracketTkn  *Token
	SemiColonTkn          *Token
}

// HaltCompilerStmt is a `__halt_compiler()` statement.
type HaltCompilerStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	HaltCompilerTkn     *Token
	OpenParenthesisTkn  *Token
	CloseParenthesisTkn *Token
	SemiColonTkn        *Token
}

// IfStmt is a `if ($Cond) $Stmt` statement.
// $ElseIf contains an entire elseif chain, if any.
// $Else may contain an else part of the statement.
type IfStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	IfTkn               *Token
	OpenParenthesisTkn  *Token
	Cond                Node
	CloseParenthesisTkn *Token
	ColonTkn            *Token
	Stmt                Node
	ElseIf              []Node
	Else                Node
	EndIfTkn            *Token
	SemiColonTkn        *Token
	AltSyntax           bool
}

// InlineHTMLStmt is a part of the script that will not be interpreted
// as a PHP script. In other words, it's everything outside of
// the <? and ?> tags.
type InlineHTMLStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	InlineHTMLTkn *Token
	Value         string
}

// InterfaceStmt is an interface declaration.
type InterfaceStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	PhpDocComment        string
	InterfaceTkn         *Token
	InterfaceName        *Identifier
	ExtendsTkn           *Token
	Extends              *InterfaceExtendsStmt
	ExtendsSeparatorTkns []*Token
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
}

// InterfaceExtendsStmt is a `extends $InterfaceNames...` statement.
// TODO: do we need this wrapper node?
// TODO: InterfaceNames could be a []*Name.
type InterfaceExtendsStmt struct {
	FreeFloating   freefloating.Collection
	Token          *Token
	Position       *position.Position
	InterfaceNames []Node
}

// LabelStmt is a `$LabelName:` statement.
type LabelStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	LabelName    *Identifier
	ColonTkn     *Token
}

// NamespaceStmt is a `namespace $NamespaceName` statement.
// If $Stmts is not nil, it's `namespace $NamespaceName { $Stmts... }`.
type NamespaceStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	NsTkn                *Token
	NamespaceName        *Name
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
	SemiColonTkn         *Token
}

// NopStmt is a `;` statement.
// It's also known as "empty statement".
// It could also be a `?>` (script closing marker).
type NopStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	SemiColonTkn *Token
}

// PropertyStmt is a `$Variable = $Expr` statement.
// It's a part of the *PropertyListStmt.
type PropertyStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	Variable      *SimpleVar
	EqualTkn      *Token
	Expr          Node
	PhpDocComment string
	PhpDoc        []phpdoc.CommentPart
}

// PropertyListStmt is a `$Modifiers $Type $Properties` statement.
// Every element in $Properties is a *PropertyStmt.
type PropertyListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	Modifiers     []*Identifier
	Type          Node
	Properties    []Node
	SeparatorTkns []*Token
	SemiColonTkn  *Token
}

// ReturnStmt is a `return $Expr` statement.
type ReturnStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ReturnTkn    *Token
	Expr         Node
	SemiColonTkn *Token
}

// StaticStmt is a `static $Vars...` statement.
// Every element in $Vars is a *StaticVarStmt.
type StaticStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	StaticTkn     *Token
	Vars          []Node
	SeparatorTkns []*Token
	SemiColonTkn  *Token
}

// StaticVarStmt is a `$Variable = $Expr`.
// It's a part of the *StaticStmt.
type StaticVarStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     *SimpleVar
	EqualTkn     *Token
	Expr         Node
}

// StmtList is a `{ $Stmts... }` statement.
// It's also known as "block statement".
type StmtList struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
}

// SwitchStmt is a `switch ($Cond) $CaseList` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endswitch`.
type SwitchStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	SwitchTkn            *Token
	OpenParenthesisTkn   *Token
	Cond                 Node
	CloseParenthesisTkn  *Token
	ColonTkn             *Token
	OpenCurlyBracketTkn  *Token
	CaseSeparatorTkn     *Token
	CaseList             *CaseListStmt
	CloseCurlyBracketTkn *Token
	EndSwitchTkn         *Token
	SemiColonTkn         *Token
	AltSyntax            bool
}

// ThrowStmt is a `throw $Expr` statement.
type ThrowStmt struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ThrowTkn     *Token
	Expr         Node
	SemiColonTkn *Token
}

// TraitStmt is a trait declaration.
type TraitStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	TraitTkn             *Token
	TraitName            *Identifier
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
	PhpDocComment        string
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
	UseTkn               *Token
	Traits               []Node
	SeparatorTkns        []*Token
	OpenCurlyBracketTkn  *Token
	TraitAdaptationList  Node
	CloseCurlyBracketTkn *Token
	SemiColonTkn         *Token
}

type TraitUseAliasStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	DoubleColonTkn *Token
	Ref            Node
	AsTkn          *Token
	Modifier       Node
	Alias          *Identifier
	SemiColonTkn   *Token
}

type TraitUsePrecedenceStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	DoubleColonTkn *Token
	Ref            Node
	InsteadofTkn   *Token
	Insteadof      []Node
	SeparatorTkns  []*Token
	SemiColonTkn   *Token
}

// TryStmt is a `try { $Stmts... } $Catches` statement.
// $Finally only presents if `finally {...}` block exists.
type TryStmt struct {
	FreeFloating         freefloating.Collection
	Position             *position.Position
	TryTkn               *Token
	OpenCurlyBracketTkn  *Token
	Stmts                []Node
	CloseCurlyBracketTkn *Token
	Catches              []Node
	Finally              Node
}

// UnsetStmt is a `unset($Vars...)` statement.
type UnsetStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	UnsetTkn            *Token
	OpenParenthesisTkn  *Token
	Vars                []Node
	SeparatorTkns       []*Token
	CloseParenthesisTkn *Token
	SemiColonTkn        *Token
}

type UseStmt struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	UseType        *Identifier
	NsSeparatorTkn *Token
	Use            *Name
	AsTkn          *Token
	Alias          *Identifier
}

type UseListStmt struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	UseTkn        *Token
	UseType       *Identifier
	Uses          []Node
	SeparatorTkns []*Token
	SemiColonTkn  *Token
}

// WhileStmt is a `while ($Cond) $Stmt` statement.
// If $AltSyntax is true, the block will begin with `:` and end with `endwhile`.
type WhileStmt struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	WhileTkn            *Token
	OpenParenthesisTkn  *Token
	Cond                Node
	CloseParenthesisTkn *Token
	ColonTkn            *Token
	Stmt                Node
	EndWhileTkn         *Token
	SemiColonTkn        *Token
	AltSyntax           bool
}
