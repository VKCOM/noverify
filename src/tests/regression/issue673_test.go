package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue673(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types=1);
$_ = ['\n' => 1, "\n" => 2];
`)
}
