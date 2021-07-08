<?php

function dupBranchBody($a) {
  switch ($a) {
    case 1:
      echo 1;
      break;
    case 2: // want `Branch 'case 2' in 'switch' is a duplicate`
      echo 1;
      break;
  }

  switch ($a) {
    case 1:
      echo 1;
      break;
    case 2: // want `Branch 'case 2' in 'switch' is a duplicate`
      echo 1;
      break;
    case 3: // want `Branch 'case 3' in 'switch' is a duplicate`
      echo 1;
      break;
  }

  switch ($a) {
    case 1:
      echo 1;
      break;
    case 2: // ok
      echo 2;
      break;
  }

  switch ($a) {
    case 1:
    case 2: // ok
      echo 2;
      break;
  }

  switch ($a) {
    case 1:
    case 2: // ok
      echo 2;
      break;
    case 3: // want `Branch 'case 3' in 'switch' is a duplicate`
      echo 2;
      break;
  }

  switch ($a) {
    case 1: case 2: // ok
      echo 2;
      break;
  }

  switch ($a) {
    case 1: case 2: // ok
      echo 2;
      break;
    default:
      echo 2;
      break;
  }

  switch ($a) {
    case 1: case 2: // ok
      echo 2;
      // fallthrough
    default:
      echo 2;
      break;
  }

  switch ($a) {
    case 1: case 2: // ok
      echo 2;
      // fallthrough
    case 3:
      echo 2;
      break;
  }
}
