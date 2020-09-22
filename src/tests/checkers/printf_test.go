package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestSprintf(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadedStubs = []string{`stubs/phpstorm-stubs/standard/standard_2.php`}
	test.AddFile(`<?php
function f($s, array $a) {
  $_ = sprintf('%d');
  $_ = sprintf('%2$d', 10); // arg not referenced
  $_ = sprintf('foo%sbar%s', $s);
  $_ = sprintf('%.2%');
  $_ = sprintf('%z', $s);
  $_ = sprintf("%'");

  $ints = [1, 2];
  $_ = sprintf("--%s--", $a);
  $_ = sprintf("%s", $ints);
  $_ = sprintf("%d", $ints);
}
`)
	test.Expect = []string{
		`%d directive refers to the args[1] which is not provided`,
		`%2$d directive refers to the args[2] which is not provided`,
		`%s directive refers to the args[2] which is not provided`,
		`argument is not referenced from the formatting string`,
		`%% directive has modifiers`,
		`unexpected format specifier z`,
		`'-flag expects a char, found end of string`,
		`potential array to string conversion`,
		`potential array to string conversion`,
	}
	test.RunAndMatch()
}

func TestPrintf(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadedStubs = []string{`stubs/phpstorm-stubs/standard/standard_2.php`}
	test.AddFile(`<?php
function f($s, array $a) {
  printf('%d');
  printf('%2$d', 10); // arg not referenced
  printf('foo%sbar%s', $s);
  printf('%.2%');
  printf('%z', $s);
  printf("%'");

  $ints = [1, 2];
  printf("--%s--", $a);
  printf("%s", $ints);
  printf("%d", $ints);
}
`)
	test.Expect = []string{
		`%d directive refers to the args[1] which is not provided`,
		`%2$d directive refers to the args[2] which is not provided`,
		`%s directive refers to the args[2] which is not provided`,
		`argument is not referenced from the formatting string`,
		`%% directive has modifiers`,
		`unexpected format specifier z`,
		`'-flag expects a char, found end of string`,
		`potential array to string conversion`,
		`potential array to string conversion`,
	}
	test.RunAndMatch()
}
