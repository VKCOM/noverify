package irfmt_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/VKCOM/noverify/src/php/parseutil"
)

func TestPrinterSingleLine(t *testing.T) {
	testCases := []string{
		`$a`,
		`$a = $b`,
		`$a = 1`,
		`$a =& $b`,
		`$a &= $b`,
		`$a |= $b`,
		`$a ^= $b`,
		`$a .= $b`,
		`$a /= $b`,
		`$a -= $b`,
		`$a %= $b`,
		`$a *= $b`,
		`$a += $b`,
		`$a **= $b`,
		`$a <<= $b`,
		`$a >>= $b`,
		`$a & $b`,
		`$a | $b`,
		`$a ^ $b`,
		`$a && $b`,
		`$a || $b`,
		`$a ?? $b`,
		`$a . $b`,
		`$a / $b`,
		`$a == $b`,
		`$a >= $b`,
		`$a > $b`,
		`$a === $b`,
		`$a and $b`,
		`$a or $b`,
		`$a xor $b`,
		`$a - $b`,
		`$a % $b`,
		`$a * $b`,
		`$a != $b`,
		`$a !== $b`,
		`$a + $b`,
		`$a ** $b`,
		`$a << $b`,
		`$a >> $b`,
		`$a <= $b`,
		`$a < $b`,
		`$a <=> $b`,
		`(array)$var`,
		`(bool)$var`,
		`(float)$var`,
		`(int)$var`,
		`(object)$var`,
		`(string)$var`,
		`(unset)$var`,
		`$var[1]`,
		`$var{1}`,
		`['Hello' => $world]`,
		`array('Hello' => $world, 2 => &$var, $var)`,
		`[...$x, $y, ...$z]`,
		`~$var`,
		`!$var`,
		`$var::CONST`,
		`clone $var`,
		`empty($var)`,
		`@$var`,
		`eval($var)`,
		`foo($a, ...$b, $c)`,
		`include 'path'`,
		`include_once 'path'`,
		`require 'path'`,
		`require_once 'path'`,
		`$var instanceof Foo`,
		`isset($a, $b)`,
		`$foo->bar($a, $b)`,
		`new Foo($a, $b)`,
		`new Foo`,
		`new Foo()`,
		`$var--`,
		`$var++`,
		`--$var`,
		`++$var`,
		`$foo->bar`,
		`['Hello' => $world, 2 => &$var, $var]`,
		`[$a, list($b, $c)]`,
		`Foo::bar($a, $b)`,
		`Foo::$bar`,
		`Foo::a`,
		`$a ?: $b`,
		`$a ? $b : $c`,
		`-$var`,
		`+$var`,
		`$$var`,
		`yield from $var`,
		`yield $var`,
		`yield $k => $var`,
		`echo $a`,
		`echo $a, $b`,
		`global $a, $b`,
		`goto FOO`,
		`use function Foo\{Bar as Baz, Quuz}`,
		`__halt_compiler()`,
		`return 1.5`,
		`throw new Exception($x)`,
		`use Foo, Bar`,
		`unset($a, $b)`,
		`exit($var)`,
		`die($var)`,
		`list($a, list($b, $c)) = $a`,
	}

	for _, code := range testCases {
		code += ";"
		root, _, err := parseutil.ParseStmt([]byte(code))
		if err != nil {
			t.Fatalf("parse %s: %v", code, err)
		}
		rootIR := irconv.ConvertNode(root)

		var buf strings.Builder
		irfmt.NewPrettyPrinter(&buf, "    ").Print(rootIR)

		want := code
		have := buf.String()
		if have != want {
			t.Errorf("results mismatch (-have +want): %s", cmp.Diff(have, want))
		}
	}
}

func TestPrinter(t *testing.T) {
	testCases := []string{
		`FOO:

`,
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

		`function foo(&$world) {

}

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

		`function f(&$foo) {

}

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
		`new class
{
    public $foo = 10;
};

`,
		`new class(1, "arg2")
{

};

`,
	}

	for _, code := range testCases {
		code = "<?php\n" + code
		root, err := parseutil.ParseFile([]byte(code))
		if err != nil {
			t.Fatalf("parse %s: %v", code, err)
		}

		rootIR := irconv.ConvertNode(root)

		var buf strings.Builder
		irfmt.NewPrettyPrinter(&buf, "    ").Print(rootIR)

		want := code
		have := buf.String()
		if have != want {
			t.Errorf("results mismatch (-have +want): %s", cmp.Diff(want, have))
		}
	}
}
