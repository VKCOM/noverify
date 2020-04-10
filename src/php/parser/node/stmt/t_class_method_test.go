package stmt_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/position"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
)

func TestSimpleClassMethod(t *testing.T) {
	src := `<? class foo{ function bar() {} }`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    33,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    33,
				},
				PhpDocComment: "",
				ClassName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  9,
						EndPos:    12,
					},
					Value: "foo",
				},
				Stmts: []node.Node{
					&stmt.ClassMethod{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  14,
							EndPos:    31,
						},
						ReturnsRef:    false,
						PhpDocComment: "",
						MethodName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  23,
								EndPos:    26,
							},
							Value: "bar",
						},
						Stmt: &stmt.StmtList{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  29,
								EndPos:    31,
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

func TestPrivateProtectedClassMethod(t *testing.T) {
	src := `<? class foo{ final private function bar() {} protected function baz() {} }`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    75,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    75,
				},
				PhpDocComment: "",
				ClassName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  9,
						EndPos:    12,
					},
					Value: "foo",
				},
				Stmts: []node.Node{
					&stmt.ClassMethod{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  14,
							EndPos:    45,
						},
						ReturnsRef:    false,
						PhpDocComment: "",
						MethodName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  37,
								EndPos:    40,
							},
							Value: "bar",
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  14,
									EndPos:    19,
								},
								Value: "final",
							},
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  20,
									EndPos:    27,
								},
								Value: "private",
							},
						},
						Stmt: &stmt.StmtList{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  43,
								EndPos:    45,
							},
							Stmts: []node.Node{},
						},
					},
					&stmt.ClassMethod{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  46,
							EndPos:    73,
						},
						ReturnsRef:    false,
						PhpDocComment: "",
						MethodName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  65,
								EndPos:    68,
							},
							Value: "baz",
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  46,
									EndPos:    55,
								},
								Value: "protected",
							},
						},
						Stmt: &stmt.StmtList{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  71,
								EndPos:    73,
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

func TestPhp7ClassMethod(t *testing.T) {
	src := `<? class foo{ public static function &bar(): void {} }`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    54,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    54,
				},
				PhpDocComment: "",
				ClassName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  9,
						EndPos:    12,
					},
					Value: "foo",
				},
				Stmts: []node.Node{
					&stmt.ClassMethod{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  14,
							EndPos:    52,
						},
						ReturnsRef:    true,
						PhpDocComment: "",
						MethodName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  38,
								EndPos:    41,
							},
							Value: "bar",
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  14,
									EndPos:    20,
								},
								Value: "public",
							},
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  21,
									EndPos:    27,
								},
								Value: "static",
							},
						},
						ReturnType: &name.Name{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  45,
								EndPos:    49,
							},
							Parts: []node.Node{
								&name.NamePart{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  45,
										EndPos:    49,
									},
									Value: "void",
								},
							},
						},
						Stmt: &stmt.StmtList{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  50,
								EndPos:    52,
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

func TestAbstractClassMethod(t *testing.T) {
	src := `<? abstract class Foo{ abstract public function bar(); }`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    56,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    56,
				},
				PhpDocComment: "",
				ClassName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  18,
						EndPos:    21,
					},
					Value: "Foo",
				},
				Modifiers: []*node.Identifier{
					{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  3,
							EndPos:    11,
						},
						Value: "abstract",
					},
				},
				Stmts: []node.Node{
					&stmt.ClassMethod{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  23,
							EndPos:    54,
						},
						ReturnsRef:    false,
						PhpDocComment: "",
						MethodName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  48,
								EndPos:    51,
							},
							Value: "bar",
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  23,
									EndPos:    31,
								},
								Value: "abstract",
							},
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  32,
									EndPos:    38,
								},
								Value: "public",
							},
						},
						Stmt: &stmt.Nop{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  53,
								EndPos:    54,
							},
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

func TestPhp7AbstractClassMethod(t *testing.T) {
	src := `<? abstract class Foo{ public function bar(): void; }`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    53,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    53,
				},
				PhpDocComment: "",
				ClassName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  18,
						EndPos:    21,
					},
					Value: "Foo",
				},
				Modifiers: []*node.Identifier{
					{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  3,
							EndPos:    11,
						},
						Value: "abstract",
					},
				},
				Stmts: []node.Node{
					&stmt.ClassMethod{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  23,
							EndPos:    51,
						},
						ReturnsRef:    false,
						PhpDocComment: "",
						MethodName: &node.Identifier{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  39,
								EndPos:    42,
							},
							Value: "bar",
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  23,
									EndPos:    29,
								},
								Value: "public",
							},
						},
						ReturnType: &name.Name{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  46,
								EndPos:    50,
							},
							Parts: []node.Node{
								&name.NamePart{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  46,
										EndPos:    50,
									},
									Value: "void",
								},
							},
						},
						Stmt: &stmt.Nop{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  50,
								EndPos:    51,
							},
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
