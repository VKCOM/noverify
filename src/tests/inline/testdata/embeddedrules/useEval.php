<?php

function useEval() {
  $hello = "Hello NoVerify!";

  eval("echo \"$hello\";"); // want `Don't use the 'eval' function`
}
