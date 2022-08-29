<?php

trait A {}
trait B {}

// class Foo {
//   private A $a;
//   public static A $a1;
//
//   public function f(A $a): A {}
//   public function f1(A $a, B $b): A {}
// }

// function f(A $a): A {}
function f1() {
  $_ = function(A $a): B {};
}


// trait Test {
//   private static ?self $instance = null;     // ok, in trait
//   public static function instance(): self {} // ok, in trait
// }
