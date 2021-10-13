package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestLiteralAsType(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * @param '!'|'?'|'$' $a
 */
function f($a) {}

/**
 * @param 'abd'|'abc' $a
 */
function f($a) {}

/**
 * @param '!='|'<'|'<='|'<>'|'='|'=='|'>'|'>='|'eq'|'ge'|'gt'|'le'|'lt'|'ne' $a
 */
function f($a) {}

/**
 * @return '!='|'<'|'<='|'<>'|'='|'=='|'>'|'>='|'eq'|'ge'|'gt'|'le'|'lt'|'ne'
 */
function f(): string {}

/**
 * @param 'abd'|int $a
 */
function f($a) {}

/**
 * @param 'abd $a
 */
function f($a) {}

/**
 * @param $a 'abd
 */
function f($a) {}
`)
	test.Expect = []string{
		`Malformed @param tag (maybe var is missing?)`,
		`Non-canonical order of variable and type`,
	}
	test.RunAndMatch()
}
