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

  /**
   * @maybe could rewrite as `$x[$i] ?? $y`
   * @pure $i
   */
  any_indexing: {
    $x[$i] !== null ? $x[$i] : $y;
    null !== $x[$i] ? $x[$i] : $y;
    $x[$i] === null ? $y : $x[$i];
    null === $x[$i] ? $y : $x[$i];
  }

  /**
   * @maybe could rewrite as `$x[$i] ?? $y`
   * @pure $i
   */
  any_array_key_exists: {
    array_key_exists($i, $x) ? $x[$i] : $y;
    !array_key_exists($i, $x) ? $y : $x[$i];
  }
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

  /** @warning === has higher precedence than ?? */
  $_ === $_ ?? $_;

  /** @warning !== has higher precedence than ?? */
  $_ !== $_ ?? $_;

  /** @warning == has higher precedence than ?? */
  $_ == $_ ?? $_;

  /** @warning != has higher precedence than ?? */
  $_ != $_ ?? $_;

  /** @warning > has higher precedence than ?? */
  $_ > $_ ?? $_;

  /** @warning >= has higher precedence than ?? */
  $_ >= $_ ?? $_;

  /** @warning < has higher precedence than ?? */
  $_ < $_ ?? $_;

  /** @warning <= has higher precedence than ?? */
  $_ <= $_ ?? $_;
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
 * @extends
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
 * @comment Report suspicious usage of bitwise operations.
 * @before  if ($isURL & $verify) ...
 * @after   if ($isURL && $verify) ...
 */
function bitwiseOps() {
  /**
   * @warning Used & bitwise operator over bool operands, perhaps && is intended?
   * @fix $x && $y
   * @type bool $x
   * @type bool $y
   */
  $x & $y;

  /**
   * @warning Used | bitwise operator over bool operands, perhaps || is intended?
   * @fix $x || $y
   * @type bool $x
   * @type bool $y
   */
  $x | $y;
}

/**
 * @comment Report call expressions that can be simplified.
 * @before  in_array($k, array_keys($this->data))
 * @after   array_key_exists($k, $this->data)
 */
function callSimplify() {
  /**
   * @maybe Could simplify to array_key_exists($key, $a)
   * @fix   array_key_exists($key, $a)
   */
  in_array($key, array_keys($a));

  /**
   * @maybe Could simplify to $x[$y]
   */
  substr($x, $y, 1);

  /**
   * @maybe Could simplify to $a[] = $v
   */
  array_push($a, $v);
}

/**
 * @comment Report not-strict-enough comparisons.
 * @before  in_array("what", $s)
 * @after   in_array("what", $s, true)
 */
function strictCmp() {
    /**
     * @warning non-strict comparison (use ===)
     */
    any_equal: {
        $_ == true;
        true == $_;
        $_ == false;
        false == $_;
        $_ == null;
        null == $_;
    }

    /**
     * @warning non-strict string comparison (use ===)
     * @type string $x
     * @type string $y
     */
     $x == $y;

    /**
     * @warning non-strict comparison (use !==)
     */
    any_not_equal: {
        $_ != true;
        true != $_;
        $_ != false;
        false != $_;
        $_ != null;
        null != $_;
    }

    /**
     * @warning non-strict string comparison (use !==)
     * @type string $x
     * @type string $y
     */
    $x != $y;

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

/**
 * @comment Report the use of curly braces for indexing.
 * @before  $x{0}
 * @after   $x[0]
 */
function indexingSyntax() {
    /**
     * @warning a{i} indexing is deprecated since PHP 7.4, use a[i] instead
     * @fix $x[$y]
     */
    $x{$y};
}

/**
 * @comment Report using an integer for $needle argument of str* functions.
 * @before  strpos("hello", 10)
 * @after   strpos("hello", chr(10))
 */
function intNeedle() {
    /**
     * @warning since PHP 7.3, passing the int parameter needle to string search functions has been deprecated, cast it explicitly to string or wrap it in a chr() function call
     * @type int $x
     */
    any: {
        strpos($_, $x);
        strrpos($_, $x);
        stripos($_, $x);
        strripos($_, $x);
        strstr($_, $x);
        strchr($_, $x);
        strrchr($_, $x);
        stristr($_, $x);
    }
}

/**
 * @extends
 */
function langDeprecated() {
    /**
     * @warning since PHP 7.3.0, the definition of case insensitive constants has been deprecated
     */
    define($_, $_, true);

    /**
     * @warning define defaults to a case sensitive constant, the third argument can be removed
     * @fix     define($_, $_);
     */
    define($_, $_, false);
}
