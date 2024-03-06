package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue78_1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	declare(strict_types=1);
global $cond;
$xs = [1, 2];
switch ($cond) {
case 0:
  trailing_exit_if($xs);
  echo "unreachable";
  break;
case 1:
  trailing_exit_foreach($xs);
  echo "unreachable";
  break;
case 2:
  trailing_exit_foreach2($xs);
  echo "unreachable";
  break;
case 3:
  trailing_throw_if($xs);
  echo "unreachable";
  break;
case 4:
  trailing_throw_foreach($xs);
  echo "unreachable";
  break;
case 5:
  trailing_throw_foreach2($xs);
  echo "unreachable";
  break;
case 6:
  trailing_exit_for($xs);
  echo "unreachable";
  break;
case 7:
  trailing_exit_while($xs);
  echo "unreachable";
  break;
case 8:
  trailing_exit_try($xs);
  echo "unreachable";
  break;
case 9:
  trailing_exit_try2($xs);
  echo "unreachable";
  break;
case 10:
  trailing_exit_catch($xs);
  echo "unreachable";
  break;
case 11:
  trailing_exit_switch($xs);
  echo "unreachable";
  break;
default:
  break;
}

class Exception {}

function trailing_exit_switch($xs) {
  switch($xs[0]) {
  case 1:
    die("ok");
  case 2:
    break;
  default:
    break;
  }
  exit;
}

function trailing_exit_try($xs) {
  try {
    if ($xs) {
      die("ok");
    }
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_try2($xs) {
  try {
    try {
      if ($xs) {
        if ($xs[0] < 1000) {
          die("ok");
        }
      }
    } catch (Exception $_) {}
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_catch($xs) {
  try {
  } catch (Exception $_) {
    die("ok");
  }
  exit;
}

function trailing_exit_if($xs) {
  if ($xs) {
    die("ok");
  }
  exit;
}

function trailing_exit_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      die("ok");
    }
  }
  exit;
}

function trailing_exit_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        die("ok");
      }
    }
  }
  exit;
}

function trailing_throw_if($xs) {
  if ($xs) {
    die("ok");
  }
  throw new Exception("oops");
}

function trailing_throw_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      die("ok");
    }
  }
  throw new Exception("oops");
}

function trailing_throw_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        die("ok");
      }
    }
  }
  throw new Exception("oops");
}

function trailing_exit_for($xs) {
  for ($i = 0; $i < 10; $i++) {
    if ($i == $xs[0]) {
      die("ok");
    }
  }
  exit;
}

function trailing_exit_while($xs) {
  while (1) {
    if ($xs[0] < 1000) {
      die("ok");
    }
    break;
  }
  exit;
}`)

	test.Expect = []string{
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
	}

	test.RunAndMatch()
}

func TestIssue78_2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
global $cond;
$xs = [1, 2];
switch ($cond) {
case 0:
  trailing_exit_if($xs);
  echo "unreachable";
  break;
case 1:
  trailing_exit_foreach($xs);
  echo "unreachable";
  break;
case 2:
  trailing_exit_foreach2($xs);
  echo "unreachable";
  break;
case 3:
  trailing_throw_if($xs);
  echo "unreachable";
  break;
case 4:
  trailing_throw_foreach($xs);
  echo "unreachable";
  break;
case 5:
  trailing_throw_foreach2($xs);
  echo "unreachable";
  break;
case 6:
  trailing_exit_for($xs);
  echo "unreachable";
  break;
case 7:
  trailing_exit_while($xs);
  echo "unreachable";
  break;
case 8:
  trailing_exit_try($xs);
  echo "unreachable";
  break;
case 9:
  trailing_exit_try2($xs);
  echo "unreachable";
  break;
case 10:
  trailing_exit_catch($xs);
  echo "unreachable";
  break;
case 11:
  trailing_exit_switch($xs);
  echo "unreachable";
  break;
default:
  break;
}

class Exception {}

function trailing_exit_switch($xs) {
  switch($xs[0]) {
  default:
    break;
  case 2:
    break;
  case 1:
    $_ = $xs[0];
  }
  exit;
}

function trailing_exit_try($xs) {
  try {
    if ($xs) {
    }
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_try2($xs) {
  try {
    try {
      if ($xs) {
        if ($xs[0] < 1000) {
        }
      }
    } catch (Exception $_) {}
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_catch($xs) {
  try {
  } catch (Exception $_) {
  }
  exit;
}

function trailing_exit_if($xs) {
  if ($xs) {
  }
  exit;
}

function trailing_exit_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
    }
  }
  exit;
}

function trailing_exit_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
      }
    }
  }
  exit;
}

function trailing_throw_if($xs) {
  if ($xs) {
  }
  throw new Exception("oops");
}

function trailing_throw_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
    }
  }
  throw new Exception("oops");
}

function trailing_throw_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
      }
    }
  }
  throw new Exception("oops");
}

function trailing_exit_for($xs) {
  for ($i = 0; $i < 10; $i++) {
    if ($i == $xs[0]) {
    }
  }
  exit;
}

function trailing_exit_while($xs) {
  while (1) {
    if ($xs[0] < 1000) {
    }
  }
  exit;
}`)

	test.Expect = []string{
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
	}

	test.RunAndMatch()
}

func TestIssue78_3(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
$xs = [1, 2];
trailing_exit_if($xs);
trailing_exit_foreach($xs);
trailing_exit_foreach2($xs);
trailing_throw_if($xs);
trailing_throw_foreach($xs);
trailing_throw_foreach2($xs);
trailing_exit_for($xs);
trailing_exit_while($xs);
trailing_exit_try($xs);
trailing_exit_try2($xs);
trailing_exit_catch($xs);
trailing_exit_switch($xs);
echo "not a dead code";

class Exception {}

function trailing_exit_switch($xs) {
  switch($xs[0]) {
  case 1:
    return "ok";
  case 2:
    break;
  default:
    break;
  }
  exit;
}

function trailing_exit_try($xs) {
  try {
    if ($xs) {
      return "ok";
    }
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_try2($xs) {
  try {
    try {
      if ($xs) {
        if ($xs[0] < 1000) {
          return "ok";
        }
      }
    } catch (Exception $_) {}
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_catch($xs) {
  try {
  } catch (Exception $_) {
    return "ok";
  }
  exit;
}

function trailing_exit_if($xs) {
  if ($xs) {
    return "ok";
  }
  exit;
}

function trailing_exit_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      return "ok";
    }
  }
  exit;
}

function trailing_exit_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        return "ok";
      }
    }
  }
  exit;
}

function trailing_throw_if($xs) {
  if ($xs) {
    return "ok";
  }
  throw new Exception("oops");
}

function trailing_throw_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      return "ok";
    }
  }
  throw new Exception("oops");
}

function trailing_throw_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        return "ok";
      }
    }
  }
  throw new Exception("oops");
}

function trailing_exit_for($xs) {
  for ($i = 0; $i < 10; $i++) {
    if ($i == $xs[0]) {
      return "ok";
    }
  }
  exit;
}

function trailing_exit_while($xs) {
  while (1) {
    if ($xs[0] < 1000) {
      return "ok";
    }
    break;
  }
  exit;
}
`)
}
