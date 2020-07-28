This document describes how to find and fix [false-positives](https://en.wikipedia.org/wiki/False_positives_and_false_negatives).

> Simply put, a **false positive** is a case where NoVerify is wrong and reports a valid PHP code as incorrect.<br>
> See [existing issues](https://github.com/VKCOM/noverify/issues?q=is%3Aissue+is%3Aopen+label%3Afalse-positive) for more context.

## Getting prepared

First of all, you need to install NoVerify locally, preferably by [building it from the source code](/docs/install.md).

Then you find a **target for the analysis**, a [PHP project](https://github.com/search?q=stars%3A%3E100+size%3A%3E1000+size%3A%3C10000+pushed%3A%3E2020-01-01+language%3APHP&type=Repositories&ref=advsearch&l=PHP&l=).

For example, we can choose [github.com/Seldaek/monolog](https://github.com/Seldaek/monolog) repository.

## Finding the **false-positive**

```bash
# 1. Clone monolog repository locally.
git clone https://github.com/Seldaek/monolog.git

# 2. Enter the downloaded directory
cd monolog

# 3. Run noverify over the source directory
noverify ./src/Monolog
````

After the 3rd step, you'll get a list of various warnings (sometimes called "reports").
Some of them are good, others - not so much.
We're looking for reports that look like a mistake on NoVerify part.

For example, look at these two reports:

```
ERROR accessLevel: Cannot access protected property \Monolog\Handler\ProcessableHandlerTrait->processors at ./src/Monolog/Handler/GroupHandler.php:64
        if ($this->processors) {
                   ^^^^^^^^^^
ERROR accessLevel: Cannot access protected method \Monolog\Handler\ProcessableHandlerTrait->processRecord() at ./src/Monolog/Handler/GroupHandler.php:65
            $record = $this->processRecord($record);
                             ^^^^^^^^^^^^^
```

Either the code is incorrect and it tries to access something it should not be accessing, or NoVerify doesn't get
the access level right in this context.

Your investigation should follow the warning location (`src/Monolog/Handler/GroupHandler.php:64`) and eventually lead to a
conclusion that `processors` and `processRecord` are defined inside the `ProcessableHandlerTrait` trait that `GroupHandler` **uses**.

When you **use** a trait, you can access its private and protected members, so NoVerify is not correct here.

[issues-209](https://github.com/VKCOM/noverify/issues/209) is an example of how such **false-positive** ticket may look like.

The main parts are:
* Code that reproduces the problem (prefer minimal reproducers, the smaller and simpler - the better).
* The incorrect linter report (in this case "cannot access protected/private method").
* The expected behavior (in case of **false-positive** it's "no reports").

## Fixing the false positive

After relevant issue is created, you can submit a code that fixes a problem.

It's suggested to start by reproducing an issue inside a testing context.

[src/linttest/regression_test.go](/src/linttest/regression_test.go) file contains a lot of test examples.

If a related issue already contains a reproducer (as it should), it's easy to add a first test case.
Following the [issue-209](https://github.com/VKCOM/noverify/issues/209) example, it could look like this:

```go
func TestIssue209(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
trait A {
  private function priv() { return 1; }
  protected function prot() { return 2; }
  /** @return int */
  public function pub() { return 3; }
}

class B {
  use A;
  /** @return int */
  public function sum() {
    return $this->priv() + $this->prot() + $this->pub();
  }
}

echo (new B)->sum(); // actual PHP prints 6
`)
}
```

Then you run the tests by using a `go test` command:

```bash
go test github.com/VKCOM/noverify/src/linttest
```

> If you're inside `noverify` directory, one can run `go test ./src/linttest` instead.

Your test **should** fail:

```
--- FAIL: TestIssue209 (0.00s)
    linttest.go:101: unexpected number of reports: expected 0, got 2
    linttest.go:124: unexpected report 0: ERROR   accessLevel: Cannot access private method \A->priv() at _file0.php:13
            return $this->priv() + $this->prot() + $this->pub();
                          ^^^^
    linttest.go:124: unexpected report 1: ERROR   accessLevel: Cannot access protected method \A->prot() at _file0.php:13
            return $this->priv() + $this->prot() + $this->pub();
                                          ^^^^
    linttest.go:135: >>> issues reported:
    linttest.go:137: ERROR   accessLevel: Cannot access private method \A->priv() at _file0.php:13
            return $this->priv() + $this->prot() + $this->pub();
                          ^^^^
    linttest.go:137: ERROR   accessLevel: Cannot access protected method \A->prot() at _file0.php:13
            return $this->priv() + $this->prot() + $this->pub();
                                          ^^^^
    linttest.go:139: <<<
FAIL
FAIL	github.com/VKCOM/noverify/src/linttest	1.174s
FAIL
```

Now you need to change NoVerify in a way that makes that test succeed.
**90%** of times you want to modify the [src/linter](/src/linter) package.

It's a good idea to start from grepping the NoVerify source code for the report message text
to quickly find a responsible code part (the one you probably need to change or at least understand).

Luckily, NoVerify checks are organized in named groups that are used as a report message prefix.
See that "accessLevel" prefix? It's the check group name.

```bash
grep -nr '"accessLevel"'
src/linter/report.go:40:			Name:    "accessLevel",
src/linter/block.go:977:		b.r.Report(e.Method, LevelError, "accessLevel", "Cannot access %s method %s->%s()", fn.AccessLevel, implClass, methodName)
src/linter/block.go:1025:		b.r.Report(e.Call, LevelError, "accessLevel", "Cannot access %s method %s::%s()", fn.AccessLevel, implClass, methodName)
src/linter/block.go:1076:		b.r.Report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s->%s", info.AccessLevel, implClass, id.Value)
src/linter/block.go:1107:		b.r.Report(e.Property, LevelError, "accessLevel", "Cannot access %s property %s::$%s", info.AccessLevel, implClass, sv.Name)
src/linter/block.go:1193:		b.r.Report(e.ConstantName, LevelError, "accessLevel", "Cannot access %s constant %s::%s", info.AccessLevel, implClass, constName.Value)
src/linttest/oop_test.go:480:	RunFilterMatch(test, "accessLevel")
```

> Note that you need to enter the NoVerify directory before running `grep`.

We've found several locations inside `src/linter/block.go` that report that problem.
They are a good starting point for the investigation.

Suppose you think that the issue is fixed. Re-run all tests:

```
go test github.com/VKCOM/noverify/src/linttest
ok  	github.com/VKCOM/noverify/src/linttest	1.244s
```

If you see this `ok` message, all tests are passed and you're golden.
Send a pull request and celebrate your awesomeness.

Your change can introduce new bugs that you may overlook during the development, this is why
it's important to run all tests as opposed to running only your new tests.

Here are some examples of commits that fixed false positives:
* [`59f5c8b9b55c03ddd936480f402ee0556b7b442a`](https://github.com/VKCOM/noverify/commit/59f5c8b9b55c03ddd936480f402ee0556b7b442a) fixes [#362](https://github.com/VKCOM/noverify/issues/362)
* [`fe953e76461d2bf97c957ee88206832526c56c2b`](https://github.com/VKCOM/noverify/commit/59f5c8b9b55c03ddd936480f402ee0556b7b442a) fixes [#183](https://github.com/VKCOM/noverify/issues/183)
* [`9271f20d8a094b4bdb38a95f388571fcd1d33f54`](https://github.com/VKCOM/noverify/commit/9271f20d8a094b4bdb38a95f388571fcd1d33f54) fixes [#182](https://github.com/VKCOM/noverify/issues/182)
