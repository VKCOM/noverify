<?php

function getTypeMisUse(mixed $var) {
  if (gettype($var) === "string") {
  }

  if (gettype($var) == "double") {
  }

  if (gettype($var) !== "array") {
  }

  if (gettype($var) != "boolean") {
  }

  if (gettype($var) === "object" && true) {
  }

  if (gettype(getTypeMisUse($var)) === "integer") {
  }

  if (gettype(getTypeMisUse($var)) != "resource") {
  }
}
