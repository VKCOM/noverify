<?php

function returnAssign(): int {
    $a = 100;
    echo $a;

    return $a = 1; // want `don't use assignment in the return statement`
}

function returnAssign2(): int {
    $a = 100;
    echo $a;

    return $a += 1; // want `don't use assignment in the return statement`
}

function returnAssignOk(): int {
    $a = 100;
    echo $a;

    return $a; // ok, no assign
}
