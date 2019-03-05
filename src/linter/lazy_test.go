package linter

import (
	"log"
	"sync"
	"testing"

	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/z7zmey/php-parser/node"
)

var once sync.Once

func testParse(t *testing.T, filename string, contents string) (rootNode node.Node, w *RootWalker) {
	once.Do(func() {
		MaxFileSize = 10000
		go MemoryLimiterThread()
	})

	var err error
	rootNode, w, err = ParseContents(filename, []byte(contents), "UTF-8", nil)
	if err != nil {
		t.Errorf("Could not parse %s: %s", filename, err.Error())
		t.Fail()
	}

	if !meta.IsIndexingComplete() {
		updateMetaInfo(filename, &w.meta)
	}

	return rootNode, w
}

func getReportsSimple(t *testing.T, contents string) []*Report {
	meta.ResetInfo()
	testParse(t, `first.php`, contents)
	meta.SetIndexingComplete(true)
	_, w := testParse(t, `first.php`, contents)
	return w.GetReports()
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

	testParse(t, `first.php`, first)
	testParse(t, `second.php`, second)

	meta.SetIndexingComplete(true)

	testParse(t, `second.php`, second)

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

func TestOverride(t *testing.T) {
	meta.ResetInfo()

	testParse(t, "meta.php", `<?php
	namespace PHPSTORM_META {
		override(\array_slice(0), type(0));
		override(\array_shift(0), elementType(0));
	}`)

	testParse(t, "std.php", `<?php
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

	lintdebug.Register(func(msg string) { log.Printf("%s", msg) })
	InitStubs()

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

	testParse(t, "test.php", contents)
	meta.SetIndexingComplete(true)
	testParse(t, "test.php", contents)

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

func TestUse(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class omg {
		public $some_property;
	}

	function doSomething($a, omg $b) {
		return function() use($b) {
			echo $b->some_property;
			echo $b->other_property;
		};
	}`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "other_property does not exist") {
		t.Errorf("No error about undefined property other_property")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}
