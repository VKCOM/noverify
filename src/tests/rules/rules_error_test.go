package rules_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/rules"
)

type ruleTest struct {
	name   string
	rule   string
	expect string
}

func TestRuleError(t *testing.T) {
	tests := []ruleTest{
		{
			name: `NamespaceWithBodyNotSupported`,
			rule: `<?php
namespace Foo {
	function boo() {}
}`,
			expect: "namespace with body is not supported",
		},
		{
			name: `MultiPartNamespaceNotSupported`,
			rule: `<?php
namespace Soo\Foo;
`,
			expect: "multi-part namespace names are not supported",
		},
		{
			name: `NameExpectsExactlyOneParam`,
			rule: `<?php
/**
 * @name Some check
 * @maybe Some
 */
$_ = foo();
`,
			expect: "@name expects exactly 1 param, got 2",
		},
		{
			name: `NameNotAllowedInFunction`,
			rule: `<?php
/**
 * @comment Some comment
 * @before  Some before
 * @after   Some after
 */
function someCheckWithInvalidRule() {
    /**
     * @name SomeName
     * @maybe Some
     */
    $_ = foo();
}
`,
			expect: "someCheckWithInvalidRule: :12: @name is not allowed inside a function",
		},
		{
			name: `LocationExpectsExactlyOneParam`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @location Somewhere in the city
 */
$_ = foo();
`,
			expect: "@location expects exactly 1 params, got 4",
		},
		{
			name: `LocationSecondParamMustBePHPGrepVariable`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @location here
 */
$_ = foo();
`,
			expect: "@location 2nd param must be a phpgrep variable",
		},
		{
			name: `ScopeExpectsExactlyOneParam`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @scope Somewhere in the city
 */
$_ = foo();
`,
			expect: "@scope expects exactly 1 params, got 4",
		},
		{
			name: `UnknownScope`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @scope city
 */
$_ = foo();
`,
			expect: "unknown @scope: city",
		},
		{
			name: `DuplicatedFix`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @fix $_ = boo();
 * @fix $_ = coo();
 */
$_ = foo();
`,
			expect: "duplicated @fix",
		},
		{
			name: `PathExpectsExactlyOneParam`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @path to hell
 */
$_ = foo();
`,
			expect: "@path expects exactly 1 param, got 2",
		},
		{
			name: `DuplicatedPath`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @path hell;
 * @path earth;
 */
$_ = foo();
`,
			expect: "duplicate @path constraint",
		},
		{
			name: `PathExpectsExactlyTwoParam`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @type bla bla bla $a
 */
$a = foo();
`,
			expect: "@type expects exactly 2 params, got 4",
		},
		{
			name: `TypeSecondParamMustBePHPGrepVariable`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @type int i_am_not_phpgrep_variable
 */
$a = foo();
`,
			expect: "@type 2nd param must be a phpgrep variable",
		},
		{
			name: `DuplicateType`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @type int $a
 * @type string $a
 */
$a = foo();
`,
			expect: "$a: duplicate type constraint",
		},
		{
			name: `BadTypeExpression`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @type <= $a
 */
$a = foo();
`,
			expect: "$a: parseType(<=): bad type expression",
		},
		{
			name: `PureExpectsExactlyOneParam`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @pure line of life
 */
$_ = foo();
`,
			expect: "@pure expects exactly 1 param, got 3",
		},
		{
			name: `PureSecondParamMustBePHPGrepVariable`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @pure a
 */
$a = foo();
`,
			expect: "@pure param must be a phpgrep variable",
		},
		{
			name: `UnknownAttribute`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 * @hello a
 */
$a = foo();
`,
			expect: "unknown attribute @hello on line 4",
		},
		{
			name: `MissingName`,
			rule: `<?php
/**
 * @maybe Some
 */
$_ = foo();
`,
			expect: "missing @name attribute",
		},
	}

	runRulesErrorTest(t, tests)
}

func runRulesErrorTest(t *testing.T, rulesTest []ruleTest) {
	t.Helper()

	for _, test := range rulesTest {
		t.Run(test.name, func(t *testing.T) {
			rparser := rules.NewParser()
			_, err := rparser.Parse("", strings.NewReader(test.rule))
			if err != nil {
				var msg string

				switch err := err.(type) {
				case *rules.ParseError:
					msg = err.Message()
				default:
					msg = err.Error()
				}

				if msg != test.expect {
					t.Errorf("unexpected error:\nwant: %s\nhave: %s", test.expect, msg)
				}
			} else if test.expect != "" {
				t.Errorf("pattern '%s' matched nothing", test.expect)
			}
		})
	}
}
