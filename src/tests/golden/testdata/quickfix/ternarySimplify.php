<?php

function ternarySimplify() {
    $z = $x ? $x : 5;
    $z = isset($x) ? $x : $y;
}