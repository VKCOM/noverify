package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue556(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/Core/Core_c.php`}
	test.AddFile(`<?php
/**
 * @param \ArrayAccess|array $v
 */
function f1($v) {
  return $v[10];
}

interface MyArrayAccess extends ArrayAccess {}

/**
 * @param MyArrayAccess|array $v
 */
function f2($v) {
  return $v[10];
}

interface MyArrayAccess2 extends MyArrayAccess {}

/**
 * @param MyArrayAccess2 $v
 */
function f3($v) {
  return $v[10];
}
`)
	test.RunAndMatch()
}
