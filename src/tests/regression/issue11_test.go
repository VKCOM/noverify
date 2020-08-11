package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue11(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class Generator {
		/** send sends a message */
		public function send();
	}

	function a($a): \Generator
	{
		yield $a;
	}

	a(42)->send(42);
	`)
}
