package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestTypeHintGood(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
abstract class FooAbstract {
    /**
     * @return string[]
     */
    abstract function f(): array;
}

class Foo extends FooAbstract {
    /**
     * @inheritdoc
     */
    function f(): array { // skipped
        return [];
    }
}

class Foo2 extends FooAbstract {
    /**
     * {@inheritdoc}
     */
    function f(): array { // skipped
        return [];
    }
}

class A {
  /**
   * @var int[]
   */
  public array $a = [];
  /**
   * @var int[]
   */
  public array $b, $c = [];

  /**
   * @param int[] $a
   */
  function f(array $a) {}

  /**
   * @param mixed[] $a
   */
  function f1(array $a) {}

  /**
   * @param mixed[] $a
   * @return mixed[]
   */
  function f2(array $a): array {}

  /**
   * @param mixed[] $a
   * @return int[]
   */
  function f3(array $a): array {}
}

/**
 * @param int[] $a
 */
function f(array $a) {}

/**
 * @param mixed[] $a
 */
function f1(array $a) {}

/**
 * @param mixed[] $a
 * @return mixed[]
 */
function f2(array $a): array {}

/**
 * @param mixed[] $a
 * @return int[]
 */
function f3(array $a): array {}

function f4() {
  // closure skipped
  $_ = function(array $a) {
    return $a;
  };
}
`)
	test.RunAndMatch()
}

func TestTypeHintBad(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
  public array $a = [];
  public array $b, $c = [];

  function f(array $a) {}
  function f1(): array {}
}

function f(array $a) {}
function f(): array {}

function f2() {
  // closures skipped
  $_ = function(array $a) {
    return $a;
  };
  $_ = function(): array {
    return $a;
  };
}
`)
	test.Expect = []string{
		`specify the type for the property in phpdoc, 'array' type hint is not precise enough`,
		`specify the type for the property in phpdoc, 'array' type hint is not precise enough`,
		`specify the type for the parameter $a in phpdoc, 'array' type hint is not precise enough`,
		`specify the return type for the function f1 in phpdoc, 'array' type hint is not precise enough`,
		`specify the type for the parameter $a in phpdoc, 'array' type hint is not precise enough`,
		`specify the return type for the function f in phpdoc, 'array' type hint is not precise enough`,
	}
	linttest.RunFilterMatch(test, "typeHint")
}
