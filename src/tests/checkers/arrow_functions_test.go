package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestArrowFunction(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types = 1);
class Boo {
  /** @return int */
  public function b() { }
}

function foo() {
  $value = 10;

  // simple function
  $_ = fn($x) => $x + 5;

  // with capture
  $_ = fn($x) => $x + $value;

  // reference
  $_ = fn&($x) => $x + $value;

  // with undefined variable
  $_ = fn($x) => $x + $undefined_variable;

  if ($value == 0) {
    $maybe_defined = 100;
  }

  // with maybe defined variable
  $_ = fn($x) => $x + $maybe_defined;

  // with unused variable
  $_ = fn($x) => $a = $x + 5;
  $_ = fn($x) => ($a = $x + 5) && $x;

  $_ = fn($x) => ($a = $x + 5) && $a + 5;

  // with PHPDoc
  /**
   * @param Boo $x
   */
  $_ = fn($x) => $x->b();

  // nested
  $_ = fn($x) => fn($y) => fn($w) => $x * $y + $w - $value;

  // nested with maybe defined variable
  $_ = fn($x) => fn($y) => fn($w) => $x * $y + $w - $maybe_defined;

  // nested with unused variable
  $_ = fn($x) => fn($y) => fn($w) => $a = $x + $y + $w;

  $_ = fn($x) => fn($y) => fn($w) => ($a = $x + 5) && $a + 5;

  // ok
  $_ = fn() => ($a = 10) && $a;

  // $a is undefined
  $_ = fn() => $a = 10 && $a;

  // arguments are not visible outside of arrow function
  echo $x; // Undefined $x
  echo $y; // Undefined $y
  echo $w; // Undefined $w
}
`)
	test.Expect = []string{
		`Cannot find referenced variable $undefined_variable`,
		`Possibly undefined variable $maybe_defined`,
		`Variable $a is unused (use $_ to ignore this inspection or specify --unused-var-regex flag)`,
		`Variable $a is unused (use $_ to ignore this inspection or specify --unused-var-regex flag)`,
		`Variable $a is unused (use $_ to ignore this inspection or specify --unused-var-regex flag)`,
		`Variable $a is unused (use $_ to ignore this inspection or specify --unused-var-regex flag)`,
		`Cannot find referenced variable $a`,
		`Cannot find referenced variable $x`,
		`Cannot find referenced variable $y`,
		`Cannot find referenced variable $w`,
		`Possibly undefined variable $maybe_defined`,
	}
	test.RunAndMatch()
}

func TestUnusedInArrowFunction(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{
		`stubs/phpstorm-stubs/standard/standard_4.php`,
		`stubs/phpstorm-stubs/standard/standard_9.php`,
	}
	test.AddFile(`<?php
	declare(strict_types = 1);

function f() {
  $a1 = [];
  $a2 = [];
  $a3 = [];
  $a4 = [];

  var_dump(array_filter([], fn($elem) => isset($a1[$elem]))); // ok
  var_dump(array_filter([], fn($elem) => isset($elem))); // $a2 not used
  var_dump(array_filter([], fn($elem) => $a3)); // ok
  var_dump(array_filter([], fn($elem) => $elem && $a4)); // ok

  $_ = fn() => ($a = 10) && $a; // ok
  $_ = fn() => ($a = 10); // $a unused
}
`)
	test.Expect = []string{
		`Variable $a2 is unused (use $_ to ignore this inspection or specify --unused-var-regex flag)`,
		`Variable $a is unused (use $_ to ignore this inspection or specify --unused-var-regex flag)`,
	}
	test.RunAndMatch()
}
