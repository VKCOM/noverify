package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue3(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class TestClass
	{
		/** get always returns "." */
		public function get(): string
		{
			return '.';
		}
	}

	function a(TestClass ...$testClasses): string
	{
		$result = '';
		foreach ($testClasses as $testClass) {
			$result .= $testClass->get();
		}

		return $result;
	}

	echo a(new TestClass()), "\n";
	echo a(); // OK to call with 0 arguments.
	`)
}
