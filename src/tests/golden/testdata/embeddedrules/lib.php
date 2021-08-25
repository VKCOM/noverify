<?php

$secret = 42;

function random() {
  global $secret;
  return $secret;
}
