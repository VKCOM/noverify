<?php
$a = 100;
$a1 = 100;
$b = 5;


// $a += $b
$a = $a + $b; // Could rewrite
$a = $a1 + $b; // Ok

// $a -= $b
$a = $a - $b; // Could rewrite
$a = $a1 - $b; // Ok

// $a *= $b
$a = $a * $b; // Could rewrite
$a = $a1 * $b; // Ok

// $a /= $b
$a = $a / $b; // Could rewrite
$a = $a1 / $b; // Ok

// $a %= $b
$a = $a % $b; // Could rewrite
$a = $a1 % $b; // Ok

// $a .= $b
$a = $a . $b; // Could rewrite
$a = $a1 . $b; // Ok

// $a &= $b
$a = $a & $b; // Could rewrite
$a = $a1 & $b; // Ok

// $a |= $b
$a = $a | $b; // Could rewrite
$a = $a1 | $b; // Ok

// $a ^= $b
$a = $a ^ $b; // Could rewrite
$a = $a1 ^ $b; // Ok

// $a <<= $b
$a = $a << $b; // Could rewrite
$a = $a1 << $b; // Ok

// $a >>= $b
$a = $a >> $b; // Could rewrite
$a = $a1 >> $b; // Ok

// $a ??= $b
$a = $a ?? $b; // Could rewrite
$a = $a1 ?? $b; // Ok