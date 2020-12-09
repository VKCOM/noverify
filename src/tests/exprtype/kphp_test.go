package exprtype_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linter/config"
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

exprtype(get_any(), 'int|mixed');
exprtype(get_any_arr(), 'int[][]|mixed[][]');
`
	runKPHPExprTypeTest(t, &exprTypeTestParams{code: code})
}

func runKPHPExprTypeTest(t *testing.T, params *exprTypeTestParams) {
	config.KPHP = true
	runExprTypeTest(t, params)
	config.KPHP = false
}
