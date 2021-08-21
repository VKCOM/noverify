package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue390(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
$cond = 1;
if ($cond && isset($a1[0])) {
  $_ = $a1;
}
if ($cond && isset($a2[0][1])) {
  $_ = $a2;
}
if (isset($a3[0]) && $cond) {
  $_ = $a3;
}
if (isset($a4[0][1]) && $cond) {
  $_ = $a4;
}

if (isset($a5[0]->x)) {
  $_ = $a5;
}
`)
	test.Expect = []string{
		`Property {mixed}->x does not exist`,
	}
	test.RunAndMatch()
}
