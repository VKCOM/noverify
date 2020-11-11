package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
)

func TestArrayAccessForAny(t *testing.T) {
	linter.KPHP = true
	linttest.SimpleNegativeTest(t, `<?php
	/** @return any */
	function get_any() {
		return [];
	}
	$any = get_any();
	$_ = $any[0];`)
	linter.KPHP = false
}
