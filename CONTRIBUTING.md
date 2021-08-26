# How to contribute

There are several ways to help out:

- create an [issue](https://github.com/vkcom/noverify/issues/) on GitHub in case you have found a bug or have a feature request
- write test cases for open bug issues
- write patches for open bug/feature issues

There are a few guidelines that we ask contributors to observe:

- The code must follow the Go coding standard (checked by a linter, see below).

- All commits messages should be formatted as
  ```
  pkgs: short desc
  
  A more detailed description.
  ```
  where `pkgs` is the name of the package or comma-separated packages in which the change occurred.

- All code changes should be covered by unit tests.


## A short description if you'd like to contribute by writing code

Below you'll find how to build a project and test it.

### Building

Make sure you have [Go](https://golang.org/dl/) version 1.16 or higher installed.  
Clone this repository and run `make build`:
```bash
git clone https://github.com/vkcom/noverify
cd noverify
make build
```

A resulting binary will be placed in the `./build` folder.

### Testing

The project uses standard tests provided by Go:
```bash
make test
```

It will run all tests from the `./tests` folder.

### Linting

We use [golangci-lint](https://github.com/golangci/golangci-lint). Its configuration file is located at `/.golangci.yml`.
```bash
make lint
```

This command will install the `golangci-lint` linter and run the analysis.

>  For convenience, there is a command `make check`, which runs the linter first, and then runs the tests.

### Releasing

We do not use complicated methods for releases. Each release is created manually:

- update the version the `Makefile`
- run the `make release` command, it will create archives with executable files in `./build`
- create a new release in GitHub with description, and upload the archives
