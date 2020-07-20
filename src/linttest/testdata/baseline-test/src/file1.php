<?php

"this line is suppressed by the baseline";

function f1() {
  return g();
}

function f2() {
  "this line is not suppressed";
}
