package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestVarShadow(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function simple($a) {
  foreach ([1, 2] as $a) {
    echo $a;
  }
}

function withList($a, $b) {
  foreach ([1, 2] as list($a, $b)) {
    echo $a, $b;
  }
}

function withDeepList($a, $b) {
  foreach ([1, 2] as list(list($a, $b), $b)) {
    echo $a, $b;
  }
}

function withSimpleKey($a, $b) {
  foreach ([1, 2] as $a => $b) {
    echo $a, $b;
  }
}

class InClass {
  public function simple($a) {
    foreach ([1, 2] as $a) {
      echo $a;
    }
  }

  public function withList($a, $b) {
    foreach ([1, 2] as list($a, $b)) {
      echo $a, $b;
    }
  }

  function withDeepList($a, $b) {
    foreach ([1, 2] as list(list($a, $b), $b)) {
      echo $a, $b;
    }
  }

  function withSimpleKey($a, $b) {
    foreach ([1, 2] as $a => $b) {
      echo $a, $b;
    }
  }
}

// In global scope.
foreach ([1, 2] as $a => $b) {
  echo $a, $b;
}
`)
	test.Expect = []string{
		`Variable $a shadow existing variable $a from current function params`,

		`Variable $a shadow existing variable $a from current function params`,
		`Variable $b shadow existing variable $b from current function params`,

		`Variable $a shadow existing variable $a from current function params`,
		`Variable $b shadow existing variable $b from current function params`,
		`Variable $b shadow existing variable $b from current function params`,

		`Variable $a shadow existing variable $a from current function params`,
		`Variable $b shadow existing variable $b from current function params`,

		// class
		`Variable $a shadow existing variable $a from current function params`,

		`Variable $a shadow existing variable $a from current function params`,
		`Variable $b shadow existing variable $b from current function params`,

		`Variable $a shadow existing variable $a from current function params`,
		`Variable $b shadow existing variable $b from current function params`,
		`Variable $b shadow existing variable $b from current function params`,

		`Variable $a shadow existing variable $a from current function params`,
		`Variable $b shadow existing variable $b from current function params`,
	}
	linttest.RunFilterMatch(test, "varShadow")
}
