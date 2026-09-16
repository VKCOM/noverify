# Type narrowing is incomplete: chained checks, null guards, and foreach/mixed edge cases

## Summary

NoVerify does not consistently narrow variable types after common PHP control-flow patterns. This leads to false positives in null-safety checks (`notNullSafetyFunctionArgumentVariable`, `notSafeCall`) even when the code is safe.

The problems are mostly in `andWalker` (condition analysis) and `handleIf` (propagating narrowed types after guard statements).

## Affected patterns

### 1. Chained conditions (`&&`) ignore previous narrowing

When several checks are combined, the right-hand side uses the wrong context and does not see types already narrowed on the left.

**Reproducer:**

```php
<?php
declare(strict_types=1);

function test(string $s): void {
    echo $s;
}

function getValue(): string|null {
    return 'hello';
}

$v = getValue();
if ($v !== null && is_string($v)) {
    test($v); // false positive: still treated as nullable
}
```

**Expected:** no warning inside the `if` body.

**Actual:** null-safety warning.

**Root cause (suspected):** in `handleTypeCheckCondition` / `handleConditionSafety`, `currentType` is taken from `falseContext` instead of the accumulated `trueContext` inside `BooleanAndExpr`.

---

### 2. Loose null comparisons are not handled

Only `=== null` / `!== null` are analyzed. `== null` / `!= null` do not narrow types.

**Reproducer:**

```php
<?php
declare(strict_types=1);

class A {}

function test(A $a): void {}

$v = null;
if ($v == null) {
    $v = new A();
    test($v); // should be safe after assignment
}
```

**Expected:** no warning after `$v = new A()`.

**Actual:** warning (same as with no narrowing).

**Root cause (suspected):** `andWalker` handles `IdenticalExpr` / `NotIdenticalExpr`, but not `EqualExpr` / `NotEqualExpr`.

---

### 3. Guard `if` with `return` / `continue` only works for `instanceof`

Early-exit guards are a common pattern:

```php
if ($v === null) {
    return; // or continue;
}
test($v); // $v should be non-null here
```

This works only when the condition is basically `instanceof` (`onlyInstanceof` flag in `handleIf`). Null checks and other narrowing are not propagated to code after the `if`.

**Reproducer:**

```php
<?php
declare(strict_types=1);

class A {
    public string $value;
}

function getA(): A|null {
    return null;
}

function test(A $a): void {
    echo $a->value;
}

function foo(): void {
    $v = getA();
    if ($v === null) {
        return;
    }
    test($v); // false positive
}
```

**Expected:** no warning.

**Actual:** `notNullSafetyFunctionArgumentVariable`.

**Root cause (suspected):** in `handleIf`, narrowed types from `falseContext` are applied only when `onlyInstanceof == true` and the `if` body always exits:

```go
if trueContext.exitFlags != 0 && onlyInstanceof && len(s.ElseIf) == 0 && s.Else == nil {
    b.ctx = falseContext
}
```

---

### 4. `foreach` with heterogeneous arrays: `mixed[]` is not narrowed

If array elements have different inferred types, NoVerify falls back to `mixed[]`. A subsequent `=== null` guard does not remove nullability.

**Reproducer:**

```php
<?php
declare(strict_types=1);

class A {
    public string $value;
}

function getA(): A|null {
    return null;
}

function test(A $a): void {
    echo $a->value;
}

foreach ([getA(), new A()] as $v) {
    if ($v === null) {
        continue;
    }
    test($v); // false positive
}
```

**Expected:** no warning (same as PHPStan/Psalm for this guard pattern with a precise array type).

**Actual:** `notNullSafetyFunctionArgumentVariable`.

**Note:** works when element types are homogeneous, e.g. `[getA(), getA()]`.

**Root cause (suspected):** `[getA(), new A()]` is inferred as `mixed[]`; `Erase("null")` does not affect `mixed`.

---

### 5. Code after `if/else` affects narrowing inside branches

Presence of a statement after an `if/else` block can change analysis inside the branches.

**Reproducer:**

```php
<?php
declare(strict_types=1);

class A {
    public string $value;
}

function test(A $a): void {}

$v = null;
if ($v === null) {
    $v = new A();
    test($v); // warning only when the line below exists
} else {
    test($v);
}
test($v);
```

**Without** the final `test($v);`: warning only in `else` (1 report).

**With** the final `test($v);`: warnings in both branches (2 reports), including after `$v = new A()`.

**Expected:** no warning in the `if` branch after assignment.

**Actual:** false positive depends on whether there is code after the whole `if/else`.

---

## Expected behavior (general)

NoVerify should narrow types similarly to PHPStan/Psalm/PHPStorm for:

| Pattern | Type after guard |
|---|---|
| `$v === null` + `return`/`continue` in `if` | non-null below |
| `$v !== null && is_string($v)` | `string` (non-null) in `if` body |
| `$v == null` / `$v != null` | same as strict comparison |
| `if (!$x instanceof T) { return; }` | `T` below (already works) |

---

## Suggested fix areas

1. **`src/linter/and_walker.go`**
   - Use accumulated `trueContext` for RHS of `&&`
   - Handle `EqualExpr` / `NotEqualExpr`
   - Unify narrowing via `VarImplicit` (like `instanceof`)

2. **`src/linter/block.go` (`handleIf`)**
   - Remove or generalize `onlyInstanceof`
   - Propagate `falseContext` narrowing after guard `if` for all narrowing conditions, not only `instanceof`
   - Merge narrowed vars into current scope (important inside loops), not only swap context pointer

3. **Foreach / `mixed`**
   - Consider PHPDoc array types
   - Or improve heterogeneous array inference instead of always falling back to `mixed[]`

---

## Environment

- NoVerify: `0.5.5` (release 0.5.5 / commit `1607147`)
- PHP: 7.x / 8.x syntax (`declare(strict_types=1)`)
- Checkers: `notNullSafetyFunctionArgumentVariable`, `notSafeCall`

---

## Suggested labels

`bug`, `type inference`, `null safety`
