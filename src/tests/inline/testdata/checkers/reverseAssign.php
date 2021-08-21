<?php

function reverseAssign($b) {
    $a = 100;

    $a += $b; // ok
    $a =+ $b; // want `Possible there should be '+='`
    $a =+$b;  // want `Possible there should be '+='`
    $a=+$b;   // ok
    $a= +$b;  // ok
    $a = +$b; // ok

    $a -= $b; // ok
    $a =- $b; // want `Possible there should be '-='`
    $a =-$b;  // want `Possible there should be '-='`
    $a=-$b;   // ok
    $a= -$b;  // ok
    $a = -$b; // ok

    echo $a;
}
