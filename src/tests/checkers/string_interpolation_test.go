package checkers

import (
	"github.com/VKCOM/noverify/src/linttest"
	"testing"
)

func TestInterpolationDeprecated1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

$name = 'PHP';

echo "Hello ${name}";
`)
	test.Expect = []string{
		`stringInterpolationDeprecated: use {$variable} instead ${variable}`,
	}
	test.RunAndMatch()
}

func TestInterpolationDeprecated2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

$name = 'PHP';

echo "Hello {$name}";
`)
	test.RunAndMatch()
}

func TestInterpolationDeprecated3(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = 1);

$name = 'PHP';
$lang = 'language';

echo "Hello {$name} ${lang}";
`)
	test.Expect = []string{
		`stringInterpolationDeprecated: use {$variable} instead ${variable}`,
	}
	test.RunAndMatch()
}
