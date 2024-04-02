package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestTypeHintGood(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
abstract class FooAbstract {
    /**
     * @return string[]
     */
    public abstract function f(): array;
}

class Foo extends FooAbstract {
    /**
     * @inheritdoc
     */
    public function f(): array { // skipped
        return [];
    }
}

class Foo2 extends FooAbstract {
    /**
     * {@inheritdoc}
     */
    public function f(): array { // skipped
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
  public function f(array $a) {}

  /**
   * @param mixed[] $a
   */
  public function f1(array $a) {}

  /**
   * @param mixed[] $a
   * @return mixed[]
   */
  public function f2(array $a): array {}

  /**
   * @param mixed[] $a
   * @return int[]
   */
  public function f3(array $a): array {}
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

  public function f(array $a) {}
  public function f1(): array {}
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
		`Specify the type for the property in PHPDoc, 'array' type hint too generic`,
		`Specify the type for the property in PHPDoc, 'array' type hint too generic`,
		`Specify the type for the parameter $a in PHPDoc, 'array' type hint too generic`,
		`Specify the return type for the function f1 in PHPDoc, 'array' type hint too generic`,
		`Specify the type for the parameter $a in PHPDoc, 'array' type hint too generic`,
		`Specify the return type for the function f in PHPDoc, 'array' type hint too generic`,
	}
	linttest.RunFilterMatch(test, "typeHint")
}
