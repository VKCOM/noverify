package linttest

import (
	"testing"
)

func TestPHPDocSyntax(t *testing.T) {
	test := NewSuite(t)
	test.AddFile(`<?php
	/**
	 * @param $x int the x param
	 * @param - $y the y param
	 * @param $z - the z param
	 * @param $a
	 * @param int
	 */
	function f($x, $y, $z, $a, $_) {
		$_ = $x;
		$_ = $y;
		$_ = $z;
	}`)
	test.Expect = []string{
		`non-canonical order of variable and type on line 2`,
		`expected a type, found '-'; if you want to express 'any' type, use 'mixed' on line 3`,
		`non-canonical order of variable and type on line 4`,
		`expected a type, found '-'; if you want to express 'any' type, use 'mixed' on line 4`,
		`malformed @param tag (maybe type is missing?) on line 5`,
		`malformed @param tag (maybe var is missing?) on line 6`,
	}
	test.RunAndMatch()
}

func TestPHPDocType(t *testing.T) {
	test := NewSuite(t)
	test.AddFile(`<?php
	/**
	 * @param [][]string $x
	 * @param double $y
	 * @return []int
	 */
	function f($x, $y) {
		$_ = $x;
		$_ = $y;
		return [1];
	}`)
	test.Expect = []string{
		`[]int type syntax: use [] after the type, e.g. T[]`,
		`[][]string type syntax: use [] after the type, e.g. T[]`,
		`use float type instead of double`,
	}
	test.RunAndMatch()
}
