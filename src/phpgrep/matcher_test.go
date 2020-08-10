package phpgrep

import (
	"fmt"
	"sync"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/irgen"
	"github.com/VKCOM/noverify/src/php/parseutil"
)

func mustParse(t testing.TB, code string) ir.Node {
	n, _, err := parseutil.Parse([]byte(code))
	if err != nil {
		t.Fatalf("parse `%s`: %v", code, err)
	}
	irnode := irgen.ConvertNode(n)
	if n, ok := irnode.(*ir.ExpressionStmt); ok {
		return n.Expr
	}
	return irnode
}

func matchInText(t *testing.T, m *Matcher, code string) bool {
	_, ok := m.Match(mustParse(t, code))
	return ok
}

type matcherTest struct {
	pattern string
	input   string
}

func mustCompile(t testing.TB, code string) *Matcher {
	var c Compiler
	matcher, err := c.Compile([]byte(code))
	if err != nil {
		t.Fatalf("pattern compilation error:\ntext: %q\nerr: %v", code, err)
	}
	return matcher
}

func runMatchTest(t *testing.T, want bool, tests []*matcherTest) {
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%v", i, want), func(t *testing.T) {
			matcher := mustCompile(t, test.pattern)
			have := matchInText(t, matcher, test.input)
			if have != want {
				t.Errorf("match results mismatch:\npattern: %q\ninput: %q\nhave: %v\nwant: %v",
					test.pattern, test.input, have, want)
			}
		})
	}
}

func TestMatchDebug(t *testing.T) {
	runMatchTest(t, true, []*matcherTest{
		{`if ($c) $_; else if ($c) {1;};`, `if ($c1) {1; 2;} else if ($c1) {1;}`},
	})
}

func TestMatchCapture(t *testing.T) {
	checkCapture := func(m *MatchData, name, want string) {
		n, ok := m.CapturedByName(name)
		if !ok {
			t.Errorf("%s not captured", name)
			return
		}
		have := irutil.FmtNode(n)
		if have != want {
			t.Errorf("%s mismatched: have %s, want %s", name, have, want)
		}
	}

	matcher := mustCompile(t, `$x = $x[$y]`)
	for i := 0; i < 5; i++ {
		result, ok := matcher.Match(mustParse(t, `$a[0] = $a[0][1]`))
		if !ok {
			t.Fatalf("pattern not matched")
		}
		checkCapture(&result, "x", "$a[0]")
		checkCapture(&result, "y", "1")
	}
}

func TestMatchConcurrent(t *testing.T) {
	matcher := mustCompile(t, `f($x, ${"*"}, $x)`)

	nodes := []ir.Node{
		mustParse(t, `1`),
		mustParse(t, `f(1, 2, 3, 4, 3, 2, 1)`),
		mustParse(t, `[0 => f(1, 2), 2 => f(1, 1)]`),
		mustParse(t, `if ($x) { f(); f([1], 2, [1]); }`),
		mustParse(t, `for (;;) { { f($x[0], $x[0], $x[0], $x[0]); } }`),
	}

	const (
		numGoroutines = 100
		numRepeats    = 200
	)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for i := 0; i < numRepeats; i++ {
				for _, n := range nodes {
					matcher.Match(n)
				}
			}
			wg.Done()
		}()
	}
}

func TestMatch(t *testing.T) {
	runMatchTest(t, true, []*matcherTest{
		{"``", "``"},
		{"`ls`", "`ls`"},
		{"`rm -rf /`", "`rm -rf /`"},

		{`${"const"}`, `null`},
		{`${"const"}`, `true`},
		{`${"const"}`, `false`},
		{`${"const"}`, `MY_CONST`},
		{`$_ = ${"const"}`, `$x = MyClass::CONST`},

		{`static $x = 10`, `static $vvv = 10`},
		{`global $x, $y`, `global $a, $b`},
		{`break $x`, `break 20`},
		{`continue $x`, `continue 20`},
		{`unset($x)`, `unset($v)`},
		{`print(1)`, `print(1)`},
		{`echo 1, 2`, `echo 1, 2`},
		{`throw new $E()`, `throw new Exception()`},

		{`!($x instanceof $y)`, `!($v instanceof MyClass)`},
		{`$x`, `$v instanceof MyClass`},
		{`$x`, `!($v instanceof MyClass)`},

		{`$x=$x`, `$x=$x`},

		{`1`, `1`},
		{`"1"`, `"1"`},
		{`'1'`, `'1'`},
		{`1.4`, `1.4`},

		{`$x & mask != 0`, `$v & mask != 0`},
		{`($x & mask) != 0`, `($v & mask) != 0`},

		{`$x`, `10`},
		{`$x`, `"abc"`},
		{`false`, `false`},
		{`NULL`, `NULL`},

		{`$x++`, `$y++`},
		{`$x--`, `$y--`},
		{`++$x`, `++$y`},
		{`--$x`, `--$y`},

		{`$x+1`, `10+1`},
		{`$x+1`, `$x+1`},
		{`$x-1`, `10-1`},
		{`$x-1`, `$x-1`},

		{`+$x`, `+1`},
		{`-$x`, `-2`},
		{`~$x`, `~$v`},
		{`!$x`, `!$v`},

		{`$f()`, `f()`},
		{`$f()`, `g()`},
		{`$f($a1, $a2)`, `f(1, 2)`},
		{`$f($a1, $a2)`, `f("sa", $t)`},

		{`$x + $x`, `1 + 1`},
		{`$x + $y`, `1 + 1`},
		{`$x | $y`, `$v1 | $v2`},
		{`$x >> $y`, `$v1 >> $v2`},
		{`$x << $y`, `$v1 << $v2`},
		{`$x and $y`, `$v1 and $v2`},
		{`$x or $y`, `$v1 or $v2`},
		{`$x xor $y`, `$v1 xor $v2`},
		{`$x != $y`, `$v1 != $v2`},
		{`$x == $y`, `$v1 == $v2`},
		{`$x === $y`, `$v1 === $v2`},
		{`$x !== $y`, `$v1 !== $v2`},
		{`$x > $y`, `$v1 > $v2`},
		{`$x >= $y`, `$v1 >= $v2`},
		{`$x < $y`, `$v1 < $v2`},
		{`$x <= $y`, `$v1 <= $v2`},
		{`$x <=> $y`, `$v1 <=> $v2`},
		{`$x && $y`, `$v1 && $v2`},
		{`$x || $y`, `$v1 || $v2`},
		{`$x ?? $y`, `$v1 ?? $v2`},
		{`$x . $y`, `$v1 . $v2`},
		{`$x / $y`, `$v1 / $v2`},
		{`$x % $y`, `$v1 % $v2`},
		{`$x * $y`, `$v1 * $v2`},
		{`$x ** $y`, `$v1 ** $v2`},

		{`$x = $x`, `$x->b = $x->b`},
		{`$x = $x`, `$x->b[0] = $x->b[0]`},
		{`$x = $x`, `$a->$x[0] = $a->$x[0]`},
		{`$x = $x`, `$x[0] = $x[0]`},
		{`$x = $x`, `T::$x = T::$x`},

		{`int($x)`, `int($v)`},
		{`array($x)`, `array($v)`},
		{`string($x)`, `string($v)`},
		{`bool($x)`, `bool($v)`},
		{`double($x)`, `double($v)`},
		{`object($x)`, `object($v)`},

		{`$$$x`, `$$$x`},
		{`$$$x`, `$$$y`},
		{`$$$x`, `$$$$x`},
		{`$$x = $$x`, `$$$x = $$$x`},
		{`$$x = $$x`, `$$x = $$x`},

		{`$x = 0`, `$v = 0`},
		{`$x += 1`, `$v += 1`},
		{`$x -= 1`, `$v -= 1`},
		{`$x =& $y`, `$x =& $y`},
		{`$x &= $y`, `$x &= $y`},
		{`$x |= $y`, `$x |= $y`},
		{`$x ^= $y`, `$x ^= $y`},
		{`$x /= $y`, `$x /= $y`},
		{`$x %= $y`, `$x %= $y`},
		{`$x *= $y`, `$x *= $y`},
		{`$x **= $y`, `$x **= $y`},
		{`$x <<= $y`, `$x <<= $y`},
		{`$x >>= $y`, `$x >>= $y`},

		{`\A\B`, `\A\B`},

		{`[]`, `[]`},
		{`array()`, `array()`},
		{`[$x, $x]`, `[1, 1]`},
		{`array($x, $x)`, `array(1, 1)`},
		{`[$k1 => 2, $k2 => 4]`, `[1 => 2, 3 => 4]`},
		{`[$k1 => 2, $k1 => 4]`, `[1 => 2, 1 => 4]`},

		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[1 => $x, 1 => $y]`},
		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[$v, 1 => $x, $v, 1 => $x, $v]`},
		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[1 => $x, 1 => $x, $v]`},
		{`[${'*'}, $k => $_, ${'*'}, $k => $_, ${'*'}]`, `[$v, 1 => $x, 1 => $x]`},

		{`{1; 2;}`, `{1; 2;}`},
		{`{$x;}`, `{1;}`},
		{`{$x;}`, `{2;}`},

		{`{${'*'};}`, `{}`},
		{`{${'*'};}`, `{1;}`},
		{`{${'*'};}`, `{1; 2;}`},
		{`{${'*'};}`, `{1; 2; 3;}`},

		{`{${'*'}; 3;}`, `{1; 2; 3;}`},
		{`{1; ${'*'};}`, `{1; 2; 3;}`},
		{`{1; ${'*'}; 3;}`, `{1; 2; 3;}`},
		{`{1; 2; ${'*'}; 3;}`, `{1; 2; 3;}`},
		{`{${'*'}; 2; ${'*'};}`, `{1; 2; 3;}`},
		{`{1; 2; 3; ${'*'};}`, `{1; 2; 3;}`},

		{`f(${'*'})`, `f()`},
		{`f(${'*'})`, `f(1)`},
		{`f(${'*'})`, `f(1, 2)`},
		{`f(${'*'})`, `f(1, 2, 3)`},
		{`f(${'*'}, 3)`, `f(1, 2, 3)`},
		{`f(${'*'}, $x, $y, $z)`, `f(1, 2, 3)`},
		{`f($x, $y, $z, ${'*'})`, `f(1, 2, 3)`},
		{`f(${'*'}, $x, ${'*'}, $y, ${'*'}, $z, ${'*'})`, `f(1, 2, 3)`},

		{`if ($cond) $_;`, `if (1 == 1) return 1;`},
		{`if ($cond) $_;`, `if (1 == 1) f();`},
		{`if ($cond) return 1;`, `if (1 == 1) return 1;`},
		{`if ($cond) { return 1; }`, `if (1 == 1) { return 1; }`},
		{`if ($_ = $_) $_`, `if ($x = f()) {}`},
		{`if ($_ = $_) $_`, `if ($x = f()) g();`},
		{`if ($cond1) $_; else if ($cond2) $_;`, `if ($c1) {} else if ($c2) {}`},
		{`if ($cond1) $_; elseif ($cond2) $_;`, `if ($c1) {} elseif ($c2) {}`},

		{`switch ($e) {}`, `switch ($x) {}`},
		{`switch ($_) {case 1: f();}`, `switch ($x) {case 1: f();}`},
		{`switch ($_) {case $_: ${'*'};}`, `switch ($x) {case 1: f1(); f2();}`},
		{`switch ($e) {default: $_;}`, `switch ($x) {default: 1;}`},

		{`strcmp($s1, $s2) > 0`, `strcmp($s1, "foo") > 0`},

		{`new $t`, `new T`},
		{`new $t()`, `new T()`},
		{`new $t($x)`, `new T(1)`},
		{`new $t($x, $y)`, `new T(1, 2)`},
		{`new $t(${'*'})`, `new T(1, 2)`},

		{`list($x, $_, $x) = f()`, `list($v, $_, $v) = f()`},
		{`list($x, $_, $x) = f()`, `list($v, , $v) = f()`},
		{`list($x) = $a`, `list($v) = [1]`},

		{`${'var'}`, `$x`},
		{`${'var'}`, `$$x`},
		{`${'x:var'} + $x`, `$x + $x`},
		{`$x + ${'x:var'}`, `$x + $x`},
		{`${'_:var'} + $_`, `$x + 1`},
		{`${'var'} + $_`, `$x + 1`},

		{`${"int"}`, `13`},
		{`${"float"}`, `3.4`},
		{`${"str"}`, `"123"`},
		{`${"num"}`, `13`},
		{`${"num"}`, `3.4`},

		{`${"expr"}`, `1`},
		{`${"expr"}`, `"124d"`},
		{`${"expr"}`, `$x`},
		{`${"expr"}`, `f(1, 5)`},
		{`${"expr"}`, `$x = [1]`},

		{`$cond ? $true : $false`, `1 ? 2 : 3`},
		{`$cond ? a : b`, `1 ? a : b`},
		{`$c1 ? $_ : $_ ? $_ : $_`, `true ? 1 : false ? 2 : 3`},
		{`$c1 ? $_ : ($_ ? $_ : $_)`, `true ? 1 : (false ? 2 : 3)`},
		{`$x ? $x : $y`, `$v ? $v : $other`},
		{`$_ == $_ ? $_ : $_ == $_ ? $_ : $_`, `$a == 1 ? 'one' : $a == 2 ? 'two' : 'other'`},

		{`$_ ?: $_`, `1 ?: 2`},

		{`isset($x)`, `isset($v)`},
		{`isset($x, $y)`, `isset($k, $v[$k])`},
		{`empty($x)`, `empty($v)`},

		{`$x->$_ = $x`, `$this->self = $this`},
		{`$x->$_ = $x`, `$this->$indirect = $this`},
		{`$x->$m()`, `$this->m()`},
		{`$x->$m(1, 2)`, `$this->m(1, 2)`},
		{`$x->ff(1, 2)`, `$this->ff(1, 2)`},

		{`$_[0]`, `$v[0]`},

		{`$c::$prop`, `C::$foo`},
		{`$c::$prop`, `C::constant`},
		{`$c::$f()`, `C::foo()`},
		{`$c::$f()`, `C::$foo()`},
		{`C::f()`, `C::f()`},
		{`C::constant`, `C::constant`},

		{`clone $v`, `clone new T()`},

		{`@$_`, `@f()`},
		{`@$_`, `@$o->method(1, 2)`},

		{`eval($_)`, `eval('1')`},

		{`exit(0)`, `exit(0)`},
		{`die(0)`, `die(0)`},

		{`include $_`, `include "foo.php"`},
		{`include_once $_`, `include_once "foo.php"`},
		{`require $_`, `require "foo.php"`},
		{`require_once $_`, `require_once "foo.php"`},

		{`__FILE__`, `__FILE__`},
		{`[$x, $x]`, `[__FILE__, __FILE__]`},

		{`"$x$y"`, `"$x$y"`},
		{`"$x 1" . $x`, `"$x 1" . "2"`},
		{`"${x}"`, `"${x}"`},

		{`function() { return $x; }`, `function() { return 10; }`},
		{`function($x) {}`, `function($arg1) {}`},
		{`function($x) use($v) {}`, `function($arg1) use($y) {}`},
		{`function() { ${"*"}; return 1; }`, `function() { return 1; }`},
		{`function() { ${"*"}; return 1; }`, `function() { f(); return 1; }`},
		{`function() { ${"*"}; return 1; }`, `function() { f(); f(); return 1; }`},

		{`${"func"}`, `function() {}`},
		{`${"func"}`, `function($x) {}`},
		{`${"func"}`, `function() { return 1; }`},

		{`1`, `1`},
		{`(1)`, `(1)`},
		{`((1))`, `((1))`},
		{`f(1)`, `f(1)`},
		{`f((1))`, `f((1))`},
		{`($foo)()`, `($foo)()`},
		{`($foo)->x`, `($foo)->x`},
	})
}

func TestMatchNegative(t *testing.T) {
	runMatchTest(t, false, []*matcherTest{
		{`1`, `2`},
		{`"1"`, `"2ed"`},
		{`'1'`, `'x'`},
		{`1.4`, `1.6`},
		{`false`, `true`},

		{`$x+1`, `10+2`},
		{`$x+1`, `$x+$x`},

		{`$x = $x`, `$x = $y`},
		{`$x = $x`, `$x[0] = $y[0]`},
		{`$x = $x`, `$x->a[0] = $y->a[0]`},
		{`$x = $x`, `$x->a[0] = $x->b[0]`},
		{`$x = $x`, `$x->a[0] = $x->a[1]`},

		{`+$x`, `-1`},
		{`-$x`, `+2`},

		{`$f()`, `f(1)`},
		{`$f()`, `g(2)`},
		{`$f($a1, $a2)`, `f()`},
		{`$f($a1, $a2)`, `f()`},

		{`$x+$x`, `1+2`},
		{`$x+$x`, `2+1`},
		{`$x+$x`, `""+1`},
		{`$x+$x`, `1+""`},

		{`$$$x`, `$x`},
		{`$$$x`, `10`},
		{`$$x = $$x`, `$$$x = $$$y`},
		{`$$x = $$x`, `$$x = $$y`},
		{`$$x = $$x`, `$$x = $x`},
		{`$$x = $$x`, `$x = $x`},

		{`[$x, $x]`, `[1, 2]`},
		{`array($x, $x)`, `array(1, 2)`},

		{`{}`, `{1;}`},
		{`{1;}`, `{}`},
		{`{1;}`, `{1; 2;}`},
		{`{1; 2;}`, `{1; 2; 3;}`},
		{`{1; 2; 3;}`, `{1; 2;}`},

		{`f(${'*'}, 4)`, `f(1, 2, 3)`},

		{`new $t`, `new T()`},
		{`new $t()`, `new T`},

		{`while ($_); {${'*'};}`, `while ($cond) {$blah;}`},
		{`for ($_; $_; $_) {${"*"};}`, `for (;;) {}`},

		{`if ($c) $_; else if ($c) $_;`, `if ($c1) {} else if ($c2) {}`},
		{`if ($c) $_; elseif ($c) $_;`, `if ($c1) {} elseif ($c2) {}`},

		{`list($x, $_, $x) = f()`, `list(,1,2) = f()`},
		{`list($x, $_, $x) = f()`, `list(2,1,) = f()`},

		{`${'x:var'}`, `1`},
		{`${'var'}`, `[10]`},
		{`${'var'}`, `THE_CONST`},
		{`${'x:var'} + $x`, `$x + 1`},
		{`$x + ${'x:var'}`, `1 + $x`},

		{`${"int"}`, `13.5`},
		{`${"float"}`, `3`},
		{`${"str"}`, `5`},
		{`${"num"}`, `$x`},
		{`${"num"}`, `"1"`},

		{`${"expr"}`, `{}`},
		{`${"expr"}`, `{{}}`},

		{`$_ == $_ ? $_ : $_ == $_ ? $_ : $_`, `$a == 1 ? 'one' : ($a == 2 ? 'two' : 'other')`},
		{`$x ? $x : $y`, `1 ?: 2`},
		{`$x ? $y : $z`, `1 ?: 2`},

		{`$x->$_ = $x`, `$this->self = $y`},

		{`$_[0]`, `$v[1]`},

		{`@$_`, `f()`},

		{`die(0)`, `exit(0)`},
		{`exit(0)`, `die(0)`},

		{`$x->$m()`, `$this->m(1)`},
		{`$x->$m(1, 2)`, `$this->m(2, 1)`},
		{`$x->ff(1, 2)`, `$this->f2(1, 2)`},

		{`C::f()`, `C2::f()`},
		{`C::f()`, `C::f2()`},
		{`C::constant`, `C::constant2`},
		{`C::constant`, `C::$prop`},

		{`__FILE__`, `__DIR__`},
		{`[$x, $x]`, `[__FILE__, __DIR__]`},
		{`[$x, $x]`, `[__DIR__, __FILE__]`},

		{`"$x$x"`, `"11"`},
		{`"$x$x"`, `'$x$x'`},

		{`int($x)`, `$v`},
		{`array($x)`, `$v`},
		{`string($x)`, `$v`},
		{`bool($x)`, `$v`},
		{`double($x)`, `$v`},
		{`object($x)`, `$v`},

		{`\A\B`, `\A\A`},
		{`\A\B`, `\B\B`},

		{`function() { return $x; }`, `function() {}`},
		{`function($x) {}`, `function() {}`},
		{`function($x) use($v) {}`, `function($arg1) use($a, $b) {}`},
		{`function() { ${"*"}; return 1; }`, `function() {}`},
		{`function() { ${"*"}; return 1; }`, `function($x) { return 1; }`},
		{`function() { ${"*"}; return 1; }`, `static function() { f(); f(); return 1; }`},

		{`!($x instanceof $y)`, `1`},
		{`$x instanceof T1`, `$v instanceof T2`},
		{`$x instanceof T`, `$x instanceof $y`},

		{`static $x = 10`, `static $vvv = 11`},
		{`global $x, $y`, `global $a`},
		{`break ${"expr"}`, `break`},
		{`continue ${"expr"}`, `continue`},
		{`unset($x)`, `unset($v, $y)`},
		{`print(1)`, `print(2)`},
		{`echo 1, 2`, `echo 1`},
		{`throw new $E()`, `throw new Exception(1)`},

		{`${"const"}`, `$v`},
		{`${"const"}`, `1`},
		{`${"const"}`, `"1"`},
		{`${"const"}`, `$x->y`},
		{`$_ = ${"const"}`, `$x = MyClass::$var`},

		{`${"func"}`, `1`},
		{`${"func"}`, `$x`},
		{`${"func"}`, `f()`},

		{`(1)`, `1`},
		{`((1))`, `(1)`},
		{`f((1))`, `f(1)`},
		{`($foo)()`, `$foo()`},
		{`($foo)->x`, `$foo->x`},
		{`(($foo))()`, `($foo)()`},
		{`(($foo))->x`, `($foo)->x`},
	})
}

func BenchmarkMatch(b *testing.B) {
	runBench := func(name, pattern string, input string) {
		b.Run(name, func(b *testing.B) {
			matcher := mustCompile(b, pattern)
			root := mustParse(b, input)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				matcher.Match(root)
			}
		})
	}

	const (
		functionCall = `f(1, 2, 'abc', $x, [FOO => BAR])`
	)

	benchmarks := []struct {
		name    string
		pattern string
		input   string
	}{
		{"positive/call_parent_ctor", `parent::__construct(${"*"})`, `parent::__construct(1, 2)`},
		{"negative/call_parent_ctor", `parent::__construct(${"*"})`, `foo([$x => $y])`},

		{"negative/const-tail", `[${"*"}, 1, 1]`, `[0,0,0,0,0,0,0,0,0]`},

		{"positive/call*", `$_(${"*"})`, functionCall},
		{"positive/call_*", `$_($_, ${"*"})`, functionCall},
		{"positive/call*_", `$_(${"*"}, $_)`, functionCall},
		{"positive/call*_*", `$_(${"*"}, $_, ${"*"})`, functionCall},
		{"negative/call_", `$_($_)`, functionCall},

		{"positive/with-1-named", `$x + 1 * $x`, `$a[0] + 1 * $a[0]`},
		{"negative/with-1-named", `$x + 1 * $x`, `$a[0] + 1 * $a[1]`},
		{"positive/with-5-named", `$x1 + $x2 + $x3 + $x4 + $x5`, `1 + 2 + 3 + 4 + 5`},
	}

	for _, bench := range benchmarks {
		runBench(bench.name, bench.pattern, bench.input)
	}
}
