package stmt_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/position"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
)

func TestSimpleFunction(t *testing.T) {
	src := `<? function foo() {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    20,
		},
		Stmts: []node.Node{
			&stmt.Function{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    20,
				},
				ReturnsRef:    false,
				PhpDocComment: "",
				FunctionName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  12,
						EndPos:    15,
					},
					Value: "foo",
				},
				Stmts: []node.Node{},
			},
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestFunctionReturn(t *testing.T) {
	src := `<? function foo() {return;}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    27,
		},
		Stmts: []node.Node{
			&stmt.Function{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    27,
				},
				ReturnsRef:    false,
				PhpDocComment: "",
				FunctionName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  12,
						EndPos:    15,
					},
					Value: "foo",
				},
				Stmts: []node.Node{
					&stmt.Return{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  19,
							EndPos:    26,
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

func TestFunctionReturnVar(t *testing.T) {
	src := `<? function foo(array $a, callable $b) {return $a;}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    51,
		},
		Stmts: []node.Node{
			&stmt.Function{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    51,
				},
				ReturnsRef:    false,
				PhpDocComment: "",
				FunctionName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  12,
						EndPos:    15,
					},
					Value: "foo",
				},
				Params: []node.Node{
					&node.Parameter{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  16,
							EndPos:    24,
						},
						ByRef:    false,
						Variadic: false,
						VariableType: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  16,
								EndPos:    21,
							},
							Value: "array",
						},
						Variable: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  22,
								EndPos:    24,
							},
							Name: "a",
						},
					},
					&node.Parameter{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  26,
							EndPos:    37,
						},
						ByRef:    false,
						Variadic: false,
						VariableType: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  26,
								EndPos:    34,
							},
							Value: "callable",
						},
						Variable: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  35,
								EndPos:    37,
							},
							Name: "b",
						},
					},
				},
				Stmts: []node.Node{
					&stmt.Return{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  40,
							EndPos:    50,
						},
						Expr: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  47,
								EndPos:    49,
							},
							Name: "a",
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

func TestRefFunction(t *testing.T) {
	src := `<? function &foo() {return 1;}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    30,
		},
		Stmts: []node.Node{
			&stmt.Function{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    30,
				},
				ReturnsRef:    true,
				PhpDocComment: "",
				FunctionName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  13,
						EndPos:    16,
					},
					Value: "foo",
				},
				Stmts: []node.Node{
					&stmt.Return{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  20,
							EndPos:    29,
						},
						Expr: &scalar.Lnumber{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  27,
								EndPos:    28,
							},
							Value: "1",
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

func TestReturnTypeFunction(t *testing.T) {
	src := `<? function &foo(): void {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    27,
		},
		Stmts: []node.Node{
			&stmt.Function{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    27,
				},
				PhpDocComment: "",
				ReturnsRef:    true,
				FunctionName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  13,
						EndPos:    16,
					},
					Value: "foo",
				},
				ReturnType: &name.Name{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  20,
						EndPos:    24,
					},
					Parts: []node.Node{
						&name.NamePart{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  20,
								EndPos:    24,
							},
							Value: "void",
						},
					},
				},
				Stmts: []node.Node{},
			},
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}
