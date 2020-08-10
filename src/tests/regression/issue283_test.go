package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue283(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait YummyProps {
  public $price = 'fair';
  protected $taste = 'good';
  private $secret = 'sauce';
}

class Borsch {
  use YummyProps;

  /** @return string */
  public function getTaste() {
    return $this->taste; // OK: using protected trait prop from embedding class
  }

  /** @return string */
  public function getSecret() {
    return $this->secret; // OK: using private trait prop from embedding class
  }
}

class SaltyBorsch extends Borsch {
  protected function getTaste2() {
    return $this->taste; // OK: using protected trait prop from inherited class
  }
  protected function getSecret2() {
    return $this->secret; // Bad: used private trait prop from inherited class
  }
}

$borsch = new Borsch();
$_ = $borsch->taste; // Bad: can't access protected trait prop from the outside
$_ = $borsch->price; // OK: can use public trait prop here

$borsch2 = new SaltyBorsch();
$_ = $borsch2->price; // OK: can also access public trait prop from derived class
`)
	test.Expect = []string{
		`Cannot access private property \Borsch->secret`,
		`Cannot access protected property \Borsch->taste`,
	}
	test.RunAndMatch()
}
