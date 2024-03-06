package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestGettypeFunction(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{
		`stubs/phpstorm-stubs/standard/standard_5.php`,
	}
	test.AddFile(`<?php
declare(strict_types=1);
function getTypeMisUse(mixed $var) {
  if (gettype($var) === "string") {
  }

  if (gettype($var) == "double") {
  }

  if (gettype($var) !== "array") {
  }

  if (gettype($var) != "boolean") {
  }

  if (gettype($var) === "object" && true) {
  }

  if (gettype(getTypeMisUse($var)) === "integer") {
  }

  if (gettype(getTypeMisUse($var)) != "resource") {
  }
}
`)
	test.Expect = []string{
		`use is_string instead of 'gettype($var) === "string"'`,
		`use is_float instead of 'gettype($var) == "double"'`,
		`use is_array instead of 'gettype($var) !== "array"'`,
		`use is_bool instead of 'gettype($var) != "boolean"'`,
		`use is_object instead of 'gettype($var) === "object"'`,
		`use is_int instead of 'gettype(getTypeMisUse($var)) === "integer"'`,
		`use is_resource instead of 'gettype(getTypeMisUse($var)) != "resource"'`,
	}

	test.RunAndMatch()
}
