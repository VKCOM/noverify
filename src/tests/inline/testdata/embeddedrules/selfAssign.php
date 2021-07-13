<?php

class Foo {
  private int $prop = 100;

  public function f(int $prop) {
    $prop = $prop; // want `Assignment to $prop itself does not make sense`
    $this->prop = $prop;
    $propNew = $prop;
    echo $propNew;
  }
}

function selfAssign() {
  $c = 100;
  $d = 100;
  $c = $c; // want `Assignment to $c itself does not make sense`
  $c = $d; // ok
}
