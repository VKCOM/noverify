# Dynamic rules

A way to create custom NoVerify extensions without writing any Go code.

## Description

Earlier, any kind of NoVerify customization and/or extension required Go source code modification.
After that, you also had to re-compile the linter to see the changes.

Dynamic rules (or just "rules") make it possible to describe new inspections in terms of [phpgrep](https://github.com/quasilyte/phpgrep)-like patterns.

Advantages:
* You don't have to know Go to create new inspections.
* PHP developers can understand the rules written in these patterns.
* No need to re-compile a linter, rules are loaded dynamically.
* Rules are declarative and you need far less NoVerify internals knowledge to write them.

Note that you can't express every single idea with this mechanism, but they work quite well for some of them. If rules are suitable for your goal, use them.

If some feature is lacking in your opinion, [tell us about it](https://github.com/VKCOM/noverify/issues/new).

Some examples of checks that can be described with rules:
* Things related to function and method calls. Forbidden functions, argument combinations and type constraints.
* Type or value-based restrictions on operators. For example, you can discourage array comparisons with `<` and `>` operators.
* Detection of unwanted language constructions like `unset cast` and `require` (instead of `require_once`).

If your idea can be expressed with syntax pattern and some filters over it, it can be expressed via rules.

See examples of [rules](/src/cmd/embeddedrules/rules.php) used in NoVerify for inspiration.

## Quick start

Before diving into the intricacies of dynamic rules, let's run a test rule on the project from [Getting started](/docs/getting_started.md).

Let's search for Yoda style comparisons, like this:

```php
if (false === $a) {} 
```

### Step 1 — create a file with a rules

Let's name the file `rules.php` and place it in the root of the project. The file will contain the following content:

```php
<?php

/**
 * Since this is not an ordinary PHP code,
 * disable all PhpStorm inspections for this file.
 *
 * @noinspection ALL
 */
    
/**
 * @comment Reports comparisons where the literal is on the left.
 * @before  false === $a
 * @after   $a === false
 */
function yodaStyle() {
  /**
   * @maybe Yoda style comparison
   * @fix $a === false
   */
  false === $a;
}
```

### Step 2 — run the analysis with new rule

When a new rule is passed to NoVerify, it is built into the general system and issued in the general report stream. In order to specify a file with rules, the `--rules` flag is used, which accepts files with rules separated by commas, or directories with rules (not recursive, in this case, all the rules that are directly in the folder will be added).

Run the following command to view the found reports for a new check.

```bash
noverify check --allow-checks='yodaStyle' --rules='rules.php' ./lib
```

You should see 22 minor warnings:

```
...
MAYBE   yodaStyle: Yoda style comparison at /mnt/c/projects/swiftmailer/lib/classes/Swift/ByteStream/FileByteStream.php:177
        if (false === $this->seekable) {
            ^^^^^^^^^^^^^^^^^^^^^^^^^
MAYBE   yodaStyle: Yoda style comparison at /mnt/c/projects/swiftmailer/lib/classes/Swift/Signers/DKIMSigner.php:371
        } elseif (false === $len) {
                  ^^^^^^^^^^^^^^
2021/07/12 22:33:00 Found 22 minor issues.
```

If so, then congratulations, you've written an extension for NoVerify!

### Step 3 — additional

If you look at the rule, you can see that it has an `@fix` attribute, which means that NoVerify can automatically fix the found occurrences.

To do this, it is enough for the command above to add the `--fix` flag:

```bash
noverify check --allow-checks='yodaStyle' --rules='rules.php' --fix ./lib
```

And now all occurrences will be replaced with direct comparison style.

### Step 4 — description of rules

Now let's dive into the theory of dynamic rules.

## Introduction

Before we get into details, here is a terminology hint: 

- **Rule** is a PHPDoc comment describing inspection properties plus a phpgrep pattern that describes a syntax to be matched by that rule;
- **Rule file** is a set of rules or, technically speaking, a sequence of PHP functions or statements that represent these rules. Every function and statements is interpreted as a [phpgrep](https://github.com/quasilyte/phpgrep/blob/master/pattern_language.md) pattern. PHPDoc comments assign metadata that is necessary to turn patterns into inspections.

Because a rule file is a valid PHP file, you can use IDE like [PhpStorm](https://www.jetbrains.com/phpstorm/) to work with them.

NoVerify accepts such files with `--rules` command-line argument. If several files are specified, they are merged. If a folder is specified, all rules from it will be added (not recursive, only the rules that are directly in the folder are added).

Let's take a look at what rule groups are.

### Rule groups

Each rule file consists of rule groups, where each group can have any number of rules. Each group is a function whose name is used as the name of the checks within the group.

For each group, you can set the following attributes:

- `@comment` — description of the checks in the group, usually a short one sentence summary, used in the documentation of the checks;
- `@before` — an example of a code that will generate an warning, used in the documentation of the check;
- `@after` — an example of a code, in which the warning is fixed and for which no warning will be issued, used in the documentation of the check.

Each group contains a set of rules.

Each rule is a [phpgrep](https://github.com/quasilyte/phpgrep/blob/master/pattern_language.md) template. To understand the patterns, read the description of the phpgrep format [here](https://github.com/quasilyte/phpgrep/blob/master/pattern_language.md).

An example of a minimal group:

```php
/**
 * @comment Description of rules.
 * @before  code with error
 * @after   code without error
 */
function nameOfCheck() {

}
```

Please note, since the name of the function is the name of the check, then if you put the function in the namespace, then the name of the check will start in the name of the namespace.

```php
namespace api_rules;

/**
 * @comment Description of rules.
 * @before  code with error
 * @after   code without error
 */
function nameOfCheck() {

}
```

The check name will be `api_rules/nameOfCheck`. It's useful if you have several rules files and want to avoid accidental diagnostic name collisions.

### Rules

Each rule is an expression or statement ([phpdoc](https://github.com/quasilyte/phpgrep/blob/master/pattern_language.md) pattern) over which PHPDoc is written with additional meta information. Each rule should have a definition of the severity level and a report message.

The `@error`, `@warning` and `@maybe` attributes are used to set the severity level and the message.

The `error` and `warning` severity levels make the rule **critical**, that is, if the linter finds them, it will exit with a non-zero status.

An example of a minimal rule with an expression:

```php
/**
 * @warning Non-strict string comparison (use ===)
 */
$x == $y;
```

An example of a minimal rule with a statement:

```php
/**
 * @warning Potentially infinite 'for' loop
 */
for ($i = $start; $i < $length; $i--) { ${"*"};}
```

Also, each rule can have other attributes, which we will talk about bellow.

#### Autofixes (`@fix`)

In the example of rule files above, you can see the `@fix` attribute, which describes what should be replaced if a match is found. If you run the linter with the `--fix` flag, it will automatically replace the match with the pattern from `@fix`.

The `@fix` pattern uses variables from the rule.

For example:

```php
/**
 * @name countUse
 * @warning Count of elements is always greater than or equal to zero, use count($arr) == 0 instead.
 * @fix count($arr) == 0
 */
count($arr) <= 0;
```

As you can see, `$arr` here is a variable that matches any expression inside `count`, so if we want this expression to appear in the fix, then we just need to write `$arr` in the `@fix` template.

#### Constraints

Sometimes it is necessary to restrict the rule by some conditions, for example, by the type of the function argument. Next, we will consider all possible constraints.

##### `@type`

The `@type` constraint allows you to restrict a rule by the type of an expression.

Let's take a closer look at the previous example:

```php
/**
 * @name strictCmp
 * @warning 3rd argument of in_array must be true when comparing strings
 * @type string $needle
 */
in_array($needle, $_);
```

As you can see, we are setting a constraint on the type of `$needle`, now if `$needle` is other than `string` then the rule will not be matched. 

> Note that `$needle` can be both a variable and an expression, this does not affect the rule, in the case of an expression, the type of that expression will be inferred.

You can also use union types:

```php
/**
 * @name strictCmp
 * @warning 3rd argument of in_array must be true when comparing strings
 * @type string|int $needle
 */
in_array($needle, $_);
```

Sometimes it may be necessary to check that an expression has the type `not int`, for this you need to add  `!` before the type name.

For example:

```php
/**
 * @name strictCmp
 * @warning 3rd argument of in_array must be true when comparing strings
 * @type !string $needle
 */
in_array($needle, $_);
```

Now if  `@needle` is of type` string`, the rule will not be applied.

> The previous rule is for example only.

Sometimes it may be necessary to use several types. There is a special attribute `@or` for this.

> This attribute is used to create sets of constraints, below we will look at usage with other constraints.

Let's take an example:

```php
/**
 * @name strictCmp
 * @warning strings must be compared using '===' operator
 * @type string $x
 * @or
 * @type string $y
 */
$x == $y;
```

In this example, the rule will be applied if `$x` or `$y` is of type `string`.

##### `@scope`

The `@scope` constraint allows you to constraint the rule by context.

There are 3 types of contexts:

* `@scope all` — default value, the rule is applied everywhere;
* `@scope root` — rule is applied only at the top level (outside of functions and methods);
* `@scope local` — rule applies only to functions and methods.

Let's take an example:

```php
/**
 * @name requireOnce
 * @maybe use 'require_once' instead of require
 * @scope root
 */
require($_);
```

Here the rule will be applied only for `require` in the file, if `require` is in a function or method, then it will not be applied.

##### `@pure`

The `@pure` restriction allows you to restrict the rule by the absence of side effects.

Thus, if the filter says `@pure $x`, then the rule will be applied only if the expression `$x` has no side effects.

Let's take an example:

```php
/**
 * @name ternarySimplify
 * @maybe Could rewrite as '$x ?: $y'
 * @pure $x
 */
$x ? $x : $y;
```

In case the expression `$x` is, for example, a call to a function that changes a global variable, then we do not want to change it to `$x ?: $y`, as this will lead to different behavior. The rule above takes this into account and if `$x` has side effects, then the rule will not be applied.

As with `@type`, you can use `@or`:

```php
/**
 * @name ternarySimplify
 * @maybe Could rewrite as '$x ?: $y'
 * @pure $x
 * @or
 * @pure $y
 */
$x ? $x : $y;
```

In this case, it will be sufficient that either `@x` or `@y` have no side effects.

> The previous rule is for example only.

##### `@or`

As mentioned above, thanks to `@or`, we can define multiple types for `@type` and also for `@pure`.

Each `@or` closes the previous set of constraints and opens a new one, that is, you can write a set of `@type` and `@pure` and write another set after it.

For example:

```php
/**
 * @name ternarySimplify
 * @maybe Could rewrite as '$x ?: $y'
 * @pure $x
 * @type string $x
 * @or
 * @pure $y
 * @type int $y
 */
$x ? $x : $y;
```

In this case, it is necessary that `$x` has no side effects and has the `string` type, or that `$y` has no side effects and has the `int` type.

> The previous rule is for example only.

> Note that you cannot use `@or` for `@scope`.

##### `@path`

The `@path` restriction allows you to restrict the rule by file path.

Thus, the rule will be applied only if there is a substring from `@path` in the file path.

For example:

```php
/**
 * @name ternarySimplify
 * @maybe Could rewrite as '$x ?: $y'
 * @pure $x
 * @path common/
 */
$x ? $x : $y;
```

This rule will now apply only to files in the path of which the `common/` folder will be.

#### Underline location (`@location`)

For every warning that NoVerify finds, it underlines the location. However, for dynamic rules, the right place is not always underlined.

For example, if there is a rule:

```php
/**
 * @name countCallCond
 * @warning count is called on every loop iteration
 */
for ($i = 0; $i < count($a); $i++) $_;
```

Which finds the loops, where the `count` function is called at each iteration.

Then the linter can find the following errors:

```
WARNING countCallCond: count is called on every loop iteration at test.php:10
for ($i = 0; $i < count($words); $i++) {
^
WARNING countCallCond: count is called on every loop iteration at test.php:15
for ($i = 0; $i < count($matches[2]); $i++) {
^
```

But. the linter is currently underlining `for`, but the reason for reporting is in the `count` function, and we would like to underline it.

To do this, there is an attribute `@location`, which takes a variable to be pointed to.

Let's change our rule:

```php
/**
 * @name countCallCond
 * @warning count is called on every loop iteration
 * @location $a
 */
for ($i = 0; $i < count($a); $i++) $_;
```

Now the errors will look like this:

```
WARNING countCallCond: count is called on every loop iteration at test.php:10
for ($i = 0; $i < count($words); $i++) {
                        ^^^^^^
WARNING countCallCond: count is called on every loop iteration at test.php:15
for ($i = 0; $i < count($matches[2]); $i++) {
                        ^^^^^^^^^^^
```

Which, however, is still not perfect, but already shows close to the desired place.

#### Combining rules with one error message

Sometimes more than one template fits the same error message.

Let's remember our rule from the introduction:

```php
/**
 * @comment Reports comparisons where the literal is on the left.
 * @before  false === $a
 * @after   $a === false
 */
function yodaStyle() {
  /**
   * @maybe Yoda style comparison
   * @fix $a === false
   */
  false === $a;
}
```

Here we only match `false`, but if there is `true`, then we want to find it as well, and the message will not be different. The option to copy the rule and replace the expression in it is not very good, since it is duplication.

To combine the rules, use the `goto` label syntax:

```php
/**
 * @comment Reports comparisons where the literal is on the left.
 * @before  false === $a
 * @after   $a === false
 */
function yodaStyle() {
  /**
   * @maybe Yoda style comparison
   * @fix $a === false
   */
  any_identical: {
    false === $a;
    true === $a;
  }
}
```

Where all rules must be wrapped in `{}`. Thus, now any expression from the list will give a warning.

> Note that you cannot write the `@fix` attribute in this case.

#### Fuzzy matching (`@strict-syntax`)

By default, NoVerify considers some constructs to be the same, for example, `array()` and `[]`, and if the rule contains `[]`, then `array()` will also match the rule.

You can disable this behavior with the `@strict-syntax` attribute.

Some normalization rules are listed below, for convenience.

| Pattern | Matches (if there is no `@strict-syntax`) | Comment |
|---|---|---|
| `array(...)` | `array(...)`, `[...]` | Array alt syntax |
| `[...]` | `array(...)`, `[...]` | Array alt syntax |
| `list(...) =` | `list(...) =`, `[...] =` | List alt syntax |
| `[...] =` | `array(...) =`, `[...] =` | List alt syntax |
| `new T` | `new T()`, `new T` | Optional constructor parens |
| `new T()` | `new T()`, `new T` | Optional constructor parens |
| `0x1` | `0x1`, `1`, `0b1` | Int literals normalization |
| `0.1` | `0.1`, `.1`, `0.10` | Float literals normalization |
| `doubleval($x)` | `doubleval($x)`, `floatval($x)` | Func alias resolving |
| `"str"` | `"str"`, `'str'` | Single and double quotes |
| `f($x, $x)` | `f(1, 1)`, `f((1), 1)`, `f(1, (1))` | Args parens reduction |
| `[$x, $x]` | `[1, 1]`, `[(1), 1]`, `[1, (1)]` | Array item parens reduction |

There is a simple rule on how to decide whether you need fuzzy matching or not:

* If you're looking for the exact syntax, use `@strict-syntax` flag
* If you're looking for some generic code pattern, don't use `@strict-syntax` flag

## Attributes reference

Rule related attributes:

| Syntax | Description |
| ------------- | ------------- |
| `@name name` | Set diagnostic name (only outside of the function group). |
| `@error message...` | Set `severity = error` and report text to `message`. |
| `@warning message...` | Set `severity = warning` and report text to `message`. |
| `@maybe message...` | Set `severity = maybe` and report text to `message`. |
| `@fix template...` | Provide a quickfix template for the rule. |
| `@scope scope_kind` | Controls where rule can be applied. `scope_kind` is `all`, `root` or `local`. |
| `@location $var` | Selects a sub-expr from a match by a matcher var that defines report cursor position. |
| `@type type_expr $var` | Adds "type equals to" filter, applied to `$var`. |
| `@pure $var` | Adds "side effect free" filter, applied to `$var`. |
| `@or` | Add a new filter set. "Closes" the previous filter set and "opens" a new one. |
| `@strict-syntax` | Sets not to use the normalization of the same constructs. |
| `@path $substr` | If specified, the rule will only work for files that contain `$substr` in the name. |

Function related attributes:

| Syntax | Description |
| --- | --- |
| `@comment text...` | Rule documentation text, usually a short one sentence summary. |
| `@before text...` | Non-compliant code example, "before the fix". |
| `@after text...` | Compliant code example, "after the fix". |
| `@extends` | Specifies that this rule extends internal linter check. Note: when used, there is no need to set `@comment`, `@before`, `@after`. |