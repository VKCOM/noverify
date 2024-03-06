<?php
declare(strict_types=1);

use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;

function f(
  int $hour,
  #[PhpStormStubsElementAvailable(from: '8.0')] int $seconds,
) {}

function f1(
  int $hour,
  #[PhpStormStubsElementAvailable(to: '7.4')] int $seconds,
) {}

function f2(
  int $hour,
  #[PhpStormStubsElementAvailable('8.0')] int $seconds,
) {}

function main() {
  f(); // want `Too few arguments for f, expecting 1, saw 0`
  f1(); // want `Too few arguments for f1, expecting 2, saw 0`
  f2(); // want `Too few arguments for f2, expecting 1, saw 0`
}
