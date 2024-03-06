package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestTriggerNonError(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
trigger_error('notice');
trigger_error('also notice', E_USER_NOTICE);
trigger_error('a warning', E_USER_WARNING);
`)
}

func TestTriggerErrorFQN(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
\trigger_error('error', E_USER_ERROR);
echo 'unreachable';
`)
	test.Expect = []string{
		`Unreachable code`,
	}
	test.RunAndMatch()
}

func TestTriggerError(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
trigger_error('error', E_USER_ERROR);
echo 'unreachable';
`)
	test.Expect = []string{
		`Unreachable code`,
	}
	test.RunAndMatch()
}

func TestTriggerErrorTransitive(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
function f($msg) {
  trigger_error($msg, E_USER_ERROR);
}

f('error');
echo 'unreachable';
`)
	test.Expect = []string{
		`Unreachable code`,
	}
	test.RunAndMatch()
}

func TestUserError(t *testing.T) {
	// user_error is a trigger_error alias.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
user_error('error', E_USER_ERROR);
echo 'unreachable';
`)
	test.Expect = []string{
		`Unreachable code`,
	}
	test.RunAndMatch()
}
