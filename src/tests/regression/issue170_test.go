package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue170(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);

global $v;

switch ($v) {
case 1:
  error(); // no break here
case 2:
  $_ = $v;
  break;
default:
  break;
}

function error() {}
`)
}
