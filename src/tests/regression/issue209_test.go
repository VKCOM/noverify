package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue209_1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types=1);
trait A {
  private function priv() { return 1; }
  protected function prot() { return 2; }
  /** @return int */
  public function pub() { return 3; }
}

class B {
  use A;
  /** @return int */
  public function sum() {
    return $this->priv() + $this->prot() + $this->pub();
  }
}

echo (new B)->sum(); // actual PHP prints 6
`)
}

func TestIssue209_2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
trait Methods {
  /***/
  public function pubMethod() {}
  protected function protMethod() {}
  private function privMethod() {}

  /***/
  public static function staticPubMethod() {}
  protected static function staticProtMethod() {}
  private static function staticPrivMethod() {}
}

class Base {
  use Methods;

  private function f() {
    $this->pubMethod();
    $this->protMethod();
    $this->privMethod();
  }
}

class Derived extends Base {
  private function f() {
    $this->pubMethod();
    $this->protMethod();
    $this->privMethod(); // Bad: can't call private
    Base::staticPubMethod();
    Base::staticProtMethod();
    parent::staticProtMethod();
    parent::staticPrivMethod(); // Bad: can't call private
  }
}

$b = new Base();
$b->pubMethod();
$b->protMethod(); // Bad: can't call from the outside
$b->privMethod(); // Bad: can't call from the outside

Base::staticPubMethod();
Derived::staticPubMethod();
Base::staticProtMethod();
`)
	test.Expect = []string{
		`Cannot access private method \Base->privMethod()`,
		`Cannot access protected method \Base->protMethod()`,
		`Cannot access private method \Base->privMethod()`,
		`Cannot access private method \Base::staticPrivMethod()`,
		`Cannot access protected method \Base::staticProtMethod()`,
	}
	test.RunAndMatch()
}
