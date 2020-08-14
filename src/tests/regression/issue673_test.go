package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue673(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
$_ = ['\n' => 1, "\n" => 2];
`)
}
