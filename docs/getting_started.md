# Getting started

In this short tutorial, we'll clone a test project and run NoVerify on it.


## Step 0 — install NoVerify if you haven't

[Installation](/docs/install.md). The easiest way is just to download a ready binary.


## Step 1 — clone a test project

Clone a repository `swiftmailer`:
```bash
git clone https://github.com/i582/swiftmailer.git
cd swiftmailer
```

## Step 2 — install dependencies from `composer.json`

```bash
composer install
```

We need to install all the dependencies so that in the future NoVerify can find the definitions of functions and classes for correct analysis.

If you are using Windows and you have encountered errors during installation, then try running the command with the `--ignore-platform-reqs` flag.

```bash
composer install --ignore-platform-reqs
```

> Without a valid `vendor` folder, NoVerify can generate many false positives.

## Step 3 — `noverify check`

Just run

```bash
noverify check ./lib
```

This will lead to an errors:

```
...
<critical> WARNING strictCmp: Non-strict string comparison (use ===) at swiftmailer/lib/classes/Swift/Signers/DomainKeySigner.php:417
        $nofws = ('nofws' == $this->canon);
                  ^^^^^^^^^^^^^^^^^^^^^^^
<critical> WARNING parentConstructor: Missing parent::__construct() call at swiftmailer/lib/classes/Swift/Attachment.php:27
    public function __construct($data = null, $filename = null, $contentType = null)
                    ^^^^^^^^^^^
2021/07/08 16:13:19 Found 119 critical and 10 minor reports
```

From the errors, you can understand on which lines NoVerify gives errors, and also understand what kind of error it is. Also, you may notice that the errors occurred in different files.

This run will analyze all files from the `./lib` folder, and it will also index the `./vendor` folder and take function and class definitions from it for analyze.

## Step 5 — let's try something

As you can see NoVerify found quite a few bugs.

### Disable or enable checks

We have quite a few `unused` errors, let's disable them.

```bash
noverify check --exclude-checks='unused' ./lib
```

Let's run a analyze for just one check. For example with `strictCmp`.

```bash
noverify check --allow-checks='strictCmp' ./lib
```

Now we only see `strictCmp` errors.

### Autofixes

NoVerify found a single place to rewrite, let's run just the `assignOp` check to see only those.

```bash
noverify check --allow-checks='assignOp' ./lib
```

Only one error were found.

```
MAYBE   assignOp: Could rewrite as `$compoundLevel ??= $this->getCompoundLevel($children)` at swiftmailer/lib/classes/Swift/Mime/SimpleMimeEntity.php:301
        $compoundLevel = $compoundLevel ?? $this->getCompoundLevel($children);
        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
```

Now let's fix them. For some checks, NoVerify can automatically fix found errors.

Run the following command for automatic fix.

```bash
noverify check --allow-checks='assignOp' --fix ./lib
```

NoVerify will fix the errors and if you run the check again:

```bash
noverify check --allow-checks='assignOp' ./lib
```

No errors will be found.

### Enable all checks

Some of the checks are disabled by default, let's run NoVerify with them. The `undefined` check can give a lot of errors, so let's turn it off.

```bash
noverify check --allow-all-checks --exclude-checks='undefined' ./lib
```

### Specify unused variable regexp

If you run a check for `unused`, you will see quite a few errors. 

```bash
noverify check --allow-checks='unused' ./lib
```

But if you look at them, you can see that most of them are variables named `$null`. Perhaps this is a way to show that the variable is not being used.

We need to match the name `null`, so a simple `^null$` regex will suffice.

Let's redefine the regex and run the analysis.

```bash
noverify check --unused-var-regex='^null$' --allow-checks='unused' ./lib
```

Now NoVerify only finds variables that do not match the regex.

The variable named `$e` is also not used in many places, it can also be disabled, but this may not be very good, since the name `$e` can be used elsewhere.

If we run a check:

```bash
noverify check --unused-var-regex='^null$|^e$' --allow-checks='unused' ./lib
```

Then only a single place will be found where the declared variable is not really used.

```
<critical> WARNING unused: Variable $name is unused (use $_ to ignore this inspection or specify --unused-var-regex flag) at swiftmailer/lib/classes/Swift/Mailer.php:73
            foreach ($message->getTo() as $address => $name) {
                                                      ^^^^^
```

In order to fix it, it is enough to rename the variable to `$null`.


## Step 5 — further reading: console options, etc.

You can read about other possible options for configuring the analysis on the [Configuration](/docs/configuration.md) page.

This project will also come in handy when you start reading the [Baseline mode](/docs/baseline.md) page.



