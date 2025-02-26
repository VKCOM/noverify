package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDeprecatedWithoutText(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class WithoutText {
  /**
   * @deprecated
   */
  public function method() {}

  /**
   * @deprecated
   */
  public static function staticMethod() {}
}

/**
 * @deprecated
 */
function funcWithoutText() {}

funcWithoutText();
(new WithoutText)->method();
WithoutText::staticMethod();

`)
	test.Expect = []string{
		"Call to deprecated function funcWithoutText",
		"Call to deprecated method {\\WithoutText}->method()",
		"Call to deprecated static method \\WithoutText::staticMethod()",
	}
	test.RunAndMatch()
}

func TestDeprecatedWithText(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class WithText {
  /**
   * @deprecated use method2() instead
   */
  public function method() {}

  /**
   * @deprecated use staticMethod2() instead
   */
  public static function staticMethod() {}
}

/**
 * @deprecated use funcWithText2() instead
 */
function funcWithText() {}

funcWithText();
(new WithText)->method();
WithText::staticMethod();

`)
	test.Expect = []string{
		"Call to deprecated function funcWithText (reason: use funcWithText2() instead)",
		"Call to deprecated method {\\WithText}->method() (reason: use method2() instead)",
		"Call to deprecated static method \\WithText::staticMethod() (reason: use staticMethod2() instead)",
	}
	test.RunAndMatch()
}

func TestDeprecatedClassAndProperty(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = "1");

/**
 * @deprecated class 
 */
class OldClass
{
/**
 * @deprecated prp
 */
public $deprecated;

public $notDeprecated;
}

$v = new OldClass();
echo $v->deprecated;
`)
	test.Expect = []string{
		"Try to create instance of the class \\OldClass that was marked as deprecated",
		"Try to call property deprecated that was marked as deprecated",
	}
	test.RunAndMatch()
}

func TestDeprecatedClassPropertyConstants(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

/**
 * @deprecated class
 */
class SomeClass {

  /**
   * @deprecated Use NEW_ONE instead
   *
   * @see SomeClass::NEW_ONE Replacement
   */
  public const OLD_ONE = 1;

  public const NEW_ONE = 1;

  /** @deprecated */
  public const OLD_TWO = 2;
  public const NEW_TWO = 2;

  /** @deprecated */
  public $a = 5;

  /**
   * @deprecated Use xyz() instead
   */
  public static function abc() {

  }

  /** Some phpdoc to hide notice */
  public static function xyz() {

  }
}

$b = new SomeClass();

echo $b->a;

echo SomeClass::OLD_ONE;

echo SomeClass::OLD_TWO;
echo SomeClass::abc();

`)
	test.Expect = []string{
		"Try to create instance of the class \\SomeClass that was marked as deprecated",
		"Try to call property a that was marked as deprecated in the class \\SomeClass",
		"Try to call constant OLD_ONE that was marked as deprecated",
		"Try to call constant OLD_TWO that was marked as deprecated",
		"Call to deprecated static method \\SomeClass::abc()",
	}
	test.RunAndMatch()
}

func TestDeprecatedCallWithChainPropertyAndExplicitType(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

class Baz {
    /**
     * @deprecated Use method getValue() instead of this property.
     */
    public string $value = 'Hello, World!';

    public function getValue(): string {
        return 'Hello from new method!';
    }
}

class Bar {
    public Baz $baz;
}

class Foo {
    public Bar $bar;
}

$a = new Foo();
$a->bar = new Bar();
$a->bar->baz = new Baz();

echo $a->bar->baz->value;

`)
	test.Expect = []string{
		"Missing PHPDoc for \\Baz::getValue public",
		"Try to call property value that was marked as deprecated",
	}
	test.RunAndMatch()
}

// TODO: This test must fail after realisation CFG and DFG, because we can type inference correctly
func TestDeprecatedCallWithChainPropertyAndWithoutExplicitType(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Baz {
    /**
     * @deprecated Use method getValue() instead of this property
     */
    public $value = 'Hello, World!';

    public function getValue() {
        return 'Hello from new method!';
    }
}

class Bar {
    public $baz;
}

class Foo {
    public $bar;
}

$a = new Foo();
$a->bar = new Bar();
$a->bar->baz = new Baz();

echo $a->bar->baz->value;
`)
	test.Expect = []string{
		"Missing PHPDoc for \\Baz::getValue public method",
	}
	test.RunAndMatch()
}
