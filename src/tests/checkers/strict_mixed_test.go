package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestStrictMixedEnabled(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
class Foo {}

function f($a) {
  $a->f();
}

function f(object $a) {
  $a->f();
}

function f() {
  $a->f();
}

function f($b) {
  [$a,] = $b;
  $a->f();
}

function f(Foo $a) {
  $a->f();
}

function f() {
  $a = null;
  $a->f();
}

function f(stdClass $a) {
  $a->f();
}
`,
	)
	test.Expect = []string{
		"Call to undefined method {mixed}->f()",
		"Call to undefined method {object}->f()",
		"Undefined variable $a",
		"Call to undefined method {undefined}->f()",
		"Call to undefined method {unknown_from_list}->f()",
		"Call to undefined method {\\Foo}->f()",
		"Call to undefined method {null}->f()",
		"Call to undefined method {\\stdClass}->f()",
	}
	test.RunAndMatch()
}

func TestStrictMixedDisabled(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {}

function f($a) {
  $a->f();
}

function f(object $a) {
  $a->f();
}

function f() {
  $a->f();
}

function f($b) {
  [$a,] = $b;
  $a->f();
}

function f(Foo $a) {
  $a->f();
}

function f() {
  $a = null;
  $a->f();
}

function f(stdClass $a) {
  $a->f();
}
`,
	)
	test.Expect = []string{
		"Undefined variable $a",
		"Call to undefined method {\\Foo}->f()",
	}
	test.RunAndMatch()
}
