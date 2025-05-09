package linter

import (
	"bytes"
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/workspace"
	"github.com/VKCOM/php-parser/pkg/position"
)

type QuickFixGenerator struct {
	file *workspace.File
}

func NewQuickFixGenerator(file *workspace.File) *QuickFixGenerator {
	return &QuickFixGenerator{file: file}
}

func (g *QuickFixGenerator) Array(arr *ir.ArrayExpr) quickfix.TextEdit {
	from := arr.Position.StartPos
	to := arr.Position.EndPos

	have := g.file.Contents()[from:to]
	have = bytes.TrimPrefix(have, []byte("array"))
	have = bytes.TrimSpace(have)
	have = bytes.TrimPrefix(have, []byte("("))
	have = bytes.TrimSuffix(have, []byte(")"))

	return quickfix.TextEdit{
		StartPos:    arr.Position.StartPos,
		EndPos:      arr.Position.EndPos,
		Replacement: fmt.Sprintf("[%s]", string(have)),
	}
}

func (g *QuickFixGenerator) NullForNotNullableProperty(prop *ir.PropertyStmt) quickfix.TextEdit {
	from := prop.Position.StartPos
	to := prop.Variable.Position.EndPos

	withoutAssign := g.file.Contents()[from:to]

	return quickfix.TextEdit{
		StartPos:    prop.Position.StartPos,
		EndPos:      prop.Position.EndPos,
		Replacement: string(withoutAssign),
	}
}

func (g *QuickFixGenerator) PhpAliasesReplace(prop *ir.Name, masterFunction string) quickfix.TextEdit {
	return quickfix.TextEdit{
		StartPos:    prop.Position.StartPos,
		EndPos:      prop.Position.EndPos,
		Replacement: masterFunction,
	}
}

func (g *QuickFixGenerator) notExplicitNullableParam(param ir.Node) quickfix.TextEdit {
	var pos *position.Position
	var value string

	switch v := param.(type) {
	case *ir.Name:
		pos = &position.Position{
			StartPos: v.Position.StartPos,
			EndPos:   v.Position.EndPos,
		}
		value = v.Value
	case *ir.Identifier:
		pos = &position.Position{
			StartPos: v.Position.StartPos,
			EndPos:   v.Position.EndPos,
		}
		value = v.Value
	}

	return quickfix.TextEdit{
		StartPos:    pos.StartPos,
		EndPos:      pos.EndPos,
		Replacement: "?" + value,
	}
}

func (g *QuickFixGenerator) GetType(node ir.Node, isFunctionName, nodeText string, isNegative bool) quickfix.TextEdit {
	pos := ir.GetPosition(node)

	if isNegative {
		isFunctionName = "!" + isFunctionName
	}

	isFunctionName = isFunctionName + "(" + nodeText + ")"

	return quickfix.TextEdit{
		StartPos:    pos.StartPos,
		EndPos:      pos.EndPos,
		Replacement: isFunctionName,
	}
}
