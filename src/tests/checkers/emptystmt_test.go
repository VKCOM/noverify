package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestEmptyStmtBadComment(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
/** @var $foo int */;
global $foo;
`)
	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed`,
	}
	test.RunAndMatch()
}

func TestEmptyStmtBadStmtEnd(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
if ($argv) {
};
while ($argv) {
};
class Foo {};
`)
	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed`,
		`Semicolon (;) is not needed here, it can be safely removed`,
		`Semicolon (;) is not needed here, it can be safely removed`,
	}
	test.RunAndMatch()
}

func TestEmptyStmtBadRequire1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
require_once 'foo.php';;
`)
	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed`,
	}
	test.RunAndMatch()
}

func TestEmptyStmtBadRequire2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
require_once 'foo.php'; ; ;
`)
	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed`,
		`Semicolon (;) is not needed here, it can be safely removed`,
	}
	test.RunAndMatch()
}

func TestEmptyStmtBadStmtSeparator(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
echo 1;
; echo 2;
echo 3; ;
`)
	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed`,
		`Semicolon (;) is not needed here, it can be safely removed`,
	}
	test.RunAndMatch()
}

func TestEmptyStmtScriptEnd1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
function f() {}
?>
`)
}

func TestEmptyStmtScriptEnd2(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
echo 123;
?>
`)
}

func TestEmptyStmtGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
if ($argv) ;
while ($argv) ;
for (;;) ;
foreach ($argv as $_) ;
do ; while ($argv);
declare (strict_types=1);
`)
}
