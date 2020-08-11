package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestRelativeName1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
namespace A;

class ClassA {}

$_ = new namespace\B\ClassB();
$_ = new namespace\ClassA();
`)

	test.AddFile(`<?php
namespace A\B;

class ClassB {}

$_ = new namespace\ClassB();
`)

	test.AddFile(`<?php
$_ = new namespace\A\B\ClassB();
`)

	test.RunAndMatch()
}
