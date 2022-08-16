package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestLinterSuppressDeprecationCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * @deprecated some
 */
function f($a) {
  return 0;
}

function g($b) {
  /**
   * @noverify-suppress deprecated
   */
  echo f(
    f(10)
  );
}

/**
 * @noverify-suppress deprecated
 */
echo f(10);
`)
	test.RunAndMatch()
}

func TestLinterSuppressAll(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function g($b) {
  /** @noverify-suppress all */
  $a = $b && $b;
}
`)
	test.RunAndMatch()
}

func TestLinterSuppressUndefinedClass(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/** @noverify-suppress undefinedClass */
echo new Foo;
`)
	test.RunAndMatch()
}

func TestLinterSuppressNotAll(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function g($b) {
  /** @noverify-suppress dupSubExpr */
  $a = $b && $b;
}
`)
	test.Expect = []string{
		`Variable $a is unused`,
	}
	test.RunAndMatch()
}

func TestLinterSuppressDeadCode(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f($b) {
  return $b;

  /**
   * @noverify-suppress deadCode
   */
  return 10;
}

function f($b) {
  return $b;

  /** @noverify-suppress deadCode */
  return 10;
}
`)
	test.RunAndMatch()
}

func TestLinterSuppressWarningInGlobalScope(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$a = 100;

/**
 * @noverify-suppress all
 */
echo $a && $a;
`)
	test.RunAndMatch()
}

func TestLinterSuppressWarnings(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {
  private string $a = "hello";
  /** */
  public function f() {}
}

function g($b) {
  /** @noverify-suppress paramClobber */
  $b = 100;

  /** @noverify-suppress dupSubExpr */
  echo $b && $b;

  /** @noverify-suppress dupBranchBody */
  if ($b) {
    echo 1;
  } else {
    echo 1;
  }

  /** @noverify-suppress accessLevel */
  echo (new Foo)->a;

  /** @noverify-suppress undefinedMethod */
  echo (new Foo)->undefined();

  /** @noverify-suppress mixedArrayKeys */
  echo [10, 20, "some" => 30];

  /** @noverify-suppress all */
  switch ($b) {
    case 1:
      echo 1;
      break;
  }

  return $b;
}
`)
	test.RunAndMatch()
}
