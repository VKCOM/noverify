<?php

function test_strpos_bad(string $s, string $sub) {
  // Give warning: single-char haystack strings are suspicious.
  $_ = strpos('/', $s);
  $_ = strpos("/", $s);
}

function test_strpos_ok(string $s, string $sub) {
  // This code is kinda OK, it checks for ($s == "x" || $s == "y").
  $_ = strpos("xy", $s);

  // Good code: needle is literal, haystack is non-constant.
  $_ = strpos($s, "x");

  // OK: both arguments are non-constant.
  $_ = strpos($s, $sub);
}

function test_stripos_bad(string $s, string $sub) {
  $_ = stripos('/', $s);
  $_ = stripos("/", $s);
}

function test_stripos_ok(string $s, string $sub) {
  $_ = stripos("xY", $s);
  $_ = stripos($s, "x");
  $_ = stripos($s, $sub);
}

function test_preg_replace_bad($pat, $subj) {
  $_ = preg_replace($pat, $subj, 'replacement');
}

function test_preg_replace_ok($pat, $subj, $repl) {
  $_ = preg_replace($pat, 'replacement', $subj);
  $_ = preg_replace($pat, $repl, $subj);
}

function test_explode_bad($s) {
  $_ = explode($s, '/');
}

function test_explode_ok($s, $delim) {
  $_ = explode('/', $s);
  $_ = explode($delim, $s);
}

function test_str_replace_bad($search, $subj) {
  $_ = str_replace($search, $subj, ' ');
  $_ = str_replace($search, $subj, '');
}

function test_str_replace_ok($search, $subj, $repl) {
  $_ = str_replace($search, 'replacement', $subj);
  $_ = str_replace($search, $repl, $subj);
}