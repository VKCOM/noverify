package name_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/position"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
)

func TestName(t *testing.T) {
	src := `<? foo();`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    9,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    9,
				},
				Expr: &expr.FunctionCall{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  3,
						EndPos:    8,
					},
					Function: &name.Name{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  3,
							EndPos:    6,
						},
						Parts: []node.Node{
							&name.NamePart{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  3,
									EndPos:    6,
								},
								Value: "foo",
							},
						},
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  6,
							EndPos:    8,
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

func TestFullyQualified(t *testing.T) {
	src := `<? \foo();`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    10,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    10,
				},
				Expr: &expr.FunctionCall{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  3,
						EndPos:    9,
					},
					Function: &name.FullyQualified{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  3,
							EndPos:    7,
						},
						Parts: []node.Node{
							&name.NamePart{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  4,
									EndPos:    7,
								},
								Value: "foo",
							},
						},
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  7,
							EndPos:    9,
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

func TestRelative(t *testing.T) {
	src := `<? namespace\foo();`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    19,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    19,
				},
				Expr: &expr.FunctionCall{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  3,
						EndPos:    18,
					},
					Function: &name.Relative{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  3,
							EndPos:    16,
						},
						Parts: []node.Node{
							&name.NamePart{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  13,
									EndPos:    16,
								},
								Value: "foo",
							},
						},
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  16,
							EndPos:    18,
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

func TestNamePartsGetter(t *testing.T) {
	expected := []node.Node{
		&name.NamePart{Value: "a"},
		&name.NamePart{Value: "b"},
	}

	plainName := &name.Name{Parts: expected}
	relativeName := &name.Relative{Parts: expected}
	fullyQualifiedName := &name.FullyQualified{Parts: expected}

	assert.DeepEqual(t, expected, plainName.GetParts())
	assert.DeepEqual(t, expected, relativeName.GetParts())
	assert.DeepEqual(t, expected, fullyQualifiedName.GetParts())
}
