package rules_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/rules"
)

type ruleErrorTest struct {
	name   string
	rule   string
	expect string
}

func TestRuleError(t *testing.T) {
	tests := []ruleErrorTest{
		{
			name: `NamespaceWithBodyNotSupported`,
			rule: `<?php
namespace Foo {
	function boo() {}
}`,
			expect: "<test>:2: namespace with body is not supported",
		},
		{
			name: `MultiPartNamespaceNotSupported`,
			rule: `<?php
namespace Soo\Foo;
`,
			expect: "<test>:2: multi-part namespace names are not supported",
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
			expect: "<test>:6: @name expects exactly 1 param, got 2",
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
			expect: "<test>:7: someCheckWithInvalidRule: <test>:12: @name is not allowed inside a function",
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
			expect: "<test>:7: @location expects exactly 1 params, got 4",
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
			expect: "<test>:7: @location 2nd param must be a phpgrep variable",
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
			expect: "<test>:7: @scope expects exactly 1 params, got 4",
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
			expect: "<test>:7: unknown @scope: city",
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
			expect: "<test>:8: duplicated @fix",
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
			expect: "<test>:7: @path expects exactly 1 param, got 2",
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
			expect: "<test>:8: duplicate @path constraint",
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
			expect: "<test>:7: @type expects exactly 2 params, got 4",
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
			expect: "<test>:7: @type 2nd param must be a phpgrep variable",
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
			expect: "<test>:8: $a: duplicate type constraint",
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
			expect: "<test>:7: $a: parseType(<=): bad type expression",
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
			expect: "<test>:7: @pure expects exactly 1 param, got 3",
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
			expect: "<test>:7: @pure param must be a phpgrep variable",
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
			expect: "<test>:7: unknown attribute @hello on line 4",
		},
		{
			name: `MissingName`,
			rule: `<?php
/**
 * @maybe Some
 */
$_ = foo();
`,
			expect: "<test>:5: missing @name attribute",
		},
		{
			name: `UnknownMatcherClass`,
			rule: `<?php
/**
 * @name Some
 * @maybe Some
 */
${"boo"} = $_;
`,
			expect: "<test>:6: pattern compilation error: unknown matcher class 'boo'",
		},
		{
			name: `VariableFromTypeNotPresentInPattern`,
			rule: `<?php
/**
 * @name Some
 * @warning Some
 * @type int $x
 * @type int $y
 */
$_ = $y;
`,
			expect: "<test>:8: @type contains a reference to a variable x that is not present in the pattern",
		},
		{
			name: `VariableFromTypeNotPresentInPatternGood`,
			rule: `<?php
/**
 * @name Some
 * @warning Some
 * @type int $x
 */
$x = 1;
`,
			expect: "",
		},
		{
			name: `VariableFromTypeNotPresentInPatternGood#2`,
			rule: `<?php
function someRules() {
	/**
	 * @warning Some
	 * @type int $x
	 */
	any: {
		$x = 1;
		$x = 2;
	}
}
`,
			expect: "",
		},
		{
			name: `VariableFromTypeNotPresentInPatternGood#3`,
			rule: `<?php
/**
 * @name Some
 * @warning Some
 * @type int $x
 */
${"x:var"} = 1;
`,
			expect: "",
		},
		{
			name: `VariableFromPureNotPresentInPattern`,
			rule: `<?php
/**
 * @name Some
 * @warning Some
 * @type int $y
 * @pure $x
 */
$_ = $y;
`,
			expect: "<test>:8: @pure contains a reference to a variable x that is not present in the pattern",
		},
		{
			name: `VariableFromLocationNotPresentInPattern`,
			rule: `<?php
/**
 * @name Some
 * @warning Some
 * @location $y
 */
(string)$x;
`,
			expect: "<test>:7: @location contains a reference to a variable y that is not present in the pattern",
		},
	}

	runRulesErrorTest(t, tests)
}

func runRulesErrorTest(t *testing.T, rulesTest []ruleErrorTest) {
	t.Helper()

	for i := range rulesTest {
		test := rulesTest[i]
		t.Run(test.name, func(t *testing.T) {
			rparser := rules.NewParser()
			_, err := rparser.Parse("<test>", strings.NewReader(test.rule))
			if err != nil {
				msg := err.Error()
				if msg != test.expect {
					t.Errorf("unexpected error:\nwant: %s\nhave: %s", test.expect, msg)
				}
			} else if test.expect != "" {
				t.Errorf("pattern '%s' matched nothing", test.expect)
			}
		})
	}
}
