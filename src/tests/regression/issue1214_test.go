package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue1214(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types = 1);

function test() {
 $special_items = null;
 $_ = [...$special_items];
}`)
}
