<?php

/**
 * @param mixed[] $array
 */
function in_array_over_array_keys(array $array) {
    $_ = in_array('abc', array_keys($array)); // bad
    $_ = array_keys('abc', $array); // good
    $_ = in_array('abc', array_keys($array), true);  // don't touch, has 3rd arg
    $_ = in_array('abc', array_keys($array), false); // don't touch, has 3rd arg
}

function some_substr(string $str, int $index) {
    $_ = substr($str, $index, 1);
    $_ = substr("hello", $index, 1);
    $_ = substr("hello", 2, 1);

    // ok, length != 1
    $_ = substr($str, $index, 10);
}

function some_array_push(array $array, int $val) {
    array_push($array, $val);
    array_push($array, 10);

    // ok, two element
    array_push($array, 100, 1200);
}
