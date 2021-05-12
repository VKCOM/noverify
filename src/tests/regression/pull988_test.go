package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPull998TraitUse(t *testing.T) {
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

func TestPull998ScalarEncapsedStringVar(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function f() {
  $a = 100;
  return "Hello ${$a}";
}
`)
}
