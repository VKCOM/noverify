# phpgrep pattern language

This file serves as a main documentation source for the pattern language used inside phpgrep.

We'll refer to it as PPL (phpgrep pattern language) for brevity.

## Overview

Syntax-wise, PPL is 100% compatible with PHP.

In fact, it only changes the semantics of some syntax constructions without adding any
new syntax forms to the PHP language. It means that phpgrep patterns can be parsed by
any parser that can handle PHP.

The patterns describe the program parts (syntax trees) that they need to match.
In places where whitespace doesn't mattern in PHP, it has no special meaning in PPL as well.

### PHP variables

PHP variables syntax, `$<id>` match any kind of node (expression or a statement) exactly once.

If same `<id>` is used multiple times, both "variables" should match the same AST.

```php
$x = $y; // Matches any assignment
$x = $x; // Matches only self-assignments
```

The special variable `$_` can be used to avoid having to give names to less important parts of the pattern
without additional restrictions that apply when variable names are identical.

```php
$_ = $_ // Matches any assignment (because $_ is special)
```

### Matcher expressions

Expressions in form of `${"<matcher>"}` or `${'<matcher>'}` are called **matcher expressions**.
The `<matcher>` determines what will be matched.

It does not matter whether you use `'` or `"`, both behave identically.

```
matcher_expr = "$" "{" quote matcher quote "}"
quote = "\"" | "'"
matcher = named_matcher | matcher_class
named_matcher = <name> ":" matcher_class
matcher_class = <see the table of supported classes below>
```

| Class | Description |
|---|---|
| `*` | Any node, 0-N times |
| `+` | Any node, 1-N times |
| `int` | Integer literal |
| `float` | Float literal |
| `num` | Integer or float literal |
| `str` | String literal |
| `const` | Constant, like true or a class constant like T::FOO |
| `var` | Variable |
| `func` | Anonymous function/closure expression |
| `expr` | Any expression |

Some examples of complete matcher expressions:
* `${'*'}` - matches any number of nodes
* `${"+"}` - matches one or more nodes
* `${'str'}` - matches any kind of string literal
* `${"x:int"}` - `x`-named matcher that matches any integer
* `$${"var"}` - matches any "variable variable", like `$$x` and `$$php`

Interesting details:
* Anonymous matchers get "_" name, so `${"var"}` is actually `${"_:var"}`
* Semantically, `$x` is `${"x:node"}` (but PPL doesn't define `node`)

### Filters

After pattern is matched, additional filters can be applied to either accept or reject the match.

Filters can only be applied to a **named** matchers.

```
filter = <name> operator argument
operator = <see list of supported ops below>
argument = <depends on the operator>
```

Filters are connected like a pipeline.
If the first filter failed, a second filter will not be executed and the match will be rejected.

This is an impossible filter list:

```
'x=1' 'x=2'
```

`x` is required to be equal to `1` and then it compared to `2`.

**or**-like behavior can be encoded in several operator arguments using `,`:

```
'x=1,2'
```

### Filtering operators

#### `~` filter

The `~` filter matches matched node source text against regular expression.

> Note: it uses **original** source text, not the printed AST node representation.
> It means that you need to take code formatting into account.

#### `!~` filter

Opposite of `~`. Matches only when given regexp is not matched.

#### `=` filter

| Class | Effect | Example |
|---|---|---|
| `*` | Sub-pattern | `x=[$_]` |
| `+` |  Sub-pattern | `x=${"var"}` |
| `int` | Value matching | `x=1,20,300` |
| `float` | Value matching | `x=5.6` |
| `num` | Value matching | `x=1,1.5` |
| `str` | Value matching | `x="foo","bar"` |
| `var` | Name matching | `x=length,len` |
| `expr` | Value matching | `x=1` |

Sub-pattern can include any valid PPL text.

Value and name matching accept a comma-separated lists of permitted values.
For strings you need to use quotes, so there is no problem with having `,` inside them.

#### `!=` filter

Opposite of `=`. Matches only when `=` would not match.
