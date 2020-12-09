<?php
/** @return string */
function retString() { return ""; }

function nonStrictComparison($x) {
  $_ = ($x == false);
  $_ = (false == $x);
  $_ = ($x == true);
  $_ = (true == $x);
  $_ = ($x == null);
  $_ = (null == $x);
  return true;
}

$_ = (nonStrictComparison(0) != false);
$_ = (false != nonStrictComparison(0));
$_ = (nonStrictComparison(0) != true);
$_ = (true != nonStrictComparison(0));
$_ = (nonStrictComparison(0) != null);
$_ = (null != nonStrictComparison(0));

function nonStrictArraySearch() {
  $a = [];

  $_ = in_array("str", $a);
  $_ = in_array(retString(), $a);
  $_ = array_search("str", $a);
  $_ = array_search(retString(), $a);
}

function nonStrictArraySearchGood() {
  $a = [];

  $_ = in_array("str", $a, true);
  $_ = in_array(retString(), $a, true);
  $_ = in_array(10, $a);
  $_ = in_array(true, $a);

  $_ = array_search("str", $a, true);
  $_ = array_search(retString(), $a, true);
  $_ = array_search(10, $a, true);
  $_ = array_search(true, $a, true);
}
