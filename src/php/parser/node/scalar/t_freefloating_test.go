package scalar_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest/assert"

	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
)

var expected freefloating.Collection = freefloating.Collection{
	freefloating.Start: []freefloating.String{
		{
			StringType: freefloating.WhiteSpaceType,
			Value:      "    ",
			Position:   nil,
		},
		{
			StringType: freefloating.CommentType,
			Value:      "//comment\n",
			Position:   nil,
		},
	},
}

var nodes = []node.Node{
	&scalar.Dnumber{
		FreeFloating: expected,
	},
	&scalar.EncapsedStringPart{
		FreeFloating: expected,
	},
	&scalar.Encapsed{
		FreeFloating: expected,
	},
	&scalar.Heredoc{
		FreeFloating: expected,
	},
	&scalar.Lnumber{
		FreeFloating: expected,
	},
	&scalar.MagicConstant{
		FreeFloating: expected,
	},
	&scalar.String{
		FreeFloating: expected,
	},
}

func TestMeta(t *testing.T) {
	for _, n := range nodes {
		actual := *n.GetFreeFloating()
		assert.DeepEqual(t, expected, actual)
	}
}
