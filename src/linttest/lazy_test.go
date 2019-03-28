package linttest_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

func TestUse(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class omg {
		public $some_property;
	}

	function doSomething($a, omg $b) {
		return function() use($b) {
			echo $b->some_property;
			echo $b->other_property;
		};
	}`)
	test.Expect = []string{"other_property does not exist"}
	test.RunAndMatch()
}

func TestOverride(t *testing.T) {
	meta.ResetInfo()

	linttest.ParseTestFile(t, "meta.php", `<?php
	namespace PHPSTORM_META {
		override(\array_slice(0), type(0));
		override(\array_shift(0), elementType(0));
	}`)

	linttest.ParseTestFile(t, "std.php", `<?php
	/**
	* @param array $array
	* @param int $offset
	* @param int $length [optional]
	* @param bool $preserve_keys [optional]
	* @return array the slice.
	*/
	function array_slice (array $array, $offset, $length = null, $preserve_keys = false) {}

	/**
	* @param array $array
	* @return mixed the shifted value, or &null; if array is
	* empty or is not an array.
	*/
	function array_shift (array &$array) {}
	`)

	linter.InitStubs()

	contents := `<?php
	function do_something() {
		$a = [];
		$a[0] = 1;
		$a[1] = 2;
		return array_slice($a, 0, 1);
	}

	function do_something2() {
		$a = [];
		$a[0] = 1;
		$a[1] = 2;
		return array_shift($a);
	}
	`

	linttest.ParseTestFile(t, "test.php", contents)
	meta.SetIndexingComplete(true)
	linttest.ParseTestFile(t, "test.php", contents)

	fn, ok := meta.Info.GetFunction(`\do_something`)
	if !ok {
		t.Errorf("Could not find function do_something")
		t.Fail()
	}

	typ := solver.ResolveTypes(fn.Typ, make(map[string]struct{}))

	if _, ok := typ[`int[]`]; !ok {
		t.Errorf("Incorrect return types: do_something() expected 'int[]', got '%s' (raw type: '%s')", typ, fn.Typ)
	}

	fn, ok = meta.Info.GetFunction(`\do_something2`)
	if !ok {
		t.Errorf("Could not find function do_something2")
		t.Fail()
	}

	typ = solver.ResolveTypes(fn.Typ, make(map[string]struct{}))

	if _, ok := typ[`int`]; !ok {
		t.Errorf("Incorrect return types: do_something2() expected 'int', got '%s' (raw type: '%s')", typ, fn.Typ)
	}
}

func TestLazy(t *testing.T) {
	meta.ResetInfo()

	first := `<?php namespace NS;
	class Test {
		public static function instance() {
			return self::$instances[0];
		}

		public static function instance2() {
			foreach (self::$instances as $instance) {
				return $instance;
			}
		}

		/** @var Test[] */
		public static $instances;
	}`

	second := `<?php function do_something() {
		return \NS\Test::instance();
	}

	function do_something2() {
		return \NS\Test::instance2();
	}
	`

	linttest.ParseTestFile(t, `first.php`, first)
	linttest.ParseTestFile(t, `second.php`, second)
	meta.SetIndexingComplete(true)
	linttest.ParseTestFile(t, `second.php`, second)

	cls, ok := meta.Info.GetClass(`\NS\Test`)
	if !ok {
		t.Errorf(`Could not find class \NS\Test`)
		t.Fail()
	}

	fn, ok := meta.Info.GetFunction(`\do_something`)
	if !ok {
		t.Errorf("Could not find function do_something")
		t.Fail()
	}

	typ := solver.ResolveTypes(fn.Typ, make(map[string]struct{}))

	if _, ok := typ[`\NS\Test`]; !ok {
		t.Errorf("Incorrect return types: class method typ: '%s' raw: '%s', resolved: %+v", cls.Methods[`instance`].Typ, fn.Typ, typ)
	}

	fn, ok = meta.Info.GetFunction(`\do_something2`)
	if !ok {
		t.Errorf("Could not find function do_something2")
		t.Fail()
	}

	typ = solver.ResolveTypes(fn.Typ, make(map[string]struct{}))

	if _, ok := typ[`\NS\Test`]; !ok {
		t.Errorf("Incorrect return types2: class method typ: '%s' raw: '%s', resolved: %+v", cls.Methods[`instance2`].Typ, fn.Typ, typ)
	}
}
