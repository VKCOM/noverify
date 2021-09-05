<?php

use JetBrains\PhpStorm\Deprecated;

#[Deprecated]
function deprecated() {}

#[Deprecated(reason: "use X instead")]
function deprecatedReason() {}

#[Deprecated(since: "8.0")]
function deprecatedSince() {}

#[Deprecated(replacement: "X()")]
function deprecatedReplacement() {}

#[Deprecated(reason: "use X instead", since: "8.0")]
function deprecatedReasonSince() {}

#[Deprecated(reason: "use X instead", replacement: "X()", since: "8.0")]
function deprecatedReasonReplacementSince() {}

/**
 * @deprecated use Y instead
 */
#[Deprecated(reason: "use X instead")]
function deprecatedReasonWithPHPDoc() {}

/**
 * @deprecated
 */
#[Deprecated(reason: "use X instead")]
function deprecatedReasonWithEmptyPHPDoc() {}

/**
 * @deprecated
 * @removed 8.0
 */
#[Deprecated(reason: "use X instead")]
function deprecatedReasonWithEmptyPHPDocRemoved() {}

/**
 * @deprecated use Y instead
 * @removed
 */
#[Deprecated(reason: "use X instead")]
function deprecatedReasonWithPHPDocEmptyRemoved() {}

function f() {
  deprecated();       // want `Call to deprecated function deprecated`
  deprecatedReason(); // want `Call to deprecated function deprecatedReason (reason: use X instead)`
  deprecatedSince();  // want `Call to deprecated function deprecatedSince (since: 8.0)`
  deprecatedReplacement(); // want `Call to deprecated function deprecatedReplacement (use X() instead)`
  deprecatedReasonSince(); // want `Call to deprecated function deprecatedReasonSince (since: 8.0, reason: use X instead)`
  deprecatedReasonReplacementSince(); // want `Call to deprecated function deprecatedReasonReplacementSince (since: 8.0, reason: use X instead, use X() instead)`
  deprecatedReasonWithPHPDoc();       // want `Call to deprecated function deprecatedReasonWithPHPDoc (reason: use X instead)`
  deprecatedReasonWithEmptyPHPDoc();  // want `Call to deprecated function deprecatedReasonWithEmptyPHPDoc (reason: use X instead)`
  deprecatedReasonWithEmptyPHPDocRemoved(); // want `Call to deprecated function deprecatedReasonWithEmptyPHPDocRemoved (reason: use X instead, removed: 8.0)`
  deprecatedReasonWithPHPDocEmptyRemoved(); // want `Call to deprecated function deprecatedReasonWithPHPDocEmptyRemoved (reason: use X instead)`
}

function instead1() {}
function instead2() {}

/**
 * @deprecated
 * @see instead1()
 */
function deprecatedWithSee() {}

/**
 * @deprecated
 * @see instead1()
 * @see instead2()
 */
function deprecatedWithSeveralSee() {}

function f1() {
  deprecatedWithSee(); // want `Call to deprecated function deprecatedWithSee (use instead1() instead)`
  deprecatedWithSeveralSee(); // want `Call to deprecated function deprecatedWithSeveralSee (use instead1() or instead2() instead)`
}
