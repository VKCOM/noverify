# Getting started

In this short tutorial, we'll clone a test project and run NoVerify on it.


## Step 0 — install NoVerify if you haven't

[Installation](/docs/install.md). The easiest way is just to download a ready binary.


## Step 1 — clone a test project

Clone a repository `swiftmailer`:
```bash
git clone https://github.com/swiftmailer/swiftmailer.git
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
<critical> WARNING strictCmp: Non-strict string comparison (use ===) at swiftmailer/lib/classes/Swift/Signers/DomainKeySigner.php:417
        $nofws = ('nofws' == $this->canon);
                  ^^^^^^^^^^^^^^^^^^^^^^^
<critical> WARNING parentConstructor: Missing parent::__construct() call at swiftmailer/lib/classes/Swift/Attachment.php:27
    public function __construct($data = null, $filename = null, $contentType = null)
                    ^^^^^^^^^^^
```

From the errors, you can understand on which lines NoVerify gives errors, and also understand what kind of error it is. Also, you may notice that the errors occurred in different files.

This run will analyze all files from the `./lib` folder, and it will also index the `./vendor` folder and take function and class definitions from it for analyze.


## Step 5 — further reading: console options, etc.

As you can see NoVerify found quite a few bugs, on the [Configuration](/docs/configuration.md) page we will look at the possible options for configuring the analysis.

This project will also come in handy when you start reading the [Baseline mode](/docs/baseline.md) page. Based on this, the basics of the mode will be explained.

