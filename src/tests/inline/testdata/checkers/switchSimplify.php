<?php

function switchSimplify($a) {
  switch ($a) { // want `Switch can be rewritten into an 'if' statement to increase readability`
    case 10:
      echo 2;
      break;
  }

  switch ($a) { // ok
    case 10:
      echo 2;
      break;
    case 11:
      echo 3;
      break;
  }

  switch ($a) { // ok
    case 10:
      echo 2;
      break;
    case 11:
      echo 3;
      break;
    case 12:
      echo 4;
      break;
  }

  switch ($a) { // want `Switch can be rewritten into an 'if' statement to increase readability`
    case 10:
      echo 2;
      break;
    default:
      echo 3;
      break;
  }

  switch ($a) { // ok
    case 10:
      echo 2;
      break;
    case 11:
      echo 4;
      break;
    default:
      echo 3;
      break;
  }
}
