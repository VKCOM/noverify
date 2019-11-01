## NoVerify [![Build Status](https://travis-ci.org/VKCOM/noverify.svg?branch=master)](https://travis-ci.org/VKCOM/noverify)

NoVerify is a PHP linter: it finds possible bugs and style violations in your code.

* NoVerify has no config: any reported issue in your PHPDoc or PHP code must be fixed.
* NoVerify aims to understand PHP code at least as well as PHPStorm does. If it behaves incorrectly or suboptimally, please [report issue](https://github.com/VKCOM/noverify/issues/new).
* This tool is written in [Go](https://golang.org/) and uses [z7zmey/php-parser](https://github.com/z7zmey/php-parser).

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
- [Using NoVerify as language server for Sublime Text](docs/sublime-plugin.md)
- [Writing new IDE/editor plugin](docs/writing-new-ide-plugin.md)

## Contribute

Just find [good first issue](https://github.com/VKCOM/noverify/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22), fix it and make pull request.
