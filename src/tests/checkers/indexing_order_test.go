package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIndexingOrderClasses(t *testing.T) {
	// This test introduces conflicting declarations.
	//
	// No matter which file traversal order we get, if we simply
	// override the metadata, we'll get warnings in either case.
	//
	// If foo/A class is recorded, the second file will have troubles
	// accessing $v->field. If bar/A is recorded, the first file will
	// have troubles creating A with 0 constructor arguments.

	test := linttest.NewSuite(t)
	test.AddNamedFile("/foo/A.php", `<?php
class A {}

$v = new A();
`)
	test.AddNamedFile("/bar/A.php", `<?php
class A {
  public $field;
  public function __construct($x) { $this->field = $x; }
}

$v = new A(1);
echo $v->field;
`)

	// TODO: this test should give no warnings.
	// Right now we ensure that the linter output doesn't depend
	// on the file traversal order, but this warnings is still out of place.
	test.Expect = []string{
		`argCount: Too few arguments for \A constructor`,
	}
	test.RunAndMatch()
}

func TestIndexingOrderTraits(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddNamedFile("/foo/A.php", `<?php
trait TA {}

class A { use TA; }

$v = new A();
`)
	test.AddNamedFile("/bar/A.php", `<?php
trait TA {
  public $field;
  public function __construct($x) { $this->field = $x; }
}

class B {
  use TA;
}

$v = new B(1);
echo $v->field;
`)

	// TODO: this test should give no warnings.
	// Right now we ensure that the linter output doesn't depend
	// on the file traversal order, but this warnings is still out of place.
	test.Expect = []string{
		`argCount: Too few arguments for \A constructor`,
	}
	test.RunAndMatch()
}

func TestIndexingOrderFuncs(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddNamedFile("/foo/A.php", `<?php
function a() {}

a();
`)
	test.AddNamedFile("/bar/A.php", `<?php
function a($x) {}

a(1);
`)

	// TODO: this test should give no warnings.
	// Right now we ensure that the linter output doesn't depend
	// on the file traversal order, but this warnings is still out of place.
	test.Expect = []string{
		`Too few arguments for a`,
	}
	test.RunAndMatch()
}
