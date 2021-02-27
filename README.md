## NoVerify 

![Build Status](https://github.com/VKCOM/noverify/workflows/Go/badge.svg)
[![Codecov](https://codecov.io/gh/i582/noverify/branch/master/graph/badge.svg)](https://codecov.io/gh/i582/noverify)

![](/docs/noverify_small.png)

NoVerify is a PHP linter: it finds possible bugs and style violations in your code.

* NoVerify has no config: any reported issue in your PHPDoc or PHP code must be fixed.
* NoVerify aims to understand PHP code at least as well as PHPStorm does. If it behaves incorrectly or suboptimally, please [report issue](https://github.com/VKCOM/noverify/issues/new).
* This tool is written in [Go](https://golang.org/) and uses [z7zmey/php-parser](https://github.com/z7zmey/php-parser).

## Features

1. Fast: analyze ~100k LOC/s (lines of code per second) on Core i7.
2. Incremental: can analyze changes in git and show only new reports. Indexing speed is ~1M LOC/s.
3. Experimental language server for VS Code and other editors that support language server protocol.
4. Auto fixes for some warnings (when -fix flag is provided).

## Default lints

NoVerify by default has the following checks:

- Unreachable code
- Array access to non-array type 
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

## User Guide

- [How to install NoVerify](docs/install.md)

Using NoVerify as linter:
- [Using NoVerify as linter / static analyser](docs/linter-usage.md)

Extending NoVerify:
- [Writing own rules quickly with PHP](docs/dynamic-rules.md)
- [Writing new checks in Go](docs/writing-checks-in-go.md)

Using NoVerify as PHP [language server](https://langserver.org):
- [Using NoVerify as language server for VSCode](docs/vscode-plugin.md)
- [Using NoVerify as language server for Sublime Text](docs/sublime-plugin.md)
- [Writing new IDE/editor plugin](docs/writing-new-ide-plugin.md)

## Contribute

Just find [good first issue](https://github.com/VKCOM/noverify/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22), fix it and make pull request.
