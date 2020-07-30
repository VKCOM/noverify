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
   * @var self
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
