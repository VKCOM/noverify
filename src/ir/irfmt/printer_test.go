package irfmt_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/VKCOM/noverify/src/php/parseutil"
)

func TestPrinter(t *testing.T) {

	testCases := []string{
		`namespace Foo {

}
abstract class Bar extends Baz
{
    public function greet()
    {
        echo 'Hello world';
    }
}
`,

		`$_ = 'hello world';

`,

		`$_ = <<<LBL
hello {$var} world
LBL;

`,

		`$_ = <<<'LBL'
hello world
LBL;

`,

		`$a = $b;

`,

		`$a =& $b;

`,

		`$a &= $b;

`,

		`$a |= $b;

`,

		`$a ^= $b;

`,

		`$a .= $b;

`,

		`$a /= $b;

`,

		`$a -= $b;

`,

		`$a %= $b;

`,

		`$a *= $b;

`,

		`$a += $b;

`,

		`$a **= $b;

`,

		`$a <<= $b;

`,

		`$a >>= $b;

`,

		`$a & $b;

`,

		`$a | $b;

`,

		`$a ^ $b;

`,

		`$a && $b;

`,

		`$a || $b;

`,

		`$a ?? $b;

`,

		`$a . $b;

`,

		`$a / $b;

`,

		`$a == $b;

`,

		`$a >= $b;

`,

		`$a > $b;

`,

		`$a === $b;

`,

		`$a and $b;

`,

		`$a or $b;

`,

		`$a xor $b;

`,

		`$a - $b;

`,

		`$a % $b;

`,

		`$a * $b;

`,

		`$a != $b;

`,

		`$a !== $b;

`,

		`$a + $b;

`,

		`$a ** $b;

`,

		`$a << $b;

`,

		`$a >> $b;

`,

		`$a <= $b;

`,

		`$a < $b;

`,

		`$a <=> $b;

`,

		`(array)$var;

`,

		`(bool)$var;

`,

		`(float)$var;

`,

		`(int)$var;

`,

		`(object)$var;

`,

		`(string)$var;

`,

		`(unset)$var;

`,

		`$var[1];

`,

		`$_ = ['Hello' => $world];

`,

		`function foo(&$world) {

}

`,

		`$_ = array('Hello' => $world, 2 => &$var, $var);

`,

		`~$var;

`,

		`!$var;

`,

		`$var::CONST;

`,

		`clone $var;

`,

		`$_ = function () use (&$foo, $bar) {

};

`,

		`namespace {
    $_ = function &(&$var) use (&$a, $b): Foo {
        $a;
    };
}

`,

		`empty($var);

`,

		`@$var;

`,

		`eval($var);

`,

		`exit($var);

`,

		`die($var);

`,

		`foo($a, ...$b, $c);

`,

		`include 'path';

`,

		`include_once 'path';

`,

		`$var instanceof Foo;

`,

		`isset($a, $b);

`,

		`list($a, list($b, $c)) = $a;

`,

		`$foo->bar($a, $b);

`,

		`new Foo($a, $b);

`,

		`new Foo;

`,

		`new Foo();

`,

		`$var--;

`,

		`$var++;

`,

		`--$var;

`,

		`++$var;

`,

		`$foo->bar;

`,

		`function f(&$foo) {

}

`,

		`require 'path';

`,

		`require_once 'path';

`,

		`['Hello' => $world, 2 => &$var, $var];

`,

		`[$a, list($b, $c)];

`,

		`Foo::bar($a, $b);

`,

		`Foo::$bar;

`,

		`$a ?: $b;

`,

		`$a ? $b : $c;

`,

		`-$var;

`,

		`+$var;

`,

		`$$var;

`,

		`yield from $var;

`,

		`yield $var;

`,

		`yield $k => $var;

`,

		`if ($b) :
    $a;
elseif ($a) :
    $b;
else :
    $c;
endif;

`,

		`namespace {
    for ($a; $b; $c) :
        $d;
    endfor;
}

`,

		`namespace {
    foreach ($var as $key => &$val) :
        $d;
    endforeach;
}

`,

		`namespace {
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

		`namespace {
    switch ($var) :
        case 'a':
            $a;
        case 'b':
            $b;
    endswitch;
}

`,

		`namespace {
    while ($a) :
        $b;
    endwhile;
}

`,

		`switch ($a) :
    case $a:
        $a;
endswitch;

`,

		`namespace {
    try {

    }
    catch (Exception | \RuntimeException $e) {
        $a;
    }
}

`,

		`class Foo
{
    public function &foo(?int &$a = null, ...$b): void
    {
        $a;
    }
}

`,

		`class Foo
{
    public function &foo(?int &$a = null, ...$b): void;
}

`,

		`namespace {
    abstract class Foo extends Bar implements Baz, Quuz
    {
        public const FOO = 'bar';
    }
}

`,

		`namespace {
    abstract class Foo extends Bar implements Baz, Quuz
    {
        public const FOO = 'bar';
    }
}

`,

		`class Foo
{
    public const FOO = 'a', BAR = 'b';
}

`,

		`for ($a = 1; $a < 5; $a++) {
    continue 1;
}

`,

		`{
    declare(FOO = 'bar') {
        ;
    }
}

`,

		`{
    declare(FOO = 'bar')
        'bar';
}

`,

		`declare(FOO = 'bar');

`,

		`switch ($a) {
    default:
        $a;
}

`,

		`namespace {
    do
        $a;
    while (1);
}

`,

		`namespace {
    do {
        $a;
    } while (1);
}

`,

		`echo $a, $b;

`,

		`$a;

`,

		`namespace {
    try {

    }

    finally {
        ;
    }
}

`,

		`namespace {
    for ($a, $b; $c, $d; $e, $f) {
        ;
    }
}

`,

		`namespace {
    for ($a; $b; $c)
        'bar';
}

`,

		`for ($a; $b; $c);

`,

		`namespace {
    foreach ($a as $b) {
        ;
    }
}

`,

		`namespace {
    foreach ($a as $k => $v)
        'bar';
}

`,

		`foreach ($a as $k => &$v);

`,

		`namespace {
    function &foo(&$var): \Foo {
        ;
    }
}

`,

		`global $a, $b;

`,

		`goto FOO;

`,

		`use function Foo\{Bar as Baz, Quuz};

`,

		`__halt_compiler();

`,

		`namespace {
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

		`namespace {
    if ($a) {
        $b;
    }

}

`,

		`if ($a);


`,

		`;
?>test
<?php
`,

		`namespace {
    interface Foo extends Bar, Baz
    {
        public function foo()
        {
            $a;
        }
    }
}

`,

		`FOO:

`,

		`namespace Foo {

}

`,

		`namespace Foo {
    $a;
}

`,

		`;

`,

		`class Foo
{
    public static $a, $b;
}

`,

		`$a = 1;

`,

		`return 1;

`,

		`class Foo
{
    static $a, $b;
}

`,

		`{
    $a;
    $b;
}

`,

		`{
    $a;
    {
        $b;
        {
            $c;
        }
    }
}

`,

		`{
    switch ($var) {
        case 'a':
            $a;
        case 'b':
            $b;
    }
}

`,

		`throw $var;

`,

		`Foo::a;

`,

		`use Foo, Bar;

`,

		`namespace {
    trait Foo
    {
        public function foo()
        {
            $a;
        }
    }
}

`,

		`namespace {
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

		`unset($a, $b);

`,

		`namespace {
    while ($a) {
        $a;
    }
}

`,

		`namespace {
    while ($a)
        $a;
}

`,

		`while ($a);

`,
	}

	for _, code := range testCases {
		runPrinterTest(t, code)
	}
}

func runPrinterTest(t *testing.T, code string) {
	t.Helper()

	code = "<?php\n" + code
	root, err := parseutil.ParseFile([]byte(code))
	if err != nil {
		t.Errorf("parse %s: %v", code, err)
		return
	}
	rootIR := irconv.NewConverter().ConvertRoot(root)

	o := bytes.NewBufferString("")
	p := irfmt.NewPrettyPrinter(o, "    ")

	p.Print(rootIR)
	want := o.String()
	have := code

	if !cmp.Equal(want, have) {
		t.Errorf("results mismatch (+ have) (- want): %s", cmp.Diff(want, have))
	}
}
