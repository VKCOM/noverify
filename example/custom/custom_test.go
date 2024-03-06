package main

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestAssignmentAsExpression(t *testing.T) {
	test := linttest.NewSuite(t)
	addCheckers(test.Config())

	test.AddFile(`<?php
	declare(strict_types=1);
	// PHPDoc annotations are not required for NoVerify in simple cases.
	function something() {
		$a = "test";
		return $a;
	}
	function in_array() {}

	function test() {
		$b = ["1", "2", "3"];

		if (in_array(something(), $b)) {
			echo "third arg true";
		}

		if (something() == $b[1]) {
			echo "must be ===";
		}
	}`)

	test.Expect = []string{
		"3rd argument of in_array must be true when comparing strings",
		"Strings must be compared using '===' operator",
	}

	test.RunAndMatch()
}
