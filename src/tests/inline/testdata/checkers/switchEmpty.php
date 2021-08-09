<?php

function switchEmpty($a) {
  switch ($a) { // want `Switch has empty body`

  }

  switch ($a) {} // want `Switch has empty body`
}
