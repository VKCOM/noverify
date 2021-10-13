package exprtype_test

import (
	"testing"
)

func TestLiteralAsType(t *testing.T) {
	code := `<?php
/**
 * @param '!'|'?'|'$' $a
 */
function f($a) {
  exprtype($a, "string");
}

/**
 * @param 'abd'|'abc' $a
 */
function f1($a) {
  exprtype($a, "string");
}

/**
 * @param '!='|'<'|'<='|'<>'|'='|'=='|'>'|'>='|'eq'|'ge'|'gt'|'le'|'lt'|'ne' $a
 */
function f2($a) {
  exprtype($a, "string");
}

/**
 * @return '!='|'<'|'<='|'<>'|'='|'=='|'>'|'>='|'eq'|'ge'|'gt'|'le'|'lt'|'ne'
 */
function f3() { return "!="; }
exprtype(f3(), "string");

/**
 * @param 'abd'|int $a
 */
function f4($a) {
	exprtype($a, "int|string");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}
