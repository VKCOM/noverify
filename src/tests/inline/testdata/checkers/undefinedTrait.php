<?php

trait DefinedTrait {}

class Foo {
  use UndefinedTrait; // want `Trait named \UndefinedTrait does not exist`
  use DefinedTrait;
}
