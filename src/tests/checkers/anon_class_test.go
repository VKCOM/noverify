package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestSimpleAnonClass(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function f() {
  $a = new class {
    /** */
    public function f() {}
  };

  $a->f();
}
`)
}

func TestAnonClassWithConstructor(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function f() {
  $a = new class(100, "s") {
    /** */
    public function f() {}

    public function __construct(int $a, string $b) {
      echo $a;
      echo $b;
    }
  };

  $a->f();
}
`)
}

func TestAnonClassWithExtends(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Boo {
  /** */
  public function b() {}
}
function f() {
  $a = new class extends Boo {
    /** */
    public function f() {}
  };

  $a->f();
  $a->b();
}
`)
}
