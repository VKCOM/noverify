<?php

function unaryRepeat() {
  $a = 100;

  echo !$a;    // ok
  echo !!$a;   // want `Several negations in a row does not make sense`
  echo !!!$a;  // want `Several negations in a row does not make sense`
  echo !!!!$a; // want `Several negations in a row does not make sense`

  echo ~$a;     // ok
  echo ~~$a;    // want `Several bitwise not (~) in a row does not make sense`
  echo ~~~$a;   // want `Several bitwise not (~) in a row does not make sense`
  echo ~~~~$a;  // want `Several bitwise not (~) in a row does not make sense`
}
