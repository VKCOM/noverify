package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestCatchDup(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyException1 extends Exception {}
class MyException2 extends Exception {}

try {
} catch (MyException1 $e) {
} catch (MyException2 $e) {
} catch (MyException1 $e) {
}
`)
	test.Expect = []string{`Duplicated catch on \MyException1`}
	test.RunAndMatch()
}

func TestCatchOrderThrowable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
try {
} catch (Throwable $e) {
} catch (Exception $e) {
}
`)
	test.Expect = []string{`Catch \Exception block will never run as it implements \Throwable which is caught above`}
	test.RunAndMatch()
}

func TestCatchOrderExtends(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyException extends Exception {}

try {
} catch (Exception $e) {
} catch (MyException $e) {
}
`)
	test.Expect = []string{
		`Catch \MyException block will never run as it extends \Exception which is caught above`,
	}
	test.RunAndMatch()
}

func TestCatchOrderExtends2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class ExceptionBase extends Exception {}
class ExceptionDerived extends ExceptionBase {}

try {
} catch (ExceptionBase $e) {
} catch (ExceptionDerived $e) {
}
`)
	test.Expect = []string{
		`Catch \ExceptionDerived block will never run as it extends \ExceptionBase which is caught above`,
	}
	test.RunAndMatch()
}

func TestCatchOrderExtendsGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class ExceptionBase extends Exception {}
class ExceptionDerived extends ExceptionBase {}

try {
} catch (ExceptionDerived $e) {
} catch (ExceptionBase $e) {
}
`)
}

func TestCatchOrderImplements(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface CustomException {}

class ExceptionBase extends Exception implements CustomException {}
class ExceptionDerived extends ExceptionBase {}

try {
} catch (CustomException $e) {
} catch (ExceptionBase $e) {
} catch (ExceptionDerived $e) {
}
`)
	test.Expect = []string{
		`Catch \ExceptionBase block will never run as it implements \CustomException which is caught above`,
		`Catch \ExceptionDerived block will never run as it implements \CustomException which is caught above`,
	}
	test.RunAndMatch()
}

func TestCatchOrderImplementsGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
interface CustomException {}

class ExceptionBase extends Exception implements CustomException {}
class ExceptionDerived extends ExceptionBase {}

try {
} catch (ExceptionDerived $e) {
} catch (ExceptionBase $e) {
} catch (CustomException $e) {
}
`)
}

func TestTryCatchVariables(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class ExceptionDerived extends Exception {}

function f() {
	try {
		$a = 100;
		$c = 100;
		$d = 100;
	} catch (ExceptionDerived $_) {
		$a = 200;
		$d = 200;
	} catch (Exception $_) {
		$a = 200;
	} finally {
		$b = 100;
	}

	echo $a; // ok
	echo $b; // from finally, ok
	echo $c; // might not defined
	echo $d; // might not defined (not all catches)

	try {
		$e = 100;
		$f = 100;
	} catch (Exception $_) {
		$e = 200;
	}

	echo $e; // ok
	echo $f; // might not defined

	try {
		$g = 100;
	} finally {
		$g = 200;
		$h = 200;
	}

	echo $g; // ok
	echo $h; // ok
}
`)
	test.Expect = []string{
		`Possibly undefined variable $c`,
		`Possibly undefined variable $d`,
		`Possibly undefined variable $f`,
	}
	test.RunAndMatch()
}

func TestTryCatchVariablesWithExit(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class ExceptionDerived extends Exception {}

function f() {
	try {
		$a = 100;
		$b = 200;
		return;
	} catch (ExceptionDerived $_) {
		$a = 200;
		$c = 200;
	} catch (Exception $_) {
		$a = 200;
		$b = 200;
		$c = 200;
	}

	echo $a; // ok
	echo $b; // might not defined
	echo $c; // ok, try end with return

	try {
		$d = 100;
		return;
	} catch (ExceptionDerived $_) {
		$d = 200;
		$e = 100;
		return;
	} catch (Exception $_) {
		$d = 200;
		return;
	} 

	echo $d; // not defined
	echo $e; // not defined

	try {
		$f = 100;
	} finally {
		$g = 100;
		return;
	}

	echo $f; // ok
	echo $g; // not defined
}
`)
	test.Expect = []string{
		`Possibly undefined variable $b`,
		`Unreachable code`,
		`Cannot find referenced variable $d`,
		`Cannot find referenced variable $e`,
		`Possibly undefined variable $g`,
	}
	test.RunAndMatch()
}
