# NoVerify options

This page is dedicated to some technical details.

- [Console options for `check` command](#console-options-for--check--command)
  * [How to disable some checks](#how-to-disable-some-checks)
  * [How to enable all checks](#how-to-enable-all-checks)
  * [How to enable only certain checks](#how-to-enable-only-certain-checks)
  * [How to exclude some files and folders from checking](#how-to-exclude-some-files-and-folders-from-checking)
  * [How to exclude some files and folders from reports](#how-to-exclude-some-files-and-folders-from-reports)
  * [How to disable file checking without changing the launch command](#how-to-disable-file-checking-without-changing-the-launch-command)
  * [How to exclude the `vendor` folder](#how-to-exclude-the--vendor--folder)
  * [How to define a list of file extensions to be interpreted as PHP extensions](#how-to-define-a-list-of-file-extensions-to-be-interpreted-as-php-extensions)
  * [How to set regexp for unused variables](#how-to-set-regexp-for-unused-variables)
  * [How to output all errors to a file](#how-to-output-all-errors-to-a-file)
  * [How to output all errors to a `json` file](#how-to-output-all-errors-to-a--json--file)
  * [How to fix some errors in automatic mode](#how-to-fix-some-errors-in-automatic-mode)
  * [How to make a check critical](#how-to-make-a-check-critical)
  * [How to change the cache directory](#how-to-change-the-cache-directory)
  * [How to disable caching](#how-to-disable-caching)
- [Hard level options](#hard-level-options)
  * [How to use dynamic rules](#how-to-use-dynamic_rules)
  * [How to use `baseline` mode](#how-to-use--baseline--mode)
  * [How to use `git diff` mode (e.g. in pre-push hook)](#how-to-use--git-diff--mode--eg-in-pre-push-hook-)
- [Other commands](#other-commands)
  * [`checkers` command](#-checkers--command)
  * [`version` command](#-version--command)

<p><br></p>

## Console options for `check` command

A full launch command line is
```bash
noverify check --option1=xxx --option2=yyy ... [folder_or_file] [folder_or_file] ...
```

When no options are specified, their default values are used.

When no folders and files are specified, the current directory `.` is assumed.

To see all the options, run the following command:

```bash
noverify check help
```

Below we will discuss the main options.

### How to disable some checks

It looks like this:

```shell
noverify check --exclude-checks='undefined, arraySyntax' ./
```

Now there will be **no** `undefined` and `arraySyntax` errors in the output.

### How to enable all checks

It looks like this:

```shell
noverify check --allow-all-checks ./
```

NoVerify has checks that are disabled by default, you can enable them with the `--allow-all-checks` flag.

### How to enable only certain checks

It looks like this:

```shell
noverify check --allow-checks='undefined, arraySyntax' ./
```

Now the output will **only** contain `undefined` and `arraySyntax` errors.

### How to exclude some files and folders from checking

It looks like this:

```bash
noverify check --index-only-files='./tests' ./
```

The `--index-only-files` option sets paths that won't be analyzed, they will be just indexed (from there the definitions of functions and classes for type inference will be taken).

### How to exclude some files and folders from reports

It looks like this:

```bash
noverify check --exclude='./src/1.php' ./
```

The `exclude` flag accepts a regular expression based on which to exclude files or folders.

Unlike `--index-only-files`, excluded files will be analyzed and errors may be found for them, but they will not be shown.

Use this flag if you do not want to see errors for some files or directories and it is convenient to use a regular expression to express the name.

### How to disable file checking without changing the launch command

It looks like this:

```php
<?php
    
/** @linter disable */

function f() {
    ...
}

...
```

It is necessary to add the comment `/** @linter disable */` to the file.

However, to prevent developers from doing this permanently, there is the `--allow-disable` flag, which determines the files in which this annotation can be used.

It looks like this:

```shell
nocolor check --allow-disable="dev_*" ./src
```

The flag sets a regular expression to determine which files are allowed.

This is usually needed if you use NoVerify in a pipeline, where changing launch commands many times in a row is not effective, but you need to give the ability to disable the linter for certain files.

### How to exclude the `vendor` folder

It looks like this:

```shell
nocolor check --ignore-vendor ./src
```

By default, if NoVerify finds a `vendor` folder, then it includes it in the index to correctly deduce types and not give errors for undefined classes and functions.

If you need to disable this behavior, then use the `--ignore-vendor` flag.

### How to define a list of file extensions to be interpreted as PHP extensions

It looks like this:

```shell
nocolor check --php-extensions='php, phtml' ./src
```

By default, NoVerify analyzes the following extensions: ` php, inc, php5, phtml`.

### How to set regexp for unused variables

It looks like this:

```shell
nocolor check --unused-var-regex='^_*' ./src
```

By default, the regexp is `^_$`. 

Sometimes variables are not used for some reason, for this they can be called `$_`, in which case NoVerify will not give a warning. However, perhaps you want NoVerify to ignore the `$_name` variables too, for example, then you need to specify the regular expression `$_*`.

### How to output all errors to a file

It looks like this:

```shell
nocolor check --output='reports.txt' ./src
```

All errors will be written to the `reports.txt` file.

### How to output all errors to a `json` file

It looks like this:

```shell
nocolor check --output-json --output='reports.json' ./src
```

All errors will be written to the `reports.json` file.

### How to fix some errors in automatic mode

It looks like this:

```shell
nocolor check --fix ./src
```

All errors that NoVerify can fix will be fixed.

If you want to fix only certain errors, then specify the `--allow-checks` flag. See [How to enable only certain checks](#how-to-enable-only-certain-checks)

### How to make a check critical

It looks like this:

```shell
nocolor check --critical='redundantCast' ./src
```

Now the appearance of the `redundantCast` error will cause NoVerify to exit with non-zero status.

By default, not all errors lead to a non-zero status, that is, if there are only `redundantCast` errors in reports, then NoVerify will exit with a zero status, since the `redundantCast` check is not critical.

If you need to make some check critical, then use the `--critical flag`, which accepts a comma-separated list of checks.

### How to change the cache directory

It looks like this:

```shell
nocolor check --cache-dir='./cache' ./src
```

By default, the directory is `$TMPDIR/noverify`. 

The cache is used to reduce the time it takes to collect information about function classes, etc. during the next launches.

### How to disable caching

It looks like this:

```shell
nocolor check --disable-cache ./src
```

<p><br></p>

## Hard level options

Next are the options for which you need to read the articles attached to them to understand how to use them.

### How to use dynamic rules

Dynamic rules are a way to add new checks to NoVerify without having to write Go code. Such rules are written in PHP. 

Read more in the article [Dynamic rules](/docs/dynamic_rules.md).

### How to use `baseline` mode

Baseline mode is necessary if you have a large codebase on which NoVerify finds a huge number of errors. Of course, it is impossible to fix them right away, so you need a way to ignore all found errors and analyze only errors found after.

Read more in the article [Baseline mode](/docs/baseline.md).

### How to use `git diff` mode (e.g. in pre-push hook)

Another way to use NoVerify for a large codebase, if NoVerify finds a large number of errors, is to run the linter only on new code. This mode uses `git`.

The changes are taken from the comparison with the previous commit, excluding changes made to `master` branch that is fetched to `ORIGIN_MASTER`.

Read more in the article [Diff mode](/docs/diff.md).

<p><br></p>

## Other commands

### `checkers` command

Shows a list of checks performed by NoVerify.

### `version` command

Shows the version of NoVerify.

