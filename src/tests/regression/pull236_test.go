package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPull236(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
$_ = new class {
  private function f() { return 10; }

  /** @return int */
  public function g() { return $this->f(); }
};
`)
}
