package node_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/position"

	"github.com/VKCOM/noverify/src/php/parser/node/expr"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
)

func TestIdentifier(t *testing.T) {
	src := `<? $foo;`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    8,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    8,
				},
				Expr: &node.SimpleVar{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  3,
						EndPos:    7,
					},
					Name: "foo",
				},
			},
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestPhp7ArgumentNode(t *testing.T) {
	src := `<? 
		foo($a, ...$b);
		$foo($a, ...$b);
		$foo->bar($a, ...$b);
		foo::bar($a, ...$b);
		$foo::bar($a, ...$b);
		new foo($a, ...$b);
		/** anonymous class */
		new class ($a, ...$b) {};
	`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 2,
			EndLine:   9,
			StartPos:  6,
			EndPos:    186,
		},
		Stmts: []node.Node{
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 2,
					EndLine:   2,
					StartPos:  6,
					EndPos:    21,
				},
				Expr: &expr.FunctionCall{
					Position: &position.Position{
						StartLine: 2,
						EndLine:   2,
						StartPos:  6,
						EndPos:    20,
					},
					Function: &name.Name{
						Position: &position.Position{
							StartLine: 2,
							EndLine:   2,
							StartPos:  6,
							EndPos:    9,
						},
						Parts: []node.Node{
							&name.NamePart{
								Position: &position.Position{
									StartLine: 2,
									EndLine:   2,
									StartPos:  6,
									EndPos:    9,
								},
								Value: "foo",
							},
						},
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 2,
							EndLine:   2,
							StartPos:  9,
							EndPos:    20,
						},
						Arguments: []node.Node{
							&node.Argument{
								Position: &position.Position{
									StartLine: 2,
									EndLine:   2,
									StartPos:  10,
									EndPos:    12,
								},
								Variadic:    false,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 2,
										EndLine:   2,
										StartPos:  10,
										EndPos:    12,
									},
									Name: "a",
								},
							},
							&node.Argument{
								Position: &position.Position{
									StartLine: 2,
									EndLine:   2,
									StartPos:  14,
									EndPos:    19,
								},
								Variadic:    true,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 2,
										EndLine:   2,
										StartPos:  17,
										EndPos:    19,
									},
									Name: "b",
								},
							},
						},
					},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 3,
					EndLine:   3,
					StartPos:  24,
					EndPos:    40,
				},
				Expr: &expr.FunctionCall{
					Position: &position.Position{
						StartLine: 3,
						EndLine:   3,
						StartPos:  24,
						EndPos:    39,
					},
					Function: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 3,
							EndLine:   3,
							StartPos:  24,
							EndPos:    28,
						},
						Name: "foo",
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 3,
							EndLine:   3,
							StartPos:  28,
							EndPos:    39,
						},
						Arguments: []node.Node{
							&node.Argument{
								Position: &position.Position{
									StartLine: 3,
									EndLine:   3,
									StartPos:  29,
									EndPos:    31,
								},
								Variadic:    false,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 3,
										EndLine:   3,
										StartPos:  29,
										EndPos:    31,
									},
									Name: "a",
								},
							},
							&node.Argument{
								Position: &position.Position{
									StartLine: 3,
									EndLine:   3,
									StartPos:  33,
									EndPos:    38,
								},
								Variadic:    true,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 3,
										EndLine:   3,
										StartPos:  36,
										EndPos:    38,
									},
									Name: "b",
								},
							},
						},
					},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 4,
					EndLine:   4,
					StartPos:  43,
					EndPos:    64,
				},
				Expr: &expr.MethodCall{
					Position: &position.Position{
						StartLine: 4,
						EndLine:   4,
						StartPos:  43,
						EndPos:    63,
					},
					Variable: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 4,
							EndLine:   4,
							StartPos:  43,
							EndPos:    47,
						},
						Name: "foo",
					},
					Method: &node.Identifier{
						Position: &position.Position{
							StartLine: 4,
							EndLine:   4,
							StartPos:  49,
							EndPos:    52,
						},
						Value: "bar",
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 4,
							EndLine:   4,
							StartPos:  52,
							EndPos:    63,
						},
						Arguments: []node.Node{
							&node.Argument{
								Position: &position.Position{
									StartLine: 4,
									EndLine:   4,
									StartPos:  53,
									EndPos:    55,
								},
								IsReference: false,
								Variadic:    false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 4,
										EndLine:   4,
										StartPos:  53,
										EndPos:    55,
									},
									Name: "a",
								},
							},
							&node.Argument{
								Position: &position.Position{
									StartLine: 4,
									EndLine:   4,
									StartPos:  57,
									EndPos:    62,
								},
								Variadic:    true,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 4,
										EndLine:   4,
										StartPos:  60,
										EndPos:    62,
									},
									Name: "b",
								},
							},
						},
					},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 5,
					EndLine:   5,
					StartPos:  67,
					EndPos:    87,
				},
				Expr: &expr.StaticCall{
					Position: &position.Position{
						StartLine: 5,
						EndLine:   5,
						StartPos:  67,
						EndPos:    86,
					},
					Class: &name.Name{
						Position: &position.Position{
							StartLine: 5,
							EndLine:   5,
							StartPos:  67,
							EndPos:    70,
						},
						Parts: []node.Node{
							&name.NamePart{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  67,
									EndPos:    70,
								},
								Value: "foo",
							},
						},
					},
					Call: &node.Identifier{
						Position: &position.Position{
							StartLine: 5,
							EndLine:   5,
							StartPos:  72,
							EndPos:    75,
						},
						Value: "bar",
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 5,
							EndLine:   5,
							StartPos:  75,
							EndPos:    86,
						},
						Arguments: []node.Node{
							&node.Argument{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  76,
									EndPos:    78,
								},
								Variadic:    false,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 5,
										EndLine:   5,
										StartPos:  76,
										EndPos:    78,
									},
									Name: "a",
								},
							},
							&node.Argument{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  80,
									EndPos:    85,
								},
								Variadic:    true,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 5,
										EndLine:   5,
										StartPos:  83,
										EndPos:    85,
									},
									Name: "b",
								},
							},
						},
					},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 6,
					EndLine:   6,
					StartPos:  90,
					EndPos:    111,
				},
				Expr: &expr.StaticCall{
					Position: &position.Position{
						StartLine: 6,
						EndLine:   6,
						StartPos:  90,
						EndPos:    110,
					},
					Class: &node.SimpleVar{
						Position: &position.Position{
							StartLine: 6,
							EndLine:   6,
							StartPos:  90,
							EndPos:    94,
						},
						Name: "foo",
					},
					Call: &node.Identifier{
						Position: &position.Position{
							StartLine: 6,
							EndLine:   6,
							StartPos:  96,
							EndPos:    99,
						},
						Value: "bar",
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 6,
							EndLine:   6,
							StartPos:  99,
							EndPos:    110,
						},
						Arguments: []node.Node{
							&node.Argument{
								Position: &position.Position{
									StartLine: 6,
									EndLine:   6,
									StartPos:  100,
									EndPos:    102,
								},
								Variadic:    false,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 6,
										EndLine:   6,
										StartPos:  100,
										EndPos:    102,
									},
									Name: "a",
								},
							},
							&node.Argument{
								Position: &position.Position{
									StartLine: 6,
									EndLine:   6,
									StartPos:  104,
									EndPos:    109,
								},
								Variadic:    true,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 6,
										EndLine:   6,
										StartPos:  107,
										EndPos:    109,
									},
									Name: "b",
								},
							},
						},
					},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 7,
					EndLine:   7,
					StartPos:  114,
					EndPos:    133,
				},
				Expr: &expr.New{
					Position: &position.Position{
						StartLine: 7,
						EndLine:   7,
						StartPos:  114,
						EndPos:    132,
					},
					Class: &name.Name{
						Position: &position.Position{
							StartLine: 7,
							EndLine:   7,
							StartPos:  118,
							EndPos:    121,
						},
						Parts: []node.Node{
							&name.NamePart{
								Position: &position.Position{
									StartLine: 7,
									EndLine:   7,
									StartPos:  118,
									EndPos:    121,
								},
								Value: "foo",
							},
						},
					},
					ArgumentList: &node.ArgumentList{
						Position: &position.Position{
							StartLine: 7,
							EndLine:   7,
							StartPos:  121,
							EndPos:    132,
						},
						Arguments: []node.Node{
							&node.Argument{
								Position: &position.Position{
									StartLine: 7,
									EndLine:   7,
									StartPos:  122,
									EndPos:    124,
								},
								Variadic:    false,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 7,
										EndLine:   7,
										StartPos:  122,
										EndPos:    124,
									},
									Name: "a",
								},
							},
							&node.Argument{
								Position: &position.Position{
									StartLine: 7,
									EndLine:   7,
									StartPos:  126,
									EndPos:    131,
								},
								Variadic:    true,
								IsReference: false,
								Expr: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 7,
										EndLine:   7,
										StartPos:  129,
										EndPos:    131,
									},
									Name: "b",
								},
							},
						},
					},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 9,
					EndLine:   9,
					StartPos:  161,
					EndPos:    186,
				},
				Expr: &expr.New{
					Position: &position.Position{
						StartLine: 9,
						EndLine:   9,
						StartPos:  161,
						EndPos:    185,
					},
					Class: &stmt.Class{
						Position: &position.Position{
							StartLine: 9,
							EndLine:   9,
							StartPos:  165,
							EndPos:    185,
						},
						PhpDocComment: "/** anonymous class */",
						ArgumentList: &node.ArgumentList{
							Position: &position.Position{
								StartLine: 9,
								EndLine:   9,
								StartPos:  171,
								EndPos:    182,
							},
							Arguments: []node.Node{
								&node.Argument{
									Position: &position.Position{
										StartLine: 9,
										EndLine:   9,
										StartPos:  172,
										EndPos:    174,
									},
									Variadic:    false,
									IsReference: false,
									Expr: &node.SimpleVar{
										Position: &position.Position{
											StartLine: 9,
											EndLine:   9,
											StartPos:  172,
											EndPos:    174,
										},
										Name: "a",
									},
								},
								&node.Argument{
									Position: &position.Position{
										StartLine: 9,
										EndLine:   9,
										StartPos:  176,
										EndPos:    181,
									},
									Variadic:    true,
									IsReference: false,
									Expr: &node.SimpleVar{
										Position: &position.Position{
											StartLine: 9,
											EndLine:   9,
											StartPos:  179,
											EndPos:    181,
										},
										Name: "b",
									},
								},
							},
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

func TestPhp7ParameterNode(t *testing.T) {
	src := `<? 
		function foo(?bar $bar=null, baz &...$baz) {}
		class foo {public function foo(?bar $bar=null, baz &...$baz) {}}
		function(?bar $bar=null, baz &...$baz) {};
		static function(?bar $bar=null, baz &...$baz) {};
	`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 2,
			EndLine:   5,
			StartPos:  6,
			EndPos:    215,
		},
		Stmts: []node.Node{
			&stmt.Function{
				Position: &position.Position{
					StartLine: 2,
					EndLine:   2,
					StartPos:  6,
					EndPos:    51,
				},
				ReturnsRef:    false,
				PhpDocComment: "",
				FunctionName: &node.Identifier{
					Position: &position.Position{
						StartLine: 2,
						EndLine:   2,
						StartPos:  15,
						EndPos:    18,
					},
					Value: "foo",
				},
				Params: []node.Node{
					&node.Parameter{
						Position: &position.Position{
							StartLine: 2,
							EndLine:   2,
							StartPos:  19,
							EndPos:    33,
						},
						ByRef:    false,
						Variadic: false,
						VariableType: &node.Nullable{
							Position: &position.Position{
								StartLine: 2,
								EndLine:   2,
								StartPos:  19,
								EndPos:    23,
							},
							Expr: &name.Name{
								Position: &position.Position{
									StartLine: 2,
									EndLine:   2,
									StartPos:  20,
									EndPos:    23,
								},
								Parts: []node.Node{
									&name.NamePart{
										Position: &position.Position{
											StartLine: 2,
											EndLine:   2,
											StartPos:  20,
											EndPos:    23,
										},
										Value: "bar",
									},
								},
							},
						},
						Variable: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 2,
								EndLine:   2,
								StartPos:  24,
								EndPos:    28,
							},
							Name: "bar",
						},
						DefaultValue: &expr.ConstFetch{
							Position: &position.Position{
								StartLine: 2,
								EndLine:   2,
								StartPos:  29,
								EndPos:    33,
							},
							Constant: &name.Name{
								Position: &position.Position{
									StartLine: 2,
									EndLine:   2,
									StartPos:  29,
									EndPos:    33,
								},
								Parts: []node.Node{
									&name.NamePart{
										Position: &position.Position{
											StartLine: 2,
											EndLine:   2,
											StartPos:  29,
											EndPos:    33,
										},
										Value: "null",
									},
								},
							},
						},
					},
					&node.Parameter{
						Position: &position.Position{
							StartLine: 2,
							EndLine:   2,
							StartPos:  35,
							EndPos:    47,
						},
						ByRef:    true,
						Variadic: true,
						VariableType: &name.Name{
							Position: &position.Position{
								StartLine: 2,
								EndLine:   2,
								StartPos:  35,
								EndPos:    38,
							},
							Parts: []node.Node{
								&name.NamePart{
									Position: &position.Position{
										StartLine: 2,
										EndLine:   2,
										StartPos:  35,
										EndPos:    38,
									},
									Value: "baz",
								},
							},
						},
						Variable: &node.SimpleVar{
							Position: &position.Position{
								StartLine: 2,
								EndLine:   2,
								StartPos:  43,
								EndPos:    47,
							},
							Name: "baz",
						},
					},
				},
				Stmts: []node.Node{},
			},
			&stmt.Class{
				Position: &position.Position{
					StartLine: 3,
					EndLine:   3,
					StartPos:  54,
					EndPos:    118,
				},
				PhpDocComment: "",
				ClassName: &node.Identifier{
					Position: &position.Position{
						StartLine: 3,
						EndLine:   3,
						StartPos:  60,
						EndPos:    63,
					},
					Value: "foo",
				},
				Stmts: []node.Node{
					&stmt.ClassMethod{
						Position: &position.Position{
							StartLine: 3,
							EndLine:   3,
							StartPos:  65,
							EndPos:    117,
						},
						PhpDocComment: "",
						ReturnsRef:    false,
						MethodName: &node.Identifier{
							Position: &position.Position{
								StartLine: 3,
								EndLine:   3,
								StartPos:  81,
								EndPos:    84,
							},
							Value: "foo",
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 3,
									EndLine:   3,
									StartPos:  65,
									EndPos:    71,
								},
								Value: "public",
							},
						},
						Params: []node.Node{
							&node.Parameter{
								Position: &position.Position{
									StartLine: 3,
									EndLine:   3,
									StartPos:  85,
									EndPos:    99,
								},
								ByRef:    false,
								Variadic: false,
								VariableType: &node.Nullable{
									Position: &position.Position{
										StartLine: 3,
										EndLine:   3,
										StartPos:  85,
										EndPos:    89,
									},
									Expr: &name.Name{
										Position: &position.Position{
											StartLine: 3,
											EndLine:   3,
											StartPos:  86,
											EndPos:    89,
										},
										Parts: []node.Node{
											&name.NamePart{
												Position: &position.Position{
													StartLine: 3,
													EndLine:   3,
													StartPos:  86,
													EndPos:    89,
												},
												Value: "bar",
											},
										},
									},
								},
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 3,
										EndLine:   3,
										StartPos:  90,
										EndPos:    94,
									},
									Name: "bar",
								},
								DefaultValue: &expr.ConstFetch{
									Position: &position.Position{
										StartLine: 3,
										EndLine:   3,
										StartPos:  95,
										EndPos:    99,
									},
									Constant: &name.Name{
										Position: &position.Position{
											StartLine: 3,
											EndLine:   3,
											StartPos:  95,
											EndPos:    99,
										},
										Parts: []node.Node{
											&name.NamePart{
												Position: &position.Position{
													StartLine: 3,
													EndLine:   3,
													StartPos:  95,
													EndPos:    99,
												},
												Value: "null",
											},
										},
									},
								},
							},
							&node.Parameter{
								Position: &position.Position{
									StartLine: 3,
									EndLine:   3,
									StartPos:  101,
									EndPos:    113,
								},
								ByRef:    true,
								Variadic: true,
								VariableType: &name.Name{
									Position: &position.Position{
										StartLine: 3,
										EndLine:   3,
										StartPos:  101,
										EndPos:    104,
									},
									Parts: []node.Node{
										&name.NamePart{
											Position: &position.Position{
												StartLine: 3,
												EndLine:   3,
												StartPos:  101,
												EndPos:    104,
											},
											Value: "baz",
										},
									},
								},
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 3,
										EndLine:   3,
										StartPos:  109,
										EndPos:    113,
									},
									Name: "baz",
								},
							},
						},
						Stmt: &stmt.StmtList{
							Position: &position.Position{
								StartLine: 3,
								EndLine:   3,
								StartPos:  115,
								EndPos:    117,
							},
							Stmts: []node.Node{},
						},
					},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 4,
					EndLine:   4,
					StartPos:  121,
					EndPos:    163,
				},
				Expr: &expr.Closure{
					Position: &position.Position{
						StartLine: 4,
						EndLine:   4,
						StartPos:  121,
						EndPos:    162,
					},
					ReturnsRef:    false,
					Static:        false,
					PhpDocComment: "",
					Params: []node.Node{
						&node.Parameter{
							Position: &position.Position{
								StartLine: 4,
								EndLine:   4,
								StartPos:  130,
								EndPos:    144,
							},
							ByRef:    false,
							Variadic: false,
							VariableType: &node.Nullable{
								Position: &position.Position{
									StartLine: 4,
									EndLine:   4,
									StartPos:  130,
									EndPos:    134,
								},
								Expr: &name.Name{
									Position: &position.Position{
										StartLine: 4,
										EndLine:   4,
										StartPos:  131,
										EndPos:    134,
									},
									Parts: []node.Node{
										&name.NamePart{
											Position: &position.Position{
												StartLine: 4,
												EndLine:   4,
												StartPos:  131,
												EndPos:    134,
											},
											Value: "bar",
										},
									},
								},
							},
							Variable: &node.SimpleVar{
								Position: &position.Position{
									StartLine: 4,
									EndLine:   4,
									StartPos:  135,
									EndPos:    139,
								},
								Name: "bar",
							},
							DefaultValue: &expr.ConstFetch{
								Position: &position.Position{
									StartLine: 4,
									EndLine:   4,
									StartPos:  140,
									EndPos:    144,
								},
								Constant: &name.Name{
									Position: &position.Position{
										StartLine: 4,
										EndLine:   4,
										StartPos:  140,
										EndPos:    144,
									},
									Parts: []node.Node{
										&name.NamePart{
											Position: &position.Position{
												StartLine: 4,
												EndLine:   4,
												StartPos:  140,
												EndPos:    144,
											},
											Value: "null",
										},
									},
								},
							},
						},
						&node.Parameter{
							Position: &position.Position{
								StartLine: 4,
								EndLine:   4,
								StartPos:  146,
								EndPos:    158,
							},
							Variadic: true,
							ByRef:    true,
							VariableType: &name.Name{
								Position: &position.Position{
									StartLine: 4,
									EndLine:   4,
									StartPos:  146,
									EndPos:    149,
								},
								Parts: []node.Node{
									&name.NamePart{
										Position: &position.Position{
											StartLine: 4,
											EndLine:   4,
											StartPos:  146,
											EndPos:    149,
										},
										Value: "baz",
									},
								},
							},
							Variable: &node.SimpleVar{
								Position: &position.Position{
									StartLine: 4,
									EndLine:   4,
									StartPos:  154,
									EndPos:    158,
								},
								Name: "baz",
							},
						},
					},
					Stmts: []node.Node{},
				},
			},
			&stmt.Expression{
				Position: &position.Position{
					StartLine: 5,
					EndLine:   5,
					StartPos:  166,
					EndPos:    215,
				},
				Expr: &expr.Closure{
					Position: &position.Position{
						StartLine: 5,
						EndLine:   5,
						StartPos:  166,
						EndPos:    214,
					},
					Static:        true,
					PhpDocComment: "",
					ReturnsRef:    false,
					Params: []node.Node{
						&node.Parameter{
							Position: &position.Position{
								StartLine: 5,
								EndLine:   5,
								StartPos:  182,
								EndPos:    196,
							},
							ByRef:    false,
							Variadic: false,
							VariableType: &node.Nullable{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  182,
									EndPos:    186,
								},
								Expr: &name.Name{
									Position: &position.Position{
										StartLine: 5,
										EndLine:   5,
										StartPos:  183,
										EndPos:    186,
									},
									Parts: []node.Node{
										&name.NamePart{
											Position: &position.Position{
												StartLine: 5,
												EndLine:   5,
												StartPos:  183,
												EndPos:    186,
											},
											Value: "bar",
										},
									},
								},
							},
							Variable: &node.SimpleVar{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  187,
									EndPos:    191,
								},
								Name: "bar",
							},
							DefaultValue: &expr.ConstFetch{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  192,
									EndPos:    196,
								},
								Constant: &name.Name{
									Position: &position.Position{
										StartLine: 5,
										EndLine:   5,
										StartPos:  192,
										EndPos:    196,
									},
									Parts: []node.Node{
										&name.NamePart{
											Position: &position.Position{
												StartLine: 5,
												EndLine:   5,
												StartPos:  192,
												EndPos:    196,
											},
											Value: "null",
										},
									},
								},
							},
						},
						&node.Parameter{
							Position: &position.Position{
								StartLine: 5,
								EndLine:   5,
								StartPos:  198,
								EndPos:    210,
							},
							Variadic: true,
							ByRef:    true,
							VariableType: &name.Name{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  198,
									EndPos:    201,
								},
								Parts: []node.Node{
									&name.NamePart{
										Position: &position.Position{
											StartLine: 5,
											EndLine:   5,
											StartPos:  198,
											EndPos:    201,
										},
										Value: "baz",
									},
								},
							},
							Variable: &node.SimpleVar{
								Position: &position.Position{
									StartLine: 5,
									EndLine:   5,
									StartPos:  206,
									EndPos:    210,
								},
								Name: "baz",
							},
						},
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

func TestCommentEndFile(t *testing.T) {
	src := `<? //comment at the end)`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: -1,
			EndLine:   -1,
			StartPos:  -1,
			EndPos:    -1,
		},
		Stmts: []node.Node{},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}
