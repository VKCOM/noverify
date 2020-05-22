<?php

/**
 * @noinspection ALL
 * @linter disable
 */

/**
 * @name ternarySimplify
 * @maybe could replace the ternary with just $cond
 * @type bool $cond
 */
$cond ? true : false;

/**
 * @name ternarySimplify
 * @maybe could rewrite as `(bool)$cond`
 * @type !bool $cond
 */
$cond ? true : false;

/**
 * @name ternarySimplify
 * @maybe could rewrite as `$x ?: $y`
 * @pure $x
 */
$x ? $x : $y;

/**
 * @name ternarySimplify
 * @maybe could rewrite as `$x ?? $y`
 */
isset($x) ? $x : $y;
