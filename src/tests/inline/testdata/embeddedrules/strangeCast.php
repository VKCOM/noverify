<?php

const A = 10;

function strangeCast() {
    $x = 100;

    $_ = $x . ""; // want `Concatenation with empty string, possible type cast, use explicit cast to string instead of concatenate with empty string`
    $_ = "" . $x; // want `Concatenation with empty string, possible type cast, use explicit cast to string instead of concatenate with empty string`
    $_ = $x . ''; // want `Concatenation with empty string, possible type cast, use explicit cast to string instead of concatenate with empty string`
    $_ = '' . $x; // want `Concatenation with empty string, possible type cast, use explicit cast to string instead of concatenate with empty string`

    $y = "10";

    $_ = 0 + $y;   // want `Addition with zero, possible type cast, use an explicit cast to int or float instead of zero addition`
    $_ = $y + 10;  // ok
    $_ = 0.0 + $y; // want `Addition with zero, possible type cast, use an explicit cast to int or float instead of zero addition`

    $string = "10";

    $_ = +$string; // want `Unary plus with non-constant expression, possible type cast, use an explicit cast to int or float instead of using the unary plus`
    $_ = +100;     // ok, constant expression
    $_ = +A;       // ok, constant expression
    $_ = -$string; // ok, unary minus
}
