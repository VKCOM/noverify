<?php

$a = 5;
function f() {
    $a++;
    return $a;
}

function ternarySimplify() {
    $z = $x ? $x : 5;
    // f not pure
    $z = f() ? f() : 5;

    $z = isset($x) ? $x : $y;
}