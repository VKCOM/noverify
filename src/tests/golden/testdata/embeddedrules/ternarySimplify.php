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

    $x_arr = [];

    sink($x_arr[10] !== null ? $x_arr[10] : $y);
    sink(null !== $x_arr[10] ? $x_arr[10] : $y);
    sink($x_arr[10] === null ? $y : $x_arr[10]);
    sink(null === $x_arr[10] ? $y : $x_arr[10]);

    sink(array_key_exists(10, $x_arr) ? $x_arr[10] : null);
    sink(! array_key_exists(10, $x_arr) ? null : $x_arr[10]);

    // ok, index not pure
    sink($x_arr[rand()] !== null ? $x_arr[rand()] : $y);
    sink(null !== $x_arr[rand()] ? $x_arr[rand()] : $y);
    sink($x_arr[rand()] === null ? $y : $x_arr[rand()]);
    sink(null === $x_arr[rand()] ? $y : $x_arr[rand()]);

    sink(array_key_exists(rand(), $x_arr) ? $x_arr[rand()] : null);
    sink(! array_key_exists(rand(), $x_arr) ? null : $x_arr[rand()]);
}

define('SOME_MASK', 0x0f);

/**
 * @param int $flags
 */
function ternarySimplify_issue540($flags) {
    sink(($flags & SOME_MASK) ? true : false);
}
