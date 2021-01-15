# Writing new IDE/editor plugin

> **Note**: only available in version **0.3.0** and below.

NoVerify implements [Language Server Protocol](https://langserver.org) for PHP, so you can write own extension for your
IDE or editor.

Use the following command to run `noverify` as language server:

```sh
$ noverify check -lang-server -cores=4
```

## PHP language server features

- Partial auto-complete for variable names, constants, functions, object properties and methods
- All reports from noverify in lint mode
- Go to definition for constants, functions, classes, methods
- Find usages for constants, functions, methods
- Show variable types on hover
