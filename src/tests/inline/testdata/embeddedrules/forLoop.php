<?php

function forLoop(array $a) {
  for ($i = 0; $i < count($a); $i--) { // want `Potentially infinite 'for' loop, because $i decreases and is always less than initial value 0 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 0; $i < count($a); $i -= 1) { // want `Potentially infinite 'for' loop, because $i decreases and is always less than initial value 0 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 0; $i < count($a); $i -= 100) { // want `Potentially infinite 'for' loop, because $i decreases and is always less than initial value 0 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 0; $i <= count($a); $i--) { // want `Potentially infinite 'for' loop, because $i decreases and is always less than initial value 0 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 0; $i <= count($a); $i -= 1) { // want `Potentially infinite 'for' loop, because $i decreases and is always less than initial value 0 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 0; $i <= count($a); $i -= 100) { // want `Potentially infinite 'for' loop, because $i decreases and is always less than initial value 0 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 100; $i > count($a); $i++) { // want `Potentially infinite 'for' loop, because $i increases and is always greater than initial value 100 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 100; $i > count($a); $i += 1) { // want `Potentially infinite 'for' loop, because $i increases and is always greater than initial value 100 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 100; $i > count($a); $i += 100) { // want `Potentially infinite 'for' loop, because $i increases and is always greater than initial value 100 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 100; $i >= count($a); $i++) { // want `Potentially infinite 'for' loop, because $i increases and is always greater than initial value 100 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 100; $i >= count($a); $i += 1) { // want `Potentially infinite 'for' loop, because $i increases and is always greater than initial value 100 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 100; $i >= count($a); $i += 100) { // want `Potentially infinite 'for' loop, because $i increases and is always greater than initial value 100 and, accordingly, count($a)`
    echo $i;
  }

  for ($i = 0; $i < count($a); $i++) { // ok
    echo $i;
  }

  for ($i = 0; $i < count($a); $i += 1) { // ok
    echo $i;
  }

  for ($i = 0; $i < count($a); $i += 100) { // ok
    echo $i;
  }

  for ($i = 0; $i <= count($a); $i++) { // ok
    echo $i;
  }

  for ($i = 0; $i <= count($a); $i += 1) { // ok
    echo $i;
  }

  for ($i = 0; $i <= count($a); $i += 100) { // ok
    echo $i;
  }

  for ($i = 100; $i > count($a); $i--) { // ok
    echo $i;
  }

  for ($i = 100; $i > count($a); $i -= 1) { // ok
    echo $i;
  }

  for ($i = 100; $i > count($a); $i -= 100) { // ok
    echo $i;
  }

  for ($i = 100; $i >= count($a); $i--) { // ok
    echo $i;
  }

  for ($i = 100; $i >= count($a); $i -= 1) { // ok
    echo $i;
  }

  for ($i = 100; $i >= count($a); $i -= 100) { // ok
    echo $i;
  }
}
