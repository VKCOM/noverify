package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestRESimplifyNamedCaptureForms(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function f($s) {
  preg_match('~(?P<abc>[0-9])~', $s);
  preg_match('~(?<abc>[0-9])~', $s);
  preg_match("~(?'abc'[0-9])~", $s); // Ignore due to the FormNamedCaptureQuote
}
`)
	test.Expect = []string{
		`May re-write '~(?P<abc>[0-9])~' as '~(?P<abc>\d)~'`,
		`May re-write '~(?<abc>[0-9])~' as '~(?<abc>\d)~'`,
	}
	test.RunAndMatch()
}

func TestRESimplifyMixed(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function f($s) {
  preg_match('/x(?:a|b|c){0,}/', $s);
  preg_match_all('/^write([-]?\d\d*)?$/i', $s);
  preg_replace('/[a-a]/', 'v', $s);
  preg_split('/[a-a]*?/', $s);
}
`)
	test.Expect = []string{
		`May re-write '/x(?:a|b|c){0,}/' as '/x[abc]*/'`,
		`May re-write '/^write([-]?\d\d*)?$/i' as '/^write(-?\d+)?$/i'`,
		`May re-write '/[a-a]/' as '/a/'`,
		`May re-write '/[a-a]*?/' as '/a*?/'`,
	}
	test.RunAndMatch()
}

func TestRESimplify(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
// (?:x) -> x
function ungroup($s) {
  preg_match('/(?:x)/', $s);
  preg_match('/(?:[abc])/', $s);
}

// xx* -> x+
function merge($s) {
  preg_match('/x[abcd][abcd]*y/', $s);
  preg_match('/axx*y/', $s);
}

// Replaces duplicated expression x with x{n}, when n is a number of duplications.
function repeat($s) {
  preg_match('/a    b/', $s);
  preg_match('/(?:foo|bar)(?:foo|bar)/', $s);
  preg_match('/.....x/', $s);
}

// Replaces the char class with equivalent expression.
function replaceCharClass($s) {
  preg_match('/[^\D]+/', $s);
  preg_match('/[^[:^word:]]+/', $s);
}

// [x] -> x
function unwrapCharClass($s) {
  preg_match('/foo[\d]{3}/', $s);
}

// \# -> #
function unescape($s) {
  preg_match('/\#/', $s);
  preg_match('/[x\#]/', $s);
  preg_match('/\>/', $s);
  preg_match('/\</', $s);
}

// a|b|c -> [abc]
function charAlt($s) {
  preg_match('/(a|b|c|d)/', $s);
  preg_match('/a|b/', $s);
}

// [a-a] -> [a]
// [a-b] -> [ab]
// [a-c] -> [abc]
function unrangeCharClass($s) {
  preg_match('/[xa-a]/', $s);
  preg_match('/[xa-b]/', $s);
  preg_match('/[1-23]/', $s);
}

// x{0,1} -> x?
// x{1,}  -> x+
// x{0,}  -> x*
// x{1}   -> x
// x{0}   ->
function unrepeat($s) {
  preg_match('/x{0}foo/', $s);
  preg_match('/x{1}/', $s);
  preg_match('/[0-9]{1,}/', $s);
  preg_match('/[0-9]{0,1}/', $s);
  preg_match('/x{0,}/', $s);
}
`)
	test.Expect = []string{
		`May re-write '/(?:x)/' as '/x/'`,
		`May re-write '/(?:[abc])/' as '/[abc]/'`,
		`May re-write '/x[abcd][abcd]*y/' as '/x[abcd]+y/'`,
		`May re-write '/axx*y/' as '/ax+y/'`,
		`May re-write '/a    b/' as '/a {4}b/'`,
		`May re-write '/(?:foo|bar)(?:foo|bar)/' as '/(?:foo|bar){2}/'`,
		`May re-write '/.....x/' as '/.{5}x/'`,
		`May re-write '/[^\D]+/' as '/\d+/'`,
		`May re-write '/[^[:^word:]]+/' as '/\w+/'`,
		`May re-write '/foo[\d]{3}/' as '/foo\d{3}/'`,
		`May re-write '/\#/' as '/#/'`,
		`May re-write '/[x\#]/' as '/[x#]/'`,
		`May re-write '/\>/' as '/>/'`,
		`May re-write '/\</' as '/</'`,
		`May re-write '/(a|b|c|d)/' as '/([abcd])/'`,
		`May re-write '/a|b/' as '/[ab]/'`,
		`May re-write '/[xa-a]/' as '/[xa]/'`,
		`May re-write '/[xa-b]/' as '/[xab]/'`,
		`May re-write '/[1-23]/' as '/[123]/'`,
		`May re-write '/x{0}foo/' as '/foo/'`,
		`May re-write '/x{1}/' as '/x/'`,
		`May re-write '/[0-9]{1,}/' as '/\d+/'`,
		`May re-write '/[0-9]{0,1}/' as '/\d?/'`,
		`May re-write '/x{0,}/' as '/x*/'`,
	}
	test.RunAndMatch()
}

func TestRESimplifyChangeDelim(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function f($s) {
  preg_match('/^http:\/\//', $s);
  preg_match('/^http:\/\/~/', $s);
  preg_match('/^http:\/\/~@/', $s);
  preg_match('@mail\@gmail\.ru@', $s); // Only 1 escaped delim, but it's not '/'
}
`)
	test.Expect = []string{
		`May re-write '/^http:\/\//' as '~^http://~'`,
		`May re-write '/^http:\/\/~/' as '@^http://~@'`,
		`May re-write '/^http:\/\/~@/' as '#^http://~@#'`,
		`May re-write '@mail\@gmail\.ru@' as '/mail@gmail\.ru/'`,
	}
	test.RunAndMatch()
}

func TestRENegativeTests(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/pcre/pcre.php`}
	test.AddFile(`<?php
function f($s) {
  // Should not suggest unescaping the delimiter.
  preg_match('/\//', $s);
  preg_match('~\~/home~', $s);
}

function _1($s) {
  preg_match('~[а-яё]~', $s);
  preg_match('/\s*\{weight=(\d+)\}\s*/', $s);
  preg_match('/\{inherits=(\d+)\}/', $s);
  preg_match('/[<>]+/', $s);
  preg_match('/[.?,!;:@#$%^&*()]+/', $s);
  preg_match('/^\d{1,7}$/', $s);
  preg_match('/^(")/', $s);
  preg_match('/( ")/', $s);
  preg_match('/unifi_devices{site="Default"} 1/', $s);
  preg_match('/unifi_devices_adopted{site="Default"} 1/', $s);
  preg_match('/unifi_devices_unadopted{site="Default"} 0/', $s);
  preg_match('/^((\").*(\"))/', $s);
  preg_match('~/api/internal/login~', $s);
  preg_match('~/api/organizations\?limit=100~', $s);
  preg_match('/[áàảãạấầẩẫậâăắằẳẵặ]/', $s);
  preg_match('/\.(com|com\.\w{2})$/', $s);
  preg_match('/\.(gov|gov\.\w{2})$/', $s);
  preg_match('/[\xC0-\xC6]/', $s);
  preg_match('/[\xE0-\xE6]/', $s);
  preg_match('/[\xC8-\xCB]/', $s);
  preg_match('/[\xE8-\xEB]/', $s);
}

function _2($s) {
  preg_match('/\bsample\b/', $s);
  preg_match('/\b(720p|1080p|hdtv|x264|dts|bluray)\b.*/', $s);
  preg_match('/(?i)\bcode\b/', $s);
  preg_match('/\p{Cyrillic}/', $s);
  preg_match('/--(?P<var_name>[\\w-]+?):\\s+?(?P<var_val>.+?);/', $s);
  preg_match('/^.+_rsa$/', $s);
  preg_match('/^.+_dsa.*$/', $s);
  preg_match('/^.+_ed25519$/', $s);
  preg_match('/^.+_ecdsa$/', $s);
  preg_match('/a[^rst]c/', $s);
  preg_match('/<li class="neirong2">信息标题.*<b>([^<]+)</', $s);
  preg_match('/<li class="neirong2">信息来源[^>]+">([^<]+)</', $s);
  preg_match('/\s*:\s*/', $s);
  preg_match('/(?m)^[ \t]*(#+)\s+/', $s);
  preg_match('/(?m)^h([0-6])\.(.*)$/', $s);
  preg_match('/<==\sPlayerInventory\.GetPlayerCardsV3\(\d*\)/', $s);
  preg_match('/^ *(#{1,6}) *([^\n]+?) *#* *(?:\n|$)/', $s);
  preg_match('/^[^\n]+/', $s);
  preg_match('/^\n+/', $s);
  preg_match('/^((?: {4}|\t)[^\n]+\n*)+/', $s);
  preg_match('/^(int)/', $s);
  preg_match('/^(\+)/', $s);
  preg_match('/[^a-zA-Z0-9]/', $s);
  preg_match('/^absolute(\.|-).*/', $s);
  preg_match('/^metric .*/', $s);
}

function _3($s) {
  preg_match('/\n={2,}/', $s);
  preg_match('/~~/', $s);
  preg_match('/[aeiou][^aeiou]/', $s);
  preg_match('/.*sses$/', $s);
  preg_match('/CIFS Session: (?P<sessions>\d+)/', $s);
  preg_match('/Share \(unique mount targets\): (?P<shares>\d+)/', $s);
  preg_match('~SMB Request/Response Buffer: (?P<smbBuffer>\d+) Pool size: (?P<smbPoolSize>\d+)~', $s);
  preg_match('~(?is)<a.+?</a>~', $s);
  preg_match('/(?is)\(.*?\)/', $s);
  preg_match('/<a href="#.+?\|/', $s);
  preg_match('/^(?:mister )(.*)$/', $s);
  preg_match('/(?ims)<!DOCTYPE.*?>/', $s);
  preg_match('/(?ims)<!--.*?-->/', $s);
  preg_match('~(?ims)<script.*?>.*?</script>~', $s);
  preg_match('/[a-z]+/', $s);
  preg_match('/[^a-z]+/', $s);
  preg_match('/Acl/', $s);
  preg_match('/Adm([^i]|$)/', $s);
  preg_match('/Aes/', $s);
  preg_match('/(es|ed|ing)$/', $s);
  preg_match('/[^[:alpha:]]/', $s);
  preg_match('/ESCAPE_([[:alnum:]]+)/', $s);
  preg_match('~[.]|,|\s/~', $s);
  preg_match('~(?imsU)\[quote(?:=[^\]]+)?\](.+)\[/quote\]~', $s);
  preg_match('/(?imU)^(.*)$/', $s);
  preg_match('/开 本：(\d+)开/', $s);
  preg_match('/^((inteiro)|(real)|(caractere)|(lógico))(\s*):/', $s);
  preg_match('/^\d*H[A-Z0-9]*$/', $s);
  preg_match('/^LI[A-Z0-9]*$/', $s);
  preg_match('/^C[A-Z0-9]*$/', $s);
  preg_match('/[a-zA-Z0-9]+@[a-zA-Z0-9.]+\.[a-zA-Z0-9]+/', $s);
  preg_match('/}\n+$/', $s);
}

function _4($s) {
  preg_match('/^\s*events\s*{/', $s);
  preg_match('/^\s*http\s*{/', $s);
  preg_match('/(?i)windows nt/', $s);
  preg_match('/(?i)windows phone/', $s);
  preg_match('~^.* ENGINE=.*/\)~', $s);
  preg_match('/[a-zA-Z0-9]+@[a-zA-Z-0-9.]+\.[a-zA-Z0-9]+/', $s);
  preg_match('/he|ll|o+/', $s);
  preg_match('/^(\S*) (\S*) (\d*) (\S*) IP(\d) (\S*)/', $s);
  preg_match('/\s*Version:\s*(.+)$/', $s);
  preg_match('/Domain Name\.*: *(.+)/', $s);
  preg_match('/(Email|EmailAddress)\(\)/', $s);
  preg_match('/\*\*[^*]*\*\*/', $s);
  preg_match('/(\*[^ ][^*]*\*)/', $s);
  preg_match('/\+[^+]*\+/', $s);
  preg_match('/\bMac[A-Za-z]{2,}[^aciozj]\b/', $s);
  preg_match('/\bMc/', $s);
  preg_match('/\b(Ma?c)([A-Za-z]+)/', $s);
  preg_match('/(?m)^Created: *(.*?)$/', $s);
  preg_match('/#\+(\w+): (.*)/', $s);
  preg_match('/^(\*+)(?: +(CANCELED|DONE|TODO))?(?: +(\[#.\]))?(?: +(.*?))?(?:(:[a-zA-Z0-9_@#%:]+:))??[ \t]*$/', $s);
  preg_match('/^[ \t]*#/', $s);
  preg_match('/(?i)\+BEGIN_(CENTER|COMMENT|EXAMPLE|QUOTE|SRC|VERSE)/', $s);
  preg_match('/(\d+\.\d+) (\w+)(\(.*) <(.+)>/', $s);
  preg_match('/(\d+\.\d+) (\w+)(\(.*)/', $s);
  preg_match('/^(\s+\S+:\s+.*)$/', $s);
  preg_match('/abc{2,5}d?e*f+/', $s);
  preg_match('/a+b+/', $s);
}

function _5($s) {
  preg_match('/(?m:^%(\.\d)?[sdfgtq]$)/', $s);
  preg_match('/(?m:^[ \t]*%(\.\d)?[sdfgtq]$)/', $s);
  preg_match('~(?i)<html.*/head>~', $s);
  preg_match('/ COLLATE ([^ ]+)/', $s);
  preg_match('/ CHARACTER SET ([^ ]+)/', $s);
  preg_match('/(?:(.+); )?InnoDB free: .*/', $s);
  preg_match('~<td><span class="label">性别：</span><span field="">([^<]+)</span></td>~', $s);
  preg_match('/\ACK2txt\n$/', $s);
  preg_match('/^\d{3}[- ]?\d{2}[- ]?\d{4}$/', $s);
  preg_match('~<arrmsg1>([\w\W]+?)</arrmsg1>~', $s);
  preg_match('/<p(.*?)>/', $s);
  preg_match('/^(?i)(\s*insert\s+into\s+)/', $s);
  preg_match('/^(?i)\((\s*\w+(?:\s*,\s*\w+){1,100})\s*\)\s*/', $s);
  preg_match('/0\d-[2-9]\d{3}-\d{4}/', $s);
  preg_match('/0\d{2}-[2-9]\d{2}-\d{4}/', $s);
  preg_match('/0\d{3}-[2-9]\d-\d{4}/', $s);
  preg_match('/0\d{4}-[2-9]-\d{4}/', $s);
  preg_match('/(?i)(refer):\s+(.*?)(\s|$)/', $s);
  preg_match('/(?m:^)(\s+)?(?i)(Whois server|whois):\s+(.*?)(\s|$)/', $s);
  preg_match('/@(?i:article){.*/', $s);
  preg_match('/\blimit \?(?:, ?\?| offset \?)?/', $s);
  preg_match('/^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,4}$/', $s);
  preg_match('/^4\d{12}(\d{3})?$/', $s);
  preg_match('/^(5[1-5]\d{4}|677189)\d{10}$/', $s);
  preg_match('/^(6011|65\d{2}|64[4-9]\d)\d{12}|(62\d{14})$/', $s);
  preg_match('/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/', $s);
  preg_match('/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}$/', $s);
  preg_match('/^\d{8}T\d{6}Z$/', $s);
  preg_match('/^\w+\(.*\)$/', $s);
  preg_match('/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}/', $s);
  preg_match('/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/', $s);
}

function _6($s) {
  preg_match('~^/video/([\w\-]{6,12})\.json$~', $s);
  preg_match('/(\n|\r|\r\n)$/', $s);
  preg_match('/(?m)^@@@@\n/', $s);
  preg_match('/agggtaaa|tttaccct/', $s);
  preg_match('/[cgt]gggtaaa|tttaccc[acg]/', $s);
  preg_match('/a[act]ggtaaa|tttacc[agt]t/', $s);
  preg_match('/(?i:http).*.git/', $s);
  preg_match('~kitsu.io/users/(.*?)/library~', $s);
  preg_match('/k8s_.*\.metadata\.name$/', $s);
  preg_match('/k8s_\w+_\w+_deployment\.spec\.selector\.match_labels$/', $s);
  preg_match('/[!&]([^=!&?[)]+)|\[\[(.*?)\]\]/', $s);
  preg_match('/\\b[A-Fa-f0-9]{32}\\b/', $s);
  preg_match('/(?i)(s?(\d{1,2}))[ex]/', $s);
  preg_match('/(?i)([ex](\d{2})(?:\d|$))/', $s);
  preg_match('/(-\s+(\d+)(?:\d|$))/', $s);
  preg_match('~ edge/(\d+)\.(\d+)~', $s);
  preg_match('/^.+@.+\..+$/', $s);
  preg_match('/^\$bgm\s+(\[.*\])$/', $s);
  preg_match('/^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$/', $s);
  preg_match('/;[^=;{}]+;/', $s);
  preg_match('/\s(if|else|while|catch)\s*([^{;]+;|;)/', $s);
  preg_match('/^[^ ]*apache2/', $s);
  preg_match('/^(?:[-+]?(?:0|[1-9]\d*))$/', $s);
  preg_match('/<(.|[\r\n])*?>/', $s);
  preg_match('/^(::f{4}:)?10\.\d{1,3}\.\d{1,3}\.\d{1,3}/', $s);
  preg_match('/(?i)\([^)]*remaster[^)]*\)$/', $s);
  preg_match('/>([a-zA-Z0-9]+@[a-zA-Z0-9.]+\.[a-zA-Z0-9]+)</', $s);
  preg_match('/(?i)\([^)]*mix[^)]*\)$/', $s);
}
`)
	test.RunAndMatch()
}
