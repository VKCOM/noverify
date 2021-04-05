<?php

$a = 5;
function f(): int {
    global $a;
    $a++;

    return $a;
}

function ternarySimplify() {
    $x = 0;
    $y = 0;

    $_ = $x ? $x : 5;
    // ok, f not pure
    $_ = f() ? f() : 5;

    $_ = isset($x) ? $x : $y;
}
