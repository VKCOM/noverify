package rules

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linter/lintapi"
)

func TestParser_Parse(t *testing.T) {
	p := NewParser()
	rset, err := p.Parse("rules.php", strings.NewReader(`<?php
/**
 * @comment This is an example of fully-documented rule.
 * @before  array(1, 2)
 * @after   [1, 2]
 */
function arraySyntax() {
  /**
   * @maybe found old array literal syntax
   */
  array(${"*"});

  /**
   * @maybe found old list assignment syntax
   */
  list(${"*"}) = $_;
}
`))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	tests := []struct {
		name string
		doc  RuleDoc
	}{
		{
			name: "arraySyntax",
			doc: RuleDoc{
				Comment: "This is an example of fully-documented rule.",
				Before:  "array(1, 2)",
				After:   "[1, 2]",
			},
		},
	}

	for _, test := range tests {
		info, ok := rset.DocByName[test.name]
		if !ok {
			t.Errorf("%s: no doc entry found", test.name)
			continue
		}
		haveDoc := info
		wantDoc := test.doc
		if diff := cmp.Diff(haveDoc, wantDoc); diff != "" {
			t.Errorf("%s: docs mismatch:\n%s", test.name, diff)
			continue
		}
	}
}

func TestParser_GroupedRule(t *testing.T) {
	p := NewParser()
	rfile := `<?php
function arrays() {
  /**
   * @warning can be replaced with isset($array[$key]) or array_key_exists($key, $array)
   */
  array_has_key: {
    in_array($key, array_keys($array));
    in_array($key, array_keys($array), true);
    in_array($key, array_keys_as_ints($array));
    in_array($key, array_keys_as_ints($array), true);
    in_array($key, array_keys_as_strings($array));
    in_array($key, array_keys_as_strings($array), true);
  }
}
`
	rset, err := p.Parse("rules.php", strings.NewReader(rfile))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(rset.Names) != 1 || rset.Names[0] != "arrays" {
		t.Fatalf("expected rule names [arrays], got %v", rset.Names)
	}

	if rset.Any.CountRules != 6 {
		t.Fatalf("expected 6 registered rules, got %d", rset.Any.CountRules)
	}

	fnCallKind := ir.GetNodeKind(&ir.FunctionCallExpr{})
	fnRules := rset.Any.RulesByKind[fnCallKind]
	if len(fnRules) != 6 {
		t.Fatalf("expected 6 function call rules, got %d", len(fnRules))
	}

	const wantMsg = "can be replaced with isset($array[$key]) or array_key_exists($key, $array)"
	for i, r := range fnRules {
		if r.Name != "arrays" {
			t.Errorf("rule %d: expected name 'arrays', got %q", i, r.Name)
		}
		if r.Message != wantMsg {
			t.Errorf("rule %d: unexpected message %q", i, r.Message)
		}
		if r.Level != lintapi.LevelWarning {
			t.Errorf("rule %d: expected warning level %d, got %d", i, lintapi.LevelWarning, r.Level)
		}
	}

	labelKind := ir.GetNodeKind(&ir.LabelStmt{})
	if labelRules := rset.Any.RulesByKind[labelKind]; len(labelRules) != 0 {
		t.Errorf("expected 0 label statement rules, got %d", len(labelRules))
	}

	stmtListKind := ir.GetNodeKind(&ir.StmtList{})
	if stmtListRules := rset.Any.RulesByKind[stmtListKind]; len(stmtListRules) != 0 {
		t.Errorf("expected 0 stmt list rules, got %d", len(stmtListRules))
	}
}
