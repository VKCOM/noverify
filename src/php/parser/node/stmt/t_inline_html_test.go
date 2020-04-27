package stmt_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest/assert"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

func TestInlineHtml(t *testing.T) {
	src := `<? ?> <div></div>`

	expected := &node.Root{
		Position: &position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  3,
			EndPos:    17,
		},
		Stmts: []node.Node{
			&stmt.Nop{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  3,
					EndPos:    5,
				},
			},
			&stmt.InlineHtml{
				Position: &position.Position{
					StartLine: 1,
					EndLine:   1,
					StartPos:  5,
					EndPos:    17,
				},
				Value: " <div></div>",
			},
		},
	}

	php7parser := php7.NewParser([]byte(src))
	php7parser.Parse()
	actual := php7parser.GetRootNode()
	assert.DeepEqual(t, expected, actual)
}
