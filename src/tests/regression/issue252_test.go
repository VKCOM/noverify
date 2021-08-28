package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue252(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {
  public $foo = 10;
}
class Bar {
  public $bar = 20;
}
function alt_foreach($arr) {
  foreach ($arr AS $key => $value):
    $_ = [$key, $value];
  endforeach;
}
function alt_if($v) {
  if ($v instanceof Foo):
    $_ = $v->foo;
  elseif ($v instanceof Bar):
    $_ = $v->bar;
  endif;
}
`)

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function alt_for() {
  for ($i = 0; $i < 10; $i++):
    $x1 = 10;
  endfor;
  $_ = $x1;
}
function alt_switch($v) {
  switch ($v):
  case 1:
    $v = 3;
  case 2:
    return $v;
  default:
    break;
  endswitch;
}`)
	test.Expect = []string{
		`Possibly undefined variable $x1`,
		`Add break or '// fallthrough' to the end of the case`,
	}
	test.RunAndMatch()
}
