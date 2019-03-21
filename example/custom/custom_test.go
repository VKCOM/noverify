package main

import (
	"log"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node"
)

func init() {
	linter.RegisterBlockChecker(func(ctx linter.BlockContext) linter.BlockChecker { return &block{ctx: ctx} })
	go linter.MemoryLimiterThread()
}

func testParse(t *testing.T, filename string, contents string) (rootNode node.Node, w *linter.RootWalker) {
	var err error
	rootNode, w, err = linter.ParseContents(filename, []byte(contents), "UTF-8", nil)
	if err != nil {
		t.Errorf("Could not parse %s: %s", filename, err.Error())
		t.Fail()
	}

	if !meta.IsIndexingComplete() {
		w.UpdateMetaInfo()
	}

	return rootNode, w
}

func singleFileReports(t *testing.T, contents string) []*linter.Report {
	meta.ResetInfo()

	testParse(t, `first.php`, contents)
	meta.SetIndexingComplete(true)
	_, w := testParse(t, `first.php`, contents)

	return w.GetReports()
}

func TestAssignmentAsExpression(t *testing.T) {
	reports := singleFileReports(t, `<?php
	// phpdoc annotations are not required for NoVerify in simple cases
	function something() {
		$a = "test";
		return $a;
	}
	function in_array() {}

	function test() {
		$b = ["1", "2", "3"];

		if (in_array(something(), $b)) {
			echo "third arg true";
		}

		if (something() == $b[1]) {
			echo "must be ===";
		}
	}
	`)

	for _, r := range reports {
		log.Printf("%s", r)
	}

	if len(reports) != 2 {
		t.Errorf("Unexpected number of reports: expected 2, got %d", len(reports))
		if len(reports) < 2 {
			t.FailNow()
		}
	}

	text := reports[0].String()

	if !strings.Contains(text, "3rd argument of in_array must be true when comparing strings") {
		t.Errorf("Wrong report text: expected '3rd argument of in_array must be true', got '%s'", text)
	}

	text = reports[1].String()

	if !strings.Contains(text, "Strings must be compared using '===' operator") {
		t.Errorf("Wrong report text: expected 'Strings must be compared using '===' operator', got '%s'", text)
	}
}
