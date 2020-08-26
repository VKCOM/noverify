<?php

/**
 * @noinspection ALL
 * @linter       disable
 */

/**
 * @comment Report ternary expressions that can be simplified.
 * @before  $x ? $x : $y
 * @after   $x ?: $y
 */
function ternarySimplify() {
  /**
   * @maybe could replace the ternary with just $cond
   * @type bool $cond
   */
  $cond ? true : false;

  /**
   * @maybe could rewrite as `(bool)$cond`
   * @type  !bool $cond
   */
  $cond ? true : false;

  /**
   * @maybe could rewrite as `$x ?: $y`
   * @fix $x ?: $y
   * @pure $x
   */
  $x ? $x : $y;

  /**
   * @maybe could rewrite as `$x ?? $y`
   * @fix $x ?? $y
   */
  isset($x) ? $x : $y;
}

/**
 * @comment Report potential operation precedence issues.
 * @before  $x & $mask == 0; // == has higher precedence than &
 * @after   ($x & $mask) == 0
 */
function precedence() {
  // Note: we report `$x & $mask != $y` as a precedence issue even
  // if it can be caught with `typecheckOp` that checks both operand
  // types (bool is not a good operand for bitwise operation).
  //
  // Reporting `invalid types, expected number found bool` is
  // not that helpful, because the root of the problem is precedence.
  // Invalid types are a result of that.

  /** @warning == has higher precedence than & */
  any_eq_bitand: {
    $_ == $_ & $_;
    $_ & $_ == $_;
  }
  /** @warning != has higher precedence than & */
  any_neq_bitand: {
    $_ != $_ & $_;
    $_ & $_ != $_;
  }
  /** @warning === has higher precedence than & */
  any_eq3_bitand: {
    $_ === $_ & $_;
    $_ & $_ === $_;
  }
  /** @warning !== has higher precedence than & */
  any_neq3_bitand: {
    $_ !== $_ & $_;
    $_ & $_ !== $_;
  }

  /** @warning == has higher precedence than | */
  any_eq_bitor: {
    $_ == $_ | $_;
    $_ | $_ == $_;
  }
  /** @warning != has higher precedence than | */
  any_neq_bitor: {
    $_ != $_ | $_;
    $_ | $_ != $_;
  }
  /** @warning === has higher precedence than | */
  any_eq3_bitor: {
    $_ === $_ | $_;
    $_ | $_ === $_;
  }
  /** @warning !== has higher precedence than | */
  any_neq3_bitor: {
    $_ !== $_ | $_;
    $_ | $_ !== $_;
  }
}

/**
 * @comment Report assignments that can be simplified.
 * @before  $x = $x + $y;
 * @after   $x += $y;
 */
function assignOp() {
  /**
   * @maybe could rewrite as `$x += $y`
   * @fix $x += $y
   * @pure $x
   */
  $x = $x + $y;

  /**
   * @maybe could rewrite as `$x -= $y`
   * @fix $x -= $y
   * @pure $x
   */
  $x = $x - $y;

  /**
   * @maybe could rewrite as `$x *= $y`
   * @fix $x *= $y
   * @pure $x
   */
  $x = $x * $y;

  /**
   * @maybe could rewrite as `$x /= $y`
   * @fix $x /= $y
   * @pure $x
   */
  $x = $x / $y;

  /**
   * @maybe could rewrite as `$x %= $y`
   * @fix $x %= $y
   * @pure $x
   */
  $x = $x % $y;

  /**
   * @maybe could rewrite as `$x &= $y`
   * @fix $x &= $y
   * @pure $x
   */
  $x = $x & $y;

  /**
   * @maybe could rewrite as `$x |= $y`
   * @fix $x |= $y
   * @pure $x
   */
  $x = $x | $y;

  /**
   * @maybe could rewrite as `$x ^= $y`
   * @fix $x ^= $y
   * @pure $x
   */
  $x = $x ^ $y;

  /**
   * @maybe could rewrite as `$x <<= $y`
   * @fix $x <<= $y
   * @pure $x
   */
  $x = $x << $y;

  /**
   * @maybe could rewrite as `$x >>= $y`
   * @fix $x >>= $y
   * @pure $x
   */
  $x = $x >> $y;

  /**
   * @maybe could rewrite as `$x .= $y`
   * @fix $x .= $y
   * @pure $x
   */
  $x = $x . $y;

  /**
   * @maybe could rewrite as `$x ??= $y`
   * @fix $x ??= $y
   * @pure $x
   */
  $x = $x ?? $y;
}
