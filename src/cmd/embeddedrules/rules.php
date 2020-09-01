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

/**
 * @comment Report potential off-by-one mistakes.
 * @before  $a[count($a)]
 * @after   $a[count($a)-1]
 */
function offBy1() {
  /**
   * @warning probably intended to use count-1 as an index
   * @fix     $a[count($a) - 1]
   */
  $a[count($a)];

  /**
   * @warning probably intended to use sizeof-1 as an index
   * @fix     $a[sizeof($a) - 1]
   */
  $a[sizeof($a)];
}

/**
 * @comment Report suspicious arguments order.
 * @before  strpos('/', $s)
 * @after   strpos($s, '/')
 */
function argsOrder() {
  /**
   * @warning potentially incorrect haystack and needle arguments order
   */
  any_haystack_needle: {
    strpos(${"char"}, ${"*"});
    stripos(${"char"}, ${"*"});
    strrpos(${"char"}, ${"*"});
    substr_count(${"str"}, ${"*"});
  }

  /**
   * @warning potentially incorrect replacement and subject arguments order
   */
  preg_replace($_, $_, ${"str"}, ${"*"});

  /**
   * @warning potentially incorrect replace and string arguments order
   */
  any_str_replace: {
    str_replace($_, $_, ${"char"}, ${"*"});
    str_replace($_, $_, "", ${"*"});
  }

  /**
   * @warning potentially incorrect delimiter and string arguments order
   */
  explode($_, ${"char"}, ${"*"});
}

/**
 * @comment Report not-strict-enough comparisons.
 * @before  in_array("what", $s)
 * @after   in_array("what", $s, true)
 */
function strictCmp() {
    /**
     * @warning non-strict comparison (use ===)
     * @type string $x
     * @type string $y
     */
    any_equal: {
        $_ == true;
        true == $_;
        $_ == false;
        false == $_;
        $_ == null;
        null == $_;
        $x == $y;
    }

    /**
     * @warning non-strict comparison (use !==)
     * @type string $x
     * @type string $y
     */
    any_not_equal: {
        $_ != true;
        true != $_;
        $_ != false;
        false != $_;
        $_ != null;
        null != $_;
        $x != $y;
    }

    /**
     * @warning 3rd argument of in_array must be true when comparing strings
     * @type string $b
     */
    in_array($b, $_);

    /**
     * @warning 3rd argument of array_search must be true when comparing strings
     * @type string $b
     */
    array_search($b, $_);
}
