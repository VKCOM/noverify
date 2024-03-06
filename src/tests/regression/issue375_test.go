package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue375(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);
function ref_sink(&$ref) {}

function f() {
  $x = [1];
  ref_sink($x[0]);
}

$x = [1, 5];
ref_sink($x[1]);

`)
}
