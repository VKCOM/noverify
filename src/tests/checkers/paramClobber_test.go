package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestParamClobberLegacyVariadic(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
function f($x, $y) {
  $args = func_get_args();
  $x = $args[0];
  $y = $args[1];
  return [$x, $y];
}
`)
}

func TestParamClobberReferenced(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
/**
 * @param mixed[] $x
 */
function f(array $x) {
  $x = $x['foo'];
  return $x;
}
`)
}

func TestParamClobberConditional(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
function f1($x, $y) {
  if ($y) {
    $x = 10;
  }
  return $x + $y;
}

function f2($x) {
  try {
    echo 123;
  } catch (Exception $_) {
    $x = "failed";
  }
  return $x;
}
`)
}

func TestParamClobberFunc(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
function f($x) {
  $x = 1343;
  return $x;
}
`)
	test.Expect = []string{
		`Param $x re-assigned before being used`,
	}
	test.RunAndMatch()
}

func TestParamClobberMethod(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
class C {
  /** f is an example method */
  public function f($x) {
    $x = 1343;
    return $x;
  }
}
`)
	test.Expect = []string{
		`Param $x re-assigned before being used`,
	}
	test.RunAndMatch()
}

func TestParamClobberClosure(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
function f() {
  return function ($x) {
    $x = 1343;
    return $x;
  };
}
`)
	test.Expect = []string{
		`Param $x re-assigned before being used`,
	}
	test.RunAndMatch()
}
