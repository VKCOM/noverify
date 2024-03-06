package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue548_1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types=1);
class A {
  private $value;
  private function method() {}
}

class B {
  private $value;
  private function method() {}
}

class C extends B {
  private $value;
  private function method() {}

  /**
   * @param A|B|C $x
   */
  private function foo($x) {
    if ($x instanceof C) {
      echo $x->value;
      echo $x->method();
    }
  }
}`)
}

func TestIssue548_2(t *testing.T) {
	// Like TestIssue548, but with different names and types order.
	linttest.SimpleNegativeTest(t, `<?php
class C {
  private $value;
  private function method() {}
}

class B {
  private $value;
  private function method() {}
}

class A extends B {
  private $value;
  private function method() {}

  /**
   * @param A|B|C $x
   */
  private function foo($x) {
    if ($x instanceof A) {
      echo $x->value;
      echo $x->method();
    }
  }
}`)
}

func TestIssue548_Magic(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class A {
  public function __call($method, $args) {}
  public function __get($prop) {}
}

class B extends A {
  /**
   * @param A|B $x
   */
  private function foo($x) {
    if ($x instanceof B) {
      echo $x->value;
      echo $x->method();
    }
  }
}`)
}

func TestIssue548_Trait1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class A {
  private $value;
  private function method() {}
}

trait T {
  private $value;
  private function method() {}
}

class B {
  use T;

  /**
   * @param A|B $x
   */
  private function foo($x) {
    if ($x instanceof B) {
      echo $x->value;
      echo $x->method();
    }
  }
}`)
}

func TestIssue548_Trait2(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class B {
  private $value;
  private function method() {}
}

trait T {
  private $value;
  private function method() {}
}

class A {
  use T;

  /**
   * @param A|B $x
   */
  private function foo($x) {
    if ($x instanceof A) {
      echo $x->value;
      echo $x->method();
    }
  }
}`)
}
