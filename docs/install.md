# How to install NoVerify

First you will need the Go toolchain (https://golang.org/).

Once Go installed, do the following command:

```sh
$ go get -u github.com/VKCOM/noverify
```

This command installs `noverify` into `$GOPATH/bin/noverify` (which expands into `$HOME/go/bin/noverify` by default).

Alternatively, you can build `noverify` with version info:

```sh
mkdir -p $GOPATH/src/github.com/VKCOM
git clone https://github.com/VKCOM/noverify.git $GOPATH/src/github.com/VKCOM

cd $GOPATH/src/github.com/VKCOM/noverify
make install
```

## Next steps

- [Using NoVerify as linter / static analyser](linter-usage.md)
- [Using NoVerify as language server for Sublime Text](sublime-plugin.md)
- [Using NoVerify as language server for VSCode](vscode-plugin.md)
