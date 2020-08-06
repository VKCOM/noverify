package phpgrep

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/cast"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
)

type matcherState struct {
	literalMatch bool

	capture []CapturedNode
}

type matcher struct {
	// root is a compiled pattern node.
	root node.Node

	// numVars is a max number of named captures this matcher may need.
	numVars int

	// Used only when -tracing build tag is specified.
	tracingBuf   *bytes.Buffer
	tracingDepth int
}

func (m *matcher) match(state *matcherState, n node.Node) (data MatchData, ok bool) {
	state.capture = state.capture[:0]
	if !m.eqNode(state, m.root, n) {
		return data, false
	}
	pos := getNodePos(n)
	if pos == nil {
		return data, false
	}
	data.Node = n
	data.Capture = state.capture
	return data, true
}

func (m *matcher) eqNameParts(xs, ys []node.Node) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i, p1 := range xs {
		p1 := p1.(*name.NamePart).Value
		p2 := ys[i].(*name.NamePart).Value
		if p1 != p2 {
			return false
		}
	}
	return true
}

func (m *matcher) eqNodeSliceNoMeta(state *matcherState, xs, ys []node.Node) bool {
	if len(xs) != len(ys) {
		return false
	}

	for i, x := range xs {
		if !m.eqNode(state, x, ys[i]) {
			return false
		}
	}

	return true
}

func (m *matcher) eqArrayItemSlice(state *matcherState, xs, ys []*expr.ArrayItem) bool {
	// FIXME.

	if len(xs) == 0 && len(ys) != 0 {
		return false
	}

	matchAny := false

	i := 0
	for i < len(xs) {
		x := xs[i]

		if matchMetaVar(x, "*") {
			matchAny = true
		}

		if matchAny {
			switch {
			// "Nothing left to match" stop condition.
			case len(ys) == 0:
				matchAny = false
				i++
			// Lookahead for non-greedy matching.
			case i+1 < len(xs) && m.eqNode(state, xs[i+1], ys[0]):
				matchAny = false
				i += 2
				ys = ys[1:]
			default:
				ys = ys[1:]
			}
			continue
		}

		if len(ys) == 0 || !m.eqNode(state, x, ys[0]) {
			return false
		}
		i++
		ys = ys[1:]
	}

	return len(ys) == 0
}

func (m *matcher) eqNodeSlice(state *matcherState, xs, ys []node.Node) bool {
	if len(xs) == 0 && len(ys) != 0 {
		return false
	}

	matchAny := false

	i := 0
	for i < len(xs) {
		x := xs[i]

		if matchMetaVar(x, "*") {
			matchAny = true
		}

		if matchAny {
			switch {
			// "Nothing left to match" stop condition.
			case len(ys) == 0:
				matchAny = false
				i++
			// Lookahead for non-greedy matching.
			case i+1 < len(xs) && m.eqNode(state, xs[i+1], ys[0]):
				matchAny = false
				i += 2
				ys = ys[1:]
			default:
				ys = ys[1:]
			}
			continue
		}

		if len(ys) == 0 || !m.eqNode(state, x, ys[0]) {
			return false
		}
		i++
		ys = ys[1:]
	}

	return len(ys) == 0
}

func (m *matcher) eqEncapsedStringPartSlice(state *matcherState, xs, ys []node.Node) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i, x := range xs {
		if !m.eqEncapsedStringPart(state, x, ys[i]) {
			return false
		}
	}
	return true
}

func (m *matcher) eqEncapsedStringPart(state *matcherState, x, y node.Node) bool {
	switch x := x.(type) {
	case *scalar.EncapsedStringPart:
		y, ok := y.(*scalar.EncapsedStringPart)
		return ok && x.Value == y.Value
	case *node.SimpleVar:
		// Match variables literally.
		y, ok := y.(*node.SimpleVar)
		return ok && x.Name == y.Name
	default:
		return m.eqNode(state, x, y)
	}
}

func (m *matcher) eqNode(state *matcherState, x, y node.Node) bool {
	if tracingEnabled && m.tracingBuf != nil {
		pad := strings.Repeat(" • ", m.tracingDepth)
		fmt.Fprintf(m.tracingBuf, "%seqNode x=%T y=%T\n", pad, x, y)
		m.tracingDepth++
		defer func() {
			m.tracingDepth--
		}()
	}

	if x == y {
		return true
	}

	switch x := x.(type) {
	case nil:
		return y == nil

	case *stmt.Expression:
		// To make it possible to match statements with $-expressions,
		// check whether expression inside x.Expr is a variable.
		if x, ok := x.Expr.(*node.SimpleVar); ok {
			return m.eqSimpleVar(state, x, y)
		}
		y, ok := y.(*stmt.Expression)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *expr.Paren:
		y, ok := y.(*expr.Paren)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *stmt.StmtList:
		y, ok := y.(*stmt.StmtList)
		return ok && m.eqNodeSlice(state, x.Stmts, y.Stmts)

	case *stmt.Function:
		return false // FIXME #23
	case *stmt.Interface:
		return false // FIXME #23
	case *stmt.Class:
		return false // FIXME #23
	case *stmt.Trait:
		return false // FIXME #23

	case *stmt.InlineHtml:
		y, ok := y.(*stmt.InlineHtml)
		return ok && x.Value == y.Value
	case *stmt.StaticVar:
		y, ok := y.(*stmt.StaticVar)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expr, y.Expr)
	case *stmt.Static:
		y, ok := y.(*stmt.Static)
		return ok && m.eqNodeSlice(state, x.Vars, y.Vars)
	case *stmt.Global:
		y, ok := y.(*stmt.Global)
		return ok && m.eqNodeSlice(state, x.Vars, y.Vars)
	case *stmt.Break:
		y, ok := y.(*stmt.Break)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *stmt.Continue:
		y, ok := y.(*stmt.Continue)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *stmt.Unset:
		y, ok := y.(*stmt.Unset)
		return ok && m.eqNodeSlice(state, x.Vars, y.Vars)
	case *expr.Print:
		y, ok := y.(*expr.Print)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *stmt.Echo:
		y, ok := y.(*stmt.Echo)
		return ok && m.eqNodeSlice(state, x.Exprs, y.Exprs)
	case *stmt.Nop:
		_, ok := y.(*stmt.Nop)
		return ok
	case *stmt.Do:
		y, ok := y.(*stmt.Do)
		return ok && m.eqNode(state, x.Cond, y.Cond) && m.eqNode(state, x.Stmt, y.Stmt)
	case *stmt.While:
		y, ok := y.(*stmt.While)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Cond, y.Cond) && m.eqNode(state, x.Stmt, y.Stmt)
	case *stmt.For:
		y, ok := y.(*stmt.For)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNodeSlice(state, x.Init, y.Init) &&
			m.eqNodeSlice(state, x.Cond, y.Cond) &&
			m.eqNodeSlice(state, x.Loop, y.Loop) &&
			m.eqNode(state, x.Stmt, y.Stmt)
	case *stmt.Foreach:
		y, ok := y.(*stmt.Foreach)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Expr, y.Expr) &&
			m.eqNode(state, x.Key, y.Key) &&
			m.eqNode(state, x.Variable, y.Variable) &&
			m.eqNode(state, x.Stmt, y.Stmt)

	case *stmt.Else:
		y, ok := y.(*stmt.Else)
		return ok && x.AltSyntax == y.AltSyntax && m.eqNode(state, x.Stmt, y.Stmt)
	case *stmt.ElseIf:
		y, ok := y.(*stmt.ElseIf)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Cond, y.Cond) && m.eqNode(state, x.Stmt, y.Stmt)
	case *stmt.If:
		y, ok := y.(*stmt.If)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNodeSliceNoMeta(state, x.ElseIf, y.ElseIf) &&
			m.eqNode(state, x.Cond, y.Cond) &&
			m.eqNode(state, x.Stmt, y.Stmt) &&
			m.eqNode(state, x.Else, y.Else)

	case *stmt.Throw:
		y, ok := y.(*stmt.Throw)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *stmt.Try:
		y, ok := y.(*stmt.Try)
		return ok && m.eqNodeSlice(state, x.Stmts, y.Stmts) &&
			m.eqNodeSlice(state, x.Catches, y.Catches) &&
			m.eqNode(state, x.Finally, y.Finally)

	case *expr.Yield:
		y, ok := y.(*expr.Yield)
		return ok && m.eqNode(state, x.Key, y.Key) && m.eqNode(state, x.Value, y.Value)
	case *expr.YieldFrom:
		y, ok := y.(*expr.YieldFrom)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *expr.InstanceOf:
		y, ok := y.(*expr.InstanceOf)
		return ok && m.eqNode(state, x.Expr, y.Expr) && m.eqNode(state, x.Class, y.Class)

	case *expr.List:
		y, ok := y.(*expr.List)
		return ok && x.ShortSyntax == y.ShortSyntax &&
			m.eqArrayItemSlice(state, x.Items, y.Items)

	case *expr.New:
		y, ok := y.(*expr.New)
		if !ok || !m.eqNode(state, x.Class, y.Class) {
			return false
		}
		if x.ArgumentList == nil || y.ArgumentList == nil {
			return x.ArgumentList == y.ArgumentList
		}
		return m.eqNodeSlice(state, x.ArgumentList.Arguments, y.ArgumentList.Arguments)

	case *stmt.Case:
		y, ok := y.(*stmt.Case)
		return ok && m.eqNode(state, x.Cond, y.Cond) && m.eqNodeSlice(state, x.Stmts, y.Stmts)
	case *stmt.Default:
		y, ok := y.(*stmt.Default)
		return ok && m.eqNodeSlice(state, x.Stmts, y.Stmts)
	case *stmt.Switch:
		y, ok := y.(*stmt.Switch)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Cond, y.Cond) &&
			m.eqNodeSlice(state, x.CaseList.Cases, y.CaseList.Cases)

	case *stmt.Return:
		y, ok := y.(*stmt.Return)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *assign.Assign:
		y, ok := y.(*assign.Assign)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Plus:
		y, ok := y.(*assign.Plus)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Reference:
		y, ok := y.(*assign.Reference)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.BitwiseAnd:
		y, ok := y.(*assign.BitwiseAnd)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.BitwiseOr:
		y, ok := y.(*assign.BitwiseOr)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.BitwiseXor:
		y, ok := y.(*assign.BitwiseXor)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Concat:
		y, ok := y.(*assign.Concat)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Coalesce:
		y, ok := y.(*assign.Coalesce)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Div:
		y, ok := y.(*assign.Div)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Minus:
		y, ok := y.(*assign.Minus)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Mod:
		y, ok := y.(*assign.Mod)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Mul:
		y, ok := y.(*assign.Mul)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.Pow:
		y, ok := y.(*assign.Pow)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.ShiftLeft:
		y, ok := y.(*assign.ShiftLeft)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *assign.ShiftRight:
		y, ok := y.(*assign.ShiftRight)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)

	case *expr.ArrayDimFetch:
		y, ok := y.(*expr.ArrayDimFetch)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Dim, y.Dim)
	case *expr.ArrayItem:
		y, ok := y.(*expr.ArrayItem)
		if !ok {
			return false
		}
		if x.Key == nil || y.Key == nil {
			return x.Key == y.Key && m.eqNode(state, x.Val, y.Val)
		}
		return m.eqNode(state, x.Key, y.Key) && m.eqNode(state, x.Val, y.Val)
	case *expr.Array:
		y, ok := y.(*expr.Array)
		return ok && x.ShortSyntax == y.ShortSyntax &&
			m.eqArrayItemSlice(state, x.Items, y.Items)

	case *node.Argument:
		y, ok := y.(*node.Argument)
		return ok && x.IsReference == y.IsReference &&
			x.Variadic == y.Variadic &&
			m.eqNode(state, x.Expr, y.Expr)
	case *expr.FunctionCall:
		y, ok := y.(*expr.FunctionCall)
		if !ok || !m.eqNode(state, x.Function, y.Function) {
			return false
		}
		return m.eqNodeSlice(state, x.ArgumentList.Arguments, y.ArgumentList.Arguments)

	case *expr.PostInc:
		y, ok := y.(*expr.PostInc)
		return ok && m.eqNode(state, x.Variable, y.Variable)
	case *expr.PostDec:
		y, ok := y.(*expr.PostDec)
		return ok && m.eqNode(state, x.Variable, y.Variable)
	case *expr.PreInc:
		y, ok := y.(*expr.PreInc)
		return ok && m.eqNode(state, x.Variable, y.Variable)
	case *expr.PreDec:
		y, ok := y.(*expr.PreDec)
		return ok && m.eqNode(state, x.Variable, y.Variable)

	case *expr.Exit:
		y, ok := y.(*expr.Exit)
		return ok && x.Die == y.Die && m.eqNode(state, x.Expr, y.Expr)

	case *expr.Include:
		y, ok := y.(*expr.Include)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.IncludeOnce:
		y, ok := y.(*expr.IncludeOnce)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.Require:
		y, ok := y.(*expr.Require)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.RequireOnce:
		y, ok := y.(*expr.RequireOnce)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.Empty:
		y, ok := y.(*expr.Empty)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.Eval:
		y, ok := y.(*expr.Eval)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.ErrorSuppress:
		y, ok := y.(*expr.ErrorSuppress)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.Clone:
		y, ok := y.(*expr.Clone)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.BitwiseNot:
		y, ok := y.(*expr.BitwiseNot)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.BooleanNot:
		y, ok := y.(*expr.BooleanNot)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.UnaryMinus:
		y, ok := y.(*expr.UnaryMinus)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *expr.UnaryPlus:
		y, ok := y.(*expr.UnaryPlus)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *expr.StaticPropertyFetch:
		switch y := y.(type) {
		case *expr.StaticPropertyFetch:
			return m.eqNode(state, x.Class, y.Class) &&
				m.eqNode(state, x.Property, y.Property)
		case *expr.ClassConstFetch:
			return nodeIsVar(x.Property) &&
				m.eqNode(state, x.Class, y.Class) &&
				m.eqNode(state, x.Property, y.ConstantName)
		default:
			return false
		}

	case *expr.ClassConstFetch:
		y, ok := y.(*expr.ClassConstFetch)
		return ok && m.eqNode(state, x.Class, y.Class) && m.eqNode(state, x.ConstantName, y.ConstantName)
	case *expr.StaticCall:
		y, ok := y.(*expr.StaticCall)
		return ok &&
			m.eqNode(state, x.Class, y.Class) &&
			m.eqNode(state, x.Call, y.Call) &&
			m.eqNodeSlice(state, x.ArgumentList.Arguments, y.ArgumentList.Arguments)

	case *expr.ShellExec:
		y, ok := y.(*expr.ShellExec)
		return ok && m.eqEncapsedStringPartSlice(state, x.Parts, y.Parts)
	case *scalar.Encapsed:
		y, ok := y.(*scalar.Encapsed)
		return ok && m.eqEncapsedStringPartSlice(state, x.Parts, y.Parts)

	case *scalar.Heredoc:
		y, ok := y.(*scalar.Heredoc)
		return ok && x.Label == y.Label && m.eqEncapsedStringPartSlice(state, x.Parts, y.Parts)
	case *scalar.MagicConstant:
		y, ok := y.(*scalar.MagicConstant)
		return ok && y.Value == x.Value
	case *scalar.Lnumber:
		y, ok := y.(*scalar.Lnumber)
		return ok && y.Value == x.Value
	case *scalar.Dnumber:
		y, ok := y.(*scalar.Dnumber)
		return ok && y.Value == x.Value
	case *scalar.String:
		y, ok := y.(*scalar.String)
		return ok && y.Value == x.Value

	case *binary.Coalesce:
		y, ok := y.(*binary.Coalesce)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Concat:
		y, ok := y.(*binary.Concat)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Div:
		y, ok := y.(*binary.Div)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Mod:
		y, ok := y.(*binary.Mod)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Mul:
		y, ok := y.(*binary.Mul)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Pow:
		y, ok := y.(*binary.Pow)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.BitwiseAnd:
		y, ok := y.(*binary.BitwiseAnd)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.BitwiseOr:
		y, ok := y.(*binary.BitwiseOr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.BitwiseXor:
		y, ok := y.(*binary.BitwiseXor)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.ShiftLeft:
		y, ok := y.(*binary.ShiftLeft)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.ShiftRight:
		y, ok := y.(*binary.ShiftRight)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.LogicalAnd:
		y, ok := y.(*binary.LogicalAnd)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.LogicalOr:
		y, ok := y.(*binary.LogicalOr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.LogicalXor:
		y, ok := y.(*binary.LogicalXor)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.BooleanAnd:
		y, ok := y.(*binary.BooleanAnd)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.BooleanOr:
		y, ok := y.(*binary.BooleanOr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.NotEqual:
		y, ok := y.(*binary.NotEqual)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.NotIdentical:
		y, ok := y.(*binary.NotIdentical)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Equal:
		y, ok := y.(*binary.Equal)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Identical:
		y, ok := y.(*binary.Identical)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Greater:
		y, ok := y.(*binary.Greater)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.GreaterOrEqual:
		y, ok := y.(*binary.GreaterOrEqual)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Smaller:
		y, ok := y.(*binary.Smaller)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.SmallerOrEqual:
		y, ok := y.(*binary.SmallerOrEqual)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Spaceship:
		y, ok := y.(*binary.Spaceship)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Plus:
		y, ok := y.(*binary.Plus)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *binary.Minus:
		y, ok := y.(*binary.Minus)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)

	case *expr.ConstFetch:
		y, ok := y.(*expr.ConstFetch)
		return ok && m.eqNode(state, x.Constant, y.Constant)
	case *name.Name:
		y, ok := y.(*name.Name)
		return ok && m.eqNameParts(x.Parts, y.Parts)
	case *name.FullyQualified:
		y, ok := y.(*name.FullyQualified)
		return ok && m.eqNameParts(x.Parts, y.Parts)
	case *node.Identifier:
		y, ok := y.(*node.Identifier)
		return ok && x.Value == y.Value
	case *node.SimpleVar:
		return m.eqSimpleVar(state, x, y)
	case *node.Var:
		return m.eqVar(state, x, y)

	case *expr.Reference:
		y, ok := y.(*expr.Reference)
		return ok && m.eqNode(state, x.Variable, y.Variable)

	case *node.Parameter:
		y, ok := y.(*node.Parameter)
		return ok && x.ByRef == y.ByRef &&
			x.Variadic == y.Variadic &&
			m.eqNode(state, x.VariableType, y.VariableType) &&
			m.eqNode(state, x.Variable, y.Variable) &&
			m.eqNode(state, x.DefaultValue, y.DefaultValue)
	case *expr.Closure:
		return m.eqClosure(state, x, y)

	case *expr.Ternary:
		return m.eqTernary(state, x, y)

	case *expr.Isset:
		y, ok := y.(*expr.Isset)
		return ok && m.eqNodeSlice(state, x.Variables, y.Variables)

	case *expr.PropertyFetch:
		y, ok := y.(*expr.PropertyFetch)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Property, y.Property)
	case *expr.MethodCall:
		y, ok := y.(*expr.MethodCall)
		return ok && m.eqNode(state, x.Variable, y.Variable) &&
			m.eqNode(state, x.Method, y.Method) &&
			m.eqNodeSlice(state, x.ArgumentList.Arguments, y.ArgumentList.Arguments)

	case *cast.Double:
		y, ok := y.(*cast.Double)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *cast.Array:
		y, ok := y.(*cast.Array)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *cast.Bool:
		y, ok := y.(*cast.Bool)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *cast.Int:
		y, ok := y.(*cast.Int)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *cast.Object:
		y, ok := y.(*cast.Object)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *cast.String:
		y, ok := y.(*cast.String)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *node.Root:
		return false

	default:
		panic(fmt.Sprintf("unhandled node: x=%T y=%T (please, fill an issue on GitHub)\n", x, y))
	}
}

func (m *matcher) matchNamed(state *matcherState, name string, y node.Node) bool {
	// Special case.
	// "_" name matches anything, always.
	// Anonymous names replaced with "_" during the compilation.
	if name == "_" {
		return true
	}

	z, ok := findNamed(state.capture, name)
	if !ok {
		// We allocate capture slice lazily and to the max capacity.
		if state.capture == nil {
			state.capture = make([]CapturedNode, 0, m.numVars)
		}
		state.capture = append(state.capture, CapturedNode{Name: name, Node: y})
		return true
	}
	if z == nil {
		return y == nil
	}

	state.literalMatch = true
	result := m.eqNode(state, z, y)
	state.literalMatch = false
	return result
}

func (m *matcher) eqTernary(state *matcherState, x *expr.Ternary, y node.Node) bool {
	if y, ok := y.(*expr.Ternary); ok {
		// To avoid matching `$x ?: $y` with `$x ? $y : $z` pattern.
		if x.IfTrue == nil || y.IfTrue == nil {
			return y.IfTrue == x.IfTrue &&
				m.eqNode(state, x.Condition, y.Condition) &&
				m.eqNode(state, x.IfFalse, y.IfFalse)
		}
		return m.eqNode(state, x.Condition, y.Condition) &&
			m.eqNode(state, x.IfTrue, y.IfTrue) &&
			m.eqNode(state, x.IfFalse, y.IfFalse)
	}

	return false
}

func (m *matcher) eqClosure(state *matcherState, x *expr.Closure, y node.Node) bool {
	if y, ok := y.(*expr.Closure); ok {
		var xUses, yUses []node.Node
		if x.ClosureUse != nil {
			xUses = x.ClosureUse.Uses
		}
		if y.ClosureUse != nil {
			yUses = y.ClosureUse.Uses
		}
		return ok && x.ReturnsRef == y.ReturnsRef &&
			x.Static == y.Static &&
			m.eqNodeSlice(state, x.Params, y.Params) &&
			m.eqNode(state, x.ReturnType, y.ReturnType) &&
			m.eqNodeSlice(state, x.Stmts, y.Stmts) &&
			m.eqNodeSlice(state, xUses, yUses)
	}

	return false
}

func (m *matcher) eqSimpleVar(state *matcherState, x *node.SimpleVar, y node.Node) bool {
	if state.literalMatch {
		y, ok := y.(*node.SimpleVar)
		return ok && x.Name == y.Name
	}
	return m.matchNamed(state, x.Name, y)
}

func (m *matcher) eqVar(state *matcherState, x *node.Var, y node.Node) bool {
	if state.literalMatch {
		y, ok := y.(*node.Var)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	}

	switch vn := x.Expr.(type) {
	case anyFunc:
		_, ok := y.(*expr.Closure)
		return ok && m.matchNamed(state, vn.name, y)
	case anyConst:
		switch y.(type) {
		case *expr.ConstFetch, *expr.ClassConstFetch:
			return m.matchNamed(state, vn.name, y)
		default:
			return false
		}
	case anyVar:
		switch y.(type) {
		case *node.SimpleVar, *node.Var:
			return m.matchNamed(state, vn.name, y)
		default:
			return false
		}
	case anyInt:
		_, ok := y.(*scalar.Lnumber)
		return ok && m.matchNamed(state, vn.name, y)
	case anyFloat:
		_, ok := y.(*scalar.Dnumber)
		return ok && m.matchNamed(state, vn.name, y)
	case anyStr:
		_, ok := y.(*scalar.String)
		return ok && m.matchNamed(state, vn.name, y)
	case anyNum:
		switch y.(type) {
		case *scalar.Lnumber, *scalar.Dnumber:
			return m.matchNamed(state, vn.name, y)
		default:
			return false
		}
	case anyExpr:
		return nodeIsExpr(y) && m.matchNamed(state, vn.name, y)
	}

	if y, ok := y.(*node.Var); ok {
		return m.eqNode(state, x.Expr, y.Expr)
	}
	return false
}
