<?php

function concatenationPrecedence_74($id) {
    echo "id: " . $id - 10;  // want `Unparenthesized expression containing both '.' and binary operator`
    echo "id: " . $id + 10;  // want `Unparenthesized expression containing both '.' and binary operator`
    echo "id: " . $id << 10; // want `Unparenthesized expression containing both '.' and binary operator`
    echo "id: " . $id >> 10; // want `Unparenthesized expression containing both '.' and binary operator`

    echo "id: " . ($id - 10);  // ok
    echo "id: " . ($id + 10);  // ok
    echo "id: " . ($id << 10); // ok
    echo "id: " . ($id >> 10); // ok

    echo $id - 10 . ": id";  // ok
    echo $id + 10 . ": id";  // ok
    echo $id << 10 . ": id"; // ok
    echo $id >> 10 . ": id"; // ok

    echo ($id - 10) . ": id";  // ok
    echo ($id + 10) . ": id";  // ok
    echo ($id << 10) . ": id"; // ok
    echo ($id >> 10) . ": id"; // ok
}
