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

## User Guide

- [How to install NoVerify](docs/install.md)

Using noverify as linter:
- [Using NoVerify as linter / static analyser](docs/linter-usage.md)

Using noverify as PHP [language server](https://langserver.org):
- [Using NoVerify as language server for VSCode](docs/vscode-plugin.md)
- [Using NoVerify as language server for Sublime Text](docs/sublime-plugin)
- [Writing new IDE/editor plugin](docs/writing-new-ide-plugin.md)
