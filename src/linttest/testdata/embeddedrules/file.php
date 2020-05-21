<?php

function sink($_) {}

function ternarySimplify($x, $y) {
    sink($x ? true : false);
    sink((bool)$x);

    sink($x > $y ? true : false);
    sink($x > $y);

    sink($x ? $x : $y);
    sink($x ?: $y);

    sink(isset($x[1]) ? $x[1] : $y);
    sink($x[1] ?? $y);
}
