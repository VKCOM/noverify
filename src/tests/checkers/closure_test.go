package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestClosureCapture(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class omg {
  public $some_property;
}

function doSomething($a, omg $b) {
  return function() use($b) {
    echo $b->some_property;
    echo $b->other_property;
  };
}`,
	)
	test.Expect = []string{
		"other_property does not exist",
	}
	test.RunAndMatch()
}

func TestClosureDoc(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {
  /**
   * @return int
   */
  public function method(): int { return 0; }
}

/**
 * @param callable(int, string): Boo|Foo $s
 */
function f(callable $s) {
  $a = $s(10);
  echo $a->method();
}
`,
	)
	test.Expect = []string{
		"Too few arguments for anonymous(int,string): Boo|Foo defined in PHPDoc, expecting 2, saw 1",
	}
	test.RunAndMatch()
}

func TestClosureInvalidDoc(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {}

/**
 * @param callable(int, string) $s
 * @param callable(int, string): Foo $s1
 * @param callable(int, string): void $s2
 */
function f(callable $s, callable $s1, callable $s2) {}
`,
	)
	test.Expect = []string{
		"Lost return type for callable(...), if the function returns nothing, specify void explicitly",
		"Type for $s can be wrote explicitly from typeHint",
	}
	test.RunAndMatch()
}
