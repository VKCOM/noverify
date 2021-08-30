package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue26_1(t *testing.T) {
	// Test that defined variable variable don't cause "undefined" warnings.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		$x = 'key';
		if (isset($$x)) {
			$_ = $x + 1;  // If $$x is isset, then $x is set as well
			$_ = $$x + 1;
			$_ = $y;      // Undefined
		}
		// After the block all vars are undefined again.
		$_ = $x;
	}`)
	test.Expect = []string{"Cannot find referenced variable $y"}
	test.RunAndMatch()
}

func TestIssue26_2(t *testing.T) {
	// Test that if $x is defined, it doesn't make $$x defined.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($x)) {
			$_ = $x;  // $x is defined
			$_ = $$x; // But $$x is not
		}
	}`)
	test.Expect = []string{"Unknown variable variable $$x used"}
	test.RunAndMatch()
}

func TestIssue26_3(t *testing.T) {
	// Test that irrelevant isset of variable-variable doesn't affect
	// other variables. Also warn for undefined variable in $$x.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($$x)) {
			$_ = $$y;
		}
	}`)
	test.Expect = []string{
		"Cannot find referenced variable $x",
		"Unknown variable variable $$y used",
	}
	test.RunAndMatch()
}

func TestIssue26_4(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($$$$x)) {
			$_ = $$$$x; // Can't track this level of indirection
		}
	}`)
	test.Expect = []string{
		"Unknown variable variable $$$x used",
		"Unknown variable variable $$$$x used",
	}
	test.RunAndMatch()
}
