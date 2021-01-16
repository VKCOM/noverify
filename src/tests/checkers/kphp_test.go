package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestArrayAccessForAny(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config.KPHP = true
	test.AddFile(`<?php
	/** @return any */
	function get_any() {
		return [];
	}
	$any = get_any();
	$_ = $any[0];`)
	test.RunAndMatch()
}
