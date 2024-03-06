<?php
declare(strict_types=1);

class Boo {
  /***/
  public function f(): int {
    return 1;
  }
}

class Foo {
  public $prop4;

  public function __construct(public Boo $prop1, private string $prop2, int $prop3) {
    echo $prop1->f();
    echo $prop2;
  }

  /***/
  public function method() {
    echo $this->prop1->f();
    echo $this->prop2;
    echo $this->prop4;
  }
}

class Goo extends Foo {}

class Doo extends Goo {
  public string $prop2;
  public Boo $prop3;
}

function f() {
  $foo = new Foo(new Boo, "", 10);
  var_dump($foo->prop1);
  var_dump($foo->prop2); // want `Cannot access private property \Foo->prop2`
  var_dump($foo->prop3); // want `Property {\Foo}->prop3 does not exist`
  var_dump($foo->prop4);

  $goo = new Goo(new Boo, "", 10);
  var_dump($goo->prop1);
  var_dump($goo->prop2); // want `Cannot access private property \Foo->prop2`
  var_dump($goo->prop3); // want `Property {\Goo}->prop3 does not exist`
  var_dump($goo->prop4);

  $doo = new Doo(new Boo, "", 10);
  var_dump($doo->prop1);
  var_dump($doo->prop2);
  var_dump($doo->prop3);
  var_dump($doo->prop4);
}
