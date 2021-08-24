package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue128(t *testing.T) {
	t.Skip()
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
class Value {
  public $x;
}

function count($arr) { return 0; }

function good($v) {
  if (isset($good) && count($good) == 1) {}

  if ($v instanceof Value && $v->x) {}
  if (isset($y) && $y instanceof Value && $y->x) {}
}

function bad1($v) {
  if (isset($bad0) && $bad0) {}
  $_ = $bad0; // Used outside of if body

  if (count($bad1) == 1 && isset($bad1)) {}
  if (isset($good) && count($bad2) == 1 && isset($bad2)) {}
  if (isset($bad3) || count($bad3) == 1) {}

  if ($v->x && $v instanceof Value) {}

  if ($y1 instanceof Value && isset($y1) && $y1->x) {}
}

$_ = $bad1;
`)
	test.Expect = []string{
		`Undefined variable $bad0`,
		`Undefined variable $bad1`, // At local scope
		`Undefined variable $bad1`, // At global scope
		`Undefined variable $bad2`,
		`Undefined variable $bad3`,
		`Property {mixed}->x does not exist`,
		`Undefined variable $y1`,
	}
	test.RunAndMatch()
}
