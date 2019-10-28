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

func TestShortArray(t *testing.T) {
	src := `<? [];`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    6,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    6,
				},
				Expr: &expr.Array{
					ShortSyntax: true,
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    5,
					},
					Items: []*expr.ArrayItem{},
				},
			},
		},
	}

	php7parser := php7.NewParser(bytes.NewBufferString(src), "test.php")
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestShortArrayItem(t *testing.T) {
	src := `<? [1];`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    7,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    7,
				},
				Expr: &expr.Array{
					ShortSyntax: true,
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    6,
					},
					Items: []*expr.ArrayItem{
						{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  5,
								EndPos:    5,
							},
							Val: &scalar.Lnumber{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  5,
									EndPos:    5,
								},
								Value: "1",
							},
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

func TestShortArrayItems(t *testing.T) {
	src := `<? [1=>1, &$b,];`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    16,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    16,
				},
				Expr: &expr.Array{
					ShortSyntax: true,
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    15,
					},
					Items: []*expr.ArrayItem{
						{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  5,
								EndPos:    8,
							},
							Key: &scalar.Lnumber{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  5,
									EndPos:    5,
								},
								Value: "1",
							},
							Val: &scalar.Lnumber{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  8,
									EndPos:    8,
								},
								Value: "1",
							},
						},
						{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  11,
								EndPos:    13,
							},
							Val: &expr.Reference{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  11,
									EndPos:    13,
								},
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  12,
										EndPos:    13,
									},
									Name: "b",
								},
							},
						},
						{},
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
