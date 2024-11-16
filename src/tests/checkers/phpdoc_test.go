package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPHPDocRefs(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`<?php
namespace A;

/**
 * @see B\ABClass OK - relative class ref
 * @see B\ABClass;
 * @see BACK-134
 * @see #4393
 * @see RFC: Generic Types and Functions
 */
function f() {
}

interface Iface {
  public function iface_method();
}
`)
	test.AddFile(`<?php
namespace A\B;

class ABclass {
  public $prop = 1;

  /**
   * @see $prop OK - refs class prop
   * @see prop OK - "$" is optional
   * @see \A\Iface::iface_method()
   */
  public static $static_prop = 2;
}

class HolyHandGrenade {
  public function hallelujah() {}
}

/**
 * @see abfunc1(), abfunc2() OK - two refs
 */
function abfunc1() {}

/**
 * @see abfunc1 OK - ref to a local symbol
 * @see \A\B\abfunc1 OK - FQN ref
 * @see abfunc1() OK - ref with parens
 * @see f1... OK - global func ref with junk
 * @see CONST1 OK - global const ref
 * @see CONST2 OK - global const ref
 * @see \CONST1 OK - FQN const ref
 * @see CONST1@ OK - global const ref with junk
 * @see \Foo::method1() OK - class method ref
 * @see HolyHandGrenade::hallelujah OK - class method ref
 * @see \Foo::FOOCONST OK - class const ref
 */
function abfunc2() {}
`)

	test.AddFile(`<?php
use A\B\HolyHandGrenade;

const CONST1 = 1;
define('CONST2', 2);

function f1() {}

class Foo {
  const FOOCONST = 'foo';

  /**
   * @see http://google.com - OK, URL
   * @see method2 OK - refs current class method
   * @see self::method2 OK - refs current class method
   * @see f1 OK - global (current namespace) function refer
   * @see \A\B\abfunc1 OK - global function ref
   * @see \A\B\ABclass OK - FQN class ref
   * @see HolyHandGrenade OK - class imported with "use"
   * @see HolyHandGrenade::hallelujah OK - imported class method
   * @see \A\B\ABclass::$prop OK - prop ref
   * @see \A\B\ABclass::$static_prop OK - static prop ref
   */
  public function method1() {
  }

  /**
   * @see foo.php
   * @see foo.js:10
   * @see self::* consts in this class
   * @see self::FOO* consts in this class
   * @see CONST1.
   * @see self::class
   * @see
   */
  public function method2() {}
}

class Bar {
}
`)

	test.AddFile(`<?php
namespace Bad;

/**
 * @see HolyHandGrenade::hallelujah... BAD - HolyHandGrenade is in other namespace
 */
function f() {}

/**
 * @see CONST43@ BAD - CONST43 is undefined
 */
class BadClass {
  /**
   * @see \NonExisting::class BAD - non-existing class
   */
  public function m() {}

  /**
   * @see invalid1, invalid2 BAD - non-existing symbol
   */
  private $prop = 10;
}
`)

	test.Expect = []string{
		`@see tag refers to unknown symbol \NonExisting::class`,
		`@see tag refers to unknown symbol HolyHandGrenade::hallelujah`,
		`@see tag refers to unknown symbol CONST43`,
		`@see tag refers to unknown symbol invalid1`,
		`@see tag refers to unknown symbol invalid2`,
	}
	linttest.RunFilterMatch(test, "invalidDocblockRef")
}

func TestPHPDocRefForConstantInClass(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`<?php
const TYPE_TEXT_GLOBAL = 0;

class FooAbstract {
  /** Text headers */
  const TYPE_TEXT_PARENT = 2;
}

class Foo extends FooAbstract {
  const TYPE_TEXT = 2;

  /**
   * Get the type of Header that this instance represents.
   *
   * @see TYPE_TEXT
   * @see TYPE_TEXT_PARENT
   * @see TYPE_TEXT_UNDEFINED
   * @see TYPE_TEXT_GLOBAL
   *
   * @return int
   */
  public function getFieldType()
  {
    return self::TYPE_TEXT;
  }
}

/**
 * @see TYPE_TEXT
 * @see TYPE_TEXT_GLOBAL
 */
function f() {}
`)
	test.Expect = []string{
		`@see tag refers to unknown symbol TYPE_TEXT_UNDEFINED`,
		`@see tag refers to unknown symbol TYPE_TEXT`,
	}
	linttest.RunFilterMatch(test, "invalidDocblockRef")
}

func TestBadParamName(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
class B1 {}
class B2 {}

/**
 * @param B1 $v2
 * @param B2 $v3
 */
function f($v1, $v2) {
}

class Bear {
  /** @param int $y */
  private function migrate($x) {}
}
`)
	test.Expect = []string{
		`@param for non-existing argument $v3`,
		`@param for non-existing argument $y`,
	}
	test.RunAndMatch()
}

func TestDeprecatedMethod(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
class Foo {
  /**
   * @deprecated use newMethod instead
   */
  public function legacyMethod1() {}

  /**
   * @deprecated
   */
  public function legacyMethod2() {}
}

(new Foo())->legacyMethod1();
function f() {
  (new Foo())->legacyMethod2();
}
`)
	test.Expect = []string{
		`Call to deprecated method {\Foo}->legacyMethod1() (reason: use newMethod instead)`,
		`Call to deprecated method {\Foo}->legacyMethod2()`,
	}
	test.RunAndMatch()
}

func TestDeprecatedStaticMethod(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
class Foo {
  /**
   * @deprecated use newMethod instead
   */
  public static function legacyMethod1() {}

  /**
   * @deprecated
   */
  public static function legacyMethod2() {}
}

Foo::legacyMethod1();
function f() {
  Foo::legacyMethod2();
}
`)
	test.Expect = []string{
		`Call to deprecated static method \Foo::legacyMethod1() (reason: use newMethod instead)`,
		`Call to deprecated static method \Foo::legacyMethod2()`,
	}
	test.RunAndMatch()
}

func TestDeprecatedFunction(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
/**
 * @deprecated use new_function instead
 */
function legacy_function1() {}

/**
 * @deprecated
 */
function legacy_function2() {}

legacy_function1();
function f() {
  legacy_function2();
}
`)
	test.Expect = []string{
		`Call to deprecated function legacy_function1 (reason: use new_function instead)`,
		`Call to deprecated function legacy_function2`,
	}
	test.RunAndMatch()
}

func TestBadPHPDocTypes(t *testing.T) {
	// If there is an incorrect phpdoc annotation,
	// don't use it as a type info.
	//
	// Before the fix, NoVerify inferred \$a and \$b to be
	// types for corresponding params, which is incorrect.

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
/**
 * @param $a
 * @param $b
 * @return int
 */
function fav_func($a, $b) {
  if ($a[0] != $b[0]) {
    return ($a[0] > $b[0]) ? -1 : 1;
  }
  return ($a[1] < $b[1]) ? -1 : 1;
}
`)
	test.Expect = []string{
		`Malformed @param $a tag (maybe type is missing?)`,
		`Malformed @param $b tag (maybe type is missing?)`,
	}
	test.RunAndMatch()
}

func TestPHPDocPresence(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
	trait TheTrait {
		public function traitPub() {}
	}
	class TheClass {
		/**
		 * This function is a good example.
		 * It's public and has a PHPDoc comment.
		 */
		public function documentedPub() {}

		// Not OK.
		public function pub() {}

		// OK.
		private function priv() {}

		// OK.
		protected function prot() {}
	}`)
	test.Expect = []string{
		`Missing PHPDoc for \TheClass::pub public method`,
		`Missing PHPDoc for \TheTrait::traitPub public method`,
	}
	test.RunAndMatch()
}

func TestPHPDocSyntax(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types = 1);
	/**
	 * @param $x int the x param
	 * @param - $y the y param
	 * @param $z - the z param
	 * @param $a
	 * @param int
	 */
	function f($x, $y, $z, $a, $_) {
		$_ = $x;
		$_ = $y;
		$_ = $z;
	}`)
	test.Expect = []string{
		`Non-canonical order of variable and type`,
		`Expected a type, found '-'; if you want to express 'any' type, use 'mixed'`,
		`Non-canonical order of variable and type`,
		`Expected a type, found '-'; if you want to express 'any' type, use 'mixed'`,
		`Malformed @param $a tag (maybe type is missing?)`,
		`Malformed @param tag (maybe var is missing?)`,
	}
	test.RunAndMatch()
}

func TestPHPDocVar(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
class Foo {
  /** @var integer $x */
  public $x;

  /** @var real */
  public $x;

  /** @var int? */
  public $x;
}
`)
	test.Expect = []string{
		`Use int type instead of integer`,
		`Use float type instead of real`,
		`Nullable syntax is ?T, not T?`,
	}
	test.RunAndMatch()
}

func TestPHPDocProperty2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
/**
 * @property integer $int
 * @property []t ts
 * @property
 * @property string
 * @property $int string
 * @property boolean[] $bools
 * @property int? $nullable
 */
class Foo {}
`)
	test.Expect = []string{
		`Use int type instead of integer`,
		`Array syntax is T[], not []T`,
		`@property ts field name must start with '$'`,
		`Non-canonical order of name and type`,
		`@property requires type and property name fields`,
		`@property requires type and property name fields`,
		`Use bool type instead of boolean`,
		`Nullable syntax is ?T, not T?`,
	}
	test.RunAndMatch()
}

func TestPHPDocType(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types = 1);
	/**
	 * @param [][]string $x1
	 * @param double $x2
	 * @param real $x3
	 * @param integer $x4
	 * @param boolean $x5
	 * @param int? $x6
	 * @return []int
	 */
	function f($x1, $x2, $x3, $x4, $x5, $x6) {
		$_ = [$x1, $x2, $x3, $x4, $x5, $x6];
		return [1];
	}
`)
	test.Expect = []string{
		`Array syntax is T[], not []T`,
		`Use float type instead of double`,
		`Use float type instead of real`,
		`Use int type instead of integer`,
		`Use bool type instead of boolean`,
		`Nullable syntax is ?T, not T?`,
		`Array syntax is T[], not []T`,
	}
	test.RunAndMatch()
}

func TestPHPDocIncorrectSyntaxOptionalTypesType(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types = 1);
	class Foo {}

	/**
	 * @param int? $x1 // error
	 * @param shape(key:int, opt?:int) $x2 // ok, is shape
	 * @param shape(key?:int, opt?:int) $x3 // ok, is shape
	 * @param Foo? $x4 // error
	 * @param string[]? $x5 // error
	 */
	function f1($x1, $x2, $x3, $x4, $x5) {
		$_ = [$x1, $x2, $x3, $x4, $x5];
	}
`)
	test.Expect = []string{
		`Nullable syntax is ?T, not T?`,
		`Nullable syntax is ?T, not T?`,
		`Nullable syntax is ?T, not T?`,
	}
	test.RunAndMatch()
}

func TestPHPDocInvalidBeginning(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
class Foo {
  /*
   * @var int
   */
  public $item = 10;

  /* @var int */
  public $item2 = 10;

  /*
   * @return int
   */
  public function f() { return 1; }

  /* @return int */
  public function f2() { return 1; }
}

/*
 * @param int $a
 */
function f($a) {
  /*
   * @var $b float
   */
  $b = 100;

  // TODO: @var string $a (ok)
  echo $a, $b;
}

/* @param int $a */
function f2($a) {
  /* @var $b float */
  $b = 100;

  // TODO: @var string $a (ok)
  echo $a, $b;
}
`)
	test.Expect = []string{
		`Multiline PHPDoc comment should start with /**, not /*`,
		`Multiline PHPDoc comment should start with /**, not /*`,
		`Multiline PHPDoc comment should start with /**, not /*`,
		`Multiline PHPDoc comment should start with /**, not /*`,
		`Multiline PHPDoc comment should start with /**, not /*`,
	}
	test.RunAndMatch()
}

func TestPHPDocTwiceNullableTypes(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);
class Foo {}

/**
 * @param ?int|null $a
 * @param ?string[]|null $b
 * @param ?Foo|null $c
 * @param shape(name: ?string, age: int)|null $d
 * @return ?int|null
 */
function f($a, $b, $c, $d) {
  echo $a;
  echo $b;
  echo $c;
  echo $d;
  return null;
}
`)
	test.Expect = []string{
		`Repeated nullable doesn't make sense`,
		`Repeated nullable doesn't make sense`,
		`Repeated nullable doesn't make sense`,
		`Repeated nullable doesn't make sense`,
	}
	test.RunAndMatch()
}
