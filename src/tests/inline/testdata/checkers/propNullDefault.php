<?php

class Boo {}
class Zoo {}

class Foo {
  private const A = null;

  /**
   * @var Boo $item
   */
  private $item = null; // want `assigning null to a not nullable property`

  /**
   * @var ?Boo $item1
   */
  public $item1 = null; // ok

  /**
   * @var ?Boo|Zoo $item11
   */
  public static $item11 = null; // ok

  /**
   * @var Boo|Zoo $item12
   */
  public $item12 = null; // want `assigning null to a not nullable property`

  /**
   * @var mixed
   */
  public static $item13 = null; // ok

  /**
   * @var int
   */
  public $item14 = null; // ok

  /**
   * @var Boo|int
   */
  public $item15 = null; // ok

  /**
   * @var Boo $item2
   * @var Boo $item3
   */
  public $item2 = null, $item3 = null; // want `assigning null to a not nullable property` and `assigning null to a not nullable property`

  public static ?Boo $item5 = null; // ok
  public ?Boo $item6 = null, $item7 = null; // ok

  public static ?int $item8 = null; // ok
  public ?int $item9 = null; // ok

  public static ?Boo $item10 = self::A; // ok
  public ?int $item101 = 10; // ok
}
