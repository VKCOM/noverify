package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDupSubExpr1(t *testing.T) {
	// Operations below are not checked by dupSubExpr.
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
function f($x) {
  $_ = [
    $x + $x,
    $x * $x,
    $x ** $x,
  ];
}
`)
}

func TestDupSubExpr2(t *testing.T) {
	// All expression below give warnings for mixed-typed values.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
function f($x) {
  $_ = [
    $x & $x, // 1
    $x | $x, // 2
    $x ^ $x, // 3
    $x and $x, // 4
    $x && $x, // 5
    $x or $x, // 6
    $x || $x, // 7
    $x xor $x, // 8
    $x - $x, // 9
    $x / $x, // 10
    $x % $x, // 11
    $x == $x, // 12
    $x === $x, // 13
    $x != $x, // 14
    $x !== $x, // 15
    $x < $x, // 16
    $x <= $x, // 17
    $x > $x, // 18
    $x >= $x, // 19
    $x <=> $x, // 20
  ];
}
`)
	test.Expect = []string{
		`Duplicated operands value in & expression`,
		`Duplicated operands value in | expression`,
		`Duplicated operands value in ^ expression`,
		`Duplicated operands value in and expression`,
		`Duplicated operands value in && expression`,
		`Duplicated operands value in or expression`,
		`Duplicated operands value in || expression`,
		`Duplicated operands value in xor expression`,
		`Duplicated operands value in - expression`,
		`Duplicated operands value in / expression`,
		`Duplicated operands value in % expression`,
		`Duplicated operands value in == expression`,
		`Duplicated operands value in === expression`,
		`Duplicated operands value in != expression`,
		`Duplicated operands value in !== expression`,
		`Duplicated operands value in < expression`,
		`Duplicated operands value in <= expression`,
		`Duplicated operands value in > expression`,
		`Duplicated operands value in >= expression`,
		`Duplicated operands value in <=> expression`,
	}
	test.RunAndMatch()
}

func TestDupSubExpr3(t *testing.T) {
	// Duplicated expression is float-typed.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
function f() {
  $x = 1.5;
  $_ = [
    $x & $x,
    $x | $x,
    $x ^ $x,
    $x and $x,
    $x && $x,
    $x or $x,
    $x || $x,
    $x xor $x,
    $x - $x,
    $x / $x,
    $x % $x,
    $x == $x,
    $x === $x,
    $x != $x,
    $x !== $x,
    $x < $x,
    $x <= $x,
    $x > $x,
    $x >= $x,
    $x <=> $x,
  ];
}
`)
	test.Expect = []string{
		`Duplicated operands value in & expression`,
		`Duplicated operands value in | expression`,
		`Duplicated operands value in ^ expression`,
		`Duplicated operands value in and expression`,
		`Duplicated operands value in && expression`,
		`Duplicated operands value in or expression`,
		`Duplicated operands value in || expression`,
		`Duplicated operands value in xor expression`,
		`Duplicated operands value in % expression`,
		`Duplicated operands value in < expression`,
		`Duplicated operands value in > expression`,
	}
	test.RunAndMatch()
}

func TestDupTernaryOperands(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
function f($cond) {
  return $cond ? 1 : 1;
}

function f2($cond) {
  if ($cond ? f(10) : f(10)) {
    return 20;
  }
}
`)
	test.Expect = []string{
		`Branches for true and false have the same operands, ternary operator is meaningless`,
		`Branches for true and false have the same operands, ternary operator is meaningless`,
	}
	test.RunAndMatch()
}

func TestDupIfElseBody(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
function f($cond) {
  if ($cond) {
    return 0;
  } else {
    return 0;
  }
}

// Nested in another if, multiple actions under both branches.
function f2($cond) {
  if ($cond) {
    if ($cond+1) {
      echo f(1);
      echo f(2);
    } else {
      echo f(1);
      echo f(2);
    }
  }
}

// Should also work for branches without {}.
function f3($cond) {
  if ($cond)
    echo 1;
  else
    echo 1;
}


// Test alt syntax.
function f2($cond) {
  if ($cond) {
    if ($cond+1):
      echo f(1);
      echo f(2);
    else:
      echo f(1);
      echo f(2);
    endif;
  }
}
`)
	test.Expect = []string{
		`Duplicated if/else actions`,
		`Duplicated if/else actions`,
		`Duplicated if/else actions`,
		`Duplicated if/else actions`,
	}
	test.RunAndMatch()
}

func TestDupIfCond1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);
const C1 = 1;
const C2 = 2;
const C3 = 3;

$glob = 0;

function purefunc($x) { return $x+1; }

function impurefunc($x) {
  global $glob;
  $glob += $x;
  return $glob;
}

function f($cond) {
  if (C1 == $cond) {
  } else if (C1 == $cond) {
  }

  if ($cond) {
  } elseif ($cond) {
  } else {}

  if ($cond+1) {
    if ($cond+4) {
    } elseif ($cond+2) {
      if ($cond+3) {}
    }
  } elseif ($cond+1) {
  }

  if ($cond+1) {
  } else if ($cond+2) {
    if ($cond+1) {
    } else if ($cond+2) {
    } elseif ($cond+3) {
    } else if ($cond+4) {
    } elseif ($cond+2) {
    }
  }
}
`)
	test.Expect = []string{
		`duplicated condition in if-else chain`,
		`duplicated condition in if-else chain`,
		`duplicated condition in if-else chain`,
		`duplicated condition in if-else chain`,
	}
	test.RunAndMatch()
}

func TestDupIfCond2(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types=1);
function f($cond) {
  if ($cond+1) {
    if ($cond+1) {
    } elseif ($cond+2) {
      if ($cond+3) {}
    }
  } elseif ($cond+2) {
  }
}
`)
}

func TestDupCaseCond(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
const C1 = 1;
const C2 = 2;
const C3 = 3;

$glob = 0;

function purefunc($x) { return $x+1; }

function impurefunc($x) {
  global $glob;
  $glob += $x;
  return $glob;
}

function f($cond) {
  switch ($cond) {
  case 'abc':
    echo 1; break;
  case 'abc':
    echo 2; break; // Bad: duplicated string literal
  }

  switch ($cond) {
  case 2:
    echo 2; break;
  case 1:
    echo 1; break;
  case 3:
    echo 3; break;
  case 1:
    echo 4; break; // Bad: duplicated int literal
  case 5:
    echo 5; break;
  }

  switch ($cond) {
  default: break;
  case C1: break;
  case C2: break;
  case C3: break;
  case C2: break; // Bad: duplicated const
  }

  // Pure function calls are tracked.
  switch ($cond) {
  case purefunc(1): break;
  case purefunc(1): break; // Bad: duplicated pure func call
  }

  // No warnings for duplicated expressions with potential side effects.
  switch ($cond) {
  case impurefunc(1): break;
  case impurefunc(1): break;
  }

  switch ($cond) {
  case 3: break; // Bad: duplicated value 3 (C3 = 3)
  case 1: break;
  case C1: break; // Bad: duplicated value 1 (C1 = 1)
  case C3: break;
  }
}
`)
	test.Expect = []string{
		`Duplicated switch case for expression 'abc'`,
		`Duplicated switch case for expression 1`,
		`Duplicated switch case for expression C2 (value: 2)`,
		`Duplicated switch case for expression purefunc(1)`,
		`Duplicated switch case for expression C1 (value: 1)`,
		`Duplicated switch case for expression C3 (value: 3)`,
	}
	linttest.RunFilterMatch(test, "dupCond")
}
