package linttest_test

import (
	"log"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
)

func TestFunctionNotOnlyExits2(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function rand() {
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
}

func TestArrayAccessForClass(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class three {}
	class five {}
	function test() {
		$a = 1==2 ? new three : new five;
		return $a['test'];
	}`)
	test.Expect = []string{"Array access to non-array type"}
	test.RunAndMatch()
}

// This test checks that expressions are evaluated in correct order.
// If order is incorrect then there would be an error that we are referencing elements of a class
// that does not implement ArrayAccess.
func TestCorrectTypes(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class three {}
	class five {}
	function test() {
		$a = ['test' => 1];
		$a = ($a['test']) ? new three : new five;
		return $a;
	}`)
}

func TestAllowReturnAfterUnreachable(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function unreachable() {
		exit;
	}

	function test() {
		unreachable();
		return;
	}`)
}

func TestFunctionReferenceParams(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function doSometing(&$result) {
		$result = 5;
	}`)
}

func TestFunctionReferenceParamsInAnonymousFunction(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function doSometing() {
		return function() use($a, &$result) {
			echo $a;
			$result = 1;
		};
	}`)
	test.Expect = []string{"Undefined variable a"}
	test.RunAndMatch()
}

func TestForeachByRefUnused(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
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
}

func TestAllowAssignmentInForLoop(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function test() {
	  for ($day = 0; $day <= 100; $day = $day + 1) {
		echo $day;
	  }
	}
	`)
}

func TestDuplicateArrayKey(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function test() {
	  return [
		  'key1' => 'something',
		  'key2' => 'other_thing',
		  'key1' => 'third_thing', // duplicate
	  ];
	}`)
	test.Expect = []string{"Duplicate array key 'key1'"}
	test.RunAndMatch()
}

func TestMixedArrayKeys(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function test() {
	  return [
		  'something',
		  'key2' => 'other_thing',
		  'key3' => 'third_thing',
	  ];
	}
	`)
	test.Expect = []string{"Mixing implicit and explicit array keys"}
	test.RunAndMatch()
}

func TestStringGlobalVarName(t *testing.T) {
	// Should not panic.
	linttest.SimpleNegativeTest(t, `<?php
	function f() {
		global ${"x"};
		global ${"${x}_{$x}"};
	}`)
}

func TestArrayLiteral(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function traditional_array_literal() {
		return array(1, 2);
	}`)
	test.Expect = []string{"Use of old array syntax"}
	test.RunAndMatch()
}

func TestIssetVarVar4(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($$$$x)) {
			$_ = $$$$x; // Can't track this level of indirection
		}
	}`)
	test.Expect = []string{
		"Unknown variable variable $$$x used",
		"Unknown variable variable $$$$x used",
	}
	test.RunAndMatch()
}

func TestIssetVarVar3(t *testing.T) {
	// Test that irrelevant isset of variable-variable doesn't affect
	// other variables. Also warn for undefined variable in $$x.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($$x)) {
			$_ = $$y;
		}
	}`)
	test.Expect = []string{
		"Undefined variable: x",
		"Unknown variable variable $$y used",
	}
	test.RunAndMatch()
}

func TestIssetVarVar2(t *testing.T) {
	// Test that if $x is defined, it doesn't make $$x defined.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($x)) {
			$_ = $x;  // $x is defined
			$_ = $$x; // But $$x is not
		}
	}`)
	test.Expect = []string{"Unknown variable variable $$x used"}
	test.RunAndMatch()
}

func TestIssetVarVar1(t *testing.T) {
	// Test that defined variable variable don't cause "undefined" warnings.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		$x = 'key';
		if (isset($$x)) {
			$_ = $x + 1;  // If $$x is isset, then $x is set as well
			$_ = $$x + 1;
			$_ = $y;      // Undefined
		}
		// After the block all vars are undefined again.
		$_ = $x;
	}`)
	test.Expect = []string{"Undefined variable: y"}
	test.RunAndMatch()
}

func TestUnused(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
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
	test.Expect = []string{
		"Unused variable g ",
		"Unused variable a ",
		"Unused variable v ",
	}
	test.RunAndMatch()
}

func TestAtVar(t *testing.T) {
	// variables declared using @var should not be overriden
	_ = linttest.GetFileReports(t, `<?php
	function test() {
		/** @var string $a */
		$a = true;
		return $a;
	}`)

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
	test := linttest.NewSuite(t)
	test.AddFile(`<?php function doExit() {
		exit;
	}

	function doSomething() {
		doExit();
		echo "Dead code";
	}`)
	test.Expect = []string{"Unreachable code"}
	test.RunAndMatch()
}

func TestFunctionNotOnlyExits(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php function rand() {
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
}

func TestFunctionJustReturns(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php function justReturn() {
		return 1;
	}

	function doSomething() {
		justReturn();
		echo "Just normal code";
	}`)
}

func TestFunctionThrowsExceptionsAndReturns(t *testing.T) {
	reports := linttest.GetFileReports(t, `<?php
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
		log.Printf("handle exitFlags: %d (%s)", fi.ExitFlags, linter.FlagsToString(fi.ExitFlags))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestSwitchBreak(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
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
}

func TestCorrectArrayTypes(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function test() {
		$a = [ 'a' => 123, 'b' => 3456 ];
		return $a['a'];
	}
	`)
	test.RunLinter()

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
