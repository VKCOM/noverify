package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue183(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types = 1);
    trait Mixin {
        public $x = 10;
    }

    class MyClass {
        use Mixin;

        /** @return int */
        public function useX() { return $this->x; }
        /** @return int */
        public function useY() { return $this->y; }
    }
`)

	test.Expect = []string{
		`Property {\MyClass}->y does not exist`,
	}

	test.RunAndMatch()
}
