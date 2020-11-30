<?php

function in_array_over_array_keys(array $array) {
    $_ = in_array('abc', array_keys($array)); // bad
    $_ = array_keys('abc', $array); // good
    $_ = in_array('abc', array_keys($array), true);  // don't touch, has 3rd arg
    $_ = in_array('abc', array_keys($array), false); // don't touch, has 3rd arg
}
