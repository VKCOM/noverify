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
	test.Expect = []string{`duplicated catch on \MyException1`}
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
	test.Expect = []string{`catch \Exception block will never run as it implements \Throwable which is caught above`}
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
	test.Expect = []string{`catch \MyException block will never run as it extends \Exception which is caught above`}
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
	test.Expect = []string{`catch \ExceptionDerived block will never run as it extends \ExceptionBase which is caught above`}
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
		`catch \ExceptionBase block will never run as it implements \CustomException which is caught above`,
		`catch \ExceptionDerived block will never run as it implements \CustomException which is caught above`,
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
