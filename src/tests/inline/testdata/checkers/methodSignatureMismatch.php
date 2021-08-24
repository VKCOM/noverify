<?php

class Foo {
  final public function f() {}
  final private function f1() {}
  final protected function f2() {}
}

class Boo extends Foo {
  public function f() {} // want `Method \Foo::f is declared final and cannot be overridden`
  public function f1() {} // ok
  public function f2() {} // want `Method \Foo::f2 is declared final and cannot be overridden`
}
