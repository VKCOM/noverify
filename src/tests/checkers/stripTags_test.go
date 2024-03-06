package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestStripTagsGood(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{"stubs/phpstorm-stubs/standard/standard_1.php"}
	test.AddFile(`<?php
declare(strict_types=1);
function f(string $s) {
  $_ = strip_tags($s, ['br', 'a']);
  $_ = strip_tags($s, ['BR']);

  $_ = strip_tags($s, ' <a><br><html> ');
  $_ = strip_tags($s, '<br>');
}
`)
	test.RunAndMatch()
}

func TestStripTagsArray(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{"stubs/phpstorm-stubs/standard/standard_1.php"}
	test.AddFile(`<?php
declare(strict_types=1);
function f(string $s) {
  return strip_tags($s, ['<br>', 'a', 'A']);
}
`)
	test.Expect = []string{
		`$allowed_tags argument: '<' and '>' are not needed for tags when using array argument`,
		`$allowed_tags argument: tag 'A' is duplicated, previously spelled as 'a'`,
	}
	test.RunAndMatch()
}

func TestStripTagsString1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{"stubs/phpstorm-stubs/standard/standard_1.php"}
	test.AddFile(`<?php
declare(strict_types=1);
function f1(string $s) {
  return strip_tags($s, '<a href="#">');
}
function f2(string $s) {
  return strip_tags($s, "<a href='#'>");
}
`)
	test.Expect = []string{
		`$allowed_tags argument: using values/attrs is an error; they make matching always fail`,
		`$allowed_tags argument: using values/attrs is an error; they make matching always fail`,
	}
	test.RunAndMatch()
}

func TestStripTagsString2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{"stubs/phpstorm-stubs/standard/standard_1.php"}
	test.AddFile(`<?php
declare(strict_types=1);
function f(string $s) {
  return strip_tags($s, '<a ><br/><br>');
}
`)
	test.Expect = []string{
		`$allowed_tags argument: tag '<a >' should not contain spaces`,
		`$allowed_tags argument: '<br/>' should be written as '<br>'`,
		`$allowed_tags argument: tag '<br>' is duplicated, previously spelled as '<br/>'`,
	}
	test.RunAndMatch()
}
