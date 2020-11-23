# Using NoVerify as language server for VSCode

[![Version](https://vsmarketplacebadge.apphb.com/version-short/EdgardMessias.php-noverify.svg)](https://marketplace.visualstudio.com/items?itemName=EdgardMessias.php-noverify)
[![Installs](https://vsmarketplacebadge.apphb.com/installs-short/EdgardMessias.php-noverify.svg)](https://marketplace.visualstudio.com/items?itemName=EdgardMessias.php-noverify)
[![Ratings](https://vsmarketplacebadge.apphb.com/rating-short/EdgardMessias.php-noverify.svg)](https://marketplace.visualstudio.com/items?itemName=EdgardMessias.php-noverify)

You can install https://marketplace.visualstudio.com/items?itemName=EdgardMessias.php-noverify using VS Code Marktplace.

For example, you can configure the following after VS Code installation:

```json
{
  "php-noverify.noverifyPath": "<noverify binary path>",
  "php-noverify.noverifyExtraArgs": [
    "check",
    "-cores=4"
  ]
}
```

After you reload VS Code, you should get NoVerify started as a language server.
