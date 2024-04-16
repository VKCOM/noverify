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
