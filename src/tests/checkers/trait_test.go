package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestTraitSingleton(t *testing.T) {
	// See #533.
	linttest.SimpleNegativeTest(t, `<?php
trait Singleton {
  /**
   * @var ?self
   */
  private static $instance = null;

  /**
   * @return self
   */
  public static function instance() {
    if (!self::$instance) {
      self::$instance = new self();
    }

    return self::$instance;
  }
}

class Foo {
  use Singleton;

  /** @return int */
  public function f() { return 42; }
}

Foo::instance()->f();
`)
}

func TestTraitAsTypeHint(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait A {
  function a() {}
}

class Foo {
  public function f(A $a) {
    $a->a();
  }
}

function f(A $a) {
  $a->a();
}

`)
	test.Expect = []string{
		`forbidden to use a trait A as a type hint for function parameter`,
		`forbidden to use a trait A as a type hint for function parameter`,
	}
	linttest.RunFilterMatch(test, "badTraitUse")
}
