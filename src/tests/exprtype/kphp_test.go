package exprtype_test

import (
	"testing"
)

func TestExprTypeAny(t *testing.T) {
	code := `<?php
/** @return any */
function get_any() {
  return 10;
}

/** @return any[][] */
function get_any_arr() {
  return [[1]];
}

exprtype(get_any(), 'mixed');
exprtype(get_any_arr(), 'mixed[][]');
`
	runKPHPExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestInstanceDeserializeType(t *testing.T) {
	code := `<?php
class Foo {
	/** Method */
	public function method() {}
}

const CLASS_NAME = "Foo";

class Boo {
	/** Method */
	public function method() {
		$text = "";
		exprtype(instance_deserialize($text, self::class), "\Boo|null");
		exprtype(instance_deserialize($text, static::class), "\Boo|null");
		exprtype(instance_deserialize($text, $this::class), "\Boo|null");
	}
}

function f() {
	$text = "";
	$className = "";

	exprtype(instance_deserialize($text, Foo::class), "\Foo|null");
	exprtype(instance_deserialize($text, "Foo"), "\Foo|null");
	exprtype(instance_deserialize($text, 10), "mixed");
	exprtype(instance_deserialize($text, $className), "mixed");
	exprtype(instance_deserialize($text, CLASS_NAME), "mixed");
}
`
	runKPHPExprTypeTest(t, &exprTypeTestParams{code: code, stubs: "<?php /* no code */"})
}

func TestInstanceCacheFetch(t *testing.T) {
	code := `<?php
class Foo {
	/** Method */
	public function method() {}
}

const CLASS_NAME = "Foo";

class Boo {
	/** Method */
	public function method() {
		$text = "";
		exprtype(instance_cache_fetch(self::class, $text), "\Boo|null");
		exprtype(instance_cache_fetch(static::class, $text), "\Boo|null");
		exprtype(instance_cache_fetch($this::class, $text), "\Boo|null");
	}
}

function f() {
	$text = "";
	$className = "";

	exprtype(instance_cache_fetch(Foo::class, $text), "\Foo|null");
	exprtype(instance_cache_fetch("Foo", $text), "\Foo|null");
	exprtype(instance_cache_fetch(10, $text), "mixed");
	exprtype(instance_cache_fetch($className, $text), "mixed");
	exprtype(instance_cache_fetch(CLASS_NAME, $text), "mixed");
}
`
	runKPHPExprTypeTest(t, &exprTypeTestParams{code: code, stubs: "<?php /* no code */"})
}

func TestArrayFirstLastType(t *testing.T) {
	code := `<?php
class Foo {}

/**
 * @return Foo[]
 */
function returnFooArray() {
	return [new Foo, new Foo, new Foo];
}

/**
 * @return mixed
 */
function returnMixed() {}

/**
 * @return mixed[]
 */
function returnMixedArray() {}

function f() {
	$a = [10, 20, 30];
	$b = array_last_element($a);
	exprtype($b, "int");
	
	$c = [new Foo, new Foo, new Foo];
	$d = array_last_element($c);
	exprtype($d, "\Foo");

	$e = returnFooArray();
	$f = array_last_element($e);
	exprtype($f, "\Foo");

	$g = returnMixed();
	$h = array_last_element($g);
	exprtype($h, "mixed");

	$i = returnMixedArray();
	$j = array_last_element($i);
	exprtype($j, "mixed");

	$k = array_last_element([10, 20]);
	exprtype($k, "int");

	$l = array_last_element(20);
	exprtype($l, "mixed");

	$m = array_last_element();
	exprtype($m, "mixed");
}

function f1() {
	$a = [10, 20, 30];
	$b = array_first_element($a);
	exprtype($b, "int");
	
	$c = [new Foo, new Foo, new Foo];
	$d = array_first_element($c);
	exprtype($d, "\Foo");

	$e = returnFooArray();
	$f = array_first_element($e);
	exprtype($f, "\Foo");

	$g = returnMixed();
	$h = array_first_element($g);
	exprtype($h, "mixed");

	$i = returnMixedArray();
	$j = array_first_element($i);
	exprtype($j, "mixed");

	$k = array_first_element([10, 20]);
	exprtype($k, "int");

	$l = array_first_element(20);
	exprtype($l, "mixed");

	$m = array_first_element();
	exprtype($m, "mixed");
}
`
	runKPHPExprTypeTest(t, &exprTypeTestParams{code: code, stubs: "<?php /* no code */"})
}

func TestArrayFilterByKey(t *testing.T) {
	code := `<?php
class Foo {}

/**
 * @return Foo[]
 */
function returnFooArray() {
	return [new Foo, new Foo, new Foo];
}

/**
 * @return mixed
 */
function returnMixed() {}

/**
 * @return mixed[]
 */
function returnMixedArray() {}

function f() {
	$a = [10, 20, 30];
	$b = array_filter_by_key($a, function() { return true; });
	exprtype($b, "int[]");
	
	$c = [new Foo, new Foo, new Foo];
	$d = array_filter_by_key($c, function() { return true; });
	exprtype($d, "\Foo[]");

	$e = returnFooArray();
	$f = array_filter_by_key($e, function() { return true; });
	exprtype($f, "\Foo[]");

	$g = returnMixed();
	$h = array_filter_by_key($g, function() { return true; });
	exprtype($h, "mixed");

	$i = returnMixedArray();
	$j = array_filter_by_key($i, function() { return true; });
	exprtype($j, "mixed[]");

	$k = array_filter_by_key([10, 20], function() { return true; });
	exprtype($k, "int[]");

	$l = array_filter_by_key([10, "hello"], function() { return true; });
	exprtype($l, "mixed[]");

	$m = array_filter_by_key([], function() { return true; });
	exprtype($m, "mixed[]");
}
`
	runKPHPExprTypeTest(t, &exprTypeTestParams{code: code, stubs: "<?php /* no code */"})
}

func runKPHPExprTypeTest(t *testing.T, params *exprTypeTestParams) {
	exprTypeTestImpl(t, params, true)
}
