<?php

/**
 * @method
 * @method int
 * @method int    method
 * @method ?string  method1
 * @method  int  method2()          <- ok
 * @method static   int   method3() <- ok
 *
 * @property
 * @property   int
 * @property int $a  <- ok
 * @property       $b  int
 * @property int c
 */
class Foo {
  /*
   * @var   int
   */
  public $a = 100;
}

/*
 * @param int $a
 */
function g($a) {
}

/**
 * @param int       $a <- ok
 *
 * @param    int    $unexisting
 * @param int
 * @param $b       int
 * @param $c
 * @param    ?int|null $d
 *
 * @param - $e the y param
 * @param    $f - the z param
 *
 * @param integer   $g
 * @param  []int   $h
 * @param   int? $a1
 *
 * @see   Foo <- ok
 * @see  FooUnExisting
 * @see       FooUnExisting
 */
function f($a, $b, $c, $d, $e, $f, $g, $h, $a1, $b1, $c1, $d1, $e1, $f1, $g1, $h1) {
  /** @var shape(foo: int) $a <- ok */
  $a2 = [];
  echo $a2;

  /** @var    shape(x[]:a) */
  $a2 = [];
  echo $a2;

  /** @var shape(x)    $a */
  $a2 = [];
  echo $a2;

  /*
   * @var int $a2
   */
  $a2 = [];
  echo $a2;
}
