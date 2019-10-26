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

func TestEmptyList(t *testing.T) {
	src := `<? list() = $b;`

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
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    9,
						},
						Items: []*expr.ArrayItem{},
					},
					Expression: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  13,
							EndPos:    14,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  13,
								EndPos:    14,
							},
							Value: "b",
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

func TestList(t *testing.T) {
	src := `<? list($a) = $b;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    17,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    17,
				},
				Expr: &assign.Assign{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    16,
					},
					Variable: &expr.List{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    11,
						},
						Items: []*expr.ArrayItem{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  9,
									EndPos:    10,
								},
								Val: &node.Variable{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  9,
										EndPos:    10,
									},
									VarName: &node.Identifier{
										Position: &position.Position{
											StartLine: 1,
											EndLine:   1,
											StartPos:  9,
											EndPos:    10,
										},
										Value: "a",
									},
								},
							},
						},
					},
					Expression: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  15,
							EndPos:    16,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  15,
								EndPos:    16,
							},
							Value: "b",
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

func TestListArrayIndex(t *testing.T) {
	src := `<? list($a[]) = $b;`

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
									StartPos:  9,
									EndPos:    12,
								},
								Val: &expr.ArrayDimFetch{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  9,
										EndPos:    12,
									},
									Variable: &node.Variable{
										Position: &position.Position{
											StartLine: 1,
											EndLine:   1,
											StartPos:  9,
											EndPos:    10,
										},
										VarName: &node.Identifier{
											Position: &position.Position{
												StartLine: 1,
												EndLine:   1,
												StartPos:  9,
												EndPos:    10,
											},
											Value: "a",
										},
									},
								},
							},
						},
					},
					Expression: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  17,
							EndPos:    18,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  17,
								EndPos:    18,
							},
							Value: "b",
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

func TestListList(t *testing.T) {
	src := `<? list(list($a)) = $b;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    23,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    23,
				},
				Expr: &assign.Assign{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    22,
					},
					Variable: &expr.List{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    17,
						},
						Items: []*expr.ArrayItem{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  9,
									EndPos:    16,
								},
								Val: &expr.List{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  9,
										EndPos:    16,
									},
									Items: []*expr.ArrayItem{
										{
											Position: &position.Position{
												StartLine: 1,
												EndLine:   1,
												StartPos:  14,
												EndPos:    15,
											},
											Val: &node.Variable{
												Position: &position.Position{
													StartLine: 1,
													EndLine:   1,
													StartPos:  14,
													EndPos:    15,
												},
												VarName: &node.Identifier{
													Position: &position.Position{
														StartLine: 1,
														EndLine:   1,
														StartPos:  14,
														EndPos:    15,
													},
													Value: "a",
												},
											},
										},
									},
								},
							},
						},
					},
					Expression: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  21,
							EndPos:    22,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  21,
								EndPos:    22,
							},
							Value: "b",
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

func TestListEmptyItem(t *testing.T) {
	src := `<? list(, $a) = $b;`

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
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    13,
						},
						Items: []*expr.ArrayItem{
							{},
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  11,
									EndPos:    12,
								},
								Val: &node.Variable{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  11,
										EndPos:    12,
									},
									VarName: &node.Identifier{
										Position: &position.Position{
											StartLine: 1,
											EndLine:   1,
											StartPos:  11,
											EndPos:    12,
										},
										Value: "a",
									},
								},
							},
						},
					},
					Expression: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  17,
							EndPos:    18,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  17,
								EndPos:    18,
							},
							Value: "b",
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

func TestListEmptyItems(t *testing.T) {
	src := `<? list(, , $a, ) = $b;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    23,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    23,
				},
				Expr: &assign.Assign{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    22,
					},
					Variable: &expr.List{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  4,
							EndPos:    17,
						},
						Items: []*expr.ArrayItem{
							{},
							{},
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  13,
									EndPos:    14,
								},
								Val: &node.Variable{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  13,
										EndPos:    14,
									},
									VarName: &node.Identifier{
										Position: &position.Position{
											StartLine: 1,
											EndLine:   1,
											StartPos:  13,
											EndPos:    14,
										},
										Value: "a",
									},
								},
							},
							{},
						},
					},
					Expression: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  21,
							EndPos:    22,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  21,
								EndPos:    22,
							},
							Value: "b",
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
