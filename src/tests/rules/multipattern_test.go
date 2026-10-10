package rules_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestRuleMultiAny(t *testing.T) {
	rfile := `<?php
function typecheckOp() {
  /**
   * @warning increment/decrement of a non-numeric type
   * @type !(int|float) $x
   */
  any_incdec: {
    $x++;
    ++$x;
    $x--;
    --$x;
  }

  /**
   * @warning don't compare arrays with numeric types
   * @type array $x
   * @type int|float $y
   * @or
   * @type int|float $x
   * @type array $y
   */
  any_arraycmp: {
    $x > $y;
    $x < $y;
    $x >= $y;
    $x <= $y;
  }
}
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	test.AddFile(`<?php
class Foo {}

function bad(string $s, Foo $foo) {
  $s++; ++$s;
  $foo--; --$foo;
}

function good_int() {
  $i = 10;
  $i++; ++$i; $i--; --$i;
}

function good_float() {
  $f = 10.53;
  $f++; ++$f; $f--; --$f;
}
`)
	test.AddFile(`<?php
function bad_arraycmp(array $a) {
  $i = 10;

  $_ = $a < $i;
  $_ = $i < $a;
  $_ = $a > $i;
  $_ = $i > $a;

  $_ = $a <= $i;
  $_ = $i <= $a;
  $_ = $a >= $i;
  $_ = $i >= $a;
}

function good_arraycmp(array $a, array $a2) {
  $_ = $a < $a2;
  $_ = $a > $a2;
  $_ = $a <= $a2;
  $_ = $a >= $a2;
}
`)
	test.Expect = []string{
		`increment/decrement of a non-numeric type`,
		`increment/decrement of a non-numeric type`,
		`increment/decrement of a non-numeric type`,
		`increment/decrement of a non-numeric type`,
		`don't compare arrays with numeric types`,
		`don't compare arrays with numeric types`,
		`don't compare arrays with numeric types`,
		`don't compare arrays with numeric types`,
		`don't compare arrays with numeric types`,
		`don't compare arrays with numeric types`,
		`don't compare arrays with numeric types`,
		`don't compare arrays with numeric types`,
	}

	test.RunRulesTest()
}

func TestRuleGroupedBlock(t *testing.T) {
	rfile := `<?php
function arrays() {
  /**
   * @warning can be replaced with isset($array[$key]) or array_key_exists($key, $array)
   */
  array_has_key: {
    in_array($key, array_keys($array));
    in_array($key, array_keys($array), true);
    in_array($key, array_keys_as_ints($array));
    in_array($key, array_keys_as_ints($array), true);
    in_array($key, array_keys_as_strings($array));
    in_array($key, array_keys_as_strings($array), true);
  }
}
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	test.AddFile(`<?php
function test_array_has_key($k, $arr) {
  in_array($k, array_keys($arr));
  in_array($k, array_keys($arr), true);
  in_array($k, array_keys_as_ints($arr));
  in_array($k, array_keys_as_ints($arr), true);
  in_array($k, array_keys_as_strings($arr));
  in_array($k, array_keys_as_strings($arr), true);
}

function test_good_arrays($k, $arr) {
  isset($arr[$k]);
  array_key_exists($k, $arr);
}

function test_labels_in_code($k, $arr) {
custom_label:
  in_array($k, array_keys($arr));
}
`)
	test.Expect = []string{
		`can be replaced with isset($arr[$k]) or array_key_exists($k, $arr)`,
		`can be replaced with isset($arr[$k]) or array_key_exists($k, $arr)`,
		`can be replaced with isset($arr[$k]) or array_key_exists($k, $arr)`,
		`can be replaced with isset($arr[$k]) or array_key_exists($k, $arr)`,
		`can be replaced with isset($arr[$k]) or array_key_exists($k, $arr)`,
		`can be replaced with isset($arr[$k]) or array_key_exists($k, $arr)`,
		`can be replaced with isset($arr[$k]) or array_key_exists($k, $arr)`,
	}

	test.RunRulesTest()
}
