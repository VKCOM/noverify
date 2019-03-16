package linter

import (
	"log"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/meta"
)

func hasReport(reports []*Report, substr string) bool {
	for _, r := range reports {
		if strings.Contains(r.String(), substr) {
			return true
		}
	}

	return false
}

func TestArrayLiteral(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function traditional_array_literal() {
		return array(1, 2);
	}`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "Use of old array syntax") {
		t.Errorf("No error about array() syntax")
	}
}

func TestUnused(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function unused_test($arg1, $arg2) {
		global $g;

		$_SERVER['test'] = 1; // superglobal, must not count as unused

		$_ = 'should not count as unused';
		$a = 10;
		foreach ([1, 2, 3] as $k => $v) {
			// $v is unused here
			echo $k;
		}
	}`)

	if len(reports) != 3 {
		t.Errorf("Unexpected number of reports: expected 3, got %d", len(reports))
	}

	if !hasReport(reports, "Unused variable g ") {
		t.Errorf("No error about unused variable g")
	}

	if !hasReport(reports, "Unused variable a ") {
		t.Errorf("No error about unused variable a")
	}

	if !hasReport(reports, "Unused variable v ") {
		t.Errorf("No error about unused variable v")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestAtVar(t *testing.T) {
	// variables declared using @var should not be overriden
	_ = getReportsSimple(t, `<?php
	function test() {
		/** @var string $a */
		$a = true;
		return $a;
	}
	`)

	fi, ok := meta.Info.GetFunction(`\test`)
	if !ok {
		t.Errorf("Could not get function test")
	}

	typ := fi.Typ
	hasBool := false
	hasString := false

	typ.Iterate(func(typ string) {
		if typ == "string" {
			hasString = true
		} else if typ == "bool" {
			hasBool = true
		}
	})

	log.Printf("$a type = %s", typ)

	if !hasBool {
		t.Errorf("Type of variable a does not have boolean type")
	}

	if !hasString {
		t.Errorf("Type of variable a does not have string type")
	}
}

func TestFunctionExit(t *testing.T) {
	reports := getReportsSimple(t, `<?php function doExit() {
		exit;
	}

	function doSomething() {
		doExit();
		echo "Dead code";
	}
	`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "Unreachable code") {
		t.Errorf("No error about unreachable code")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestFunctionNotOnlyExits(t *testing.T) {
	reports := getReportsSimple(t, `<?php function rand() {
		return 4;
	}

	function doExit() {
		if (rand()) {
			exit;
		} else {
			return;
		}
	}

	function doSomething() {
		doExit();
		echo "Not always dead code";
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestFunctionNotOnlyExits2(t *testing.T) {
	reports := getReportsSimple(t, `<?php function rand() {
		return 4;
	}

	class RuntimeException {}

	class Something {
		public static function doExit() {
			if (rand()) {
				throw new \RuntimeException("OMG");
			}

			return rand();
		}
	}

	function doSomething() {
		Something::doExit();
		echo "Not always dead code";
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestFunctionJustReturns(t *testing.T) {
	reports := getReportsSimple(t, `<?php function justReturn() {
		return 1;
	}

	function doSomething() {
		justReturn();
		echo "Just normal code";
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestFunctionThrowsExceptionsAndReturns(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class Exception {}

	function handle($b) {
		if ($b === 1) {
			return $b;
		}

		switch ($b) {
			case "a":
				throw new \Exception("a");

			default:
				throw new \Exception("default");
		}
	}

	function doSomething() {
		handle(1);
		echo "This code is reachable\n";
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	fi, ok := meta.Info.GetFunction(`\handle`)

	if ok {
		log.Printf("handle exitFlags: %d (%s)", fi.ExitFlags, FlagsToString(fi.ExitFlags))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestSwitchBreak(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function doSomething($a) {
		switch ($a) {
		case 1:
			echo "One\n";
			break;
		default:
			echo "Two";
			break;
		}

		echo "Three";
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestArrayAccessForClass(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class three {}
	class five {}
	function test() {
		$a = 1==2 ? new three : new five;
		return $a['test'];
	}`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "Array access to non-array type") {
		t.Errorf("No error about array access to non-array type")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

// This test checks that expressions are evaluated in correct order.
// If order is incorrect then there would be an error that we are referencing elements of a class
// that does not implement ArrayAccess.
func TestCorrectTypes(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class three {}
	class five {}
	function test() {
		$a = ['test' => 1];
		$a = ($a['test']) ? new three : new five;
		return $a;
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestAllowReturnAfterUnreachable(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function unreachable() {
		exit;
	}

	function test() {
		unreachable();
		return;
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestFunctionReferenceParams(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function doSometing(&$result) {
		$result = 5;
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestFunctionReferenceParamsInAnonymousFunction(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function doSometing() {
		return function() use($a, &$result) {
			echo $a;
			$result = 1;
		};
	}`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "Undefined variable a") {
		t.Errorf("No error about undefined variable $a")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestForeachByRefUnused(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class SomeClass {
		public $a;
	}

	/**
	 * @param SomeClass[] $some_arr
	 */
	function doSometing() {
		$some_arr = [];

		foreach ($some_arr as $var) {
			$var->a = 1;
		}

		foreach ($some_arr as &$var2) {
			$var2->a = 2;
		}
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestCorrectArrayTypes(t *testing.T) {
	meta.ResetInfo()

	first := `<?php
	function test() {
		$a = [ 'a' => 123, 'b' => 3456 ];
		return $a['a'];
	}
	`

	testParse(t, `first.php`, first)
	meta.SetIndexingComplete(true)
	testParse(t, `first.php`, first)

	fn, ok := meta.Info.GetFunction(`\test`)
	if !ok {
		t.Errorf("Could not find function test")
		t.Fail()
	}

	if l := fn.Typ.Len(); l != 1 {
		t.Errorf("Unexpected number of types: %d, excepted 1", l)
	}

	if !fn.Typ.IsInt() {
		t.Errorf("Wrong type: %s, excepted int", fn.Typ)
	}
}

func TestAllowAssignmentInForLoop(t *testing.T) {
	reports := getReportsSimple(t, `<?php

	function test() {
	  for ($day = 0; $day <= 100; $day = $day + 1) {
		echo $day;
	  }
	}
	`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestDuplicateArrayKey(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function test() {
	  return [
		  'key1' => 'something',
		  'key2' => 'other_thing',
		  'key1' => 'third_thing', // duplicate
	  ];
	}
	`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "Duplicate array key 'key1'") {
		t.Errorf("No error about duplicate array key 'key1'")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestMixedArrayKeys(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	function test() {
	  return [
		  'something',
		  'key2' => 'other_thing',
		  'key3' => 'third_thing',
	  ];
	}
	`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "Mixing implicit and explicit array keys") {
		t.Errorf("No error about mixed keys")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}
