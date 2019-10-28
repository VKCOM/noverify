package expr_test

import (
	"bytes"
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestShortList(t *testing.T) {
	src := `<? [$a] = $b;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    13,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    13,
				},
				Expr: &assign.Assign{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    12,
					},
					Variable: &expr.List{
						ShortSyntax: true,
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    7,
						},
						Items: []*expr.ArrayItem{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  5,
									EndPos:    6,
								},
								Val: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  5,
										EndPos:    6,
									},
									Name: "a",
								},
							},
						},
					},
					Expression: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  11,
							EndPos:    12,
						},
						Name: "b",
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

func TestShortListArrayIndex(t *testing.T) {
	src := `<? [$a[]] = $b;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    15,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    15,
				},
				Expr: &assign.Assign{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    14,
					},
					Variable: &expr.List{
						ShortSyntax: true,
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    9,
						},
						Items: []*expr.ArrayItem{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  5,
									EndPos:    8,
								},
								Val: &expr.ArrayDimFetch{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  5,
										EndPos:    8,
									},
									Variable: &node.SimpleVar{
										Position: &position.Position{
											StartLine: 1,
											EndLine:   1,
											StartPos:  5,
											EndPos:    6,
										},
										Name: "a",
									},
								},
							},
						},
					},
					Expression: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  13,
							EndPos:    14,
						},
						Name: "b",
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

func TestShortListList(t *testing.T) {
	src := `<? [list($a)] = $b;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    19,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    19,
				},
				Expr: &assign.Assign{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    18,
					},
					Variable: &expr.List{
						ShortSyntax: true,
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    13,
						},
						Items: []*expr.ArrayItem{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  5,
									EndPos:    12,
								},
								Val: &expr.List{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  5,
										EndPos:    12,
									},
									Items: []*expr.ArrayItem{
										{
											Position: &position.Position{
												StartLine: 1,
												EndLine:   1,
												StartPos:  10,
												EndPos:    11,
											},
											Val: &node.SimpleVar{
												Position: &position.Position{
													StartLine: 1,
													EndLine:   1,
													StartPos:  10,
													EndPos:    11,
												},
												Name: "a",
											},
										},
									},
								},
							},
						},
					},
					Expression: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  17,
							EndPos:    18,
						},
						Name: "b",
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
