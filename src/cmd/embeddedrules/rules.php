<?php

/**
 * @noinspection ALL
 * @linter disable
 */

// TODO: when #323 is implemented, warning messages can suggest
// concrete code fixes.

// TODO: add @pure annotation to make sure that matched expression
// is free of side effects.

/**
 * @name ternarySimplify
 * @maybe can simplify to $cond
 * @type bool $cond
 */
$cond ? true : false;

/**
 * @name ternarySimplify
 * @maybe can simplify to (bool)$cond
 * @type !bool $cond
 */
$cond ? true : false;

/**
 * @name ternarySimplify
 * @maybe use ?: shorthand for $a?$a:$b case
 */
$x ? $x : $y;

/**
 * @name ternarySimplify
 * @maybe could use ?? (null coalesce operator)
 */
isset($x) ? $x : $y;
