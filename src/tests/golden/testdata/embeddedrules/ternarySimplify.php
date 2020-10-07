<?php

function sink($_) {}

function ternarySimplify(int $x, $y) {
    sink($x ? true : false);
    sink((bool)$x);

    sink($x > $y ? true : false);
    sink($x > $y);

    sink($x ? $x : $y);
    sink($x ?: $y);

    sink(isset($x[1]) ? $x[1] : $y);
    sink($x[1] ?? $y);

    sink(random() ? random() : $y);
}

define('SOME_MASK', 0x0f);

/**
 * @param int $flags
 */
function ternarySimplify_issue540($flags) {
    sink(($flags & SOME_MASK) ? true : false);
}
