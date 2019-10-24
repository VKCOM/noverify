package expr_test

import (
	"bytes"
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestArrayDimFetch(t *testing.T) {
	src := `<? $a[1];`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    9,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    9,
				},
				Expr: &expr.ArrayDimFetch{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    8,
					},
					Variable: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    5,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  4,
								EndPos:    5,
							},
							Value: "a",
						},
					},
					Dim: &scalar.Lnumber{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  7,
							EndPos:    7,
						},
						Value: "1",
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

func TestArrayDimFetchNested(t *testing.T) {
	src := `<? $a[1][2];`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    12,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    12,
				},
				Expr: &expr.ArrayDimFetch{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    11,
					},
					Variable: &expr.ArrayDimFetch{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    8,
						},
						Variable: &node.Variable{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  4,
								EndPos:    5,
							},
							VarName: &node.Identifier{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  4,
									EndPos:    5,
								},
								Value: "a",
							},
						},
						Dim: &scalar.Lnumber{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  7,
								EndPos:    7,
							},
							Value: "1",
						},
					},
					Dim: &scalar.Lnumber{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  10,
							EndPos:    10,
						},
						Value: "2",
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
