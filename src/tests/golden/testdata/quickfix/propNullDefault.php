<?php

class Boo {}
class Zoo {}

class Foo {
  /**
   * @var Boo $item
   */
  public static $item = null;

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
  private static $item12 = null;

  /**
   * @var Boo|Zoo $item13
   */
  public $item13=NULL;

  /**
   * @var Boo|Zoo $item14
   */
  public static $item14= null;

  /**
   * @var Boo|Zoo $item14
   */
  public $item15 /** bla bla */= null;

  /**
   * @var Boo $item2
   * @var Boo $item3
   */
  public $item2 = null, $item3 = null;

  /**
   * @var Boo $item21
   * @var Boo $item31
   */
  public $item21=null, $item31= null;

  public static ?Boo $item5 = null;
  public ?Boo $item6 = null, $item7 = null;

  public static ?int $item8 = null;
}
