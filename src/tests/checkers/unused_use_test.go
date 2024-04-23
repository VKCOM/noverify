package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestUsedUse1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

use Random\RandomException;

use Random\DeadCode;

use function My\Full\functionName;

use function My\Full\functionName as func;

$a = new Random\RandomException;

`)
	test.Expect = []string{
		"Unused `use` statement",
		"Unused `use` statement",
	}
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestUsedUse2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

use Random\RandomException;

use Random\DeadCode;

use function My\Full\functionName;

use function My\Full\functionName as func;

$a = new Random\RandomException;
My\Full\functionName();

`)
	test.Expect = []string{
		"Unused `use` statement",
	}
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestUsedUse3(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

use Random\RandomException;

use Random\DeadCode;

use function My\Full\functionName;

use function My\Full\functionName as func;

$a = new Random\RandomException;
My\Full\functionName();
$b = new Random\DeadCode;

`)

	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestUsedUse4(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

use QQ\WW\SomeQQClass;
use QQ\WW\SomeQQClass1 as SomeQQClass2;

/**
 * @mixin SomeQQClass
 */
class SomeClass {
  /** */
  public function method()
  {
    echo $this->methodQQ1();
  }
}

`)
	test.Expect = []string{
		"Unused `use` statement",
	}

	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestUsedUse5(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

use QQ\WW\SomeQQClass;
use QQ\WW\SomeQQClass1 as SomeQQClass2;

/**
 * @mixin SomeQQClass
 */
class SomeClass {
  /** */
  public function method()
  {
    echo $this->methodQQ1();
  }
}

/** 
 * @mixin SomeQQClass2
 */
class SomeClass2 {
  /** */
  public function method3()
  {
    echo "";
  }
}

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestUnusedUseDeprecatedAttribute(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`
<?php

use JetBrains\PhpStorm\Deprecated;

#[Deprecated]
function deprecated() {}

#[Deprecated(reason: "use X instead")]
function deprecatedReason() {}

deprecated();
deprecatedReason();

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")

}

func TestExtends(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`
<?php

declare(strict_types=1);

namespace A\B\C;

use A\B\C;

class A extends C {
}

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestMethodParams(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`
<?php

declare(strict_types=1);

namespace A\B\C;

use A\B\D;

class TestClass
{
    public static function testParamValue(D $value) : void
    {

    }
}

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestFunctionParams(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`
<?php

declare(strict_types=1);

namespace A\B\C;

use A\B\D;

class TestClass
{
    public static function testParamValue(D $value) : void
    {

    }
}

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestInterfaceImpl(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`
<?php

namespace A\B\C;

use A\B\CInterface;

abstract class AbstractClass implements CInterface
{

}

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestCases(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`
<?php

declare(strict_types=1);

namespace Brick\Math\Internal;

use Brick\Math\Exception\RoundingNecessaryException;
use Brick\Math\RoundingMode;

function divRound(int $roundingMode) : void
    {
        switch ($roundingMode) {
            case RoundingMode::UNNECESSARY:
                    throw RoundingNecessaryException::roundingNecessary();
                break;

            case RoundingMode::UP:
                break;

            case RoundingMode::DOWN:
                break;

            default:
                throw new \InvalidArgumentException('Invalid rounding mode.');
        }
    }

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")
}

func TestTraits(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`
<?php

namespace A\B\C;

use A\B\C\D\SomeTrait;
use A\B\C\D\SomeTrait2;


class SomeClass extends C
{
    use SomeTrait;
    use SomeTrait2;
}

`)
	linttest.RunFilterMatch(test, "unusedUseStatements")
}
