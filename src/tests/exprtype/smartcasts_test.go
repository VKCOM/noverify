package exprtype_test

import "testing"

func TestIsString(t *testing.T) {
	code := `<?php
/**
 * @param mixed $a
 * @param mixed $b
 */
function f($a, $b) {
  if (is_string($a)) {
    exprtype($a, "string");
  }

  if (is_string($a)) {
    exprtype($a, "string");
  } else {
    exprtype($a, "mixed");
  }

  if (!is_string($a)) {
    exprtype($a, "mixed");
  } else {
    exprtype($a, "string");
  }

  if (is_string($a) && is_string($b)) {
    exprtype($a, "string");
    exprtype($b, "string");
  }

  if (is_string($a) || is_string($b)) {
    exprtype($a, "string");
    exprtype($b, "string");
  }

  if (is_string($a) && !is_string($b)) {
    exprtype($a, "string");
    exprtype($b, "mixed");
  }

  if (is_string($a) && !is_string($b)) {
    exprtype($a, "string");
    exprtype($b, "mixed");
  } else {
    exprtype($a, "mixed");
    exprtype($b, "string");
  }
}

/**
 * @param string $a
 */
function f1($a) {
  if (is_string($a)) {
    exprtype($a, "string");
  } else {
    exprtype($a, "mixed");
  }
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestIsStringWithIsInt(t *testing.T) {
	code := `<?php
class Foo {}
/**
 * @param mixed $a
 * @param Foo   $b
 */
function f($a, $b) {
  if (is_string($a)) {
    exprtype($a, "string");
  } else if (is_int($a)) {
    exprtype($a, "int");
  } else {
    exprtype($a, "mixed");
  }

  if (is_string($b)) {
    exprtype($b, "string");
  } else if (is_int($b)) {
    exprtype($b, "int");
  } else {
    exprtype($b, "\Foo");
  }
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestIsStringWithReturn(t *testing.T) {
	code := `<?php
class Foo {}
/**
 * @param mixed $a
 * @param ?Foo  $b
 */
function f($a, $b) {
  if (!is_string($a)) {
    return;
  }

  exprtype($a, "string");

  if (is_null($b)) {
    return;
  }

  exprtype($b, "\Foo");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestIsInt(t *testing.T) {
	code := `<?php
/**
 * @param mixed $a
 * @param mixed $b
 */
function f($a, $b) {
  if (is_int($a)) {
    exprtype($a, "int");
  }

  if (is_int($a)) {
    exprtype($a, "int");
  } else {
    exprtype($a, "mixed");
  }

  if (!is_int($a)) {
    exprtype($a, "mixed");
  } else {
    exprtype($a, "int");
  }

  if (is_int($a) && is_int($b)) {
    exprtype($a, "int");
    exprtype($b, "int");
  }

  if (is_int($a) || is_int($b)) {
    exprtype($a, "int");
    exprtype($b, "int");
  }

  if (is_int($a) && !is_int($b)) {
    exprtype($a, "int");
    exprtype($b, "mixed");
  }

  if (is_int($a) && !is_int($b)) {
    exprtype($a, "int");
    exprtype($b, "mixed");
  } else {
    exprtype($a, "mixed");
    exprtype($b, "int");
  }
}

/**
 * @param int $a
 */
function f1($a) {
  if (is_int($a)) {
    exprtype($a, "int");
  } else {
    exprtype($a, "mixed");
  }
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}
