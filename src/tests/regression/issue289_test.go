package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue289(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types=1);
class Foo { public $value = 11; }

$xs = [0, new Foo()];

/* @var Foo $foo */
$foo = $xs[1];
$_ = $foo->value;
`)
}
