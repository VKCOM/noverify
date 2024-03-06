package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue16(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
	declare(strict_types = 1);
	interface DateTimeInterface {
		public function format($fmt);
	}

	interface OtherInterface {
		public function useless();
	}

	interface TestInterface
	{
		const TEST = 1;

		public function getCreatedAt(): \DateTimeInterface;
	}

	interface TestExInterface extends OtherInterface, TestInterface
	{
	}

	function a(TestExInterface $testInterface): string
	{
		echo TestExInterface::TEST;
		return $testInterface->getCreatedAt()->format('U');
	}

	function b(TestExInterface $testInterface) {
		echo TestExInterface::TEST2;
		return $testInterface->nonexistent()->format('U');
	}`)
	test.Expect = []string{
		`Call to undefined method {\TestExInterface}->nonexistent()`,
		"Call to undefined method {mixed}->format()",
		"Class constant \\TestExInterface::TEST2 does not exist",
	}
	linttest.RunFilterMatch(test, "undefinedMethod", "undefinedConstant")
}
