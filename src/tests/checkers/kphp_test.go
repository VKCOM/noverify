package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestArrayAccessForAny(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().KPHP = true
	test.AddFile(`<?php
	/** @return any */
	function get_any() {
		return [];
	}
	$any = get_any();
	$_ = $any[0];`)
	test.RunAndMatch()
}

func TestDifferentTypesAsUndefinedClass(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().KPHP = true
	test.AddFile(`<?php
/**
 * @param tuple(int, string) $a
 * @param shape(key: int, val: string) $b
 * @param kmixed $c
 * @param any $d
 * @param future<int> $e
 */
function f($a, $b, $c, $d, $e) {}
`)
	test.RunAndMatch()
}
