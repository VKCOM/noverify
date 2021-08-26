# Diff mode

Just like [baseline](/docs/baseline.md), diff mode is used to analyze not the whole project at once, but only new code. Diff mode works based on `git` capabilities, which means that if you are using a different version control system, then you will not be able to use this mode.

## How it works

To use diff mode, it is advisable to understand how it works.

The principle of mode is as follows:

- When you create a new branch from the `master` and make changes to it and then run the analysis, the linter does the following:
  - Analyzes the `master` in the state from which the branch was created and collects all reports
  - Analyzes the current state of the branch and collects all reports
  - Finds reports that are not in the reports of the master and gives only them

Thus, the linter will display only those reports that appeared in the new or changed code from the branch.

## Usage

Let's take a look at how to use it. In contrast to the baseline mode, the use here is not trivial.

In [Getting started](/docs/getting_started.md) we used a test project, let's run diff mode in it.

First of all, let's create a `noverify.sh` script with the following content:

```bash
#!/bin/sh

# Preparing the beginning and end of the commits that will be analyzed
commit_end=`git rev-parse --abbrev-ref HEAD`
commit_begin=`git rev-parse -q --verify origin/$commit_end`

if [ -z "$commit_begin" ]; then
    commit_begin=ORIGIN_MASTER
fi

# Call noverify
noverify check\
    --git=.git\
    --git-commit-from=$commit_begin\
    --git-commit-to=$commit_end\
    --git-work-tree=.\
```

This script will prepare all the required arguments and run NoVerify.

Now run the following command:

```shell
sh noverify.sh
```

You should see the following:

```
12:47:55 Started
12:47:56 merge base between ORIGIN_MASTER and e2806f133ddf3a9fbc9ba60b8b45ae3bd230c875 is e2806f133ddf3a9fbc9ba60b8b45ae3bd230c875
12:47:56 merge base between ORIGIN_MASTER and master is e2806f133ddf3a9fbc9ba60b8b45ae3bd230c875
12:47:56 Indexing complete in 209.678292ms
12:47:56 Parsed old files versions for 6.794541ms
12:47:56 Indexed files versions for 43.191042ms
12:47:56 Parsed new file versions in 7.226459ms
12:47:56 Computed reports diff for 1.25µs
12:47:56 No issues found. Your code is perfect.
```

As you can see, no errors were found, since we are analyzing the `master` without changes.

Let's create a new branch as if we want to change something:

```shell
git checkout -b "fix_important_bug"

# Current state:
master
--- 0 ---- 1
           | 
           2
           fix_important_bug 

```

And add a new file:

```shell
vim test_file.php
```

With the following content:

```php
<?php
  
function f() {
  $a = 100; // Variable $a is not used
}
```

Save and try to run the linter again:

```shell
sh noverify.sh
```

The linter will now find a new error:

```
<critical> WARNING unused: Variable $a is unused (use $_ to ignore this inspection or specify --unused-var-regex flag) at /Users/petrmakhnev/swiftmailer/test_file.php:4
  $a = 100;
  ^^
```

If we create a commit:

```shell
git add ./test_file.php
git commit -m "fix"

# Current state:
master
--- 0 ---- 1
           | 
           2 ---- 3
           fix_important_bug 
```

Then the linter will also continue to find the error above. The behavior for uncommitted files is determined by the `--git-include-untracked` flag, which is `true` by default (See [Flags](#diff-mode-flags) section).

Let's say we understand that there is a mistake and it needs to be corrected. Let's fix the file `test_file.php`:

```php
<?php
  
function f() {
  $a = 100;
  echo $a;
}
```

And create a commit:

```shell
git add ./test_file.php
git commit -m "fix of fix"

# Current state:
master
--- 0 ---- 1
           | 
           2 ---- 3 ---- 4
           fix_important_bug 
```

Save and try to run the linter again:

```shell
sh noverify.sh
```

Now the linter will not find errors, as the previous error was fixed.

<p><br></p>

As you can see, the linter analyzes the latest state in the branch and compares it to the `master` from which this branch was created.

Let's consider a situation when during the development of our branch, the `master` branch was updated:

```shell
# Current state:
master
--- 0 ---- 1 ---- 5 ---- 6
           | 
           2 ---- 3 ---- 4
           fix_important_bug 
```

As you can see a few commits have been added to the `master` branch, however, the linter will still compare against the `master`  at commit 1.

If you want it to compare with the last commit from the `master`, then you need to do `git merge`:

```shell
# Current state:
master
--- 0 ---- 1 ---- 5 ---- 6
           |             |
           2 ---- 3 ---- 4 ---- 5
           fix_important_bug 
```

And in this case, commits 5 and 6 will be compared.

After the branch is merged with the `master`, then for the new branch everything starts over and all previous changes are not taken into account.

```shell
# Current state:
master
--- 0 ---- 1 ---- 5 ---- 6 ---- 7 ---- 8 --- 9
           |             |      |            |
           2 ---- 3 ---- 4 ---- 5            10 ---- 11
           fix_important_bug                 fix_of_fix_important_bug
```

This will analyze the `master` at commit 9 and the branch at commit 11.

Along with flags specifically for diff mode, you can use all other flags as in normal mode.

## Diff mode flags

### `--git-include-untracked` (enabled by default)

This flag enables the analysis of new and uncommitted files.

Thus, if you added or changed a file, but did not commit it, then with this flag the linter will analyze it.

### `--git-work-tree`

This flag sets the working directory in which the `--git-include-untracked` flag will be taken into account.

Thus, if it is set only for a certain folder, then changes in new or uncommitted files will be checked only if they are inside this folder.

### `--git-commit-from` и  `--git-commit-to`

These flags are responsible for the range of commits for analysis.

### `--git-skip-fetch`

This flag disables the automatic fetching of the `master: ORIGIN_MASTER`.

This flag is used if you have previously executed it, for example with the command:

```shell
git fetch --no-tags -q origin master:ORIGIN_MASTER
```

### `--gitignore`

This flag enables the use of  .gitignore` files to parse files that were not excluded by them.

### `--git-author-whitelist`

This flag sets a comma-separated list of authors for which the linter will run.

Thus, if this list contains a specific user, then the linter will be launched only on the changes made by him.

### `--git-disable-compensate-master`

This flag turns off the standard behavior and allows you to parse changes directly between commits in `--git-commit-from` and `--git-commit-to`.

