package stmt_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestTrait(t *testing.T) {
	src := `<? trait Foo {}`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    15,
		},
		Stmts: []node.Node{
			&stmt.Trait{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    15,
				},
				PhpDocComment: "",
				TraitName: &node.Identifier{
					Position: &position.Position{
						StartLine: 1,
						EndLine:   1,
						StartPos:  9,
						EndPos:    12,
					},
					Value: "Foo",
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
