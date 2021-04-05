package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDisableDeadCode(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function f() {
#ifndef KPHP
    return 1;
#endif

    /** @noinspection PhpUnreachableStatementInspection */
    return 10;
}

function f1() {
#ifndef KPHP
    return 1;
#endif

    /** @noinspection PhpUnreachableStatementInspection */
	$a = 100;

    return $a;
}

function f2() {
#ifndef KPHP
    return 1;
#endif

#ifndef KPHP2
	/** @noinspection PhpUnreachableStatementInspection */
    return 1;
#endif

    /** @noinspection PhpUnreachableStatementInspection */
	$a = 100;

    return $a;
}
`)
}
