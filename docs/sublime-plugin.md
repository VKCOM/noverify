# Using NoVerify as language server for Sublime Text

> **Note**: only available in version **0.3.0** and below.

You can install https://github.com/tomv564/LSP using Package Control. Here is an example config for NoVerify (replaces
phpls):

```json
{
  "clients": {
    "phpls": {
      "command": ["/path/to/noverify", "check", "-cores=4", "-lang-server"],
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
