<?php

class BitwiseExample1 {}
class BitwiseExample2 {}

function andor_bool_bad() {
  $x = 10;
  $_ = (($x > 0) & ($x != 15)); // Bad 1
  $_ = (($x == 1) | ($x == 2)); // Bad 2
}

function andor_bool_good() {
  $x = 10;
  $_ = (($x > 0) && ($x != 15));
  $_ = (($x == 1) || ($x == 2));

  $_ = $x & 10;
  $_ = 10 & $x;
  $_ = $x | 10;
  $_ = 10 | $x;
}

/**
 * @param int|float $x
 * @param float|int $y
 */
function bitwise_number($x, $y) {
  $_ = $x ^ $y;
}
