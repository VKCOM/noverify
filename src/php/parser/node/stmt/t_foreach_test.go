package stmt_test

import (
	"bytes"
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/position"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
)

func TestForeach(t *testing.T) {
	src := `<? foreach ($a as $v) {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    24,
		},
		Stmts: []node.Node{
			&stmt.Foreach{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    24,
				},
				Expr: &node.Variable{
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
				Variable: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    20,
						},
						Value: "v",
					},
				},
				Stmt: &stmt.StmtList{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  23,
						EndPos:    24,
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

func TestForeachExpr(t *testing.T) {
	src := `<? foreach ([] as $v) {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    24,
		},
		Stmts: []node.Node{
			&stmt.Foreach{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    24,
				},
				Expr: &expr.Array{
					ShortSyntax: true,
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  13,
						EndPos:    14,
					},
					Items: []*expr.ArrayItem{},
				},
				Variable: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    20,
						},
						Value: "v",
					},
				},
				Stmt: &stmt.StmtList{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  23,
						EndPos:    24,
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

func TestAltForeach(t *testing.T) {
	src := `<? foreach ($a as $v) : endforeach;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    35,
		},
		Stmts: []node.Node{
			&stmt.Foreach{
				AltSyntax: true,
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    35,
				},
				Expr: &node.Variable{
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
				Variable: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    20,
						},
						Value: "v",
					},
				},
				Stmt: &stmt.StmtList{
					Position: &position.Position{
						StartLine: -1,
						EndLine:   -1,
						StartPos:  -1,
						EndPos:    -1,
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

func TestForeachWithKey(t *testing.T) {
	src := `<? foreach ($a as $k => $v) {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    30,
		},
		Stmts: []node.Node{
			&stmt.Foreach{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    30,
				},
				Expr: &node.Variable{
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
				Key: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    20,
						},
						Value: "k",
					},
				},
				Variable: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  25,
						EndPos:    26,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  25,
							EndPos:    26,
						},
						Value: "v",
					},
				},
				Stmt: &stmt.StmtList{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  29,
						EndPos:    30,
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

func TestForeachExprWithKey(t *testing.T) {
	src := `<? foreach ([] as $k => $v) {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    30,
		},
		Stmts: []node.Node{
			&stmt.Foreach{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    30,
				},
				Expr: &expr.Array{
					ShortSyntax: true,
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  13,
						EndPos:    14,
					},
					Items: []*expr.ArrayItem{},
				},
				Key: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    20,
						},
						Value: "k",
					},
				},
				Variable: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  25,
						EndPos:    26,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  25,
							EndPos:    26,
						},
						Value: "v",
					},
				},
				Stmt: &stmt.StmtList{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  29,
						EndPos:    30,
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

func TestForeachWithRef(t *testing.T) {
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
				Expr: &node.Variable{
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
				Key: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    20,
						},
						Value: "k",
					},
				},
				Variable: &expr.Reference{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  25,
						EndPos:    27,
					},
					Variable: &node.Variable{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  26,
							EndPos:    27,
						},
						VarName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  26,
								EndPos:    27,
							},
							Value: "v",
						},
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

func TestForeachWithList(t *testing.T) {
	src := `<? foreach ($a as $k => list($v)) {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    36,
		},
		Stmts: []node.Node{
			&stmt.Foreach{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    36,
				},
				Expr: &node.Variable{
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
				Key: &node.Variable{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  19,
						EndPos:    20,
					},
					VarName: &node.Identifier{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    20,
						},
						Value: "k",
					},
				},
				Variable: &expr.List{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  25,
						EndPos:    32,
					},
					Items: []*expr.ArrayItem{
						&expr.ArrayItem{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  30,
								EndPos:    31,
							},
							Val: &node.Variable{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  30,
									EndPos:    31,
								},
								VarName: &node.Identifier{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  30,
										EndPos:    31,
									},
									Value: "v",
								},
							},
						},
					},
				},
				Stmt: &stmt.StmtList{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  35,
						EndPos:    36,
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
