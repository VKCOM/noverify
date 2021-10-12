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
function simpleRule() {
  /**
   * @maybe Some error message
   */
  $a++;
}
