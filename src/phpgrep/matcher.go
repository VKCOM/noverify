package phpgrep

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
)

type matcherState struct {
	literalMatch bool

	capture []CapturedNode
}

type matcher struct {
	// root is a compiled pattern node.
	root ir.Node

	// numVars is a max number of named captures this matcher may need.
	numVars int

	caseSensitive bool

	// Used only when -tracing build tag is specified.
	tracingBuf   *bytes.Buffer
	tracingDepth int
}

func (m *matcher) match(state *matcherState, n ir.Node) (data MatchData, ok bool) {
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

func (m *matcher) eqNodeSliceNoMeta(state *matcherState, xs, ys []ir.Node) bool {
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

func (m *matcher) eqArrayItemSlice(state *matcherState, xs, ys []*ir.ArrayItemExpr) bool {
	// FIXME.

	if len(xs) == 0 && len(ys) != 0 {
		return false
	}

	matchAny := false

	var backXS []*ir.ArrayItemExpr
	var backYS []*ir.ArrayItemExpr

	maybeBacktrack := func(matched bool) bool {
		if !matched && backXS != nil {
			if tracingEnabled && m.tracingBuf != nil {
				pad := strings.Repeat(" • ", m.tracingDepth)
				fmt.Fprintf(m.tracingBuf, "%sbacktrack!\n", pad)
			}
			return m.eqArrayItemSlice(state, backXS, backYS)
		}
		return matched
	}

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
				backXS = xs
				backYS = ys[1:]
				matchAny = false
				i += 2
				ys = ys[1:]
			default:
				ys = ys[1:]
			}
			continue
		}

		if len(ys) == 0 || !m.eqNode(state, x, ys[0]) {
			return maybeBacktrack(false)
		}
		i++
		ys = ys[1:]
	}

	return maybeBacktrack(len(ys) == 0)
}

func (m *matcher) eqNodeSlice(state *matcherState, xs, ys []ir.Node) bool {
	if len(xs) == 0 && len(ys) != 0 {
		return false
	}

	matchAny := false

	var backXS []ir.Node
	var backYS []ir.Node

	maybeBacktrack := func(matched bool) bool {
		if !matched && backXS != nil {
			if tracingEnabled && m.tracingBuf != nil {
				pad := strings.Repeat(" • ", m.tracingDepth)
				fmt.Fprintf(m.tracingBuf, "%sbacktrack!\n", pad)
			}
			return m.eqNodeSlice(state, backXS, backYS)
		}
		return matched
	}

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
				backXS = xs
				backYS = ys[1:]
				matchAny = false
				i += 2
				ys = ys[1:]
			default:
				ys = ys[1:]
			}
			continue
		}

		if len(ys) == 0 || !m.eqNode(state, x, ys[0]) {
			return maybeBacktrack(false)
		}
		i++
		ys = ys[1:]
	}

	return maybeBacktrack(len(ys) == 0)
}

func (m *matcher) eqEncapsedStringPartSlice(state *matcherState, xs, ys []ir.Node) bool {
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

func (m *matcher) eqEncapsedStringPart(state *matcherState, x, y ir.Node) bool {
	switch x := x.(type) {
	case *ir.EncapsedStringPart:
		y, ok := y.(*ir.EncapsedStringPart)
		return ok && x.Value == y.Value
	case *ir.SimpleVar:
		// Match variables literally.
		y, ok := y.(*ir.SimpleVar)
		return ok && x.Name == y.Name
	default:
		return m.eqNode(state, x, y)
	}
}

func (m *matcher) eqNodeWithCase(state *matcherState, x, y ir.Node) bool {
	switch x := x.(type) {
	case *ir.Name:
		y, ok := y.(*ir.Name)
		if !ok {
			return false
		}
		if m.caseSensitive {
			return x.Value == y.Value
		}
		return strings.EqualFold(x.Value, y.Value)

	case *ir.Identifier:
		y, ok := y.(*ir.Identifier)
		if !ok {
			return false
		}
		if m.caseSensitive {
			return x.Value == y.Value
		}
		return strings.EqualFold(x.Value, y.Value)

	default:
		return m.eqNode(state, x, y)
	}
}

func (m *matcher) eqNode(state *matcherState, x, y ir.Node) bool {
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

	case *ir.ExpressionStmt:
		// To make it possible to match statements with $-expressions,
		// check whether expression inside x.Expr is a variable.
		if x, ok := x.Expr.(*ir.SimpleVar); ok {
			return m.eqSimpleVar(state, x, y)
		}
		y, ok := y.(*ir.ExpressionStmt)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *ir.ParenExpr:
		y, ok := y.(*ir.ParenExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *ir.StmtList:
		y, ok := y.(*ir.StmtList)
		return ok && m.eqNodeSlice(state, x.Stmts, y.Stmts)

	case *ir.FunctionStmt:
		return false // FIXME #23
	case *ir.InterfaceStmt:
		return false // FIXME #23
	case *ir.ClassStmt:
		return false // FIXME #23
	case *ir.TraitStmt:
		return false // FIXME #23
	case *ir.AnonClassExpr:
		return false

	case *ir.InlineHTMLStmt:
		y, ok := y.(*ir.InlineHTMLStmt)
		return ok && x.Value == y.Value
	case *ir.StaticVarStmt:
		y, ok := y.(*ir.StaticVarStmt)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expr, y.Expr)
	case *ir.StaticStmt:
		y, ok := y.(*ir.StaticStmt)
		return ok && m.eqNodeSlice(state, x.Vars, y.Vars)
	case *ir.GlobalStmt:
		y, ok := y.(*ir.GlobalStmt)
		return ok && m.eqNodeSlice(state, x.Vars, y.Vars)
	case *ir.BreakStmt:
		y, ok := y.(*ir.BreakStmt)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.ContinueStmt:
		y, ok := y.(*ir.ContinueStmt)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.UnsetStmt:
		y, ok := y.(*ir.UnsetStmt)
		return ok && m.eqNodeSlice(state, x.Vars, y.Vars)
	case *ir.PrintExpr:
		y, ok := y.(*ir.PrintExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.EchoStmt:
		y, ok := y.(*ir.EchoStmt)
		return ok && m.eqNodeSlice(state, x.Exprs, y.Exprs)
	case *ir.NopStmt:
		_, ok := y.(*ir.NopStmt)
		return ok
	case *ir.DoStmt:
		y, ok := y.(*ir.DoStmt)
		return ok && m.eqNode(state, x.Cond, y.Cond) && m.eqNode(state, x.Stmt, y.Stmt)
	case *ir.WhileStmt:
		y, ok := y.(*ir.WhileStmt)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Cond, y.Cond) && m.eqNode(state, x.Stmt, y.Stmt)
	case *ir.ForStmt:
		y, ok := y.(*ir.ForStmt)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNodeSlice(state, x.Init, y.Init) &&
			m.eqNodeSlice(state, x.Cond, y.Cond) &&
			m.eqNodeSlice(state, x.Loop, y.Loop) &&
			m.eqNode(state, x.Stmt, y.Stmt)
	case *ir.ForeachStmt:
		y, ok := y.(*ir.ForeachStmt)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Expr, y.Expr) &&
			m.eqNode(state, x.Key, y.Key) &&
			m.eqNode(state, x.Variable, y.Variable) &&
			m.eqNode(state, x.Stmt, y.Stmt)

	case *ir.ElseStmt:
		y, ok := y.(*ir.ElseStmt)
		return ok && x.AltSyntax == y.AltSyntax && m.eqNode(state, x.Stmt, y.Stmt)
	case *ir.ElseIfStmt:
		y, ok := y.(*ir.ElseIfStmt)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Cond, y.Cond) && m.eqNode(state, x.Stmt, y.Stmt)
	case *ir.IfStmt:
		y, ok := y.(*ir.IfStmt)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNodeSliceNoMeta(state, x.ElseIf, y.ElseIf) &&
			m.eqNode(state, x.Cond, y.Cond) &&
			m.eqNode(state, x.Stmt, y.Stmt) &&
			m.eqNode(state, x.Else, y.Else)

	case *ir.ThrowStmt:
		y, ok := y.(*ir.ThrowStmt)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.CatchStmt:
		y, ok := y.(*ir.CatchStmt)
		return ok && m.eqSimpleVar(state, x.Variable, y.Variable) &&
			m.eqNodeSlice(state, x.Types, y.Types) &&
			m.eqNodeSlice(state, x.Stmts, y.Stmts)
	case *ir.TryStmt:
		y, ok := y.(*ir.TryStmt)
		return ok && m.eqNodeSlice(state, x.Stmts, y.Stmts) &&
			m.eqNodeSlice(state, x.Catches, y.Catches) &&
			m.eqNode(state, x.Finally, y.Finally)

	case *ir.YieldExpr:
		y, ok := y.(*ir.YieldExpr)
		return ok && m.eqNode(state, x.Key, y.Key) && m.eqNode(state, x.Value, y.Value)
	case *ir.YieldFromExpr:
		y, ok := y.(*ir.YieldFromExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *ir.InstanceOfExpr:
		y, ok := y.(*ir.InstanceOfExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr) &&
			m.eqNodeWithCase(state, x.Class, y.Class)

	case *ir.ListExpr:
		y, ok := y.(*ir.ListExpr)
		return ok && x.ShortSyntax == y.ShortSyntax &&
			m.eqArrayItemSlice(state, x.Items, y.Items)

	case *ir.NewExpr:
		y, ok := y.(*ir.NewExpr)
		if !ok || !m.eqNodeWithCase(state, x.Class, y.Class) {
			return false
		}
		if x.Args == nil {
			return y.Args == nil
		}
		if y.Args == nil {
			return x.Args == nil
		}
		return m.eqNodeSlice(state, x.Args, y.Args)

	case *ir.CaseStmt:
		y, ok := y.(*ir.CaseStmt)
		return ok && m.eqNode(state, x.Cond, y.Cond) && m.eqNodeSlice(state, x.Stmts, y.Stmts)
	case *ir.DefaultStmt:
		y, ok := y.(*ir.DefaultStmt)
		return ok && m.eqNodeSlice(state, x.Stmts, y.Stmts)
	case *ir.SwitchStmt:
		y, ok := y.(*ir.SwitchStmt)
		return ok && x.AltSyntax == y.AltSyntax &&
			m.eqNode(state, x.Cond, y.Cond) &&
			m.eqNodeSlice(state, x.CaseList.Cases, y.CaseList.Cases)

	case *ir.ReturnStmt:
		y, ok := y.(*ir.ReturnStmt)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *ir.Assign:
		y, ok := y.(*ir.Assign)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignPlus:
		y, ok := y.(*ir.AssignPlus)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignReference:
		y, ok := y.(*ir.AssignReference)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignBitwiseAnd:
		y, ok := y.(*ir.AssignBitwiseAnd)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignBitwiseOr:
		y, ok := y.(*ir.AssignBitwiseOr)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignBitwiseXor:
		y, ok := y.(*ir.AssignBitwiseXor)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignConcat:
		y, ok := y.(*ir.AssignConcat)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignCoalesce:
		y, ok := y.(*ir.AssignCoalesce)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignDiv:
		y, ok := y.(*ir.AssignDiv)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignMinus:
		y, ok := y.(*ir.AssignMinus)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignMod:
		y, ok := y.(*ir.AssignMod)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignMul:
		y, ok := y.(*ir.AssignMul)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignPow:
		y, ok := y.(*ir.AssignPow)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignShiftLeft:
		y, ok := y.(*ir.AssignShiftLeft)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)
	case *ir.AssignShiftRight:
		y, ok := y.(*ir.AssignShiftRight)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Expression, y.Expression)

	case *ir.ArrayDimFetchExpr:
		y, ok := y.(*ir.ArrayDimFetchExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Dim, y.Dim) && x.CurlyBrace == y.CurlyBrace
	case *ir.ArrayItemExpr:
		y, ok := y.(*ir.ArrayItemExpr)
		if !ok {
			return false
		}
		if x.Key == nil || y.Key == nil {
			return x.Key == y.Key && m.eqNode(state, x.Val, y.Val)
		}
		return m.eqNode(state, x.Key, y.Key) && m.eqNode(state, x.Val, y.Val)
	case *ir.ArrayExpr:
		y, ok := y.(*ir.ArrayExpr)
		return ok && x.ShortSyntax == y.ShortSyntax &&
			m.eqArrayItemSlice(state, x.Items, y.Items)

	case *ir.Argument:
		y, ok := y.(*ir.Argument)
		return ok && x.IsReference == y.IsReference &&
			x.Variadic == y.Variadic &&
			m.eqNode(state, x.Expr, y.Expr)
	case *ir.FunctionCallExpr:
		y, ok := y.(*ir.FunctionCallExpr)
		if !ok || !m.eqNodeWithCase(state, x.Function, y.Function) {
			return false
		}
		return m.eqNodeSlice(state, x.Args, y.Args)

	case *ir.PostIncExpr:
		y, ok := y.(*ir.PostIncExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable)
	case *ir.PostDecExpr:
		y, ok := y.(*ir.PostDecExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable)
	case *ir.PreIncExpr:
		y, ok := y.(*ir.PreIncExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable)
	case *ir.PreDecExpr:
		y, ok := y.(*ir.PreDecExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable)

	case *ir.ExitExpr:
		y, ok := y.(*ir.ExitExpr)
		return ok && x.Die == y.Die && m.eqNode(state, x.Expr, y.Expr)

	case *ir.ImportExpr:
		y, ok := y.(*ir.ImportExpr)
		return ok && x.Func == y.Func && m.eqNode(state, x.Expr, y.Expr)
	case *ir.EmptyExpr:
		y, ok := y.(*ir.EmptyExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.EvalExpr:
		y, ok := y.(*ir.EvalExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.ErrorSuppressExpr:
		y, ok := y.(*ir.ErrorSuppressExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.CloneExpr:
		y, ok := y.(*ir.CloneExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.BitwiseNotExpr:
		y, ok := y.(*ir.BitwiseNotExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.BooleanNotExpr:
		y, ok := y.(*ir.BooleanNotExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.UnaryMinusExpr:
		y, ok := y.(*ir.UnaryMinusExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	case *ir.UnaryPlusExpr:
		y, ok := y.(*ir.UnaryPlusExpr)
		return ok && m.eqNode(state, x.Expr, y.Expr)

	case *ir.StaticPropertyFetchExpr:
		switch y := y.(type) {
		case *ir.StaticPropertyFetchExpr:
			return m.eqNodeWithCase(state, x.Class, y.Class) &&
				m.eqNode(state, x.Property, y.Property)
		case *ir.ClassConstFetchExpr:
			return nodeIsVar(x.Property) &&
				m.eqNodeWithCase(state, x.Class, y.Class) &&
				m.eqNode(state, x.Property, y.ConstantName)
		default:
			return false
		}

	case *ir.ClassConstFetchExpr:
		y, ok := y.(*ir.ClassConstFetchExpr)
		return ok && m.eqNodeWithCase(state, x.Class, y.Class) && m.eqNode(state, x.ConstantName, y.ConstantName)
	case *ir.StaticCallExpr:
		y, ok := y.(*ir.StaticCallExpr)
		return ok &&
			m.eqNodeWithCase(state, x.Class, y.Class) &&
			m.eqNodeWithCase(state, x.Call, y.Call) &&
			m.eqNodeSlice(state, x.Args, y.Args)

	case *ir.ShellExecExpr:
		y, ok := y.(*ir.ShellExecExpr)
		return ok && m.eqEncapsedStringPartSlice(state, x.Parts, y.Parts)
	case *ir.Encapsed:
		y, ok := y.(*ir.Encapsed)
		return ok && m.eqEncapsedStringPartSlice(state, x.Parts, y.Parts)

	case *ir.Heredoc:
		y, ok := y.(*ir.Heredoc)
		return ok && x.Label == y.Label && m.eqEncapsedStringPartSlice(state, x.Parts, y.Parts)
	case *ir.MagicConstant:
		y, ok := y.(*ir.MagicConstant)
		return ok && y.Value == x.Value
	case *ir.Lnumber:
		y, ok := y.(*ir.Lnumber)
		return ok && y.Value == x.Value
	case *ir.Dnumber:
		y, ok := y.(*ir.Dnumber)
		return ok && y.Value == x.Value
	case *ir.String:
		y, ok := y.(*ir.String)
		return ok && y.Value == x.Value

	case *ir.CoalesceExpr:
		y, ok := y.(*ir.CoalesceExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.ConcatExpr:
		y, ok := y.(*ir.ConcatExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.DivExpr:
		y, ok := y.(*ir.DivExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.ModExpr:
		y, ok := y.(*ir.ModExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.MulExpr:
		y, ok := y.(*ir.MulExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.PowExpr:
		y, ok := y.(*ir.PowExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.BitwiseAndExpr:
		y, ok := y.(*ir.BitwiseAndExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.BitwiseOrExpr:
		y, ok := y.(*ir.BitwiseOrExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.BitwiseXorExpr:
		y, ok := y.(*ir.BitwiseXorExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.ShiftLeftExpr:
		y, ok := y.(*ir.ShiftLeftExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.ShiftRightExpr:
		y, ok := y.(*ir.ShiftRightExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.LogicalAndExpr:
		y, ok := y.(*ir.LogicalAndExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.LogicalOrExpr:
		y, ok := y.(*ir.LogicalOrExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.LogicalXorExpr:
		y, ok := y.(*ir.LogicalXorExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.BooleanAndExpr:
		y, ok := y.(*ir.BooleanAndExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.BooleanOrExpr:
		y, ok := y.(*ir.BooleanOrExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.NotEqualExpr:
		y, ok := y.(*ir.NotEqualExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.NotIdenticalExpr:
		y, ok := y.(*ir.NotIdenticalExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.EqualExpr:
		y, ok := y.(*ir.EqualExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.IdenticalExpr:
		y, ok := y.(*ir.IdenticalExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.GreaterExpr:
		y, ok := y.(*ir.GreaterExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.GreaterOrEqualExpr:
		y, ok := y.(*ir.GreaterOrEqualExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.SmallerExpr:
		y, ok := y.(*ir.SmallerExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.SmallerOrEqualExpr:
		y, ok := y.(*ir.SmallerOrEqualExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.SpaceshipExpr:
		y, ok := y.(*ir.SpaceshipExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.PlusExpr:
		y, ok := y.(*ir.PlusExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)
	case *ir.MinusExpr:
		y, ok := y.(*ir.MinusExpr)
		return ok && m.eqNode(state, x.Left, y.Left) && m.eqNode(state, x.Right, y.Right)

	case *ir.ConstFetchExpr:
		y, ok := y.(*ir.ConstFetchExpr)
		return ok && m.eqNode(state, x.Constant, y.Constant)
	case *ir.Name:
		y, ok := y.(*ir.Name)
		return ok && x.Value == y.Value
	case *ir.Identifier:
		y, ok := y.(*ir.Identifier)
		return ok && x.Value == y.Value
	case *ir.SimpleVar:
		return m.eqSimpleVar(state, x, y)
	case *ir.Var:
		return m.eqVar(state, x, y)

	case *ir.ReferenceExpr:
		y, ok := y.(*ir.ReferenceExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable)

	case *ir.Parameter:
		y, ok := y.(*ir.Parameter)
		return ok && x.ByRef == y.ByRef &&
			x.Variadic == y.Variadic &&
			m.eqNode(state, x.VariableType, y.VariableType) &&
			m.eqNode(state, x.Variable, y.Variable) &&
			m.eqNode(state, x.DefaultValue, y.DefaultValue)
	case *ir.ClosureExpr:
		return m.eqClosure(state, x, y)

	case *ir.TernaryExpr:
		return m.eqTernary(state, x, y)

	case *ir.IssetExpr:
		y, ok := y.(*ir.IssetExpr)
		return ok && m.eqNodeSlice(state, x.Variables, y.Variables)

	case *ir.PropertyFetchExpr:
		y, ok := y.(*ir.PropertyFetchExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable) && m.eqNode(state, x.Property, y.Property)
	case *ir.MethodCallExpr:
		y, ok := y.(*ir.MethodCallExpr)
		return ok && m.eqNode(state, x.Variable, y.Variable) &&
			m.eqNodeWithCase(state, x.Method, y.Method) &&
			m.eqNodeSlice(state, x.Args, y.Args)

	case *ir.TypeCastExpr:
		y, ok := y.(*ir.TypeCastExpr)
		return ok && x.Type == y.Type && m.eqNode(state, x.Expr, y.Expr)

	case *ir.Root:
		return false

	default:
		panic(fmt.Sprintf("unhandled node: x=%T y=%T (please, fill an issue on GitHub)\n", x, y))
	}
}

func (m *matcher) matchNamed(state *matcherState, name string, y ir.Node) bool {
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

func (m *matcher) eqTernary(state *matcherState, x *ir.TernaryExpr, y ir.Node) bool {
	if y, ok := y.(*ir.TernaryExpr); ok {
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

func (m *matcher) eqClosure(state *matcherState, x *ir.ClosureExpr, y ir.Node) bool {
	if y, ok := y.(*ir.ClosureExpr); ok {
		var xUses, yUses []ir.Node
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

func (m *matcher) eqSimpleVar(state *matcherState, x *ir.SimpleVar, y ir.Node) bool {
	if state.literalMatch {
		y, ok := y.(*ir.SimpleVar)
		return ok && x.Name == y.Name
	}
	return m.matchNamed(state, x.Name, y)
}

func (m *matcher) eqVar(state *matcherState, x *ir.Var, y ir.Node) bool {
	if state.literalMatch {
		y, ok := y.(*ir.Var)
		return ok && m.eqNode(state, x.Expr, y.Expr)
	}

	switch vn := x.Expr.(type) {
	case anyFunc:
		_, ok := y.(*ir.ClosureExpr)
		return ok && m.matchNamed(state, vn.name, y)
	case anyConst:
		switch y.(type) {
		case *ir.ConstFetchExpr, *ir.ClassConstFetchExpr:
			return m.matchNamed(state, vn.name, y)
		default:
			return false
		}
	case anyVar:
		switch y.(type) {
		case *ir.SimpleVar, *ir.Var:
			return m.matchNamed(state, vn.name, y)
		default:
			return false
		}
	case anyInt:
		_, ok := y.(*ir.Lnumber)
		return ok && m.matchNamed(state, vn.name, y)
	case anyFloat:
		_, ok := y.(*ir.Dnumber)
		return ok && m.matchNamed(state, vn.name, y)
	case anyStr:
		_, ok := y.(*ir.String)
		return ok && m.matchNamed(state, vn.name, y)
	case anyStr1:
		y, ok := y.(*ir.String)
		return ok && len(y.Value) == 1 && m.matchNamed(state, vn.name, y)
	case anyNum:
		switch y.(type) {
		case *ir.Lnumber, *ir.Dnumber:
			return m.matchNamed(state, vn.name, y)
		default:
			return false
		}
	case anyExpr:
		return nodeIsExpr(y) && m.matchNamed(state, vn.name, y)
	case anyCall:
		switch y.(type) {
		case *ir.FunctionCallExpr, *ir.StaticCallExpr, *ir.MethodCallExpr:
			return m.matchNamed(state, vn.name, y)
		default:
			return false
		}
	}

	if y, ok := y.(*ir.Var); ok {
		return m.eqNode(state, x.Expr, y.Expr)
	}
	return false
}
