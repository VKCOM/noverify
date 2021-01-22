package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
)

func TestArrayAccessForAny(t *testing.T) {
	config := linter.NewConfig()
	config.KPHP = true
	test := linttest.NewSuite(t)
	test.Linter = linter.NewLinter(config)
	test.AddFile(`<?php
	/** @return any */
	function get_any() {
		return [];
	}
	$any = get_any();
	$_ = $any[0];`)
	test.RunAndMatch()
}
