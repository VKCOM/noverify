package exprtype_test

import (
	"testing"
)

func TestExprTypeUnionParam(t *testing.T) {
	code := `<?php
class Foo {}
class Boo {}

function f(int|string $a) {
  exprtype($a, 'int|string');
}

function f(int|string|float $a) {
  exprtype($a, 'float|int|string');
}

function f(int|Foo|float $a) {
  exprtype($a, '\Foo|float|int');
}

function f(Boo|Foo $a) {
  exprtype($a, '\Boo|\Foo');
}
`
	runPHP8ExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeUnionReturn(t *testing.T) {
	code := `<?php
class Foo {}
class Boo {}

function f1(): int|string {}
exprtype(f1(), 'int|string');

function f2(): int|string|float {}
exprtype(f2(), 'float|int|string');

function f3(): int|Foo|float {}
exprtype(f3(), '\Foo|float|int');

function f4(): Boo|Foo {}
exprtype(f4(), '\Boo|\Foo');
`
	runPHP8ExprTypeTest(t, &exprTypeTestParams{code: code})
}

func runPHP8ExprTypeTest(t *testing.T, params *exprTypeTestParams) {
	exprTypeTestImpl(t, params, false)
}
