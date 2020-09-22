package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestREVetParse(t *testing.T) {
	// Don't add too much test cases here as we don't want to
	// be dependent on the regexp parser library error messages
	// too much. We prove that we react on parse errors and that's enough.
	test := linttest.NewSuite(t)
	test.LoadedStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function parseErrors($s) {
  preg_match('', $s);
  preg_match('aba', $s);
  preg_match('<foo([abc)+', $s);
  preg_match('#foo([abc)+#', $s);
}

`)
	test.Expect = []string{
		`parse error: empty pattern: can't find delimiters`,
		`parse error: 'a' is not a valid delimiter`,
		`parse error: can't find '>' ending delimiter`,
		`parse error: unterminated '['`,
	}
	test.RunAndMatch()
}

func TestREVet(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadedStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function charRange($s) {
	preg_match('~[$-%]~', $s);
	preg_match('~[ -!]~', $s);
	preg_match('~[❤-❥]~', $s);
}

function altDups($s) {
	preg_match('~x|x~', $s);
	preg_match('~([a-z]|[a-z]|[0-9])~', $s);
	preg_match('~(foo)+|(bar)+|(foo)+~');
	preg_match_all('/(?:\*[\dAGC]+)+|(?:\/[\dAGC]+)+|(?:\+[\d]+)+|(?:\-[\d]+)+/', $s, $matches);
}

function charClassDuplicates($s) {
	preg_match('~x[aba]y~', $s);

	preg_match('~[\141a]~', $s);
	preg_match('~[a\x61]~', $s);
	preg_match('~[^a\x{61}]~', $s);

	preg_match('~[a-cb]~', $s);
	preg_match('~[^a-ba-b]~', $s);

	preg_match('~[\x{61}-\x{63}c]~', $s);
	preg_match('~[\x61-\x63c]~', $s);
	preg_match('~[\141-\143c]~', $s);

	preg_match('~[\d5]~', $s);
	preg_match('~[5-6\d]~', $s);
	preg_match('~[\w_]~', $s);
	preg_match('~[\w%a-d]~', $s);
	preg_match('~[\Dg]~', $s);
	preg_match('~[\D❤5]~', $s);
	preg_match('~[\s\t]~', $s);
	preg_match('~[\n\s]~', $s);

	preg_match('~[1-52-34]~', $s);
	preg_match('~[42-31-5]~', $s);

	preg_match('~[\w\W❤]~', $s);
}

function repeatedQuantifier($s) {
	preg_match('~(a+)+~', $s);
	preg_match('~(?:[ab]*)+~', $s);
	preg_match('~((ab)+)*~', $s);
}

function redundantFlags($s) {
	preg_match('~(?m)(?m)~', $s);
	preg_match('~(?ims:(?i:foo))(?im:bar)~', $s);
	preg_match('~(?i)(?ims:flags1)(?m:flags2)~', $s);
	preg_match('~((?m)(?m:a|b(?s:foo))(?i)x)~', $s);

	preg_match('~(?i)foo~i');
}

function redundantFlagClear($s) {
	preg_match('/(?-i)x/', $s);
	preg_match('/(?i:foo)(?-i)bar/', $s);
	preg_match('/(?i:(?m:fo(?-i)o))(?-mi)bar/', $s);
	preg_match('/(?i-ii)/', $s);
	preg_match('/(?:(?i)foo)(?-i)/', $s);
	preg_match('/((?i)(?-i))(?-i)/', $s);
	preg_match('/(?:(?i)(?-i))(?-i)/', $s);
	preg_match('/(?m-s)(?:tags)/(\S+)$/', $s);
}

function suspiciousAltAnchor($s) {
	preg_match('~^foo|bar|baz~', $s);
	preg_match('~(?:^a|bc)c~', $s);
	preg_match('~foo|bar|baz$~', $s);
	preg_match('~c(?:a|bc$)~', $s);
}

function danglingCaret() {
	preg_match('~a^~', $s);
	preg_match('~a^b~', $s);
	preg_match('~^^foo~', $s);
	preg_match('~foo?|bar^~', $s);
	preg_match('~(?i:a)^foo~', $s);
	preg_match('~(?i)^(?:foo|bar|^baz)~', $s);
	preg_match('~(?i)^(?m)foobar^baz~', $s);
	preg_match('~(?i:foo|((?:f|b|(foo|^bar)^)))~', $s);
	preg_match('~(?i)(?m)\n^foo|bar|baz~', $s);
}
`)
	test.Expect = []string{
		`suspicious char range '$-%' in [$-%]`,
		`suspicious char range ' -!' in [ -!]`,
		`suspicious char range '❤-❥' in [❤-❥]`,

		`'x' is duplicated in x|x`,
		`'[a-z]' is duplicated in [a-z]|[a-z]|[0-9]`,
		`'(foo)+' is duplicated in (foo)+|(bar)+|(foo)+`,

		`'a' is duplicated in [aba]`,

		`'\141' intersects with 'a' in [\141a]`,
		`'a' intersects with '\x61' in [a\x61]`,
		`'a' intersects with '\x{61}' in [^a\x{61}]`,
		`'a-c' intersects with 'b' in [a-cb]`,
		`'a-b' is duplicated in [^a-ba-b]`,
		`'\x{61}-\x{63}' intersects with 'c' in [\x{61}-\x{63}c]`,
		`'\x61-\x63' intersects with 'c' in [\x61-\x63c]`,
		`'\141-\143' intersects with 'c' in [\141-\143c]`,
		`'\d' intersects with '5' in [\d5]`,
		`'\d' intersects with '5-6' in [5-6\d]`,
		`'\w' intersects with '_' in [\w_]`,
		`'\w' intersects with 'a-d' in [\w%a-d]`,
		`'\D' intersects with 'g' in [\Dg]`,
		`'\D' intersects with '❤' in [\D❤5]`,
		`'\s' intersects with '\t' in [\s\t]`,
		`'\s' intersects with '\n' in [\n\s]`,
		`'1-5' intersects with '2-3' in [1-52-34]`,
		`'1-5' intersects with '2-3' in [42-31-5]`,
		`'\W' intersects with '❤' in [\w\W❤]`,

		`repeated greedy quantifier in (a+)+`,
		`repeated greedy quantifier in (?:[ab]*)+`,
		`repeated greedy quantifier in ((ab)+)*`,

		`redundant flag m in (?m)`,
		`redundant flag i in (?i:foo)`,
		`redundant flag i in (?ims:flags1)`,
		`redundant flag m in (?m:a|b(?s:foo))`,
		`redundant flag i in (?i)`,

		`clearing unset flag i in (?-i)`,
		`clearing unset flag i in (?-i)`,
		`clearing unset flag m in (?-mi)`,
		`clearing unset flag i in (?-mi)`,
		`clearing unset flag i in (?i-ii)`,
		`clearing unset flag i in (?-i)`,
		`clearing unset flag i in (?-i)`,
		`clearing unset flag i in (?-i)`,
		`clearing unset flag s in (?m-s)`,

		`^ applied only to 'foo' in ^foo|bar|baz`,
		`^ applied only to 'a' in ^a|bc`,
		`$ applied only to 'baz' in foo|bar|baz$`,
		`$ applied only to 'bc' in a|bc$`,

		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
		`dangling or redundant ^, maybe \^ is intended?`,
	}
	linttest.RunFilterMatch(test, `regexpVet`)
}

func TestREVet_2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadedStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function f($s) {
  preg_match('/(?m)^([\s\t]*)([\*\-\+]|\d\.)\s+/', $s);
  preg_match('~^(www.|https://|http://)*[A-Za-z0-9._%+\-]+\.[com|org|edu|net]{3}$~', $s);
  preg_match('/^[\w\d]{3,30}$/', $s);
}
`)
	test.Expect = []string{
		`'\s' intersects with '\t' in [\s\t]`,
		`'e' is duplicated in [com|org|edu|net]`,
		`'\w' intersects with '\d' in [\w\d]`,
	}
	test.RunAndMatch()
}

func TestREVet_3(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadedStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function f($s) {
  preg_match('~^each\s+(\$[\w]*)(?:\s*,\s*(\$[\w0-9\-_]*))?\s+in\s+(.+)$~', $s);
  preg_match('~(?i)(?:for=)([^(;|,| )]+)~', $s);
  preg_match('~[^\w\d\-_ ]~', $s);
  preg_match('~^[\w+-\.]+$~', $s);
  preg_match('~^([\w\.\-_+]+)$~', $s);
  preg_match('~Usage: docker \\\\[OPTIONS\\\\] COMMAND~', $s);
  preg_match('~(ok|FAIL)\s+(.+)[\s]+(\d+\.\d+(s| seconds))([\s\t]+coverage:\s+(\d+\.\d+)\% of statements)?~', $s);
  preg_match('~UPSTREAM: (revert: [a-f0-9]{7,}: )?(([\w\.-]+\/[\w-\.-]+)?: )?(\d+:|<carry>:|<drop>:)~', $s);
  preg_match('~\A\w{1}:[/\/]~', $s);
}
`)
	test.Expect = []string{
		`'\w' intersects with '0-9' in [\w0-9\-_]`,
		`'|' is duplicated in [^(;|,| )]`,
		`'\w' intersects with '\d' in [^\w\d\-_ ]`,
		`suspicious char range '+-\.' in [\w+-\.]`,
		`'\w' intersects with '_' in [\w\.\-_+]`,
		`'O' is duplicated in [OPTIONS\\]`,
		`'\s' intersects with '\t' in [\s\t]`,
		`'-' is duplicated in [\w-\.-]`,
		`'/' intersects with '\/' in [/\/]`,
	}
	linttest.RunFilterMatch(test, `regexpVet`)
}

func TestREVet_4(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadedStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function goodFlags($s) {
	preg_match('~(?:(?i)foo)(?i)bar~', $s);
	preg_match('~(?m)(?i)(?-m)(?-i)(?m)(?i)~', $s);
	preg_match('~(?ms:(?i:foo))(?im:bar)~', $s);
	preg_match('~(?i)(?ms:flags1)(?m:flags2)~', $s);
	preg_match('~((?m)(?i:a|b(?s:foo))(?i)x)~', $s);
	preg_match('~(?i)yy(?-i)x~', $s);
	preg_match('~(?i-i)~', $s);
	preg_match('~(?i:foo)(?i)bar~', $s);
	preg_match('~(?i:(?m:fo(?-i)o))(?mi)x(?-mi)bar~', $s);
	preg_match('~(?i-i)(?i)~', $s);
	preg_match('~(?:(?i)foo)(?i)x(?-i)~', $s);

	preg_match('~(?-i)foo~i', $s);
}

function goodAnchors($s) {
	preg_match('~^~', $s);
	preg_match('~^foo~', $s);
	preg_match('~^foo?|bar~', $s);
	preg_match('~^foo|^bar~', $s);
	preg_match('~(^a|^b)~', $s);
	preg_match('~(?i)^foo~', $s);
	preg_match('~(?i)((?m)a|^foo)b~', $s);
	preg_match('~(?i)(?m)\bfoo|bar|^baz~', $s);
	preg_match('~(?i)^(?m)foo|bar|baz~', $s);
	preg_match('~(?i:foo|((?:f|^b|(foo|^bar))))~', $s);
	preg_match('~(?i)^(?m)foo|bar|^baz~', $s);
	preg_match('~(?i)(?:)(^| )\S+~', $s);
}
`)
	test.RunAndMatch()
}
