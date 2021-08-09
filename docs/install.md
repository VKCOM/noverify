# Installation

## Ready binaries — the easiest way

Go to the [Releases](https://github.com/vkcom/noverify/releases) page and download the latest version for your OS.

Check that it launches correctly:

```bash
noverify version
```

*(here and then, we suppose that the `noverify` binary is available by name)*

You're done! Proceed to the [Getting started](/docs/getting_started.md) page.

## With `go get`

Make sure you have [Go](https://golang.org/dl/) version 1.16 or higher installed.  
Run the following command:

```sh
$ go install -u github.com/VKCOM/noverify
```

NoVerify will be installed to `$GOPATH/bin/noverify`, which usually expands to `$HOME/go/bin/noverify`.

For convenience, you can add this folder to the **PATH**.

## Build from source

Make sure you have [Go](https://golang.org/dl/) version 1.16 or higher installed.  
Clone this repository and run `make build`:

```bash
git clone https://github.com/vkcom/noverify
cd noverify
make build
```

Optionally, you can pass a name of the binary:

```bash
make build BIN_NAME=noverify.bin
```

A resulting binary will be placed in the `./build` folder.

## Next steps

- [Using NoVerify as linter / static analyser](/docs/getting_started.md)
- [Using NoVerify as language server for Sublime Text](sublime-plugin.md)
- [Using NoVerify as language server for VSCode](vscode-plugin.md)