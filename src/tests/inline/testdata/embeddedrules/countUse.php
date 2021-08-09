<?php

function countUse() {
  $arr = [1, 2, 3];

  // always true cases

  if (count($arr) >= 0) { // want `this expression is always true`
    echo 1;
  }

  if (0 <= count($arr)) { // want `this expression is always true`
    echo 1;
  }

  if (count($arr) > 0) { // ok
    echo 1;
  }

  if (0 < count($arr)) { // ok
    echo 1;
  }

  if (count($arr) != 0) { // ok
    echo 1;
  }

  if (0 != count($arr)) { // ok
    echo 1;
  }

  if (count($arr) !== 0) { // ok
    echo 1;
  }

  if (0 !== count($arr)) { // ok
    echo 1;
  }

  if (count($arr)) { // ok
    echo 1;
  }

  // always false cases

  if (count($arr) < 0) { // want `this expression is always false`
    echo 1;
  }

  if (0 > count($arr)) { // want `this expression is always false`
    echo 1;
  }

  if (count($arr) <= 0) { // want `use count($arr) == 0 instead`
    echo 1;
  }

  if (0 >= count($arr)) { // want `use 0 == count($arr) instead`
    echo 1;
  }

  if (count($arr) == 0) { // ok
    echo 1;
  }

  if (0 == count($arr)) { // ok
    echo 1;
  }

  if (count($arr) === 0) { // ok
    echo 1;
  }

  if (0 === count($arr)) { // ok
    echo 1;
  }

  if (!count($arr)) { // ok
    echo 1;
  }
}
