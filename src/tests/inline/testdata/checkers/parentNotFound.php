<?php

class Foo {
  public function f() {
    parent::b(); // want `Cannot call method on parent as this class does not extend another`
    self::f(); // ok
  }
}

interface IFoo {}

class Boo {
  public function b() {}
}

class Foo1 extends Boo {
  public function f() {
    parent::b(); // ok
    self::b(); // ok
  }
}

class Foo2 extends Boo implements IFoo {
  public function f() {
    parent::b(); // ok
    self::b(); // ok
  }
}

class Foo3 implements IFoo {
  public function f() {
    parent::b(); // want `Cannot call method on parent as this class does not extend another`
    self::f(); // ok
  }
}
