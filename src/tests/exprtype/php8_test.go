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

func TestExprReturnTypeAware(t *testing.T) {
	code := `<?php
use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;

#[LanguageLevelTypeAware(["8.0" => "string[]"], default: "string[]|false")]
function f() {}
function f1() {}

exprtype(f(), "false|string[]");
exprtype(f1(), "void");

class Foo {
  #[LanguageLevelTypeAware(["8.0" => "string[]"], default: "string[]|false")]
  public static function f() {}
  public static function f1() {}
}

exprtype(Foo::f(), "false|string[]");
exprtype(Foo::f1(), "void");
`
	runPHP8ExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestParamTypeAware(t *testing.T) {
	code := `<?php
use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;

function f(
  #[LanguageLevelTypeAware(["8.0" => "string[]"], default: "string[]|false")] $param,
  $param1,
  #[LanguageLevelTypeAware(["8.0" => "string[]"], default: "string[]|false")] int $param2,
) {
  exprtype($param, "false|string[]");
  exprtype($param1, "mixed");
  exprtype($param2, "false|int|string[]");
}

class Foo {
  public function f(
    #[LanguageLevelTypeAware(["8.0" => "string[]"], default: "string[]|false")] $param,
    $param1,
    #[LanguageLevelTypeAware(["8.0" => "string[]"], default: "string[]|false")] int $param2,
  ) {
    exprtype($param, "false|string[]");
    exprtype($param1, "mixed");
    exprtype($param2, "false|int|string[]");
  }
}
`
	runPHP8ExprTypeTest(t, &exprTypeTestParams{code: code})
}

func runPHP8ExprTypeTest(t *testing.T, params *exprTypeTestParams) {
	exprTypeTestImpl(t, params, false)
}
