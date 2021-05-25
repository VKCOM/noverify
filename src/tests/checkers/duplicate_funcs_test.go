package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDuplicateFuncs(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function alert() {
    echo 1;
}
`)
	test.AddFile(`<?php
function Alert() {
    echo 1;
}

alert();
`)
	test.Expect = []string{}
	test.RunAndMatch()
}
