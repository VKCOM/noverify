package lintdoc

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/linter"
)

func TestRenderCheckDocumentation(t *testing.T) {
	runTest := func(info linter.CheckInfo, expect string) {
		t.Helper()
		var buf strings.Builder
		err := RenderCheckDocumentation(&buf, info)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if diff := cmp.Diff(buf.String(), expect); diff != "" {
			t.Errorf("%#v:\n%s", info, diff)
		}
	}

	runTest(linter.CheckInfo{
		Name:    "shortExample",
		Comment: "Report nothing, but test short info rendering.",
	}, `shortExample checker documentation

Report nothing, but test short info rendering.`)

	runTest(linter.CheckInfo{
		Name:    "fullExample",
		Comment: "Report nothing, but test full info rendering.",
		Before:  `ereg($pat, $s)`,
		After:   `preg_match($pat, $s)`,
	}, `fullExample checker documentation

Report nothing, but test full info rendering.

Non-compliant code:
ereg($pat, $s)

Compliant code:
preg_match($pat, $s)`)

	runTest(linter.CheckInfo{
		Name:    "fullExampleMultiLine",
		Comment: "Report nothing, but test full info rendering.",
		Before: `class Foo {
  public function Foo($v) { $this->v = $v; }
}`,
		After: `class Foo {
  public function __construct($v) { $this->v = $v; }
}`,
	}, `fullExampleMultiLine checker documentation

Report nothing, but test full info rendering.

Non-compliant code:
class Foo {
  public function Foo($v) { $this->v = $v; }
}

Compliant code:
class Foo {
  public function __construct($v) { $this->v = $v; }
}`)

	runTest(linter.CheckInfo{
		Name:     "ternarySimplify",
		Comment:  `Report ternary expressions that can be simplified.`,
		Before:   `$x ? $x : $y`,
		After:    `$x ?: $y`,
		Quickfix: true,
	}, `ternarySimplify checker documentation (auto fix available)

Report ternary expressions that can be simplified.

Non-compliant code:
$x ? $x : $y

Compliant code:
$x ?: $y`)
}
