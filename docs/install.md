# Installation

## Composer — the easiest way

> Can only be installed via Composer 2. See https://blog.packagist.com/composer-2-0-is-now-available/

Run the following command:

```shell
composer require --dev vkcom/noverify
```

After NoVerify is installed as a dependency, run the following command to download the binary.

```shell
./vendor/bin/noverify-get
```

The ready-to-run binary will be placed in the same folder and can be launched with the next command:

```shell
./vendor/bin/noverify
```

By default, the latest available version is downloaded, but you can install other versions, see help command for details.

For example:

```shell
./vendor/bin/noverify-get --version 0.3.0
```

### Troubleshooting

#### Composer

`vkcom/noverify` package requires the `ext-zip` extension installed on the system, if you receive an error that it is not installed, then install it with the following command (replace the version with the PHP version you need):

On Ubuntu:

```
sudo apt install php8.0-zip
```

On macOS:

```
brew update
brew install php@8.0
brew link php@8.0
brew link php@8.0 --force
```

#### noverify-get

If you get an error "not supported arch" or "not supported os", then create a new [issue](https://github.com/VKCOM/noverify/issues/new) in which describe what values the script displayed and this version can be added to releases.

#### Other

Create an [issue](https://github.com/VKCOM/noverify/issues/new) if you have any problems with the installation.

## Ready binaries

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

```shell
$ go get github.com/VKCOM/noverify
```

NoVerify will be installed to `$GOPATH/bin/noverify`, which usually expands to `$HOME/go/bin/noverify`.

For convenience, you can add this folder to the **PATH**.

## Build from source

Make sure you have [Go](https://golang.org/dl/) version 1.16 or higher installed.  
Clone this repository and run `make build`:

```shell
git clone https://github.com/vkcom/noverify
cd noverify
make build
```

Optionally, you can pass a name of the binary:

```shell
make build BIN_NAME=noverify.bin
```

A resulting binary will be placed in the `./build` folder.

## Next steps

- [Using NoVerify as linter / static analyser](/docs/getting_started.md)
- [Using NoVerify as language server for Sublime Text](sublime-plugin.md)
- [Using NoVerify as language server for VSCode](vscode-plugin.md)