package exprtype_test

import (
	"testing"
)

func TestClosureCallbackArgumentsTypes(t *testing.T) {
	code := `<?php
function usort($array, $callback) {}
function uasort($array, $callback) {}
function array_map($callback, $array, $_ = null) {}
function array_walk($callback, $array) {}
function array_walk_recursive($callback, $array) {}
function array_filter($array, $callback) {}
function array_reduce($array, $callback) {}
function some_function_without_model($callback, $array) {}

class Foo { public function f() {} }
class Boo { public function b() {} }

/** @return Foo[] */
function return_foo() { return []; }

/** @return Boo[] */
function return_boo() { return []; }

$foo_array = return_foo();
$boo_array = return_boo();

usort($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

uasort($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

array_map(function($a) {
  exprtype($a, "\Foo");
}, $foo_array);

array_walk($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

// wrong number of args
array_walk($foo_array, function($a, $b, $c) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
  exprtype($c, "mixed");
});

array_walk_recursive($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

array_filter($foo_array, function($a) {
  exprtype($a, "\Foo");
});

array_reduce($foo_array, function($carry, $item) {
  exprtype($carry, "\Foo");
  exprtype($item, "\Foo");
});

// mixed array, but function args with type hints
$mixed_arr = [];
usort($mixed_arr, function(Foo $a, Foo $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

$mixed_arr_2 = [];
usort($mixed_arr_2, function(int $a, int $b) {
  exprtype($a, "int");
  exprtype($b, "int");
});

// mixed array, but not all function args have type hints
$mixed_arr_3 = [];
usort($mixed_arr_3, function(Foo $a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

// non mixed array, but function args with type hints
$non_mixed_arr = [1, 2, 3];
usort($non_mixed_arr, function(int $a, int $b) {
  exprtype($a, "int");
  exprtype($b, "int");
});

// non mixed array, but not all function args have type hints
$non_mixed_arr_2 = [new Foo, new Foo, new Foo];
usort($non_mixed_arr_2, function(Foo $a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

some_function_without_model(function($b) {
  exprtype($b, "mixed");
}, $d);

// Not supported
function callback($a) {
  exprtype($a, "mixed");
}

// Not supported
$callback = function($a) {
  exprtype($a, "mixed");
};

array_map($callback, $d);
array_map('callback', $d);
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestClosureCallbackArgumentsPossibleErrorVariations(t *testing.T) {
	code := `<?php
function usort($array, $callback) {}
function uasort($array, $callback) {}
function array_map($callback, $array, $_ = null) {}
function array_walk($callback, $array) {}
function array_walk_recursive($callback, $array) {}
function array_filter($array, $callback) {}
function array_reduce($array, $callback) {}

class Foo { public function f() {} }

/** @return Foo[] */
function return_foo() { return []; }

$foo_array = return_foo();

// more arguments than necessary
usort($foo_array, function($a, $b, $c) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
  exprtype($c, "mixed");
});

// less arguments than necessary
uasort($foo_array, function($a) {
  exprtype($a, "\Foo");
});

// more arguments than necessary
array_map(function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
}, $foo_array);

// less arguments than necessary
array_map(function() {

}, $foo_array);

// more arguments than necessary
array_walk($foo_array, function($a, $b, $c) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
  exprtype($c, "mixed");
});

// less arguments than necessary
array_walk($foo_array, function() {
});

// more arguments than necessary
array_filter($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

// less arguments than necessary
array_filter($foo_array, function() {
});

// more arguments than necessary
array_reduce($foo_array, function($carry, $item, $c) {
  exprtype($carry, "\Foo");
  exprtype($item, "\Foo");
  exprtype($c, "mixed");
});

// less arguments than necessary
array_reduce($foo_array, function() {
});
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestNestedClosureCallback(t *testing.T) {
	code := `<?php
function array_map($callback, array $arr1, array $_ = null) {}
function usort($array, $callback) {}

class Foo { public function f() {} }
/** @return Foo[][] */

function return_foo() { return []; }

$foo_array = return_foo();

array_map(function($a) {
  exprtype($a, "\Foo[]");
  array_map(function($a) {
    exprtype($a, "\Foo");
  }, $a);
  exprtype($a, "\Foo[]");
}, $foo_array);

usort($foo_array, function($a, $b) {
  exprtype($a, "\Foo[]");
  exprtype($b, "\Foo[]");

  array_map(function($a) {
    exprtype($a, "\Foo");
  }, $a);

  array_map(function($b) {
    exprtype($b, "\Foo");
  }, $b);

  exprtype($a, "\Foo[]");
  exprtype($b, "\Foo[]");
});
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestClosureExprType(t *testing.T) {
	code := `<?php
class Foo {
  public function method() {}
}

function func() {
  $f = function(): Foo { return new Foo; };
  $foo = $f();
  
  exprtype($f, "\Closure$(exprtype.php,func):7$");
  exprtype($foo, "\Foo");
}

function func1() {
  $f1 = function(): float {
    if (1) {
	  return new Foo;
	}
    return 10; 
  };

  $foo1 = $f1();

  exprtype($f1, "\Closure$(exprtype.php,func1):15$");
  exprtype($foo1, "float");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestCallableDoc(t *testing.T) {
	code := `<?php
class Foo {
  /**
   * @return int
   */
  public function method(): int { return 0; }
}

class Boo {
  /**
   * @return int
   */
  public function method(): int { return 0; }
}

/**
 * @param callable(): Foo $s
 */
function f1(callable $s) {
  $a = $s();
  exprtype($a, "\Foo");
}

/**
 * @param callable(Foo): Foo $s
 */
function f2(callable $s) {
  $a = $s(new Foo);
  exprtype($a, "\Foo");
}

/**
 * @param callable(int): Foo $s
 */
function f3(callable $s) {
  $a = $s(10);
  exprtype($a, "\Foo");
}

/**
 * @param callable(int, string): Foo|Boo $s
 */
function f4(callable $s) {
  $a = $s(10, "ss");
  exprtype($a, "\Boo|\Foo");
}

/**
 * @param callable(): callable(): Foo $s
 * @param callable(): callable(): Foo|Boo $s1
 */
function f5(callable $s, callable $s1) {
  $a = $s();
  $a1 = $s1();
  exprtype($a, "\Closure$():Foo");
  exprtype($a1, "\Closure$():Foo/Boo");
  $b = $a();
  $b1 = $a1();
  exprtype($b, "\Foo");
  exprtype($b1, "\Boo|\Foo");
}

/**
 * @return callable(): callable(): Foo
 */
function f6(): callable {
  return function() { return function() { return new Foo; }; };
}

function f7() {
  $a = f6();
  exprtype($a, "\Closure$():callable(): Foo|callable");
  $b = $a();
  exprtype($b, "\Closure$():Foo");
  $c = $b();
  exprtype($c, "\Foo");
}

function f8() {
  /**
   * @var callable(): Foo $a
   */
  $a = null;

  $b = $a();
  exprtype($b, "\Foo");
}

/**
* @var callable(): Foo $a
*/
$a = null;
$b = $a();
exprtype($b, "\Foo");

/**
 * @param callable(int, string) $s
 */
function f9(callable $s) {
  $a = $s(10, "ss");
  exprtype($a, "mixed");
}

/**
 * @param callable(int, string): Foo $s
 */
function f10(callable $s) {
  if ($s() instanceof Boo) {
    exprtype($s(), "\Boo");
  }
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestClosurePHPDoc(t *testing.T) {
	code := `<?php
class Foo {
  /**
   * @return int
   */
  public function method(): int { return 0; }
}

function f() {
  /**
   * @param int $a
   * @return callable(int): Foo
   */
  $b = function(int $a): callable { return function (){}; };
  exprtype($b, "\Closure$(exprtype.php,f):14$");
  $c = $b(10);
  exprtype($c, "\Closure$(int):Foo|callable");
  $d = $c();
  exprtype($d, "\Foo");
}

function f1() {
  $b = 
     /**
      * @param int $a
      * @return callable(int): Foo
      */
     function(int $a): callable { return function (){}; };
  exprtype($b, "\Closure$(exprtype.php,f1):28$");
  $c = $b(10);
  exprtype($c, "\Closure$(int):Foo|callable");
  $d = $c();
  exprtype($d, "\Foo");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}
