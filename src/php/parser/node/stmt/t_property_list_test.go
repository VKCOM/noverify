package stmt_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest/assert"
	"github.com/VKCOM/noverify/src/php/parser/node/name"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestProperty(t *testing.T) {
	src := `<? class foo {var $a;}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    22,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    22,
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
					&stmt.PropertyList{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  14,
							EndPos:    21,
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  14,
									EndPos:    17,
								},
								Value: "var",
							},
						},
						Properties: []node.Node{
							&stmt.Property{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  18,
									EndPos:    20,
								},
								PhpDocComment: "",
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  18,
										EndPos:    20,
									},
									Name: "a",
								},
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

func TestProperties(t *testing.T) {
	src := `<? class foo {public static $a, $b = 1;}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    40,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    40,
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
					&stmt.PropertyList{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  14,
							EndPos:    39,
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
						Properties: []node.Node{
							&stmt.Property{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  28,
									EndPos:    30,
								},
								PhpDocComment: "",
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  28,
										EndPos:    30,
									},
									Name: "a",
								},
							},
							&stmt.Property{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  32,
									EndPos:    38,
								},
								PhpDocComment: "",
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  32,
										EndPos:    34,
									},
									Name: "b",
								},
								Expr: &scalar.Lnumber{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  37,
										EndPos:    38,
									},
									Value: "1",
								},
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

func TestProperties2(t *testing.T) {
	src := `<? class foo {public static $a = 1, $b;}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    40,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    40,
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
					&stmt.PropertyList{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  14,
							EndPos:    39,
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
						Properties: []node.Node{
							&stmt.Property{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  28,
									EndPos:    34,
								},
								PhpDocComment: "",
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  28,
										EndPos:    30,
									},
									Name: "a",
								},
								Expr: &scalar.Lnumber{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  33,
										EndPos:    34,
									},
									Value: "1",
								},
							},
							&stmt.Property{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  36,
									EndPos:    38,
								},
								PhpDocComment: "",
								Variable: &node.SimpleVar{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
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
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}

func TestPropertyType(t *testing.T) {
	src := `<? class foo {var bar $a;}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    26,
		},
		Stmts: []node.Node{
			&stmt.Class{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    26,
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
					&stmt.PropertyList{
						Position: &position.Position{
							StartLine: 1,
							EndLine:   1,
							StartPos:  14,
							EndPos:    25,
						},
						Modifiers: []*node.Identifier{
							{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  14,
									EndPos:    17,
								},
								Value: "var",
							},
						},
						Type: &name.Name{
							Position: &position.Position{
								StartLine: 1,
								EndLine:   1,
								StartPos:  18,
								EndPos:    21,
							},
							Parts: []node.Node{
								&name.NamePart{
									Position: &position.Position{
										StartLine: 1,
										EndLine:   1,
										StartPos:  18,
										EndPos:    21,
									},
									Value: "bar",
								},
							},
						},
						Properties: []node.Node{
							&stmt.Property{
								Position: &position.Position{
									StartLine: 1,
									EndLine:   1,
									StartPos:  22,
									EndPos:    24,
								},
								PhpDocComment: "",
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
