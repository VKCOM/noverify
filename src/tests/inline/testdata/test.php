<?php

function f() {
    $x = 100;

    $y; // want `expression evaluated but not used` and `Undefined variable: y`
    $x; // want `expression evaluated but not used`

    $_ = new Foo; // want `Type \Foo not found`
    $_ = new Foo;
}
