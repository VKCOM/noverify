## NoVerify [![Build Status](https://travis-ci.org/VKCOM/noverify.svg?branch=master)](https://travis-ci.org/VKCOM/noverify)

NoVerify is a linter for PHP: it searches for potential problems in your code.
It allows to write your own rules as well and it has no config: all reports
it generates are potential errors that must be fixed, or some PHPDoc annotations
must be written.

This tool is written in Go (https://golang.org/) using PHP parser from z7zmey (https://github.com/z7zmey/php-parser).

It aims to understand PHP code at least as well as PHPStorm does, which is
not an easy task. Please open issues for any behaviour you find to be incorrect or
suboptimal.

## Features

1. Fast: analyze ~100k LOC/s (lines of code per second) on Core i7
2. Incremental: can analyze changes in git and show only new reports. Indexing speed is ~1M LOC/s.
3. Experimental language server for VS Code and other editors that support language server protocol.

## Default lints

NoVerify by default has the following checks:

- Unreachable code
- Array access to non-array type (beta)
- Too few arguments when calling a function/method
- Call to undefined function/method
- Fetching of undefined constant/class property
- Class not found
- PHPDoc is incorrect
- Undefined variable
- Variable not always defined
- Case without "break;"
- Syntax error
- Unused variable
- Incorrect access to private/protected elements
- Incorrect implementation of IteratorAggregate interface
- Incorrect array definition, e.g. duplicate keys

## Custom lints

You can write your own checks that can use type information from NoVerify
and check for complex things, e.g. enforcing that strings are compared only
using === operator. See [example](/example) folder to see some examples of custom checks. 

## Installation

In order to install NoVerify, you will need the Go toolchain (https://golang.org/).

> Optionally, you could also get PHPStorm stubs (https://github.com/JetBrains/phpstorm-stubs).

Once go is installed, you need to execute the following:

```sh
$ go get -u github.com/VKCOM/noverify
```

Your noverify binary will be located at `$GOPATH/bin/noverify`, usually this
translates to `$HOME/go/bin/noverify`.

## Usage

### Analyze full project

In order to get reports for all files in repository, run the following:

```sh
$ noverify -cache-dir=$HOME/tmp/cache/noverify /path/to/your/project/root
```

Cache dir is optional, but recommended. Next launch would be much faster with cache if you specify some cache directory.

By default, "embedded" phpstorm-stubs are used.
If there is some error during the NoVerify run, like "failed to load embedded stubs", try
to provide explicit (non-empty) `-stubs-dir` argument. That argument expects a path to a cloned
phpstorm-stubs repository. You can use either the [upstream version](https://github.com/JetBrains/phpstorm-stubs) or [VKCOM fork](https://github.com/VKCOM/phpstorm-stubs) that contains
several fixes that are important for static analysis.

Running NoVerify with custom phpstorm-stubs can look like this:

```sh
$ noverify -stubs-dir=/path/to/phpstorm-stubs -cache-dir=$HOME/tmp/cache/noverify /path/to/your/project/root
```

The command will print you some progress messages and reports like that:

```
ERROR   Class not found \Symfony\Component\Filesystem\Filesystem at vendor/symfony/http-kernel/Tests/KernelTest.php:35
        $fs = new Filesystem();
                  ^^^^^^^^^^
```

Command exit code will be 2 if there are reports found with non-MAYBE level.
There are several severity levels for the reports: ERROR, WARNING, INFO, HINT, UNUSED, MAYBE, SYNTAX.

### Analyze only git diff (e.g. in pre-push hook)

It is possible to only show new reports in changed code when it has been changed using git. Only changed files will be checked in this mode unless `-git-full-diff` option is specified. Changes are compared to previous commit, excluding changes made to `master` branch that is fetched to ORIGIN_MASTER.

Here is an example of command to check for changes that you are going to push:

```sh
#!/bin/sh

# Prepare git arguments
git fetch --no-tags -q origin master:ORIGIN_MASTER
ref=`git rev-parse --abbrev-ref HEAD`
prev_ref=`git rev-parse -q --verify origin/$ref`
if [ -z "$prev_ref" ]; then
    prev_ref=ORIGIN_MASTER
fi

# Call noverify
noverify\
    -git=.git\
    -git-skip-fetch\
    -git-commit-from=$prev_ref\
    -git-commit-to=$ref\
    -git-ref=refs/heads/$ref\
    -git-work-tree=.\
    -cache-dir=$HOME/tmp/cache/noverify
    -index-only-files='generated/a.php,generated/web'
```

Here is the short summary of options used here:
 - `-git` specifies path to .git directory
 - `-git-skip-fetch` is a flag to disable automatic fetch of `master:ORIGIN_MASTER` (it is already done in this hook)
 - `-git-commit-from` and `-git-commit-to` specify range of commits to analyze
 - `-git-ref` is name of pushed branch
 - `-git-work-tree` is an optional parameter that you can specify if you want to be able to analyze uncommited changes too
 - `-cache-dir` is an optional directory for cache (greatly increases indexing speed)
 - `-index-only-files` is index-only targets (see below)

If you have files that are not a part of a git repository (i.e. they are ignored),
you need to specify those files explicitly via `-index-only-files`.

### Disable some reports

There are multiple ways to disable linter for certain files and lines:

- Write `/** @linter disable */` PHPDoc annotation in the start of a file and add this file to `-allow-disable` regex
- Add files or directories into `-exclude` regex (e.g. `-exclude='vendor/|tests/'` or `-exclude="vendor|tests"` for Windows)
- Enter `@linter disable` in a commit message to disable checks for this commit only (diff mode only).

There is also check-specific disabling mechanism. Every annotated warning can be disabled using
`-exclude-checks` argument, which is a comma-separated list of checks to be disabled.

Given this PHP file (`hello.php`):

```php
<?php
$x = array($v, 2);
```

By default, NoVerify would report 2 issues:

```sh
$ noverify hello.php
MAYBE   arraySyntax: Use of old array syntax (use short form instead) at /home/quasilyte/CODE/php/hello.php:3
$x = array($v, 2);
     ^^^^^^^^^^^^
ERROR   undefined: Undefined variable: v at /home/quasilyte/CODE/php/hello.php:3
$x = array($v, 2);
           ^^
```

The `arraySyntax` and `undefined` are so-called "check names" which you can use to disable associated reports.

```sh
$ noverify -exclude-checks arraySyntax,undefined hello.php
# No warnings
```

### Using in CI / using explicit checks enable list

For CI purposes it's usually more reliable to use an explicit list of checks to be executed,
so updating a linter doesn't break your build only because new checks were added.

The `-allow-checks` argument by default includes all stable checks, but you can override
it by passing your own comma-separated list of check names instead:

```sh
# Run only 2 checks, undefined and deadCode.
$ noverify -allow-checks undefined,deadCode hello.php
```

You can use it in combination with `-exclude-checks`.
Exclusion rules are applied after inclusion rules are applied.

### Language server mode (experimental)

If you want to launch noverify in language server mode, launch it in your IDE/editor extension like the following:

```sh
$ noverify -lang-server -cores=4 -cache-dir=/path/to/cache
```

## Visual Studio Code integration
There is no official extension for VS Code that supports this mode, so you will need to take, for example, https://marketplace.visualstudio.com/items?itemName=felixfbecker.php-intellisense VS Code extension and replace `extension.js` to the one provided in this repo.

For example, execute the following after VS Code installation:

```sh
$ vim extension.js # replace /path/to/cache and /path/to/phpstorm-stubs to proper values
$ cp extension.js ~/.vscode/extensions/felixfbecker.php-intellisense-*/out/extension.js
```

After you reload VS Code, you should get noverify started as a language server.

## Sublime Text intergration

You can install https://github.com/tomv564/LSP using Package Control. Here is an example config for noverify (replaces phpls):

```json
{
  "clients": {
    "phpls": {
      "command": ["/path/to/noverify", "-cache-dir=/path/to/cache", "-cores=4", "-lang-server"],
      "scopes": ["source.php", "embedding.php"],
      "syntaxes": ["Packages/PHP/PHP.sublime-syntax"],
      "languageId": "php"
    }
  },
  "log_stderr": true,
  "only_show_lsp_completions": true
}
```

You can then enable `phpls` for current project and enjoy all supported features.

## Features

Language server features:
- Partial auto-complete for variable names, constants, functions, object properties and methods
- All reports from noverify in lint mode
- Go to definition for constants, functions, classes, methods
- Find usages for constants, functions, methods
- Show variable types on hover
