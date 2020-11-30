<?php

function foo(array $a) { return 0; }

$_ = array(isset($x) ? $x : $y); // expressions in global scope
$_ = foo(array($x ? $x : 5));
$_ = [
    array(foo(array($x ? $x : 5))),
    array(foo([$x ? $x : 5])),
    [foo(array($x ? $x : 5))],
    [[$x ? $x : 5]],
];

function f() {
    $_ = [
        array(foo(array($x ? $x : 5))),
        array(foo([$x ? $x : 5])),
        [foo(array($x ? $x : 5))],
        [[$x ? $x : 5]],
    ];
    $_ = array(isset($x) ? $x : $y); // and in block
}
