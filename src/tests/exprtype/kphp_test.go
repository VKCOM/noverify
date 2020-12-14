package exprtype_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linter"
)

func TestExprTypeAny(t *testing.T) {
	code := `<?php
/** @return any */
function get_any() {
  return 10;
}

/** @return any[][] */
function get_any_arr() {
  return [[1]];
}

exprtype(get_any(), 'mixed');
exprtype(get_any_arr(), 'mixed[][]');
`
	runKPHPExprTypeTest(t, &exprTypeTestParams{code: code})
}

func runKPHPExprTypeTest(t *testing.T, params *exprTypeTestParams) {
	linter.KPHP = true
	runExprTypeTest(t, params)
	linter.KPHP = false
}
