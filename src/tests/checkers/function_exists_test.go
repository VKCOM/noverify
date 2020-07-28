package linttest_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestFunctionExists1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function function_exists($name) { return 1 == 2; }

if (function_exists('\foo')) {
  \foo();
}
if (function_exists('bar')) {
  bar("a", "b");
}
if (function_exists('a\b\baz')) {
  a\b\baz(1);
  if (function_exists('f2')) {
    a\b\baz(f2(1));
  }
}
`)
}

func TestFunctionExists2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function function_exists($name) { return 1 == 2; }

function f($cond) {
  if (!function_exists('\foo')) {
    \foo();
  }
  if ($cond && !function_exists('bar')) {
    bar("a", "b");
  }
  if ($cond || !function_exists('a\b\baz')) {
    a\b\baz(1);
  }
}
`)
	test.Expect = []string{
		`Call to undefined function \foo`,
		`Call to undefined function bar`,
		`Call to undefined function a\b\baz`,
	}
	test.RunAndMatch()
}

func TestFunctionExists3(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function function_exists($name) { return 1 == 2; }

if (function_exists('\foo')) {
}
if (function_exists('bar')) {
}
if (function_exists('a\b\baz')) {
}

if (function_exists('a\b\baz')) {
  a\b\baz(1);
  if (function_exists('f2')) {
  }
  f2();
}

\foo();
bar("a", "b");
a\b\baz(1);
`)
	test.Expect = []string{
		`Call to undefined function f2`,
		`Call to undefined function \foo`,
		`Call to undefined function bar`,
		`Call to undefined function a\b\baz`,
	}
	test.RunAndMatch()
}
