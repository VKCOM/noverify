<?php

function assignPlus($a1, $b) {
  global $a;
  // $a += $b
  $a = $a + $b; // Could rewrite
  $a = $a1 + $b; // Ok
}

function assignMinus($a1, $b) {
  global $a;
  // $a -= $b
  $a = $a - $b; // Could rewrite
  $a = $a1 - $b; // Ok
}

function assignMul($a1, $b) {
  global $a;
  // $a *= $b
  $a = $a * $b; // Could rewrite
  $a = $a1 * $b; // Ok
}

function assignDiv($a1, $b) {
  global $a;
  // $a /= $b
  $a = $a / $b; // Could rewrite
  $a = $a1 / $b; // Ok
}

function assignMod($a1, $b) {
  global $a;
  // $a %= $b
  $a = $a % $b; // Could rewrite
  $a = $a1 % $b; // Ok
}

function assignConcat($a1, $b) {
  global $a;
  // $a .= $b
  $a = $a . $b; // Could rewrite
  $a = $a1 . $b; // Ok
}

function assignBitAnd($a1, $b) {
  global $a;
  // $a &= $b
  $a = $a & $b; // Could rewrite
  $a = $a1 & $b; // Ok
}

function assignBitOr($a1, $b) {
  global $a;
  // $a |= $b
  $a = $a | $b; // Could rewrite
  $a = $a1 | $b; // Ok
}

function assignXor($a1, $b) {
  global $a;
  // $a ^= $b
  $a = $a ^ $b; // Could rewrite
  $a = $a1 ^ $b; // Ok
}

function assignShiftLeft($a1, $b) {
  global $a;
  // $a <<= $b
  $a = $a << $b; // Could rewrite
  $a = $a1 << $b; // Ok
}

function assignShiftRight($a1, $b) {
  global $a;
  // $a >>= $b
  $a = $a >> $b; // Could rewrite
  $a = $a1 >> $b; // Ok
}

function assignNullCoalesce($a1, $b) {
  global $a;
  // $a ??= $b
  $a = $a ?? $b; // Could rewrite
  $a = $a1 ?? $b; // Ok
}
