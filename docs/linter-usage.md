# Using NoVerify as linter / static analyser

## Analyze full project

In order to get reports for all files in repository, run the following:

```sh
$ noverify check -cache-dir=$HOME/tmp/cache/noverify /path/to/your/project/root
```

The cache directory is optional, by default it is already set to `$TMPDIR/noverify`, but if you want to change the folder for storing the cache, then use the `-cache-dir` flag.

By default, "embedded" phpstorm-stubs are used.
If there is some error during the NoVerify run, like "failed to load embedded stubs", try
to provide explicit (non-empty) `-stubs-dir` argument. That argument expects a path to a cloned
phpstorm-stubs repository. You can use either the [upstream version](https://github.com/JetBrains/phpstorm-stubs) or [VKCOM fork](https://github.com/VKCOM/phpstorm-stubs) that contains
several fixes that are important for static analysis.

Running NoVerify with custom phpstorm-stubs can look like this:

```sh
$ noverify check -stubs-dir=/path/to/phpstorm-stubs /path/to/your/project/root
```

The command will print you some progress messages and reports like that:

```
ERROR   Class not found \Symfony\Component\Filesystem\Filesystem at vendor/symfony/http-kernel/Tests/KernelTest.php:35
        $fs = new Filesystem();
                  ^^^^^^^^^^
```

Command exit code will be 2 if there are reports found with non-MAYBE level.
There are several severity levels for the reports: ERROR, WARNING, INFO, HINT, UNUSED, MAYBE, SYNTAX.

## Analyze only git diff (e.g. in pre-push hook)

It is possible to only show new reports in changed code when it has been changed using git. Only changed files will be checked in this mode unless `-git-full-diff` option is specified. Changes are compared to previous commit, excluding changes made to `master` branch that is fetched to ORIGIN_MASTER.

Here is an example of command to check for changes that you are going to push:

```sh
#!/bin/sh

# Prepare git arguments
git fetch --no-tags -q origin master:ORIGIN_MASTER
ref=`git rev-parse --abbrev-ref HEAD`
prev_ref=`git rev-parse -q --verify origin/$ref`
if [ -z "$prev_ref" ]; then
    prev_ref=ORIGIN_MASTER
fi

# Call noverify
noverify check\
    -git=.git\
    -git-skip-fetch\
    -git-commit-from=$prev_ref\
    -git-commit-to=$ref\
    -git-ref=refs/heads/$ref\
    -git-work-tree=.\
    -cache-dir=$HOME/tmp/cache/noverify
    -index-only-files='generated/a.php,generated/web'
```

Here is the short summary of options used here:
 - `-git` specifies path to .git directory
 - `-git-skip-fetch` is a flag to disable automatic fetch of `master:ORIGIN_MASTER` (it is already done in this hook)
 - `-git-commit-from` and `-git-commit-to` specify range of commits to analyze
 - `-git-ref` is name of pushed branch
 - `-git-work-tree` is an optional parameter that you can specify if you want to be able to analyze uncommited changes too
 - `-cache-dir` is an optional directory for cache (greatly increases indexing speed)
 - `-index-only-files` is index-only targets (see below)

If you have files that are not a part of a git repository (i.e. they are ignored),
you need to specify those files explicitly via `-index-only-files`.

## Disable some reports

There are multiple ways to disable linter for certain files and lines:

- Write `/** @linter disable */` PHPDoc annotation in the start of a file and add this file to `-allow-disable` regex
- Add files or directories into `-exclude` regex (e.g. `-exclude='vendor/|tests/'` or `-exclude="vendor|tests"` for Windows)
- Enter `@linter disable` in a commit message to disable checks for this commit only (diff mode only).

There is also check-specific disabling mechanism. Every annotated warning can be disabled using
`-exclude-checks` argument, which is a comma-separated list of checks to be disabled.

Given this PHP file (`hello.php`):

```php
<?php
$x = array($v, 2);
```

By default, NoVerify would report 2 issues:

```sh
$ noverify check hello.php
MAYBE   arraySyntax: Use the short form '[]' instead of the old 'array()' at /home/quasilyte/CODE/php/hello.php:3
$x = array($v, 2);
     ^^^^^^^^^^^^
ERROR   undefined: Undefined variable $v at /home/quasilyte/CODE/php/hello.php:3
$x = array($v, 2);
           ^^
```

The `arraySyntax` and `undefined` are so-called "check names" which you can use to disable associated reports.

```sh
$ noverify check -exclude-checks arraySyntax,undefined hello.php
# No warnings
```

## Using in CI / using explicit checks enable list

For CI purposes it's usually more reliable to use an explicit list of checks to be executed,
so updating a linter doesn't break your build only because new checks were added.

The `-allow-checks` argument by default includes all stable checks, but you can override
it by passing your own comma-separated list of check names instead:

```sh
# Run only 2 checks, undefined and deadCode.
$ noverify check -allow-checks undefined,deadCode hello.php
```

You can use it in combination with `-exclude-checks`.
Exclusion rules are applied after inclusion rules are applied.
