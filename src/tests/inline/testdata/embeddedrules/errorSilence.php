<?php

@include_once "file.php"; // want `Don't use @, silencing errors is bad practice`

function g(): int { return 1; }
function g1(): array { return [1]; }

function errorSilence() {
  @include_once "file.php"; // want `Don't use @, silencing errors is bad practice`

  echo @g(); // want `Don't use @, silencing errors is bad practice`
  echo @g1()[123]; // want `Don't use @, silencing errors is bad practice`
}
