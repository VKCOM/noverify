package ir

func (n *Assign) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignBitwiseAnd) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignBitwiseOr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignBitwiseXor) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignCoalesce) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignConcat) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignDiv) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignMinus) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignMod) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignMul) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignPlus) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignPow) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignReference) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignShiftLeft) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *AssignShiftRight) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expression != nil {
		n.Expression.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BitwiseAndExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BitwiseOrExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BitwiseXorExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BooleanAndExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BooleanOrExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *CoalesceExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ConcatExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *DivExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *EqualExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *GreaterExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *GreaterOrEqualExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *IdenticalExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *LogicalAndExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *LogicalOrExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *LogicalXorExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *MinusExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ModExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *MulExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *NotEqualExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *NotIdenticalExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PlusExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PowExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ShiftLeftExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ShiftRightExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *SmallerExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *SmallerOrEqualExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *SpaceshipExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Left != nil {
		n.Left.Walk(v)
	}
	if n.Right != nil {
		n.Right.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ArrayCastExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BoolCastExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *DoubleCastExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *IntCastExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ObjectCastExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *StringCastExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *UnsetCastExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ArrayExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Items {
		if n.Items[i] != nil {
			n.Items[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ArrayDimFetchExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Dim != nil {
		n.Dim.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ArrayItemExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Key != nil {
		n.Key.Walk(v)
	}
	if n.Val != nil {
		n.Val.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ArrowFunctionExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Params {
		if n.Params[i] != nil {
			n.Params[i].Walk(v)
		}
	}
	if n.ReturnType != nil {
		n.ReturnType.Walk(v)
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BitwiseNotExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *BooleanNotExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ClassConstFetchExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Class != nil {
		n.Class.Walk(v)
	}
	if n.ConstantName != nil {
		n.ConstantName.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *CloneExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ClosureExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Params {
		if n.Params[i] != nil {
			n.Params[i].Walk(v)
		}
	}
	if n.ClosureUse != nil {
		n.ClosureUse.Walk(v)
	}
	if n.ReturnType != nil {
		n.ReturnType.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ClosureUseExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Uses {
		if n.Uses[i] != nil {
			n.Uses[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ConstFetchExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Constant != nil {
		n.Constant.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *EmptyExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ErrorSuppressExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *EvalExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ExitExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *FunctionCallExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Function != nil {
		n.Function.Walk(v)
	}
	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *IncludeExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *IncludeOnceExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *InstanceOfExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	if n.Class != nil {
		n.Class.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *IssetExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Variables {
		if n.Variables[i] != nil {
			n.Variables[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ListExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Items {
		if n.Items[i] != nil {
			n.Items[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *MethodCallExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Method != nil {
		n.Method.Walk(v)
	}
	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *NewExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Class != nil {
		n.Class.Walk(v)
	}
	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ParenExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PostDecExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PostIncExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PreDecExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PreIncExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PrintExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PropertyFetchExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Property != nil {
		n.Property.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ReferenceExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *RequireExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *RequireOnceExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ShellExecExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Parts {
		if n.Parts[i] != nil {
			n.Parts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *StaticCallExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Class != nil {
		n.Class.Walk(v)
	}
	if n.Call != nil {
		n.Call.Walk(v)
	}
	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *StaticPropertyFetchExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Class != nil {
		n.Class.Walk(v)
	}
	if n.Property != nil {
		n.Property.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *TernaryExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Condition != nil {
		n.Condition.Walk(v)
	}
	if n.IfTrue != nil {
		n.IfTrue.Walk(v)
	}
	if n.IfFalse != nil {
		n.IfFalse.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *UnaryMinusExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *UnaryPlusExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *YieldExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Key != nil {
		n.Key.Walk(v)
	}
	if n.Value != nil {
		n.Value.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *YieldFromExpr) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *FullyQualifiedName) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Parts {
		if n.Parts[i] != nil {
			n.Parts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *Name) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Parts {
		if n.Parts[i] != nil {
			n.Parts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *NamePart) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *RelativeName) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Parts {
		if n.Parts[i] != nil {
			n.Parts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *Argument) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ArgumentList) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Arguments {
		if n.Arguments[i] != nil {
			n.Arguments[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *Identifier) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *Nullable) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *Parameter) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.VariableType != nil {
		n.VariableType.Walk(v)
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.DefaultValue != nil {
		n.DefaultValue.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *Root) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *SimpleVar) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *Var) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *Dnumber) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *Encapsed) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Parts {
		if n.Parts[i] != nil {
			n.Parts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *EncapsedStringPart) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *Heredoc) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Parts {
		if n.Parts[i] != nil {
			n.Parts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *Lnumber) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *MagicConstant) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *String) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *BreakStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *CaseStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Cond != nil {
		n.Cond.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *CaseListStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Cases {
		if n.Cases[i] != nil {
			n.Cases[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *CatchStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Types {
		if n.Types[i] != nil {
			n.Types[i].Walk(v)
		}
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ClassStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.ClassName != nil {
		n.ClassName.Walk(v)
	}
	for i := range n.Modifiers {
		if n.Modifiers[i] != nil {
			n.Modifiers[i].Walk(v)
		}
	}
	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}
	if n.Extends != nil {
		n.Extends.Walk(v)
	}
	if n.Implements != nil {
		n.Implements.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ClassConstListStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Modifiers {
		if n.Modifiers[i] != nil {
			n.Modifiers[i].Walk(v)
		}
	}
	for i := range n.Consts {
		if n.Consts[i] != nil {
			n.Consts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ClassExtendsStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.ClassName != nil {
		n.ClassName.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ClassImplementsStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.InterfaceNames {
		if n.InterfaceNames[i] != nil {
			n.InterfaceNames[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ClassMethodStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.MethodName != nil {
		n.MethodName.Walk(v)
	}
	for i := range n.Modifiers {
		if n.Modifiers[i] != nil {
			n.Modifiers[i].Walk(v)
		}
	}
	for i := range n.Params {
		if n.Params[i] != nil {
			n.Params[i].Walk(v)
		}
	}
	if n.ReturnType != nil {
		n.ReturnType.Walk(v)
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ConstListStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Consts {
		if n.Consts[i] != nil {
			n.Consts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ConstantStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.ConstantName != nil {
		n.ConstantName.Walk(v)
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ContinueStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *DeclareStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Consts {
		if n.Consts[i] != nil {
			n.Consts[i].Walk(v)
		}
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *DefaultStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *DoStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	if n.Cond != nil {
		n.Cond.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *EchoStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Exprs {
		if n.Exprs[i] != nil {
			n.Exprs[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ElseStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ElseIfStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Cond != nil {
		n.Cond.Walk(v)
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ExpressionStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *FinallyStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ForStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Init {
		if n.Init[i] != nil {
			n.Init[i].Walk(v)
		}
	}
	for i := range n.Cond {
		if n.Cond[i] != nil {
			n.Cond[i].Walk(v)
		}
	}
	for i := range n.Loop {
		if n.Loop[i] != nil {
			n.Loop[i].Walk(v)
		}
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ForeachStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	if n.Key != nil {
		n.Key.Walk(v)
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *FunctionStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.FunctionName != nil {
		n.FunctionName.Walk(v)
	}
	for i := range n.Params {
		if n.Params[i] != nil {
			n.Params[i].Walk(v)
		}
	}
	if n.ReturnType != nil {
		n.ReturnType.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *GlobalStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Vars {
		if n.Vars[i] != nil {
			n.Vars[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *GotoStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Label != nil {
		n.Label.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *GroupUseStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.UseType != nil {
		n.UseType.Walk(v)
	}
	if n.Prefix != nil {
		n.Prefix.Walk(v)
	}
	for i := range n.UseList {
		if n.UseList[i] != nil {
			n.UseList[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *HaltCompilerStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *IfStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Cond != nil {
		n.Cond.Walk(v)
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	for i := range n.ElseIf {
		if n.ElseIf[i] != nil {
			n.ElseIf[i].Walk(v)
		}
	}
	if n.Else != nil {
		n.Else.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *InlineHTMLStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *InterfaceStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.InterfaceName != nil {
		n.InterfaceName.Walk(v)
	}
	if n.Extends != nil {
		n.Extends.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *InterfaceExtendsStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.InterfaceNames {
		if n.InterfaceNames[i] != nil {
			n.InterfaceNames[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *LabelStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.LabelName != nil {
		n.LabelName.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *NamespaceStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.NamespaceName != nil {
		n.NamespaceName.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *NopStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	v.LeaveNode(n)
}

func (n *PropertyStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *PropertyListStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Modifiers {
		if n.Modifiers[i] != nil {
			n.Modifiers[i].Walk(v)
		}
	}
	if n.Type != nil {
		n.Type.Walk(v)
	}
	for i := range n.Properties {
		if n.Properties[i] != nil {
			n.Properties[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *ReturnStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *StaticStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Vars {
		if n.Vars[i] != nil {
			n.Vars[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *StaticVarStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Variable != nil {
		n.Variable.Walk(v)
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *StmtList) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *SwitchStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Cond != nil {
		n.Cond.Walk(v)
	}
	if n.CaseList != nil {
		n.CaseList.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *ThrowStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Expr != nil {
		n.Expr.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *TraitStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.TraitName != nil {
		n.TraitName.Walk(v)
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *TraitAdaptationListStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Adaptations {
		if n.Adaptations[i] != nil {
			n.Adaptations[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *TraitMethodRefStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Trait != nil {
		n.Trait.Walk(v)
	}
	if n.Method != nil {
		n.Method.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *TraitUseStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Traits {
		if n.Traits[i] != nil {
			n.Traits[i].Walk(v)
		}
	}
	if n.TraitAdaptationList != nil {
		n.TraitAdaptationList.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *TraitUseAliasStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Ref != nil {
		n.Ref.Walk(v)
	}
	if n.Modifier != nil {
		n.Modifier.Walk(v)
	}
	if n.Alias != nil {
		n.Alias.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *TraitUsePrecedenceStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Ref != nil {
		n.Ref.Walk(v)
	}
	for i := range n.Insteadof {
		if n.Insteadof[i] != nil {
			n.Insteadof[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *TryStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Stmts {
		if n.Stmts[i] != nil {
			n.Stmts[i].Walk(v)
		}
	}
	for i := range n.Catches {
		if n.Catches[i] != nil {
			n.Catches[i].Walk(v)
		}
	}
	if n.Finally != nil {
		n.Finally.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *UnsetStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	for i := range n.Vars {
		if n.Vars[i] != nil {
			n.Vars[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *UseStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.UseType != nil {
		n.UseType.Walk(v)
	}
	if n.Use != nil {
		n.Use.Walk(v)
	}
	if n.Alias != nil {
		n.Alias.Walk(v)
	}
	v.LeaveNode(n)
}

func (n *UseListStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.UseType != nil {
		n.UseType.Walk(v)
	}
	for i := range n.Uses {
		if n.Uses[i] != nil {
			n.Uses[i].Walk(v)
		}
	}
	v.LeaveNode(n)
}

func (n *WhileStmt) Walk(v Visitor) {
	if !v.EnterNode(n) {
		return
	}
	if n.Cond != nil {
		n.Cond.Walk(v)
	}
	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}
	v.LeaveNode(n)
}
