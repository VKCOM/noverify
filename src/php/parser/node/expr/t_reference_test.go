package expr_test

import (
	"bytes"
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestForeachWithRef(t *testing.T) {
	t.Helper()
	src := `<? foreach ($a as $k => &$v) {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    31,
		},
		Stmts: []node.Node{
			&stmt.Foreach{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    31,
				},
				Expr: &node.SimpleVar{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  13,
						EndPos:    14,
					},
					Name: "a",
				},
				Key: &node.SimpleVar{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					Name: "k",
				},
				Variable: &expr.Reference{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  25,
						EndPos:    27,
					},
					Variable: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  26,
							EndPos:    27,
						},
						Name: "v",
					},
				},
				Stmt: &stmt.StmtList{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  30,
						EndPos:    31,
					},
					Stmts: []node.Node{},
				},
			},
		},
	}

	php7parser := php7.NewParser(bytes.NewBufferString(src), "test.php")
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}
