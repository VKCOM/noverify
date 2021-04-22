package inline

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestInline(t *testing.T) {
	linttest.RunInlineTest(t, "./testdata")
}
