<?php

/**
 * @noinspection ALL
 */

#region POD

/**
 * @return string // want `type in typehint and phpdoc not compatibl: cannot use string as int`
 */
function f(): int {
  return 2;
}

///**
// * @return string|int
// */
//function f1(): int {
//  return 2;
//}
//
///**
// * @return string
// */
//function f2(): Foo { // want `type in typehint and phpdoc not compatibl: cannot use string as class \Foo`
//  return 2;
//}
//
///**
// * @return int[][]
// */
//function f4(): array {
//  return 2;
//}
//
///**
// * @return Roo|Boo
// */
//function f5(): string { // want `type in typehint and phpdoc not compatibl: none of the possible types (\Boo|\Roo) are compatible with string`
//  return 2;
//}
//
///**
// * @return int
// */
//function f6(): ?int { // want `type in typehint and phpdoc not compatibl: cannot use type int as nullable type int|null`
//  return 2;
//}
//
///**
// * @return ?int
// */
//function f7(): int { // want `type in typehint and phpdoc not compatibl: cannot use type int as nullable type int|null`
//  return 2;
//}
//
///**
// * @return false
// */
//function f8(): bool {
//  return 2;
//}
//
///**
// * @return true
// */
//function f9(): bool {
//  return 2;
//}
//
///**
// * @return Boo|Foo|string
// */
//function f10(): mixed {
//  return 2;
//}
//
///**
// * @return Closure
// */
//function f11(): callable {
//  return 2;
//}
//
///**
// * @return int
// */
//function f12(): callable { // want `type in typehint and phpdoc not compatibl: cannot use callable as int`
//  return 2;
//}
//
///**
// * @return string
// */
//function f13(): callable {
//  return 2;
//}
//
///**
// * @return int
// */
//function f14(): iterable { // want `type in typehint and phpdoc not compatibl: cannot use iterable as int`
//  return 2;
//}
//
//#endregion
//
//#region Classes
//
//class A {
//}
//
//class B {
//}
//
//class BExtendsA extends A {
//}
//
///**
// * @return A
// */
//function cf1(): A { // ok
//  return 2;
//}
//
///**
// * @return A
// */
//function cf2(): NonExisting { // want `type in typehint and phpdoc not compatibl: cannot use class \A as class \NonExisting`
//  return 2;
//}
//
///**
// * @return A
// */
//function cf3(): B { // want `type in typehint and phpdoc not compatibl: cannot use class \A as class \B`
//  return 2;
//}
//
///**
// * @return A
// */
//function cf4(): BExtendsA { // ok
//  return 2;
//}
//
//#endregion