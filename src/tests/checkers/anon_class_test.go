package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestSimpleAnonClass(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
		declare(strict_types = 1);
function f() {
  $a = new class {
    /** */
    public function f() {}
  };

  $a->f();
}
`)
}

func TestAnonClassAsInterface(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);
interface IFace {}

function f(IFace $if) {}

f(new class implements IFace {});
`)
}

func TestAnonClassFromDocumentation(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types = 1);
class Outer {
  private $prop = 1;
  protected $prop2 = 2;

  protected function func1() {
    return 3;
  }

  /** */
  public function func2() {
    return new class($this->prop) extends Outer {
      private $prop3;

      /** */
      public function __construct($prop) {
        $this->prop3 = $prop;
      }

      /** */
      public function func3() {
        return $this->prop2 + $this->prop3 + $this->func1();
      }
    };
  }
}

echo (new Outer)->func2()->func3();
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestAnonClassWithConstructor(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);
function f() {
  $a = new class(100, "s") {
    /** */
    public function f() {}

    public function __construct(int $a, string $b) {
      echo $a;
      echo $b;
    }
  };

  $a->f();
}
`)
}

func TestAnonClassWithExtends(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);
class Boo {
  /** */
  public function b() {}
}

function f() {
  $a = new class extends Boo {
    /** */
    public function f() {}
  };

  $a->f();
  $a->b();
}
`)
}

func TestAnonClassWithImplements(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);
interface IBoo {
  /** */
  public function b() {}
}

function f() {
  $a = new class implements IBoo {
    /** */
    public function b() {}
  };

  $a->b();
}
`)
}

func TestAnonClassWithSeveralImplements(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);

interface IBoo {
  /** */
  public function b() {}
}

interface IFoo {
  /** */
  public function f() {}
}

function f() {
  $a = new class implements IBoo {
    /** */
    public function b() {}

    /** */
    public function f() {}
  };

  $a->b();
  $a->f();
}
`)
}

func TestAnonClassWithImplementsError(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface IBoo {
  public function b() {}
}

function f() {
  $a = new class implements IBoo {
    public function f() {}
  };

  $a->f();
}
`)
	test.Expect = []string{
		`Class \anon$(_file0.php):7$ must implement \IBoo::b method`,
	}
	linttest.RunFilterMatch(test, "unimplemented")
}

func TestAnonClassWithConstructorArgsMismatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f() {
  $a = new class(10) {
    public function __construct(int $a, string $b) {}
    public function f() {}
  };

  $a->f();

  $a = new class {
    public function __construct(int $a) {}
    public function f() {}
  };

  $a->f();

  $a = new class(1, 2, 3) {
    public function __construct(int ...$a) {}
    public function f() {}
  };

  $a->f();

  $a = new class {
    public function __construct(int ...$a) {}
    public function f() {}
  };

  $a->f();

  $a = new class {
    public function __construct(string $b, int ...$a) {}
    public function f() {}
  };

  $a->f();

  $a = new class("hello") {
    public function __construct(string $b, int ...$a) {}
    public function f() {}
  };

  $a->f();
}
`)
	test.Expect = []string{
		`Too few arguments for \anon$(_file0.php):3$ constructor, expecting 2, saw 1`,
		`Too few arguments for \anon$(_file0.php):10$ constructor, expecting 1, saw 0`,
		`Too few arguments for \anon$(_file0.php):31$ constructor, expecting 1, saw 0`,
	}
	linttest.RunFilterMatch(test, "argCount")
}

func TestAnonClassAsReturn(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f() {
  return new class {
    public function f() {}
  };
}

f()->f();
f()->f1();
`)
	test.Expect = []string{
		`Call to undefined method {\anon$(_file0.php):3$}->f1()`,
	}
	linttest.RunFilterMatch(test, "undefinedMethod")
}

func TestAnonClassInsideOther(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f() {
  return new class {
    public function f() {
      $a = new class {
         public function g() {}
      };

      $a->g();

      return $a;
    }
  };
}

f()->f();
f()->f1();
f()->f()->g();
f()->f()->g1();
`)
	test.Expect = []string{
		`Call to undefined method {\anon$(_file0.php):3$}->f1()`,
		`Call to undefined method {\anon$(_file0.php):5$}->g1()`,
	}
	linttest.RunFilterMatch(test, "undefinedMethod")
}

func TestAnonClassInsideOtherInsideClass(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
namespace Test;

class Foo {
  public static function f() {
    return new class {
      public function f() {
        $a = new class {
           public function g() {}
        };
  
        $a->g();
  
        return $a;
      }
    };
  }
}

Foo::f()->f();
Foo::f()->f1();
Foo::f()->f()->g();
Foo::f()->f()->g1();
`)
	test.Expect = []string{
		`Call to undefined method {\Test\anon$(_file0.php):6$}->f1()`,
		`Call to undefined method {\Test\anon$(_file0.php):8$}->g1()`,
	}
	linttest.RunFilterMatch(test, "undefinedMethod")
}

func TestAnonClassWithTrait(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait A {
  public function traitMethod() {}
}

function f() {
  return new class {
    use A;

    public function f() {}
  };
}

f()->f();
f()->traitMethod();
f()->f1();
`)
	test.Expect = []string{
		`Call to undefined method {\anon$(_file0.php):7$}->f1()`,
	}
	linttest.RunFilterMatch(test, "undefinedMethod")
}
