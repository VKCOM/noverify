package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
)

func pathRulesSetInit(t *testing.T) *linttest.Suite {
	t.Helper()
	linterConfig := linter.NewConfig("8.1")
	linterConfig.ProjectPath = "dev/"
	linterConfig.StrictMixed = true

	linterConfig.PathRules = linter.BuildRuleTree(map[string]*linter.PathRuleSet{
		"disable-emptyStmt": {
			Enabled:  map[string]bool{},
			Disabled: map[string]bool{"emptyStmt": true},
		},
		"enable-emptyStmt": {
			Enabled: map[string]bool{
				"emptyStmt": true,
			},
			Disabled: map[string]bool{},
		},
		"mixed/foo.php": {
			Enabled: map[string]bool{
				"emptyStmt": true,
			},
			Disabled: map[string]bool{
				"undefinedFunction": true,
			},
		},
		"star/*/tests": {
			Enabled: map[string]bool{
				"emptyStmt": true,
			},
			Disabled: map[string]bool{},
		},
		"mixed/*/tests": {
			Enabled: map[string]bool{
				"emptyStmt": true,
			},
			Disabled: map[string]bool{
				"undefinedFunction": true,
			},
		},
	})

	var suite = linttest.NewSuite(t)
	suite.IgnoreUndeclaredChecks()
	suite.UseConfig(linterConfig)
	return suite
}

func TestDisablePath(t *testing.T) {
	test := pathRulesSetInit(t)
	code := `<?php
require_once 'foo.php';;
        `
	test.AddNamedFile("disable-emptyStmt/foo.php", code)

	test.RunAndMatch()
}

func TestEnablePath(t *testing.T) {
	test := pathRulesSetInit(t)
	code := `<?php
require_once 'foo.php';;
        `
	test.AddNamedFile("dev/enable-emptyStmt/foo.php", code)

	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed at`,
	}
	test.RunAndMatch()
}

func TestStarPath(t *testing.T) {
	test := pathRulesSetInit(t)
	code := `<?php
require_once 'foo.php';;
        `
	test.AddNamedFile("star/something/another/tests/foo.php", code)

	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed at`,
	}
	test.RunAndMatch()
}

func TestMixedStarPath(t *testing.T) {
	test := pathRulesSetInit(t)
	code := `<?php
require_once 'foo.php';;
        `

	test.AddNamedFile("mixed/something/another/tests/foo.php", code)

	code = `<?php
function function_exists($name) { return 1 == 2; }

function f($cond) {
  if (!function_exists('\foo')) {
    \foo();
  }
  if ($cond && !function_exists('bar')) {
    bar("a", "b");
  }
  if ($cond || !function_exists('a\b\baz')) {
    a\b\baz(1);
  }
}
`
	test.AddNamedFile("mixed/something/another/tests/foo2.php", code)

	test.Expect = []string{
		`Semicolon (;) is not needed here, it can be safely removed at`,
	}
	test.RunAndMatch()
}

func TestMixedRulesPath(t *testing.T) {
	test := pathRulesSetInit(t)
	mergedCode := `<?php
require_once 'foo.php';;
function function_exists($name) { return 1 == 2; }

function f($cond) {
  if (!function_exists('\foo')) {
    \foo();
  }
  if ($cond && !function_exists('bar')) {
    bar("a", "b");
  }
  if ($cond || !function_exists('a\b\baz')) {
    a\b\baz(1);
  }
}
`

	test.AddNamedFile("mixed/foo.php", mergedCode)

	test.Expect = []string{
		"Semicolon (;) is not needed here, it can be safely removed at",
	}
	test.RunAndMatch()
}

func TestIsRuleEnabledWithAbsolutePaths(t *testing.T) {
	pathRules := map[string]*linter.PathRuleSet{
		"/A": {
			Enabled:  map[string]bool{"rule1": true},
			Disabled: map[string]bool{},
		},
		"/A/B": {
			Enabled:  map[string]bool{},
			Disabled: map[string]bool{"rule1": true},
		},
		"/A/B/C": {
			Enabled:  map[string]bool{"rule2": true},
			Disabled: map[string]bool{},
		},
	}

	ruleTree := linter.BuildRuleTree(pathRules)

	tests := []struct {
		name      string
		filePath  string
		checkRule string
		want      bool
	}{
		{
			name:      "Absolute path with rule1 disabled in /A/B",
			filePath:  "/A/B/D/file.php",
			checkRule: "rule1",
			want:      false,
		},
		{
			name:      "Absolute path with rule2 enabled in /A/B/C",
			filePath:  "/A/B/C/file.php",
			checkRule: "rule2",
			want:      true,
		},
		{
			name:      "Absolute path with rule1 enabled in /A",
			filePath:  "/A/X/Y/file.php",
			checkRule: "rule1",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linter.IsRuleEnabledForPath(ruleTree, tt.filePath, tt.checkRule)
			if got != tt.want {
				t.Errorf("IsRuleEnabledForPath(%q, %q) = %v; want %v", tt.filePath, tt.checkRule, got, tt.want)
			}
		})
	}
}

func TestIsRuleEnabledWithRelativeComponents(t *testing.T) {
	pathRules := map[string]*linter.PathRuleSet{
		"A/B/C": {
			Enabled:  map[string]bool{"rule1": true},
			Disabled: map[string]bool{},
		},
	}

	ruleTree := linter.BuildRuleTree(pathRules)

	tests := []struct {
		name      string
		filePath  string
		checkRule string
		want      bool
	}{
		{
			name:      "Path with '.' components",
			filePath:  "A/B/./C/file.php",
			checkRule: "rule1",
			want:      true,
		},
		{
			name:      "Path with '..' components",
			filePath:  "A/B/D/../C/file.php",
			checkRule: "rule1",
			want:      true,
		},
		{
			name:      "Path leading to different directory due to '..'",
			filePath:  "A/B/C/../D/file.php",
			checkRule: "rule1",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linter.IsRuleEnabledForPath(ruleTree, tt.filePath, tt.checkRule)
			if got != tt.want {
				t.Errorf("IsRuleEnabledForPath(%q, %q) = %v; want %v", tt.filePath, tt.checkRule, got, tt.want)
			}
		})
	}
}

func TestIsRuleEnabledWithLeadingAndTrailingSlashes(t *testing.T) {
	pathRules := map[string]*linter.PathRuleSet{
		"/A/B/C/": {
			Enabled:  map[string]bool{"rule1": true},
			Disabled: map[string]bool{},
		},
	}

	ruleTree := linter.BuildRuleTree(pathRules)

	tests := []struct {
		name      string
		filePath  string
		checkRule string
		want      bool
	}{
		{
			name:      "Path with trailing slash in rule and file path",
			filePath:  "/A/B/C/file.php",
			checkRule: "rule1",
			want:      true,
		},
		{
			name:      "Path with trailing slash in rule only",
			filePath:  "/A/B/C/file.php",
			checkRule: "rule1",
			want:      true,
		},
		{
			name:      "Path with trailing slash in file path only",
			filePath:  "/A/B/C//file.php",
			checkRule: "rule1",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linter.IsRuleEnabledForPath(ruleTree, tt.filePath, tt.checkRule)
			if got != tt.want {
				t.Errorf("IsRuleEnabledForPath(%q, %q) = %v; want %v", tt.filePath, tt.checkRule, got, tt.want)
			}
		})
	}
}

func TestIsRuleEnabledWithConflictingRules(t *testing.T) {
	pathRules := map[string]*linter.PathRuleSet{
		"A": {
			Enabled:  map[string]bool{"rule1": true},
			Disabled: map[string]bool{},
		},
		"A/B": {
			Enabled:  map[string]bool{},
			Disabled: map[string]bool{"rule1": true},
		},
		"A/B/C": {
			Enabled:  map[string]bool{"rule1": true},
			Disabled: map[string]bool{},
		},
	}

	ruleTree := linter.BuildRuleTree(pathRules)

	tests := []struct {
		name      string
		filePath  string
		checkRule string
		want      bool
	}{
		{
			name:      "Rule1 re-enabled in A/B/C after being disabled in A/B",
			filePath:  "A/B/C/file.php",
			checkRule: "rule1",
			want:      true,
		},
		{
			name:      "Rule1 remains disabled in A/B/D",
			filePath:  "A/B/D/file.php",
			checkRule: "rule1",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linter.IsRuleEnabledForPath(ruleTree, tt.filePath, tt.checkRule)
			if got != tt.want {
				t.Errorf("IsRuleEnabled(%q, %q) = %v; want %v", tt.filePath, tt.checkRule, got, tt.want)
			}
		})
	}
}

func TestWildcardMiddleSegmentBug(t *testing.T) {
	pathRules := map[string]*linter.PathRuleSet{
		"*/tests/*": {
			Enabled:  map[string]bool{},
			Disabled: map[string]bool{"rule1": true},
		},
	}

	ruleTree := linter.BuildRuleTree(pathRules)

	tests := []struct {
		name      string
		filePath  string
		checkRule string
		want      bool
	}{
		{
			name:      "Wildcard */tests/* should disable rule1 in tests/A/B/file.php",
			filePath:  "tests/A/B/file.php",
			checkRule: "rule1",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linter.IsRuleEnabledForPath(ruleTree, tt.filePath, tt.checkRule)
			if got != tt.want {
				t.Errorf("IsRuleEnabledForPath(%q, %q) = %v; want %v", tt.filePath, tt.checkRule, got, tt.want)
			}
		})
	}
}
