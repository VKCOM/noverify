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

function bitwise_non_number(string $s1, string $s2, array $a) {
  $_ = $s1 & $s2; // Bad 1
  $_ = $a | $s1;  // Bad 2
  $_ = $s2 ^ [];  // Bad 3

  $obj1 = new BitwiseExample1();
  $obj2 = new BitwiseExample2();
  return $obj1 & $obj2;
}

/**
 * @param int|float $x
 * @param float|int $y
 */
function bitwise_number($x, $y) {
  $_ = $x ^ $y;
}
