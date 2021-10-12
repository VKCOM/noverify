<?php

function simpleRuleTest(): int {
  $a = 100;
  $a++; // want `Some error message`
  return $a;
}
