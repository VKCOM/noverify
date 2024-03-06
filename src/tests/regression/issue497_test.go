package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue497(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types = 1);
class T {}

/**
 * @param shape(a:int) $x
 * @return T<int>
 */
function f($x) {
  $v = $x['a'];
  return [$v];
}
`)
}
