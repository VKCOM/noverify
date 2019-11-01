# Using NoVerify as language server for VSCode

There is no official extension for VS Code that supports this mode, so you will need to take, for example, https://marketplace.visualstudio.com/items?itemName=felixfbecker.php-intellisense VS Code extension and replace `extension.js` to the one provided in this repo.

For example, execute the following after VS Code installation:

```sh
$ vim extension.js # replace /path/to/cache and /path/to/phpstorm-stubs to proper values
$ cp extension.js ~/.vscode/extensions/felixfbecker.php-intellisense-*/out/extension.js
```

After you reload VS Code, you should get NoVerify started as a language server.
