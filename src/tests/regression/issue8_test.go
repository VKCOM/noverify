package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue8(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types=1);
	class Magic
	{
		public function __get($a);
		public function __set($a, $b);
		public function __call($a, $b);
	}

	class MagicStatic {
		public static function __callStatic($a, $b);
	}

	function test() {
		$m = new Magic;
		echo $m->some_property;
		$m->another_property = 3;
		$m->call_something();
		MagicStatic::callSomethingStatic();
	}`)
}
