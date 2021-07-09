<?php

function switchDefaultShouldBeFirstOrLastRule($a) {
  switch ($a) {
    case 100:
      echo 2;
      break;
    default: // want `'default' should be first or last to improve readability`
      echo 100;
      break;
    case 101:
      echo 3;
      break;
  }

  switch ($a) {
    case 100:
      echo 2;
      break;
    case 102:
      echo 4;
      break;
    default: // want `'default' should be first or last to improve readability`
      echo 100;
      break;
    case 101:
      echo 3;
      break;
  }

  switch ($a) {
    case 100:
      echo 2;
      break;
    case 102:
      echo 4;
      break;
    default: // want `'default' should be first or last to improve readability`
      echo 100;
      break;
    case 103:
      echo 5;
      break;
    case 101:
      echo 3;
      break;
  }

  switch ($a) {
    case 100:
      echo 2;
      break;
    default: // want `'default' should be first or last to improve readability`
    case 101:
      echo 3;
      break;
  }

  switch ($a) {
    case 100:
      echo 2;
      break;
    case 101:
    default: // ok
      echo 3;
      break;
  }

  switch ($a) {
    case 102:
    case 101:
      echo 3;
      break;
    default: // ok
      echo 100;
      break;
  }

  switch ($a) {
    case 100:
      echo 2;
      break;
    case 101:
      echo 3;
      break;
    default: // ok
      echo 100;
      break;
  }

  switch ($a) {
    default: // ok
      echo 100;
      break;
    case 100:
      echo 2;
      break;
    case 101:
      echo 3;
      break;
  }

  switch ($a) {
    default: // ok
      echo 100;
      break;
    case 102:
    case 100:
      echo 2;
      break;
  }
}

function switchShouldContainDefault($a) {
  switch ($a) { // want `Add 'default' branch to avoid unexpected unhandled condition values`
    case 102:
    case 103:
    case 101:
      echo 3;
      break;
  }

  switch ($a) { // want `Add 'default' branch to avoid unexpected unhandled condition values`
    case 102:
    case 100:
      echo 2;
      break;
    case 101:
      echo 3;
      break;
  }

  switch ($a) { // ok
    case 102:
    case 101:
      echo 3;
      break;
    default:
      echo 4;
      break;
  }

  switch ($a) { // ok
    case 101:
    case 102:
    default:
      echo 4;
      break;
  }
}
