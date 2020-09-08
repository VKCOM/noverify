<?php

function precedence_lhs($x, $mask) {
  $_ = 0 == $mask & $x;
  $_ = 0 != $mask & $x;
  $_ = 0 === $mask & $x;
  $_ = 0 !== $mask & $x;

  $_ = 0 == $mask | $x;
  $_ = 0 != $mask | $x;
  $_ = 0 === $mask | $x;
  $_ = 0 !== $mask | $x;
}

function precedence_rhs($x, $mask) {
  $_ = $x & $mask == 0;
  $_ = $x & $mask != 0;
  $_ = $x & $mask === 0;
  $_ = $x & $mask !== 0;

  $_ = $x | $mask == 0;
  $_ = $x | $mask != 0;
  $_ = $x | $mask === 0;
  $_ = $x | $mask !== 0;
}

function precedence_foo() { return 10; }

function precedence_rhs_good($x, $mask) {
  $_ = ($x & $mask) == 0;
  $_ = ($x & $mask) != 0;
  $_ = ($x & $mask) === 0;
  $_ = ($x & $mask) !== 0;

  $_ = ($x | $mask) == 0;
  $_ = ($x | $mask) != 0;
  $_ = ($x | $mask) === 0;
  $_ = ($x | $mask) !== 0;

  $_ = 0x02 | (($x & $mask) != 0);
  $_ = 0x02 & (precedence_foo() !== 0);
}

function precedence_lhs_good($x, $mask) {
  $_ = 0 == ($mask & $x);
  $_ = 0 != ($mask & $x);
  $_ = 0 === ($mask & $x);
  $_ = 0 !== ($mask & $x);

  $_ = 0 == ($mask | $x);
  $_ = 0 != ($mask | $x);
  $_ = 0 === ($mask | $x);
  $_ = 0 !== ($mask | $x);

  $_ = (($x & $mask) != 0) | 0x02;
  $_ = (precedence_foo() !== 0) & 0x02;
}
