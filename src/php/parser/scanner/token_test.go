package scanner_test

import (
	"reflect"
	"testing"

	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/scanner"
)

func TestToken(t *testing.T) {
	tkn := &scanner.Token{
		Value: `foo`,
		Position: position.Position{
			StartLine: 1,
			EndLine:   1,
			StartPos:  0,
			EndPos:    3,
		},
	}

	c := []freefloating.String{
		{
			Value:      "test comment",
			StringType: freefloating.CommentType,
			Position:   nil,
		},
	}

	tkn.FreeFloating = c

	if !reflect.DeepEqual(tkn.FreeFloating, c) {
		t.Errorf("comments are not equal\n")
	}

	if tkn.String() != `foo` {
		t.Errorf("token value is not equal\n")
	}
}
