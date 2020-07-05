package quickfix

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestQuickfixNested(t *testing.T) {
	// Test that we don't end up with broken code if one replacement
	// occurs inside another one.

	// Imagine that we had a `$f($x)` -> `$x` rule.
	// There are at least 3 possible outcomes:
	// 1. g(10); -- outer call replaced
	// 2. 10;    -- both calls replaced
	// 3. f(10); -- inner call replaced
	//
	// We don't care which one will be chosen as long as the
	// result is consistent and doesn't result in some garbage.
	input := `f(g(10));`
	fix1 := TextEdit{
		StartPos:    0,
		EndPos:      8,
		Replacement: `g(10)`,
	}
	fix2 := TextEdit{
		StartPos:    2,
		EndPos:      7,
		Replacement: `10`,
	}

	runTest := func(want string, fixes []TextEdit) {
		var buf bytes.Buffer
		writeFixes(&buf, []byte(input), fixes)
		if buf.String() != want {
			t.Errorf("%q %v:\nhave: `%s`\nwant: `%s`", input, fixes, buf.String(), want)
		}
	}

	runTest(`g(10);`, []TextEdit{fix1})
	runTest(`g(10);`, []TextEdit{fix1, fix2})
	runTest(`f(10);`, []TextEdit{fix2})
	runTest(`f(10);`, []TextEdit{fix2, fix1})

	// Also test that we can tolerate duplicated rewrites.
	runTest(`g(10);`, []TextEdit{fix1, fix1, fix2, fix1})
	runTest(`f(10);`, []TextEdit{fix2, fix1, fix2, fix2})
}

func TestQuickfix(t *testing.T) {
	tests := map[string][]string{
		// No replacements.
		``:      {},
		` abc `: {},

		// Empty replacements.
		`$a$`:       {``},
		`$a$$b$`:    {``, ``},
		`1$a$2$b$3`: {``, ``},

		// Replacement and original of the same length.
		`  $a$$b$ `:  {`x`, `y`},
		` $a$ $b$  `: {`x`, `y`},

		// Replacement longer than original.
		`$aa$ $bb$ $cc$`:     {`xxx`, `yyy`, `zzz`},
		`  $aa$  $bb$ $cc$`:  {`xxx`, `yyy`, `zzz`},
		` $aa$ $bb$  $cc$  `: {`xxx`, `yyy`, `zzz`},

		// Original longer than replacement.
		`$aa$ $bbb$ $cc$`:    {`x`, `yy`, `z`},
		` $aa$  $bbb$ $cc$ `: {`x`, `yy`, `z`},
		`$aa$ $bbb$  $cc$`:   {`x`, `yy`, `z`},

		// Mixed lengths.
		` $a$    $bb$  $ccc$ `:    {`xx`, `y`, `zz`},
		` $aaa$    $bb$   $cc$  `: {`xxx`, `y`, `z`},
		` $aaa$ $bbb$ $cccc$ $d$`: {`x`, `yyyyyyy`, `zzz`, ``},
		` $d$$aaa$  $bbb$$cccc$`:  {`xxxxxx`, `y`, ``, `zz`},
		`$aaaaaaaaaaaaa$$b$`:      {`x`, `yyyy`},
		`$a$$bbbbb$`:              {`yyyy`, `x`},

		// Multiline.
		`<?php
foo = $x ?
       x :
       y$;`: {`x ?: y`},
	}

	type testCase struct {
		input string
		want  string
		fixes []TextEdit
	}
	varRegex := regexp.MustCompile(`\$[^$]+?\$`)
	createTestCase := func(s string, replacements []string) testCase {
		i := -1
		withReplacements := varRegex.ReplaceAllStringFunc(s, func(x string) string {
			i++
			return replacements[i]
		})

		fixes := make([]TextEdit, len(replacements))

		for i, m := range varRegex.FindAllStringIndex(s, -1) {
			fixup := i * len("$") * 2
			begin := m[0]
			end := m[1]
			fixes[i] = TextEdit{
				StartPos:    begin - fixup,
				EndPos:      (end - len("$")*2) - fixup,
				Replacement: replacements[i],
			}
		}

		return testCase{
			input: strings.ReplaceAll(s, "$", ""),
			want:  withReplacements,
			fixes: fixes,
		}
	}

	for s, replacements := range tests {
		test := createTestCase(s, replacements)
		var buf bytes.Buffer
		writeFixes(&buf, []byte(test.input), test.fixes)
		if buf.String() != test.want {
			t.Errorf("%q %q:\nhave: `%s`\nwant: `%s`", test.input, replacements, buf.String(), test.want)
		}
	}
}
