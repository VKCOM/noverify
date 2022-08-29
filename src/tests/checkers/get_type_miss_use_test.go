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
function getTypeMissUse(mixed $var){
    if (gettype($var) === "string"){
    }

    if (gettype($var) == "double"){
    }

    if (gettype($var) !== "array"){
    }

    if (gettype($var) != "boolean"){
    }

	if (gettype($var) != "object"){
    }
}
`)
	test.Expect = []string{
		`use is_string instead of gettype()`,
		`use is_float instead of gettype()`,
		`use is_array instead of gettype()`,
		`use is_bool instead of gettype()`,
		`use is_object instead of gettype()`,
	}

	test.RunAndMatch()
}
