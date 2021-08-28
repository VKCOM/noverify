<?php

namespace GlobalClasses {
  class Boo {}
  class IBoo {}
}

namespace ErrorsInTypehint {
  use GlobalClasses\Boo;
  use GlobalClasses\IBoo;

  use GlobalClasses\Boo as BooG;
  use GlobalClasses\IBoo as IBooG;

  function definedClassWithUse(Boo $a, IBoo $b) {}
  function definedClassWithUseAlias(BooG $a, IBooG $b) {}

  class Foo {}
  interface IFoo {}

  function definedClass(Foo $a, IFoo $b) {}
  function undefinedClass(
    Foo1 $a,  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`
    IFoo1 $b, // want `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`
  ) {}

  function nullableUndefinedClass(
    ?Foo1 $a,  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`
    ?IFoo1 $b, // want `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`
  ) {}

  function unionUndefinedClass(
    Foo|Foo1 $a,  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`
    IFoo1|Foo $b, // want `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`
  ) {}

  function returnDefinedClass(): Foo {}
  function returnDefinedIface(): IFoo {}

  function returnUndefinedClass(): Foo1 {}  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`
  function returnUndefinedIface(): IFoo1 {} // want `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`

  function returnNullableUndefinedClass(): ?Foo1 {}  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`

  function returnUnionUndefinedClass(): Foo|Foo1 {}  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`

  function returnUnionOfUndefinedClass(): IFoo1|Foo1 {}  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist` and `Class or interface named \ErrorsInTypehint\Foo1 does not exist`

  class Test {
    public Foo $a;
    public IFoo $b;

    public Foo1 $c;  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`
    public IFoo1 $d; // want `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`

    public ?Foo1 $c;  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`
    public ?IFoo1 $d; // want `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`

    public Foo1|Foo $e;  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist`
    public Foo|IFoo1 $f; // want `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`

    public Foo1|IFoo1 $g;  // want `Class or interface named \ErrorsInTypehint\Foo1 does not exist` and `Class or interface named \ErrorsInTypehint\IFoo1 does not exist`
  }

  trait SingletonSelf {
    private static ?self $instance = null;
    public static function instance(): self {}
  }
}

namespace ErrorsInPHPDoc {
  class Foo {}
  interface IFoo {}

  /**
   * @param Foo  $a
   * @param IFoo $b
   */
  function definedClass($a, $b) {}

  /**
   * @param Foo1  $a // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
   * @param IFoo1 $b // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
   */
  function undefinedClass($a, $b) {}

  /**
   * @param ?Foo1  $a // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
   * @param ?IFoo1 $b // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
   */
  function nullableUndefinedClass($a, $b) {}

  /**
   * @param Foo|Foo1  $a // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
   * @param IFoo1|Foo $b // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
   */
  function unionUndefinedClass($a, $b) {}

  /**
   * @return Foo
   */
  function returnDefinedClass() {}
  /**
   * @return IFoo
   */
  function returnDefinedIface() {}

  /**
   * @return Foo1 // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
   */
  function returnUndefinedClass() {}
  /**
   * @return IFoo1 // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
   */
  function returnUndefinedIface() {}

  /**
   * @return ?Foo1 // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
   */
  function returnNullableUndefinedClass() {}

  /**
   * @return Foo|Foo1  // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
   */
  function returnUnionUndefinedClass() {}

  /**
   * @return IFoo1|Foo1 // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist` and `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
   */
  function returnUnionOfUndefinedClass() {}

  class Test {
    /**
     * @var Foo
     */
    public $a;
    /**
     * @var IFoo
     */
    public $b;

    /**
     * @var Foo1 // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
     */
    public $c;
    /**
     * @var IFoo1 // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
     */
    public $d;

    /**
     * @var ?Foo1 // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
     */
    public $c;
    /**
     * @var ?IFoo1 // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
     */
    public $d;

    /**
     * @var Foo1|Foo // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
     */
    public $e;
    /**
     * @var Foo|IFoo1 // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
     */
    public $f;

    /**
     * @var Foo1|IFoo1 // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist` and `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
     */
    public $g;
  }

  function f($a) {
    /**
     * @var Foo
     */
    $a = 100;

    /**
     * @var Foo $a
     */
    $a = 100;

    /**
     * @var IFoo
     */
    $a = 100;

    /**
     * @var IFoo $a
     */
    $a = 100;


    /**
     * @var Foo1 // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
     */
    $a = 100;

    /**
     * @var Foo1 $a // want `Class or interface named \ErrorsInPHPDoc\Foo1 does not exist`
     */
    $a = 100;

    /**
     * @var IFoo1 // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
     */
    $a = 100;

    /**
     * @var IFoo1 $a // want `Class or interface named \ErrorsInPHPDoc\IFoo1 does not exist`
     */
    $a = 100;

    echo $a;
  }

  trait SingletonSelf {
    /** @var ?self */
    private static $instance = null;

    /** @return self */
    public static function instance() {}
  }

  trait SingletonStatic {
    /** @var ?static */
    private static $instance = null;

    /** @return static */
    public static function instance() {}
  }
}
