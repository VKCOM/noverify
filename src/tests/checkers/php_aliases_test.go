package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPhpAliases(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = "1")
$_ = join("", []);
`)
	test.Expect = []string{
		`Call to undefined function join`,
		`Use implode instead of 'join'`,
	}
	test.RunAndMatch()
}

func TestPhpAliasesFunctionCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types = "1")
$_ = join("", []);

function test($d){
ocicollmax();
}

test(join("", []));
`)
	test.Expect = []string{
		`Use OCICollection::max instead of 'ocicollmax'`,
		`Call to undefined function ocicollmax`,
		`Use implode instead of 'join'`,
		`Call to undefined function join`,
		`Use implode instead of 'join'`,
		`Call to undefined function join`,
	}
	test.RunAndMatch()
}
