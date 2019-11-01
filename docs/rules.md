## Dynamic rules

A way to create custom NoVerify extensions without writing any Go code.

> Warning: this feature is very new. Changes are imminent!

### Description

Earlier, any kind of NoVerify customization and/or extension required Go source code modification.
After that, you also had to re-compile the linter to see the changes.

Dynamic rules (or just "rules") make it possible to describe new inspections in terms of [phpgrep](https://github.com/quasilyte/phpgrep)-like patterns.

Advantages:
* You don't have to know Go to create new inspections.
* PHP developers can understand the rules written in these patterns.
* No need to re-compile a linter, rules are loaded dynamically.
* Rules are declarative and you need far less NoVerify internals knowledge to write them.

They might also be more stable when they'll be fully released.

Note that you can't express every single idea with this mechanism, but they work
quite well for some of them. If rules are suitable for your goal, use them.
If some feature is lacking in your opinion, [tell us about it](https://github.com/VKCOM/noverify/issues/new).

Some examples of checks that can be described with rules:
* Things related to function and method calls.
  Forbidden functions, argument combinations and type constraints.
* Type or value-based restrictions on operators. For example,
  you can discourage array comparisons with `<` and `>` operators.
* Detection of unwanted language constructions like `unset cast` and
  `require` (instead of `require_once`).

If your idea can be expressed with syntax pattern and some filters over it, it can be expressed via rules.

### Introduction

Before we get into details, here is a terminology hint: a rule is a phpdoc comment describing inspection properties plus a phpgrep pattern that describes a syntax to be matched by that rule.

The **rule file** is a set of rules or, technically speaking, a sequence of PHP statements that represent these rules. Every statement is interpreted as a phpgrep pattern. Phpdoc comments assign metadata that is necessary to turn patterns into inspections.

Because a rule file is a valid PHP file, you can use IDE like [PhpStorm](https://www.jetbrains.com/phpstorm/) to work with them.

NoVerify accepts such files with `-rules` command-line argument. If several files are specified, they are merged.

A single rules file can look like this:

```php
<?php

/**
 * Since this is not an ordinary PHP code,
 * disable all PhpStorm inspections for this file.
 *
 * @noinspection ALL
 */

/** @warning use 'count' instead of 'sizeof' */
sizeof($_);

/**
 * It should also be noted that you can document
 * your rules within the same comment.
 *
 * @warning 3rd argument of in_array must be true when comparing strings
 * @type string $needle
 */
in_array($needle, $_);
```

An example above contains 2 rules. The first rule asks to use [count](https://www.php.net/manual/ru/function.count.php) instead of [sizeof](https://www.php.net/manual/ru/function.sizeof.php).
The second rule wants you to use `$strict=true` when using [in_array](https://www.php.net/manual/ru/function.in-array.php) function with string-typed `$needle`.

The only mandatory attribute is rule **category** that combines severity level and report message text.

There are 4 categories right now: `error`, `warning`, `info` and `maybe`.<br>
`error` and `warning` makes a rules **critical**, so linter will exit with non-zero status if it matches.

All other available attributes are matching **constraints**. Constraints that can be repeated can be called **filters**.

If you look at the example again, you'll note that `in_array` uses a `@type` filter.
Because of that, a rule only matches when the type of the matcher variable is a string.
A rule can have several filters and they are connected with and-like operators: all of them must be satisfied.

If you need or-like filter connection, there is a special operator-like attribute `@or` that makes it possible
to have several **sets of filters** for a single rule. If any of these filter sets are satisfied, the match will be accepted.

Here is an example of `@or` usage:

```php
/**
 * @warning strings must be compared using '===' operator
 * @type string $x
 * @or
 * @type string $y
 */
$x == $y;
```

That constraint makes a rule match only when either of `==` operands have a `string` type.

As an example of unrepeatable constraints, there is a `@scope` attribute:

```php
/**
 * @maybe prefer require_once over require
 * @scope root
 */
require($_);

/**
 * @maybe prefer include_once over include
 * @scope root
 */
include($_);
```

It assigns a scope constraint that controls in what context that rule should be applied.

There are currently 3 kinds of scope:
* `@scope all` - default value, rule works everywhere.
* `@scope root` - run rule only on the top level (outside of functions/methods).
* `@scope local` - run rule only inside functions/methods.

When we say "unrepeatable", it means that you can't have several `@scope` attributes even
if you use `@or`. It would be shared between all filter sets.

With NoVerify builtin inspections, every issue report is prefixed with a check name, like `unused` or `undefined`.

If a rule has `@name <string>` attribute, that is as a rule report tag.

Otherwise, a rule is called **anonymous** and instead of some dull placeholder, you'll get a
`filename:line` marker, where `filename` is a rule file that defines that rule and `line` is a
line that contains the pattern being matched. It's useful not only during the rule debugging but
also for tracking the rule definition location (especially useful with multiple rule files).

Here is an example of how anonumous report can look like:

```
WARNING rules.php:28: strings must be compared using '===' at www/super_file.php:123
if (substr($w, 0, 3) == 'www') {
    ^^^^^^^^^^^^^^^^^^^^^^^^^
```

To disable a rule, just comment it out:

```php
// /** @warning use 'count' instead of 'sizeof' */
// sizeof($_);
```

Rules can match not only expressions but statements as well.

This rule, for example, finds all for loops that call `count` on every iteration:

```php
/**
 * @info count is called on every loop iteration
 */
for ($i = 0; $i < count($a); $i++) $_;
```

> We don't encourage anyone to rewrite these kinds of loops. It's just an example.

Here are possible reports for that rule:

```
INFO rules.php:13: count is called on every loop iteration at www/foo.php:693
for ($i = 0; $i < count($words); $i++) {
^
INFO rules.php:13: count is called on every loop iteration at www/foo.php:58
for ($i = 0; $i < count($matches[2]); $i++) {
^
```

There is a slight issue with this rule. The "cursor" (that `^` char in the report) highlihts the `for` statement beginning,
although it would be better to point to the `count()` position. There is a `@location` attribute for that:

```php
/**
 * @info count is called on every loop iteration
 * @location $a
 */
for ($i = 0; $i < count($a); $i++) $_;
```

Now result is different:

```
INFO rules.php:14: count is called on every loop iteration at www/foo.php:693
for ($i = 0; $i < count($words); $i++) {
                        ^^^^^^
INFO rules.php:14: count is called on every loop iteration at www/foo.php:58
for ($i = 0; $i < count($matches[2]); $i++) {
                        ^^^^^^^^^^^
```

### Working with types

TODO.

### Attributes reference

| Syntax | Description |
| ------------- | ------------- |
| `@error message...` | Set severity=error and report text to `message`. |
| `@warning message...` | Set severity=warning and report text to `message`. |
| `@info message...` | Set severity=info and report text to `message`. |
| `@maybe message...` | Set severity=maybe and report text to `message`. |
| `@scope scope_kind` | Controls where rule can be applied. `scope_kind` is `all`, `root` or `local`. |
| `@location $var` | Selects a sub-expr from a match by a matcher var that defines report cursor position. |
| `@type type_expr $var` | Adds "type equals to" filter, applied to `$var`. |
| `@or` | Add a new filter set. "Closes" the previous filter set and "opens" a new one. |

### Creating a new rule + debugging it

TODO.

### More examples

```php
/**
 * @info excessive int cast: expression is already int-typed
 * @location $x
 * @type int $x
 */
(int)$x;

/**
 * @info excessive string cast: expression is already string-typed
 * @location $x
 * @type string $x
 */
(string)$x;

/**
 * @warning array to string conversion
 * @type array $x
 */
(string)$x;
```

### Development notes

Dynamic rules features that are being developed and discussed:
* [Multi-pattern rules syntax](https://github.com/VKCOM/noverify/issues/276)
