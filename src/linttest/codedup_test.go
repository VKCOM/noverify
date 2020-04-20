package linttest_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDupIfCond(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
const C1 = 1;
const C2 = 2;
const C3 = 3;

$glob = 0;

function purefunc($x) { return $x+1; }

function impurefunc($x) {
  global $glob;
  $glob += $x;
  return $glob;
}

function f($cond) {
  if ($cond + 1) {
  } elseif ($cond + 1) {
  }

  if ($cond) {
  } elseif ($cond) {
  } else {}
}
`)
	test.RunAndMatch()
}

func TestDupCaseCond(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
const C1 = 1;
const C2 = 2;
const C3 = 3;

$glob = 0;

function purefunc($x) { return $x+1; }

function impurefunc($x) {
  global $glob;
  $glob += $x;
  return $glob;
}

function f($cond) {
  switch ($cond) {
  case 'abc':
    echo 1; break;
  case 'abc':
    echo 2; break; // Bad: duplicated string literal
  }

  switch ($cond) {
  case 2:
    echo 2; break;
  case 1:
    echo 1; break;
  case 3:
    echo 3; break;
  case 1:
    echo 4; break; // Bad: duplicated int literal
  case 5:
    echo 5; break;
  }

  switch ($cond) {
  default: break;
  case C1: break;
  case C2: break;
  case C3: break;
  case C2: break; // Bad: duplicated const
  }

  // Pure function calls are tracked.
  switch ($cond) {
  case purefunc(1): break;
  case purefunc(1): break; // Bad: duplicated pure func call
  }

  // No warnings for duplicated expressions with potential side effects.
  switch ($cond) {
  case impurefunc(1): break;
  case impurefunc(1): break;
  }
}
`)
	test.Expect = []string{
		`duplicated switch case #2`,
		`duplicated switch case #4`,
		`duplicated switch case #5`,
		`duplicated switch case #2`,
	}
	test.RunAndMatch()
}
