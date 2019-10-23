package stmt_test

import (
	"bytes"
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/position"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
)

func TestConstList(t *testing.T) {
	src := `<? const FOO = 1, BAR = 2;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    26,
		},
		Stmts: []node.Node{
			&stmt.ConstList{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    26,
				},
				Consts: []node.Node{
					&stmt.Constant{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  10,
							EndPos:    16,
						},
						PhpDocComment: "",
						ConstantName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  10,
								EndPos:    12,
							},
							Value: "FOO",
						},
						Expr: &scalar.Lnumber{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  16,
								EndPos:    16,
							},
							Value: "1",
						},
					},
					&stmt.Constant{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    25,
						},
						PhpDocComment: "",
						ConstantName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  19,
								EndPos:    21,
							},
							Value: "BAR",
						},
						Expr: &scalar.Lnumber{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  25,
								EndPos:    25,
							},
							Value: "2",
						},
					},
				},
			},
		},
	}

	php7parser := php7.NewParser(bytes.NewBufferString(src), "test.php")
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}
