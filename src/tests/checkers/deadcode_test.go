package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDeadCodeInNullCoalesce(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function retNullable(): ?string {
	return "";
}

function f() {
  $_ = $q ?? 10; // ok

  $a = 100;
  $_ = $a ?? 10; // dead code

  $c = retNullable();
  $_ = $c ?? 10; // dead code, but imprecise type

  $d = 10;
  if (1) {
    $d = null;
  }
  $_ = $d ?? 10; // ok

  $input = [1.5,2,3];
  $_ = $input["test"] ?? 10; // ok
  $_ = (int)$input["test"] ?? 10; // dead code
  $_ = (int)($input["test"] ?? 10); // ok

  global $x;
  $_ = $x ?? 10; // ok
}

`)
	test.Expect = []string{
		`$a is not nullable, right side of the expression is unreachable`,
		`(int)$input["test"] is not nullable, right side of the expression is unreachable`,
	}
	linttest.RunFilterMatch(test, "deadCode")
}
