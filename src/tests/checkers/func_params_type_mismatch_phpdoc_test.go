package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestFunctionParamTypeMismatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
// Functions testing parameter type mismatch between phpdoc and typehint

/**
 * Function with correct non-nullable parameter: both phpdoc and typehint are non-nullable.
 *
 * @param int $a
 */
function funcCorrect1(int $a) {
}

/**
 * Function with correct nullable parameter: both phpdoc and typehint are nullable.
 *
 * @param int|null $b
 */
function funcCorrect2(?int $b) {
}

/**
 * Function with error: phpdoc specifies non-nullable int, but typehint is nullable (?int).
 *
 * @param int $c
 */
function funcError1(?int $c) {
}

/**
 * Function with error: phpdoc specifies nullable int (int|null), but typehint is non-nullable int.
 *
 * @param int|null $d
 */
function funcError2(int $d) {
}
`)
	test.Expect = []string{
		`param $c miss matched with phpdoc type <<int>>`,
		`param $d miss matched with phpdoc type <<int|null>>`,
	}
	test.RunAndMatch()
}

func TestMethodParamTypeMismatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
// Class methods testing parameter type mismatch between phpdoc and typehint

class Sample {
  /**
   * Method with correct non-nullable parameter: both phpdoc and typehint match.
   *
   * @param string $name
   */
  public function methodCorrect(string $name) {
  }

  /**
   * Method with error: phpdoc specifies nullable string (string|null), but typehint is non-nullable.
   *
   * @param string|null $title
   */
  public function methodError1(string $title) {
  }

  /**
   * Method with error: phpdoc specifies non-nullable string, but typehint is nullable (?string).
   *
   * @param string $desc
   */
  public function methodError2(?string $desc) {
  }
}
`)
	test.Expect = []string{
		`param $title miss matched with phpdoc type <<string|null>>`,
		`param $desc miss matched with phpdoc type <<string>>`,
	}
	test.RunAndMatch()
}

func TestIgnoreReturnTypeMismatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
// This test ensures that only parameter types are checked,
// and any mismatches in return types are ignored.

/**
 * Function with a mismatch in return type between phpdoc and typehint.
 * The parameter type is correct and should not trigger an error.
 *
 * @param int $a
 * @return int|null
 */
function returnMismatch(int $a): int {
    return $a > 0 ? $a : null;
}

// Call the function to trigger linting
returnMismatch(5);
`)
	test.Expect = []string{
		`Expression evaluated but not used`,
	}
	test.RunAndMatch()
}

func TestArrayParamTypeMismatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
// Functions testing array parameter types

/**
 * Function with correct array parameter:
 * phpdoc specifies "array" and typehint is array.
 *
 * @param array $a
 */
function funcArrayCorrect1(array $a) {
}

/**
 * Function with correct array parameter:
 * phpdoc specifies "int[]" and typehint is array.
 *
 * @param int[] $b
 */
function funcArrayCorrect2(array $b) {
}

/**
 * Function with error: phpdoc specifies non-nullable "array",
 * but typehint is nullable array (?array).
 *
 * @param array $c
 */
function funcArrayError1(?array $c) {
}

`)
	test.Expect = []string{
		`param $c miss matched with phpdoc type <<array>>`,
	}
	test.RunAndMatch()
}

func TestByReferenceCorrect(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * Clear standalone line tokens.
 *
 * @param array $nodes  Parsed nodes
 * @param array $tokens Tokens to be parsed
 *
 * @return array|null Resulting indent token, if any
 */
function clearStandaloneLines(array &$nodes, array &$tokens) {
}

`)
	// No errors expected.
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestByReferenceMismatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * Incorrect by-reference parameter:
 * PHPDoc specifies int, but typehint is array.
 *
 * @param int $a
 */
function byRefError(array &$a) {
}
`)
	test.Expect = []string{
		`param $a miss matched with phpdoc type <<int>>`,
	}
	test.RunAndMatch()
}

func TestCallableCorrect(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * Function with callable parameter.
 *
 * @param callable(int, string): Boo|Foo $s
 */
function testCallable(callable $s) {
  // Dummy call.
  $s(1, "test");
}

testCallable(function($a, $b) {
  return new stdClass();
});
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestInterfaceCorrect(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface Responsible {}

/**
 * Uses Responsible interface value.
 *
 * @param \Responsible $r
 */
function reference_iface(Responsible $r) {
}

reference_iface(new class implements Responsible {});
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestUnionMismatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * Function with error: phpdoc specifies nullable int (int|null),
 * but typehint is non-nullable int.
 *
 * @param int|null $d
 */
function funcError2(int $d) {
}

funcError2(5);
`)
	test.Expect = []string{
		`param $d miss matched with phpdoc type <<int|null>>`,
	}
	test.RunAndMatch()
}

func TestNullableCorrect(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * Function with correct nullable parameter.
 *
 * @param int|null $d
 */
function nullableCorrect(?int $d) {
}

nullableCorrect(null);
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestArrayCorrect(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * Function with array parameters.
 *
 * @param array $a
 * @param int[] $b
 */
function funcArray(array $a, array $b) {
}

funcArray([], []);
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestBoolSynonym(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * Function with boolean type synonym.
 *
 * @param boolean $flag
 */
function testBool(bool $flag) {
}
`)
	test.Expect = []string{
		`Use bool type instead of boolean`,
	}
	test.RunAndMatch()
}

func TestAliasNormalization(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
namespace Test;

use A as B;

/**
 * Function with aliased type.
 *
 * @param \A $item
 */
function testAlias(B $item) {
}

`)
	test.Expect = []string{
		`Class or interface named \A does not exist`,
		`Class or interface named \A does not exist`,
	}
	test.RunAndMatch()
}
