<?php

class Foo {}
interface IFoo {}

function definedClass(Foo $a, IFoo $b) {}
function undefinedClass(
  Foo1 $a,  // want `Type \Foo1 not found`
  IFoo1 $b, // want `Type \IFoo1 not found`
) {}

function nullableUndefinedClass(
  ?Foo1 $a,  // want `Type \Foo1 not found`
  ?IFoo1 $b, // want `Type \IFoo1 not found`
) {}

function unionUndefinedClass(
  Foo|Foo1 $a,  // want `Type \Foo1 not found`
  IFoo1|Foo $b, // want `Type \IFoo1 not found`
) {}

function returnDefinedClass(): Foo {}
function returnDefinedIface(): IFoo {}

function returnUndefinedClass(): Foo1 {}  // want `Type \Foo1 not found`
function returnUndefinedIface(): IFoo1 {} // want `Type \IFoo1 not found`

function returnNullableUndefinedClass(): ?Foo1 {}  // want `Type \Foo1 not found`

function returnUnionUndefinedClass(): Foo|Foo1 {}  // want `Type \Foo1 not found`

function returnUnionOfUndefinedClass(): IFoo1|Foo1 {}  // want `Type \Foo1 not found` and `Type \Foo1 not found`

class Test {
  public Foo $a;
  public IFoo $b;

  public Foo1 $c;  // want `Type \Foo1 not found`
  public IFoo1 $d; // want `Type \IFoo1 not found`

  public ?Foo1 $c;  // want `Type \Foo1 not found`
  public ?IFoo1 $d; // want `Type \IFoo1 not found`

  public Foo1|Foo $e;  // want `Type \Foo1 not found`
  public Foo|IFoo1 $f; // want `Type \IFoo1 not found`

  public Foo1|IFoo1 $g;  // want `Type \Foo1 not found` and `Type \IFoo1 not found`
}
