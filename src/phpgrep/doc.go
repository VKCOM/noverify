// Package phpgrep is a library for searching PHP code
// using syntax trees.
//
// Inspired by mvdan/gogrep.
//
// TODO(quasilyte): add the actual overview.
package phpgrep

// TODO(quasilyte): limitations of the currently used PHP parser imply some
// unwanted side-effects. Enumerating them here.
//
// - We're missing parenthesis. We can't match with respect to them.
//
// - Can't match ";" (the empty statement).

// TODO(quasilyte): unimplemented features.
//
// - Replace functionality.
//
// - Handle case sensitivity carefully (provide an option?).
//
// - stmt.Expression vs normal expressions named captures (should they match?).
//
// - Multi-statements matching.
//   To match something like `while ($_); {${"*"};}` we need to
//   continue matching rest of the pattern parts instead of stopping
//   when `while ($_);` part is matched.

// List of things that are hard (or impossible) to represent via patterns.
//
// - `$lhs <op> $rhs`; we can't express <op>.
//
// - Find a switch inside which "continue" is used.
//   The problem is that there is no way to properly recurse.
//
// - Negative assertions, like "switch that doesn't have a default case".
//
// - Find duplicated switch case conditions.
//   Can't use ${'*'} for "any case", since syntax expects only
//   T_CASE and T_DEFAULT there, resulting in a syntax error.
//   See issue #1.
//
// - $c::$constant doesn't match class const fetch,
//   since it's static property fetch.
