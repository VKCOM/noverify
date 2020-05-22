package phpgrep

import (
	"fmt"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
)

func matchInText(m *matcher, code []byte) bool {
	root, _, err := parsePHP7(code)
	if err != nil {
		return false
	}
	if x, ok := root.(*stmt.Expression); ok {
		root = x.Expr
	}
	return m.match(root)
}

func findInText(m *matcher, code []byte, callback func(*MatchData) bool) {
	root, _, err := parsePHP7(code)
	if err != nil {
		return
	}
	m.findAST(root, callback)
}

type matcherTest struct {
	pattern string
	input   string
}

func mustCompile(t testing.TB, c *Compiler, code string) *Matcher {
	matcher, err := c.Compile([]byte(code))
	if err != nil {
		t.Fatalf("pattern compilation error:\ntext: %q\nerr: %v", code, err)
	}
	return matcher
}

func TestFind(t *testing.T) {
	runFindTest := func(t *testing.T, pattern, code string, wantMatches []string) {
		var c Compiler
		matcher := mustCompile(t, &c, pattern)
		var haveMatches []string
		findInText(&matcher.m, []byte(code), func(m *MatchData) bool {
			pos := m.Node.GetPosition()
			posFrom := pos.StartPos
			posTo := pos.EndPos
			haveMatches = append(haveMatches, string(code[posFrom:posTo]))
			return true
		})
		if len(haveMatches) != len(wantMatches) {
			t.Errorf("matches count mismatch:\nhave: %d\nwant: %d",
				len(haveMatches), len(wantMatches))
			t.Log("have:")
			for _, have := range haveMatches {
				t.Log(have)
			}
			t.Log("want:")
			for _, want := range wantMatches {
				t.Log(want)
			}
			return
		}
		for i, have := range haveMatches {
			want := wantMatches[i]
			if have != want {
				t.Errorf("match mismatch:\nhave: %q\nwant: %q", have, want)
			}
		}
	}

	runFindTest(t, `$x+1`, `<?php $x+1;`, []string{`$x+1`})

	runFindTest(t, `$x = $x`, `<?php
            $x = $x; $z1 = 10; $y = $y; $z2 = 20; $x = $y;
        `, []string{
		`$x = $x`,
		`$y = $y`,
	})

	// TODO: uncomment when parentheses are handled correctly.
	// runFindTest(t, `($x)`, `<?php
	//     $x + $y; ($x1 + $y1); (($x2 + $y2));
	// `, []string{
	// 	`($x1 + $y1)`,
	// 	`(($x2 + $y2))`,
	// 	`($x2 + $y2)`,
	// })
}

func runMatchTest(t *testing.T, want bool, tests []*matcherTest) {
	var c Compiler
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%v", i, want), func(t *testing.T) {
			matcher := mustCompile(t, &c, test.pattern)
			have := matchInText(&matcher.m, []byte(test.input))
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

		{`1`, `(1)`},
		{`(1)`, `(1)`},
		{`((1))`, `((1))`},
		{`f(1)`, `f(1)`},
		{`f((1))`, `f((1))`},
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
		{`for ($_; $_; $_) {${"*"};}`, `for ($i = 0; $i < 10; $i++) { echo $; }`},

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

		{`$_ == $_ ? $_ : $_ == $_ ? $_ : $_`, `$a == 1 ? ('one' : $a == 2 ? ('two' : 'other'))`},
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
		{`function($x) use($v) {}`, `function($arg1) use() {}`},
		{`function() { ${"*"}; return 1; }`, `function() {}`},
		{`function() { ${"*"}; return 1; }`, `function($x) { return 1; }`},
		{`function() { ${"*"}; return 1; }`, `static function() { f(); f(); return 1; }`},

		{`!($x instanceof $y)`, `1`},
		{`$x instanceof T1`, `$v instanceof T2`},
		{`$x instanceof T`, `$x instance of $y`},

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

		// TODO: uncomment when parentheses are handled correctly.
		// {`(1)`, `1`},
		// {`((1))`, `(1)`},
		// {`f((1))`, `f(1)`},
	})
}

func BenchmarkFind(b *testing.B) {
	var c Compiler

	runBench := func(name, pattern string, input []byte) {
		b.Run(name, func(b *testing.B) {
			matcher := mustCompile(b, &c, pattern)
			root, _, err := parsePHP7(input)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				matcher.m.match(root)
			}
		})
	}

	// f(), f($x), ..., f($x{i})
	lotsOfCalls := []byte("<?php\n")
	for i := 0; i < 50; i++ {
		call := "f(" + strings.Repeat("$x,", i) + ");\n"
		lotsOfCalls = append(lotsOfCalls, []byte(call)...)
	}

	benchmarks := []struct {
		name    string
		pattern string
		input   []byte
	}{
		// Benchmarking list matching.
		{"positive/call*", `$_(${"*"})`, lotsOfCalls},
		{"positive/call_*", `$_($_, ${"*"})`, lotsOfCalls},
		{"positive/call*_", `$_(${"*"}, $_)`, lotsOfCalls},
		{"positive/call*_*", `$_(${"*"}, $_, ${"*"})`, lotsOfCalls},
		{"negative/call_", `$_($_)`, lotsOfCalls},

		// Benchmarking named variables.
		{"positive/with-1-named", `$x`, benchmarkInput},
		{"positive/with-5-named", `$x1 + $x2 + $x3 + $x4 + $x5`, benchmarkInput},
		{"negative/with-0-named", `1 + 7 - 103`, benchmarkInput},
	}

	for _, bench := range benchmarks {
		runBench(bench.name, bench.pattern, bench.input)
	}
}

var benchmarkInput = []byte(`<?php

use N\{ClassName,
  AnotherClassName,
  OneMoreClassName};

namespace A {
  function foo() {
    return 0;
  }

  function bar($x,
    $y, int $z = 1) {
    $x = 0;
// $x = 1
    do {
      $y += 1;
    } while ($y < 10);
    if (true)
      $x = 10;
    elseif ($y < 10)
      $x = 5;
    elseif (true)
      $x = 5;
    for ($i = 0; $i < 10; $i++)
      $yy = $x > 2 ? 1 : 2;
    while (true)
      $x = 0;
    do {
      $x += 1;
    } while (true);
    foreach (["a" => 0, "b" => 1,
              "c" => 2] as $e1) {
      echo $e1;
    }
    $count = 10;
    $x     = ["x", "y",
      [1 => "abc",
       2 => "def", 3 => "ghi"]];
    $zz    = [0.1, 0.2,
      0.3, 0.4];
    $x     = [
      0   => "zero",
      123 => "one two three",
      25  => "two five",
    ];
    bar(0, bar(1,
      "b"));
  }

  abstract class Foo extends FooBaseClass implements Bar1, Bar2, Bar3 {

    var $numbers = ["one", "two", "three", "four", "five", "six"];
    var $v = 0;
    public $path = "root";

    const FIRST  = 'first';
    const SECOND = 0;
    const Z      = -1;

    function bar($v,
      $w = "a") {
      $y      = $w;
      $result = foo("arg1",
        "arg2",
        10);
      switch ($v) {
        case 0:
          return 1;
        case 1:
          echo '1';
          break;
        case 2:
          break;
        default:
          $result = 10;
      }
      return $result;
    }

    public static function fOne($argA, $argB, $argC, $argD, $argE, $argF, $argG, $argH) {
      $x = $argA + $argB + $argC + $argD + $argE + $argF + $argG + $argH;
      list($field1, $field2, $field3, $filed4, $field5, $field6) = explode(",", $x);
      fTwo($argA, $argB, $argC, fThree($argD, $argE, $argF, $argG, $argH));
      $z      = $argA == "Some string" ? "yes" : "no";
      $colors = ["red", "green", "blue", "black", "white", "gray"];
      $count  = count($colors);
      for ($i = 0; $i < $count; $i++) {
        $colorString = $colors[$i];
      }
    }

    function fTwo($strA, $strB, $strC, $strD) {
      if ($strA == "one" || $strB == "two" || $strC == "three") {
        return $strA + $strB + $strC;
      }
      $x = $foo->one("a", "b")->two("c", "d", "e")->three("fg")->four();
      $y = a()->b()->c();
      return $strD;
    }

    function fThree($strA, $strB, $strC, $strD, $strE) {
      try {
      } catch (Exception $e) {
        foo();
      } finally {
        // do something
      }
      return $strA + $strB + $strC + $strD + $strE;
    }

    protected abstract function fFour();

  }
}

function f() {}

$_ = f(1 + 2 + 3 + 4 + 5);
$_ = f(f() + f() + f() + f() + f());
`)
