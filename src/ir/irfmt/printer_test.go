package irfmt_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/VKCOM/noverify/src/linttest"
)

func TestPrinter(t *testing.T) {

	testCases := []string{
		`<?php
namespace Foo {

}
abstract class Bar extends Baz
{
    public function greet()
    {
        echo 'Hello world';
    }
}
`,

		`<?php
$_ = 'hello world';

`,

		`<?php
$_ = <<<LBL
hello {$var} world
LBL;

`,

		`<?php
$_ = <<<'LBL'
hello world
LBL;

`,

		`<?php
$a = $b;

`,

		`<?php
$a =& $b;

`,

		`<?php
$a &= $b;

`,

		`<?php
$a |= $b;

`,

		`<?php
$a ^= $b;

`,

		`<?php
$a .= $b;

`,

		`<?php
$a /= $b;

`,

		`<?php
$a -= $b;

`,

		`<?php
$a %= $b;

`,

		`<?php
$a *= $b;

`,

		`<?php
$a += $b;

`,

		`<?php
$a **= $b;

`,

		`<?php
$a <<= $b;

`,

		`<?php
$a >>= $b;

`,

		`<?php
$a & $b;

`,

		`<?php
$a | $b;

`,

		`<?php
$a ^ $b;

`,

		`<?php
$a && $b;

`,

		`<?php
$a || $b;

`,

		`<?php
$a ?? $b;

`,

		`<?php
$a . $b;

`,

		`<?php
$a / $b;

`,

		`<?php
$a == $b;

`,

		`<?php
$a >= $b;

`,

		`<?php
$a > $b;

`,

		`<?php
$a === $b;

`,

		`<?php
$a and $b;

`,

		`<?php
$a or $b;

`,

		`<?php
$a xor $b;

`,

		`<?php
$a - $b;

`,

		`<?php
$a % $b;

`,

		`<?php
$a * $b;

`,

		`<?php
$a != $b;

`,

		`<?php
$a !== $b;

`,

		`<?php
$a + $b;

`,

		`<?php
$a ** $b;

`,

		`<?php
$a << $b;

`,

		`<?php
$a >> $b;

`,

		`<?php
$a <= $b;

`,

		`<?php
$a < $b;

`,

		`<?php
$a <=> $b;

`,

		`<?php
(array)$var;

`,

		`<?php
(bool)$var;

`,

		`<?php
(float)$var;

`,

		`<?php
(int)$var;

`,

		`<?php
(object)$var;

`,

		`<?php
(string)$var;

`,

		`<?php
(unset)$var;

`,

		`<?php
$var[1];

`,

		`<?php
$_ = ['Hello' => $world];

`,

		`<?php
function foo(&$world) {

}

`,

		`<?php
$_ = array('Hello' => $world, 2 => &$var, $var);

`,

		`<?php
~$var;

`,

		`<?php
!$var;

`,

		`<?php
$var::CONST;

`,

		`<?php
clone $var;

`,

		`<?php
$_ = function () use (&$foo, $bar) {

};

`,

		`<?php
namespace {
    $_ = function &(&$var) use (&$a, $b): Foo {
        $a;
    };
}

`,

		`<?php
empty($var);

`,

		`<?php
@$var;

`,

		`<?php
eval($var);

`,

		`<?php
exit($var);

`,

		`<?php
die($var);

`,

		`<?php
foo($a, ...$b, $c);

`,

		`<?php
include 'path';

`,

		`<?php
include_once 'path';

`,

		`<?php
$var instanceof Foo;

`,

		`<?php
isset($a, $b);

`,

		`<?php
list($a, list($b, $c)) = $a;

`,

		`<?php
$foo->bar($a, $b);

`,

		`<?php
new Foo($a, $b);

`,

		`<?php
new Foo;

`,

		`<?php
new Foo();

`,

		`<?php
$var--;

`,

		`<?php
$var++;

`,

		`<?php
--$var;

`,

		`<?php
++$var;

`,

		`<?php
$foo->bar;

`,

		`<?php
function f(&$foo) {

}

`,

		`<?php
require 'path';

`,

		`<?php
require_once 'path';

`,

		`<?php
['Hello' => $world, 2 => &$var, $var];

`,

		`<?php
[$a, list($b, $c)];

`,

		`<?php
Foo::bar($a, $b);

`,

		`<?php
Foo::$bar;

`,

		`<?php
$a ?: $b;

`,

		`<?php
$a ? $b : $c;

`,

		`<?php
-$var;

`,

		`<?php
+$var;

`,

		`<?php
$$var;

`,

		`<?php
yield from $var;

`,

		`<?php
yield $var;

`,

		`<?php
yield $k => $var;

`,

		`<?php
if ($b) :
    $a;
elseif ($a) :
    $b;
else :
    $c;
endif;

`,

		`<?php
namespace {
    for ($a; $b; $c) :
        $d;
    endfor;
}

`,

		`<?php
namespace {
    foreach ($var as $key => &$val) :
        $d;
    endforeach;
}

`,

		`<?php
namespace {
    if ($a) :
        $d;
    elseif ($b) :
        $b;
    elseif ($c) :
    else :
        $b;
    endif;
}

`,

		`<?php
namespace {
    switch ($var) :
        case 'a':
            $a;
        case 'b':
            $b;
    endswitch;
}

`,

		`<?php
namespace {
    while ($a) :
        $b;
    endwhile;
}

`,

		`<?php
switch ($a) :
    case $a:
        $a;
endswitch;

`,

		`<?php
namespace {
    try {

    }
    catch (Exception | \RuntimeException $e) {
        $a;
    }
}

`,

		`<?php
class Foo
{
    public function &foo(?int &$a = null, ...$b): void
    {
        $a;
    }
}

`,

		`<?php
class Foo
{
    public function &foo(?int &$a = null, ...$b): void;
}

`,

		`<?php
namespace {
    abstract class Foo extends Bar implements Baz, Quuz
    {
        public const FOO = 'bar';
    }
}

`,

		`<?php
namespace {
    abstract class Foo extends Bar implements Baz, Quuz
    {
        public const FOO = 'bar';
    }
}

`,

		`<?php
class Foo
{
    public const FOO = 'a', BAR = 'b';
}

`,

		`<?php
for ($a = 1; $a < 5; $a++) {
    continue 1;
}

`,

		`<?php
{
    declare(FOO = 'bar') {
        ;
    }
}

`,

		`<?php
{
    declare(FOO = 'bar')
        'bar';
}

`,

		`<?php
declare(FOO = 'bar');

`,

		`<?php
switch ($a) {
    default:
        $a;
}

`,

		`<?php
namespace {
    do
        $a;
    while (1);
}

`,

		`<?php
namespace {
    do {
        $a;
    } while (1);
}

`,

		`<?php
echo $a, $b;

`,

		`<?php
$a;

`,

		`<?php
namespace {
    try {

    }

    finally {
        ;
    }
}

`,

		`<?php
namespace {
    for ($a, $b; $c, $d; $e, $f) {
        ;
    }
}

`,

		`<?php
namespace {
    for ($a; $b; $c)
        'bar';
}

`,

		`<?php
for ($a; $b; $c);

`,

		`<?php
namespace {
    foreach ($a as $b) {
        ;
    }
}

`,

		`<?php
namespace {
    foreach ($a as $k => $v)
        'bar';
}

`,

		`<?php
foreach ($a as $k => &$v);

`,

		`<?php
namespace {
    function &foo(&$var): \Foo {
        ;
    }
}

`,

		`<?php
global $a, $b;

`,

		`<?php
goto FOO;

`,

		`<?php
use function Foo\{Bar as Baz, Quuz};

`,

		`<?php
__halt_compiler();

`,

		`<?php
namespace {
    if ($a)
        $b;
    elseif ($c) {
        $d;
    }
    elseif ($e);
    else
        $f;
}

`,

		`<?php
namespace {
    if ($a) {
        $b;
    }

}

`,

		`<?php
if ($a);


`,

		`<?php
;
?>test
<?php
`,

		`<?php
namespace {
    interface Foo extends Bar, Baz
    {
        public function foo()
        {
            $a;
        }
    }
}

`,

		`<?php
FOO:

`,

		`<?php
namespace Foo {

}

`,

		`<?php
namespace Foo {
    $a;
}

`,

		`<?php
;

`,

		`<?php
class Foo
{
    public static $a, $b;
}

`,

		`<?php
$a = 1;

`,

		`<?php
return 1;

`,

		`<?php
class Foo
{
    static $a, $b;
}

`,

		`<?php
{
    $a;
    $b;
}

`,

		`<?php
{
    $a;
    {
        $b;
        {
            $c;
        }
    }
}

`,

		`<?php
{
    switch ($var) {
        case 'a':
            $a;
        case 'b':
            $b;
    }
}

`,

		`<?php
throw $var;

`,

		`<?php
Foo::a;

`,

		`<?php
use Foo, Bar;

`,

		`<?php
namespace {
    trait Foo
    {
        public function foo()
        {
            $a;
        }
    }
}

`,

		`<?php
namespace {
    try {
        $a;
    }
    catch (Exception | \RuntimeException $e) {
        $b;
    }
    finally {
        ;
    }
}

`,

		`<?php
unset($a, $b);

`,

		`<?php
namespace {
    while ($a) {
        $a;
    }
}

`,

		`<?php
namespace {
    while ($a)
        $a;
}

`,

		`<?php
while ($a);

`,
	}

	for _, code := range testCases {
		runPrinterTest(t, code)
	}
}

func runPrinterTest(t *testing.T, code string) {
	rootNode, _ := linttest.ParseTestFile(t, "printer_test.php", code)

	o := bytes.NewBufferString("")
	p := irfmt.NewPrettyPrinter(o, "    ")

	p.Print(rootNode)
	want := o.String()
	have := code

	if !cmp.Equal(want, have) {
		t.Errorf("results mismatch (+ have) (- want): %s", cmp.Diff(want, have))
	}
}
