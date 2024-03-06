package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue37(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types=1);
	class Foo {
		public $a;
		public $b;
	}

	/**
	 * @param Foo[] $arr
	 */
	function f($arr) {
		$ads_ids = array_keys($arr);
		foreach ($ads_ids as $num => $ad_id) {
			if ($num + 1 < count($ads_ids)) {
				$second_ad_id = $ads_ids[$num + 1];
				$arr[$ad_id]->a = $arr[$second_ad_id]->b;
			}
		}
	}`)
	linttest.RunFilterMatch(test, "unused")
}
