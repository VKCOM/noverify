package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestFewFunctionArgs(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {
  public function instanceFunction(int $x) {}
  public static function staticFunction(int $x) {}

  public static function staticFunctionWithDefault(int $x, string $z = "") {}
  public static function staticFunctionWithSeveralDefault(int $x, string $z = "", float $b = 10.5) {}

  public static function staticFunctionWithVarArg(int $x, string ...$z) {}
}

function f1(int $x) {}

/**
 * @param callable(int): Foo $a 
 */
function main(callable $a) {
  f1();
  f1(10);

  Foo::staticFunction();
  Foo::staticFunction(10); // ok

  Foo::staticFunctionWithDefault();
  Foo::staticFunctionWithDefault(10); // ok
  Foo::staticFunctionWithDefault(10, "sss"); // ok

  Foo::staticFunctionWithSeveralDefault();
  Foo::staticFunctionWithSeveralDefault(10); // ok
  Foo::staticFunctionWithSeveralDefault(10, "sss"); // ok
  Foo::staticFunctionWithSeveralDefault(10, "sss", 10); // ok

  Foo::staticFunctionWithVarArg(); 
  Foo::staticFunctionWithVarArg(10, "sss"); // ok
  Foo::staticFunctionWithVarArg(10, "sss", "10"); // ok

  (new Foo)->instanceFunction();
  (new Foo)->instanceFunction(10); // ok

  $a();
  $a(1); // ok

  $b = function(int $a) {};
  $b();
  $b(10); // ok
}
`)
	test.Expect = []string{
		`Too few arguments for f1, expecting 1, saw 0`,
		`Too few arguments for Foo::staticFunction, expecting 1, saw 0`,
		`Too few arguments for Foo::staticFunctionWithDefault, expecting 1, saw 0`,
		`Too few arguments for Foo::staticFunctionWithSeveralDefault, expecting 1, saw 0`,
		`Too few arguments for Foo::staticFunctionWithVarArg, expecting 1, saw 0`,
		`Too few arguments for Foo::instanceFunction, expecting 1, saw 0`,
		`Too few arguments for anonymous(int): Foo defined in PHPDoc, expecting 1, saw 0`,
		`Too few arguments for anonymous(...) defined on line 43, expecting 1, saw 0`,
	}
	linttest.RunFilterMatch(test, "argCount")
}
