# Getting started

In this short tutorial, we'll clone a dummy project and run NoVerify on it.


## Step 0 — install NoVerify if you haven't

[Installation](/docs/install.md). The easiest way is just to download a ready binary.


## Step 1 — clone a dummy project

Clone a repository `noverify-dummy-project`:
```bash
git clone https://github.com/vkcom/noverify-dummy-project
cd noverify-dummy-project
```

## Step 2 — install dependencies from `composer.json`

```bash
composer install
```

We need to install all the dependencies so that in the future NoVerify can find the definitions of functions and classes for correct analysis.

## Step 3 — `noverify check`

Just run

```bash
noverify check ./src
```

This will lead to an errors:

```
<critical> ERROR   undefined: Call to undefined method {\Foo}->method() at error.php:12
  echo $a->method();
           ^^^^^^
<critical> ERROR   undefined: Call to undefined method {\Foo}->method() at error.php:12
  echo $a->method();
           ^^^^^^
```

Why did this error appear? From the errors, you can understand on which lines NoVerify gives errors, and also understand what kind of error it is. Also, you may notice that the errors occurred in different files.

This run will analyze all files from the `./src` folder, and it will also index the `./vendor` folder and take function and class definitions from it.


## Step 4 — further reading: console options, etc.

They can be found in the [Configuration](/docs/configuration.md) page.

