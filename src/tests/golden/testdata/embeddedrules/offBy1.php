<?php

function test_countindex_bad(array $xs, array $tabs) {
  // Potentially results in undefined index access.
  // In some very rare cases this code might work,
  // but it doesn't look like a good practive to rely on it.
  $_ = $xs[count($xs)];
  $_ = $xs[sizeof($xs)];

  if ($tabs[count($tabs)] == "") {
    // This expression is not detected. :(
    // See #713.
    unset($tabs[count($tabs)]);
  }
}

function test_countindex_good(array $xs) {
  $_ = $xs[count($xs)-1];
  $_ = $xs[sizeof($xs)-1];
  $_ = $xs[0];
}
