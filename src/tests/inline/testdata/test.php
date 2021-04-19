<?php

/**
 * @see A
 */
function f() { // want `@see tag refers to unknown symbol A`
    $x = 100;

    $y; // want `expression evaluated but not used` and `Undefined variable: y`
    $x; // want `expression evaluated but not used`

    $_ = new Foo; // want `Type \Foo not found`
}
