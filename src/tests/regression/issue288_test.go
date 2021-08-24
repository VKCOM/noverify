package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue288(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
class Box {
  public $item1;
  public $item2;
}

$_ = isset($a) ? $a[0] : 0;
$_ = isset($b) && isset($a) ? $a[0] + $b : 0;
$_ = isset($a[0]) ? $a[0] : 0;
$_ = isset($a[0]) ? $a : [0];

$_ = isset($b1) ? 0 : $b1;
$_ = isset($b2[0]) ? 0 : $b2;
$_ = isset($b3[0]) ? 0 : $b3;


function f($x, $y) {
  $_ = $x instanceof Box ? $x->item1 : 0;
  $_ = $y instanceof Box ? 0 : $y->item2;
}

$x = new Box();
$_ = $badvar ? 0 : 1;
$_ = isset($x) && isset($y) ? $x : 0;
$_ = $x instanceof Box ? 0 : 1;
`)
	test.Expect = []string{
		`Undefined variable $badvar`,
		`Undefined variable $b1`,
		`Undefined variable $b2`,
		`Undefined variable $b3`,
		`Property {mixed}->item2 does not exist`,
	}
	test.RunAndMatch()
}
