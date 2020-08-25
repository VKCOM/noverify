package rules

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser(t *testing.T) {
	p := NewParser()
	rset, err := p.Parse("rules.php", strings.NewReader(`<?php
/**
 * @comment This is an example of fully-docummented rule.
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
				Comment: "This is an example of fully-docummented rule.",
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
