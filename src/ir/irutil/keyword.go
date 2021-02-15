package irutil

import (
	"bytes"

	"github.com/z7zmey/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/ir"
)

// Keywords returns one or two tokens that
// contain the keywords for the passed node.
func Keywords(n ir.Node) []*token.Token {
	switch n := n.(type) {
	case *ir.FunctionStmt:
		return []*token.Token{n.FunctionTkn}
	case *ir.DefaultStmt:
		return []*token.Token{n.DefaultTkn}
	case *ir.CaseStmt:
		return []*token.Token{n.CaseTkn}
	case *ir.CloneExpr:
		return []*token.Token{n.CloneTkn}
	case *ir.ConstListStmt:
		return []*token.Token{n.ConstTkn}
	case *ir.GotoStmt:
		return []*token.Token{n.GotoTkn}
	case *ir.ThrowStmt:
		return []*token.Token{n.ThrowTkn}
	case *ir.YieldExpr:
		return []*token.Token{n.YieldTkn}
	case *ir.YieldFromExpr:
		tok := n.YieldFromTkn

		parts := bytes.Split(tok.Value, []byte(" "))

		yieldVal := parts[0]
		yieldPos := n.Position
		yieldPos.EndLine = yieldPos.StartLine
		yieldPos.EndPos = yieldPos.StartPos + len(yieldVal)
		yieldTkn := &token.Token{
			ID:           tok.ID,
			Value:        yieldVal,
			Position:     yieldPos,
			FreeFloating: tok.FreeFloating,
		}

		fromVal := parts[1]
		if len(parts) == 3 {
			fromVal = parts[2]
		}
		fromPos := n.Position
		fromPos.StartLine = fromPos.EndLine
		fromPos.StartPos = fromPos.EndPos - len(fromVal)
		fromTkn := &token.Token{
			ID:       tok.ID,
			Value:    fromVal,
			Position: fromPos,
		}

		return []*token.Token{yieldTkn, fromTkn}
	case *ir.BreakStmt:
		return []*token.Token{n.BreakTkn}
	case *ir.ReturnStmt:
		return []*token.Token{n.ReturnTkn}
	case *ir.ForeachStmt:
		return []*token.Token{n.ForeachTkn}
	case *ir.ForStmt:
		return []*token.Token{n.ForTkn}
	case *ir.WhileStmt:
		return []*token.Token{n.WhileTkn}
	case *ir.DoStmt:
		return []*token.Token{n.DoTkn}
	case *ir.TryStmt:
		return []*token.Token{n.TryTkn}
	case *ir.CatchStmt:
		return []*token.Token{n.CatchTkn}
	case *ir.FinallyStmt:
		return []*token.Token{n.FinallyTkn}
	case *ir.NewExpr:
		return []*token.Token{n.NewTkn}
	case *ir.GlobalStmt:
		return []*token.Token{n.GlobalTkn}
	case *ir.ContinueStmt:
		return []*token.Token{n.ContinueTkn}
	case *ir.InterfaceStmt:
		return []*token.Token{n.InterfaceTkn}
	case *ir.ClassImplementsStmt:
		return []*token.Token{n.ImplementsTkn}
	case *ir.ClassExtendsStmt:
		return []*token.Token{n.ExtendsTkn}
	case *ir.TraitStmt:
		return []*token.Token{n.TraitTkn}
	case *ir.TraitUseStmt:
		return []*token.Token{n.UseTkn}
	case *ir.NamespaceStmt:
		return []*token.Token{n.NsTkn}
	case *ir.ImportExpr:
		return []*token.Token{n.ImportTkn}
	case *ir.IfStmt:
		return []*token.Token{n.IfTkn}
	case *ir.ElseStmt:
		return []*token.Token{n.ElseTkn}
	case *ir.ElseIfStmt:
		if !n.Merged {
			return []*token.Token{n.IfTkn, n.ElseTkn}
		}

		return []*token.Token{n.ElseIfTkn}
	}

	return nil
}
