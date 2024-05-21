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
 * @deprecated property
 */
public $prp;
}

`)
	test.Expect = []string{
		"Has deprecated class OldClass",
		"Has deprecated field in class OldClass",
	}
	test.RunAndMatch()
}
