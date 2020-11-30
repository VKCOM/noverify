<?php

$array = ['a' => 1, 'b' => 2, 'c' => 3];

if (in_array('z', array_keys($array))) { // 2
    echo "contains z key\n";
}

function f1($k, $xs) {
    return in_array($k . '_id', array_keys($xs['data'][0])); // 1
}

function f2($k, $xs) {
    return in_array($k . '_id', $xs['data'][0]);
}
