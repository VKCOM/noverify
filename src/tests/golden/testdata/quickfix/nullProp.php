<?php

class Boo {}
class Zoo {}

class Foo {
  /**
   * @var Boo $item
   */
  public $item = null;

  /**
   * @var ?Boo $item1
   */
  public $item1 = null;

  /**
   * @var ?Boo|Zoo $item11
   */
  public $item11 = null;

  /**
   * @var Boo|Zoo $item12
   */
  public $item12 = null;

  /**
   * @var Boo $item2
   * @var Boo $item3
   */
  public $item2 = null, $item3 = null;

  public ?Boo $item5 = null;
  public ?Boo $item6 = null, $item7 = null;

  public ?int $item8 = null;
}
