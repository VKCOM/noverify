package stmt_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestAltIf(t *testing.T) {
	src := `<?
		if ($a) :
		endif;
	`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 2,
			EndLine:   3,
			StartPos:  5,
			EndPos:    23,
		},
		Stmts: []node.Node{
			&stmt.If{
				AltSyntax: true,
				Position: &position.Position{
					StartLine: 2,
					EndLine:   3,
					StartPos:  5,
					EndPos:    23,
				},
				Cond: &node.SimpleVar{
					Position: &position.Position{
						StartLine: 2,
						EndLine:   2,
						StartPos:  9,
						EndPos:    11,
					},
					Name: "a",
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

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestAltElseIf(t *testing.T) {
	src := `<?
		if ($a) :
		elseif ($b):
		endif;
	`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 2,
			EndLine:   4,
			StartPos:  5,
			EndPos:    38,
		},
		Stmts: []node.Node{
			&stmt.If{
				AltSyntax: true,
				Position: &position.Position{
					StartLine: 2,
					EndLine:   4,
					StartPos:  5,
					EndPos:    38,
				},
				Cond: &node.SimpleVar{
					Position: &position.Position{
						StartLine: 2,
						EndLine:   2,
						StartPos:  9,
						EndPos:    11,
					},
					Name: "a",
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
				ElseIf: []node.Node{
					&stmt.ElseIf{
						AltSyntax: true,
						Position: &position.Position{
							StartLine: 3,
							EndLine:   -1,
							StartPos:  17,
							EndPos:    -1,
						},
						Cond: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 3,
								EndLine:   3,
								StartPos:  25,
								EndPos:    27,
							},
							Name: "b",
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
			},
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestAltElse(t *testing.T) {
	src := `<?
		if ($a) :
		else:
		endif;
	`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 2,
			EndLine:   4,
			StartPos:  5,
			EndPos:    31,
		},
		Stmts: []node.Node{
			&stmt.If{
				AltSyntax: true,
				Position: &position.Position{
					StartLine: 2,
					EndLine:   4,
					StartPos:  5,
					EndPos:    31,
				},
				Cond: &node.SimpleVar{
					Position: &position.Position{
						StartLine: 2,
						EndLine:   2,
						StartPos:  9,
						EndPos:    11,
					},
					Name: "a",
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
				Else: &stmt.Else{
					AltSyntax: true,
					Position: &position.Position{
						StartLine: 3,
						EndLine:   -1,
						StartPos:  17,
						EndPos:    -1,
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
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestAltElseElseIf(t *testing.T) {
	src := `<?
		if ($a) :
		elseif ($b):
		elseif ($c):
		else:
		endif;
	`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 2,
			EndLine:   6,
			StartPos:  5,
			EndPos:    61,
		},
		Stmts: []node.Node{
			&stmt.If{
				AltSyntax: true,
				Position: &position.Position{
					StartLine: 2,
					EndLine:   6,
					StartPos:  5,
					EndPos:    61,
				},
				Cond: &node.SimpleVar{
					Position: &position.Position{
						StartLine: 2,
						EndLine:   2,
						StartPos:  9,
						EndPos:    11,
					},
					Name: "a",
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
				ElseIf: []node.Node{
					&stmt.ElseIf{
						AltSyntax: true,
						Position: &position.Position{
							StartLine: 3,
							EndLine:   -1,
							StartPos:  17,
							EndPos:    -1,
						},
						Cond: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 3,
								EndLine:   3,
								StartPos:  25,
								EndPos:    27,
							},
							Name: "b",
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
					&stmt.ElseIf{
						AltSyntax: true,
						Position: &position.Position{
							StartLine: 4,
							EndLine:   -1,
							StartPos:  32,
							EndPos:    -1,
						},
						Cond: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 4,
								EndLine:   4,
								StartPos:  40,
								EndPos:    42,
							},
							Name: "c",
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
				Else: &stmt.Else{
					AltSyntax: true,
					Position: &position.Position{
						StartLine: 5,
						EndLine:   -1,
						StartPos:  47,
						EndPos:    -1,
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
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}
