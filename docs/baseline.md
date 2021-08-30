# Baseline mode

When launching a linter for the first time on a large project where the linter has never been used before, you will see a huge number of errors. It is impossible and not always necessary to fix them all. In order to start from scratch, it is necessary to suppress all existing errors and show only the errors that appeared after.

For this, the **Baseline mode** is used. It creates a map of errors and suppresses them, showing only new ones.

In [Getting started](/docs/getting_started.md), we already ran NoVerify on a test project. There were about 130 errors in there, so let's create a baseline file to suppress them.

Run the following command:

```bash
noverify check --output-baseline --output='baseline.json' ./lib
```

`baseline.json` with an error map will be created in the project folder.

Now, to run the analysis with this map in mind, use the `--baseline` flag:

```bash
noverify check --baseline='baseline.json' ./lib
```

After starting, you will see that no errors were found.

Let's add the following lines to the `__construct` function in the `swiftmailer/lib/classes/Swift/Mime/SimpleMimeEntity.php` file to check that the linter only shows an error in them:

```php
$a = [1];
echo $a[count($a)];
```

And run the analysis:

```bash
noverify check --baseline='baseline.json' ./lib
```

NoVerify will only find a new error:

```
<critical> WARNING offBy1: Probably intended to use count-1 as an index at swiftmailer/lib/classes/Swift/Mime/SimpleMimeEntity.php:98
        echo $a[count($a)];
             ^^^^^^^^^^^^^
```

<p><br></p>

## Conservative baseline

Baseline mode cannot be 100% accurate, so sometimes there can be situations where adding code to a file will result in errors being found in old code that has already been suppressed. If this happens to you often, then try using a conservative baseline.

To do this, add the `--conservative-baseline` flag in all the previous commands.

> The baseline file also needs to be regenerated, since a baseline file from normal mode will not work for conservative mode.

