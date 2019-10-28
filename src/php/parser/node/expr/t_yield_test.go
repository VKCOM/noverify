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

func TestYield(t *testing.T) {
	src := `<? yield;`

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
				Expr: &expr.Yield{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    8,
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

func TestYieldVal(t *testing.T) {
	src := `<? yield $a;`

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
				Expr: &expr.Yield{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    11,
					},
					Value: &node.SimpleVar{
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
	}

	php7parser := php7.NewParser(bytes.NewBufferString(src), "test.php")
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestYieldKeyVal(t *testing.T) {
	src := `<? yield $a => $b;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    18,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    18,
				},
				Expr: &expr.Yield{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    17,
					},
					Key: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  10,
							EndPos:    11,
						},
						Name: "a",
					},
					Value: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  16,
							EndPos:    17,
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

func TestYieldExpr(t *testing.T) {
	src := `<? yield 1;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  4,
			EndPos:    11,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  4,
					EndPos:    11,
				},
				Expr: &expr.Yield{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    10,
					},
					Value: &scalar.Lnumber{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  10,
							EndPos:    10,
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

func TestYieldKeyExpr(t *testing.T) {
	src := `<? yield $a => 1;`

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
				Expr: &expr.Yield{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    16,
					},
					Key: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  10,
							EndPos:    11,
						},
						Name: "a",
					},
					Value: &scalar.Lnumber{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  16,
							EndPos:    16,
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

func TestYieldFrom(t *testing.T) {
	src := `<? yield from $a;`

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
				Expr: &expr.YieldFrom{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  4,
						EndPos:    16,
					},
					Expr: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  15,
							EndPos:    16,
						},
						Name: "a",
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
