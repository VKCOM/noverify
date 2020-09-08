<?php

$secret = 42;

function random() {
  global $secret;
  return $secret;
}

$_ = new class {
  // This is a bad code, but we're skipping the anon class
  // bodies right now.
  public function f(string $s) {
    return strpos('/', $s);
  }
};