package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue778(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
trait FooStatic {
    /** @return void */
    static function f() {
        self::g(); // g() is expected to be defined outside of the FooStatic trait
    }
}

trait FooInstance {
    /** @return void */
    function f() {
        $this->g(); // g() is expected to be defined outside of the FooInstance trait
    }
}
`)
}
