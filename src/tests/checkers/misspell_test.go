package checkers_test

import (
	"testing"

	"github.com/client9/misspell"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
)

//nolint:misspell // misspelled on purpose
func TestMisspellPhpdocPositive(t *testing.T) {
	config := linter.NewConfig()
	config.TypoFixer = misspell.New()
	test := linttest.NewSuite(t)
	test.Linter = linter.NewLinter(config)
	test.AddFile(`<?php
/**
 * This function is a pure perfektion.
 */
function f1() {}

/**
 * This class is our performace secret.
 */
class c1 {
  /**
   * This constant comment is very informitive.
   */
  const Foo = 0;

  /**
   * This method is never called, this is why it's inexpencive.
   */
  private static function secret() {}

  /**
   * This property is not inefficeint.
   */
  private $prop = 1;
}
`)
	test.Expect = []string{
		`"perfektion" is a misspelling of "perfection"`,
		`"performace" is a misspelling of "performance"`,
		`"informitive" is a misspelling of "informative"`,
		`"inexpencive" is a misspelling of "inexpensive"`,
		`"inefficeint" is a misspelling of "inefficient"`,
	}
	test.RunAndMatch()
}

//nolint:misspell // misspelled on purpose
func TestMisspellPhpdocNegative(t *testing.T) {
	config := linter.NewConfig()
	config.TypoFixer = misspell.New()
	test := linttest.NewSuite(t)
	test.Linter = linter.NewLinter(config)
	test.AddFile(`<?php
interface Responsable {}

/**
 * Uses Responsable interface value.
 * @param \Responsable $r
 */
function reference_iface(Responsable $r) {
}

/**
 * Don't warn on emails.
 * This function is a pure perfektion@gmail.com.
 */
function f1() {}

/**
 * Don't warn on hosts.
 * This class is our performace.io.org secret.
 */
class c1 {
  /**
   * Don't warn on paths.
   * This method is never called, this is why it's /a/b/inexpencive.
   */
  private static function secret() {}
}
`)
	test.RunAndMatch()
}

//nolint:misspell // misspelled on purpose
func TestMisspellNamePositive(t *testing.T) {
	config := linter.NewConfig()
	config.TypoFixer = misspell.New()
	test := linttest.NewSuite(t)
	test.Linter = linter.NewLinter(config)
	test.AddFile(`<?php
function unconditionnally_rollback() {}

function f($notificaton) {
}

class c {
  private function m($flag_normallized) {}
}

class Mocrotransactions {
}

function f_overpoweing() {
}

class d {
  private function set_persistance() {}
}
`)
	test.Expect = []string{
		`"unconditionnally" is a misspelling of "unconditionally"`,
		`"notificaton" is a misspelling of "notification"`,
		`"normallized" is a misspelling of "normalized"`,
		`"Mocrotransactions" is a misspelling of "Microtransactions"`,
		`"overpoweing" is a misspelling of "overpowering"`,
		`"persistance" is a misspelling of "persistence"`,
	}
	test.RunAndMatch()
}

//nolint:misspell // misspelled on purpose
func TestMisspellNameNegative(t *testing.T) {
	config := linter.NewConfig()
	config.TypoFixer = misspell.New()
	test := linttest.NewSuite(t)
	test.Linter = linter.NewLinter(config)
	test.AddFile(`<?php
function includ() {
}

class impelments {}

class PostRedirect {}
`)
	test.RunAndMatch()
}
