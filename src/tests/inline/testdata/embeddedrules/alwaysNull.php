<?php

class Foo {
  public int $prop = 100;
  public function method(): bool {return true; }
  public function method2($a, $b): bool { return true; }
}

function alwaysNull() {
  $obj = new Foo;

  if ($obj == null && $obj->method()) { echo 1;} // want `'$obj' is always 'null', maybe you meant '$obj != null && ...' or '$obj == null || ...'`
  if ($obj == null && $obj->method2(100, 200)) { echo 1;} // want `'$obj' is always 'null', maybe you meant '$obj != null && ...' or '$obj == null || ...'`
  if ($obj == null && $obj->prop) { echo 1; } // want `'$obj' is always 'null', maybe you meant '$obj != null && ...' or '$obj == null || ...'`

  if ($obj != null || $obj->method()) { echo 1;} // want `'$obj' is always 'null', maybe you meant '$obj == null || ...' or '$obj != null && ...'`
  if ($obj != null || $obj->method2(100, 200)) { echo 1;} // want `'$obj' is always 'null', maybe you meant '$obj == null || ...' or '$obj != null && ...'`
  if ($obj != null || $obj->prop) { echo 1; } // want `'$obj' is always 'null', maybe you meant '$obj == null || ...' or '$obj != null && ...'`

  if ($obj == null || $obj->method()) { echo 1;}
  if ($obj == null || $obj->method2(100, 200)) { echo 1;}
  if ($obj == null || $obj->prop) { echo 1; }

  if ($obj != null && $obj->method()) { echo 1;}
  if ($obj != null && $obj->method2(100, 200)) { echo 1;}
  if ($obj != null && $obj->prop) { echo 1; }
}
