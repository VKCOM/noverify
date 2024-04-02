# phpgrep

[![Go Report Card](https://goreportcard.com/badge/github.com/quasilyte/phpgrep)](https://goreportcard.com/report/github.com/quasilyte/phpgrep)
[![GoDoc](https://godoc.org/github.com/quasilyte/phpgrep?status.svg)](https://godoc.org/github.com/quasilyte/phpgrep)
[![Build Status](https://travis-ci.org/quasilyte/phpgrep.svg?branch=master)](https://travis-ci.org/quasilyte/phpgrep)

> TODO: clarify that this is a customized for integration fork of the [phpgrep](https://github.com/quasilyte/phpgrep).

Syntax-aware grep for PHP code.

This repository is used for the library and command-line tool development.
A good source for additional utilities and ready-to-run recipes is [phpgrep-contrib](https://github.com/quasilyte/phpgrep-contrib) repository.

## Overview

`phpgrep` is both a library and a command-line tool.

Library can be used to perform syntax-aware PHP code matching inside Go programs
while binary utility can be used from your favorite text editor or terminal emulator.

It's very close to [structural search and replace](https://www.jetbrains.com/help/phpstorm/structural-search-and-replace.html)
in PhpStorm, but better suited for standalone usage.

In many ways, inspired by [github.com/mvdan/gogrep/](https://github.com/mvdan/gogrep/).

See also: ["phpgrep: syntax aware code search"](https://speakerdeck.com/quasilyte/phpgrep-syntax-aware-code-search).

## Quick start

To install `phpgrep` binary under your `$(go env GOPATH)/bin`:

```bash
go install -v github.com/quasilyte/phpgrep/cmd/phpgrep@latest
```

If `$GOPATH/bin` is under your system `$PATH`, `phpgrep` command should be available after that.<br>
This should print the help message:

```bash
$ phpgrep -help
Usage: phpgrep [flags...] target pattern [filters...]
Where:
  flags are command-line flags that are listed in -help (see below)
  target is a file or directory name where search is performed
  pattern is a string that describes what is being matched
  filters are optional arguments bound to the pattern

Examples:
  # Find f calls with a single variable argument.
  phpgrep file.php 'f(${"var"})'
  # Like previous example, but searches inside entire
  # directory recursively and variable names are restricted
  # to $id, $uid and $gid.
  # Also uses -v flag that makes phpgrep output more info.
  phpgrep -v ~/code/php 'f(${"x:var"})' 'x=id,uid,gid'

Exit status:
  0 if something is matched
  1 if nothing is matched
  2 if error occured

# ... rest of output
```

Create a test file `hello.php`:

```php
<?php
function f(...$xs) {}
f(10);
f(20);
f(30);
f($x);
f();
```

Run `phpgrep` over that file:

```bash
# phpgrep hello.php 'f(${"x:int"})' 'x!=20'
hello.php:3: f(10)
hello.php:5: f(30)
```

We found all `f` calls with a **single** argument `x` that is `int` literal **not equal** to 20.

Next thing to learn is `${"*"}` matcher.

Suppose you need to match all `foo` function calls that have `null` argument.<br>
`foo` is variadic, so it's unknown where that argument can be located.

This pattern will match `null` arguments at any position: `foo(${"*"}, null, ${"*"})`.

Read [pattern language docs](/pattern_language.md) to learn more about how to write search patterns.

## Recipes

This section contains ready-to-use `phpgrep` patterns.

`srcdir` is a target source directory (can also be a single filename).

### Useful recipes

```bash
# Find arrays with at least 1 duplicated key.
$ phpgrep srcdir '[${"*"}, $k => $_, ${"*"}, $k => $_, ${"*"}]'

# Find where ?: can be applied.
$ phpgrep srcdir '$x ? $x : $y' # Use `$x ?: $y` instead

# Find potential operator precedence issues.
$ phpgrep srcdir '$x & $mask == $y' # Should be ($x & $mask) == $y
$ phpgrep srcdir '$x & $mask != $y' # Should be ($x & $mask) != $y

# Find calls where func args are misplaced.
$ phpgrep srcdir 'stripos(${"str"}, $_)'
$ phpgrep srcdir 'explode($_, ${"str"}, ${"*"})

# Find new calls without parentheses.
$ phpgrep srcdir 'new $t'

# Find all if statements with a body without {}.
$ phpgrep srcdir 'if ($cond) $x' 'x!~^\{'
# Or without regexp.
$ phpgrep srcdir 'if ($code) ${"expr"}'

# Find all error-supress operator usages.
$ phpgrep srcdir '@$_'
```

### Miscellaneous recipes

```bash
# Find all function calls that have at least one var-argument that has _id suffix.
$ phpgrep srcdir '$f(${"*"}, ${"x:var"}, ${"*"})' 'x~.*_id$'
```
