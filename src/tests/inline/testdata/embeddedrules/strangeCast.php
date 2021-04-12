<?php

function strangeCast() {
    $x = 100;

    $_ = $x.""; // want `concatenation with empty string, possible type cast, use explicit cast to string instead of concatenate with empty string`
    $_ = $x.''; // want `concatenation with empty string, possible type cast, use explicit cast to string instead of concatenate with empty string`

    $y = "10";

    $_ = $y + 0; // want `addition with zero, possible type cast, use an explicit cast to int or float instead of zero addition`
    $_ = $y + 10; // ok
    $_ = $y + 0.0; // want `addition with zero, possible type cast, use an explicit cast to int or float instead of zero addition`

    $string = "10";

    $_ = +$string; // want `unary plus, possible type cast, use an explicit cast to int or float instead of using the unary plus`
    $_ = -$string; // ok
}
