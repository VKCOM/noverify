package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue327(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);
function sink(...$args) {}

function f() {
  return <<<'SQL'
    SELECT login, password FROM user WHERE login LIKE '%admin%'
  SQL;
}

function f2($x) {
  sink(<<<STR
    $x $x
    STR, 1, 2);

  sink(<<<STR
    abc
  STR);

  sink(<<<"STR"
  STRNOTEND
    abc
  STR);
}
`)
}
