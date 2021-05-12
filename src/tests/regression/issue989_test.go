package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue989TraitUse(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
trait Trait1 {}

function f() {
  $_ = new class() {
    use Trait1 {
      getClient as public;
    }
  };

  return 1;
}
`)
}

func TestIssue989ScalarEncapsedStringVar(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function f() {
  $a = 100;
  return "Hello ${$a}";
}
`)
}

func TestIssue989InterpretStrings(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
$unicode = "\u{0000000000aA19}";
$unicode = "\u{00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041}";
`)
}
