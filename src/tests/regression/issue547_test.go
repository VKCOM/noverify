package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue547(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/standard/standard_3.php`}
	test.AddFile(`<?php
putenv("A=1");
\putenv("B=2");
`)
	test.RunAndMatch()
}
