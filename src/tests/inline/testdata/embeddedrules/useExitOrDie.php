<?php

function useExitOrDie() {
  exit; // want `Don't use the 'exit' function`
  exit(1); // want `Don't use the 'exit' function`

  die; // want `Don't use the 'die' function`
  die("die"); // want `Don't use the 'die' function`
}
