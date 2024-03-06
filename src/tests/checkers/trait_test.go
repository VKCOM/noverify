package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestTraitSingleton(t *testing.T) {
	// See #533.
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types = 1);
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
trait A {}
trait B {}

class Foo {
  private A $a;
  public static A $a1;

  public function f(A $a): A {}
  public function f1(A $a, B $b): A {}
}

function f(A $a): A {}
function f1(A $a, B $b): A {
  $_ = function(A $a): B {};
}

trait Test {
  private static ?self $instance = null;     // ok, in trait
  public static function instance(): self {} // ok, in trait
}
`)
	test.Expect = []string{
		`Cannot use trait A as a typehint for property type`,
		`Cannot use trait A as a typehint for property type`,
		`Cannot use trait A as a typehint for return type`,
		`Cannot use trait A as a typehint for parameter type`,
		`Cannot use trait A as a typehint for return type`,
		`Cannot use trait A as a typehint for parameter type`,
		`Cannot use trait B as a typehint for parameter type`,
		`Cannot use trait A as a typehint for return type`,
		`Cannot use trait A as a typehint for parameter type`,
		`Cannot use trait A as a typehint for return type`,
		`Cannot use trait A as a typehint for parameter type`,
		`Cannot use trait B as a typehint for parameter type`,
		`Cannot use trait B as a typehint for closure return type`,
		`Cannot use trait A as a typehint for parameter type`,
	}
	linttest.RunFilterMatch(test, "badTraitUse")
}
