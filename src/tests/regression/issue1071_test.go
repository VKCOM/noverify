package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue1071FunctionWithBackSlash(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{
		"stubs/phpstorm-stubs/standard/standard_8.php",
		"stubs/phpstorm-stubs/meta/.phpstorm.meta.php",
	}
	test.AddFile(`<?php
	declare(strict_types = 1);
class Foo {
  /**
   * @return int
   */
  public function f(): int { return 0; }
}

function f() {
  $a = [new Foo];
  $b = reset($a);
  $b1 = \reset($a);
  echo $b->f();
  echo $b1->f();
}
`)
	test.RunAndMatch()
}
