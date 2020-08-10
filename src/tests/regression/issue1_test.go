package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	interface TestInterface
	{
		const TEST = '1';
	}

	class TestClass implements TestInterface
	{
		/** get returns interface constant */
		public function get()
		{
			return self::TEST;
		}
	}`)
}
