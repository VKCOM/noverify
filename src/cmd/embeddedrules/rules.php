<?php

/**
 * @noinspection ALL
 * @linter disable
 */

function ternarySimplify() {
  /**
   * @maybe could replace the ternary with just $cond
   * @type bool $cond
   */
  $cond ? true : false;

  /**
   * @maybe could rewrite as `(bool)$cond`
   * @type !bool $cond
   */
  $cond ? true : false;

  /**
   * @maybe could rewrite as `$x ?: $y`
   * @pure $x
   */
  $x ? $x : $y;

  /** @maybe could rewrite as `$x ?? $y` */
  isset($x) ? $x : $y;
}