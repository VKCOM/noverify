package expr_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestPrint(t *testing.T) {
	src := `<? print($a);`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    13,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    13,
				},
				Expr: &expr.Print{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  3,
						EndPos:    12,
					},
					Expr: &expr.Paren{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  8,
							EndPos:    12,
						},
						Expr: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  9,
								EndPos:    11,
							},
							Name: "a",
						},
					},
				},
			},
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}
