# Checkers

## Brief statistics

| Total checks | Checks enabled by default | Disabled checks by default | Autofixable checks |
| ------------ | ------------------------- | -------------------------- | ------------------ |
| 117           | 99                        | 18                         | 15                 |

## Table of contents
 - Enabled by default
   - [`accessLevel` checker](#accesslevel-checker)
   - [`alwaysNull` checker](#alwaysnull-checker)
   - [`argCount` checker](#argcount-checker)
   - [`argsOrder` checker](#argsorder-checker)
   - [`arraySyntax` checker (autofixable)](#arraysyntax-checker)
   - [`assignOp` checker (autofixable)](#assignop-checker)
   - [`badTraitUse` checker](#badtraituse-checker)
   - [`bareTry` checker](#baretry-checker)
   - [`bitwiseOps` checker (autofixable)](#bitwiseops-checker)
   - [`callSimplify` checker (autofixable)](#callsimplify-checker)
   - [`callStatic` checker](#callstatic-checker)
   - [`caseBreak` checker](#casebreak-checker)
   - [`caseContinue` checker](#casecontinue-checker)
   - [`catchOrder` checker](#catchorder-checker)
   - [`concatenationPrecedence` checker](#concatenationprecedence-checker)
   - [`constCase` checker (autofixable)](#constcase-checker)
   - [`countUse` checker (autofixable)](#countuse-checker)
   - [`deadCode` checker](#deadcode-checker)
   - [`deprecated` checker](#deprecated-checker)
   - [`discardExpr` checker](#discardexpr-checker)
   - [`discardVar` checker](#discardvar-checker)
   - [`dupArrayKeys` checker](#duparraykeys-checker)
   - [`dupBranchBody` checker](#dupbranchbody-checker)
   - [`dupCatch` checker](#dupcatch-checker)
   - [`dupCond` checker](#dupcond-checker)
   - [`dupGlobal` checker](#dupglobal-checker)
   - [`dupSubExpr` checker](#dupsubexpr-checker)
   - [`emptyStmt` checker](#emptystmt-checker)
   - [`emptyStringCheck` checker](#emptystringcheck-checker)
   - [`forLoop` checker](#forloop-checker)
   - [`implicitModifiers` checker](#implicitmodifiers-checker)
   - [`indexingSyntax` checker (autofixable)](#indexingsyntax-checker)
   - [`intNeedle` checker](#intneedle-checker)
   - [`intOverflow` checker](#intoverflow-checker)
   - [`invalidDocblock` checker](#invaliddocblock-checker)
   - [`invalidDocblockRef` checker](#invaliddocblockref-checker)
   - [`invalidDocblockType` checker](#invaliddocblocktype-checker)
   - [`invalidExtendClass` checker](#invalidextendclass-checker)
   - [`invalidNew` checker](#invalidnew-checker)
   - [`keywordCase` checker](#keywordcase-checker)
   - [`linterError` checker](#lintererror-checker)
   - [`magicMethodDecl` checker](#magicmethoddecl-checker)
   - [`maybeUndefined` checker](#maybeundefined-checker)
   - [`methodSignatureMismatch` checker](#methodsignaturemismatch-checker)
   - [`misspellComment` checker](#misspellcomment-checker)
   - [`misspellName` checker](#misspellname-checker)
   - [`mixedArrayKeys` checker](#mixedarraykeys-checker)
   - [`nameMismatch` checker](#namemismatch-checker)
   - [`nestedTernary` checker](#nestedternary-checker)
   - [`newAbstract` checker](#newabstract-checker)
   - [`nonPublicInterfaceMember` checker](#nonpublicinterfacemember-checker)
   - [`notExplicitNullableParam` checker (autofixable)](#notexplicitnullableparam-checker)
   - [`notNullSafetyFunctionArgumentArrayDimFetch` checker](#notnullsafetyfunctionargumentarraydimfetch-checker)
   - [`notNullSafetyFunctionArgumentConstFetch` checker](#notnullsafetyfunctionargumentconstfetch-checker)
   - [`notNullSafetyFunctionArgumentFunctionCall` checker](#notnullsafetyfunctionargumentfunctioncall-checker)
   - [`notNullSafetyFunctionArgumentList` checker](#notnullsafetyfunctionargumentlist-checker)
   - [`notNullSafetyFunctionArgumentPropertyFetch` checker](#notnullsafetyfunctionargumentpropertyfetch-checker)
   - [`notNullSafetyFunctionArgumentStaticFunctionCall` checker](#notnullsafetyfunctionargumentstaticfunctioncall-checker)
   - [`notNullSafetyFunctionArgumentVariable` checker](#notnullsafetyfunctionargumentvariable-checker)
   - [`notNullSafetyFunctionCall` checker](#notnullsafetyfunctioncall-checker)
   - [`notNullSafetyPropertyFetch` checker](#notnullsafetypropertyfetch-checker)
   - [`notNullSafetyStaticFunctionCall` checker](#notnullsafetystaticfunctioncall-checker)
   - [`notNullSafetyVariable` checker](#notnullsafetyvariable-checker)
   - [`offBy1` checker (autofixable)](#offby1-checker)
   - [`oldStyleConstructor` checker](#oldstyleconstructor-checker)
   - [`paramClobber` checker](#paramclobber-checker)
   - [`parentConstructor` checker](#parentconstructor-checker)
   - [`phpAliases` checker (autofixable)](#phpaliases-checker)
   - [`precedence` checker](#precedence-checker)
   - [`printf` checker](#printf-checker)
   - [`redundantGlobal` checker](#redundantglobal-checker)
   - [`regexpSimplify` checker](#regexpsimplify-checker)
   - [`regexpSyntax` checker](#regexpsyntax-checker)
   - [`regexpVet` checker](#regexpvet-checker)
   - [`reverseAssign` checker](#reverseassign-checker)
   - [`selfAssign` checker](#selfassign-checker)
   - [`stdInterface` checker](#stdinterface-checker)
   - [`strangeCast` checker](#strangecast-checker)
   - [`strictCmp` checker](#strictcmp-checker)
   - [`stringInterpolationDeprecated` checker](#stringinterpolationdeprecated-checker)
   - [`stripTags` checker](#striptags-checker)
   - [`switchEmpty` checker](#switchempty-checker)
   - [`switchSimplify` checker](#switchsimplify-checker)
   - [`syntax` checker](#syntax-checker)
   - [`ternarySimplify` checker (autofixable)](#ternarysimplify-checker)
   - [`unaryRepeat` checker (autofixable)](#unaryrepeat-checker)
   - [`undefinedClass` checker](#undefinedclass-checker)
   - [`undefinedConstant` checker](#undefinedconstant-checker)
   - [`undefinedFunction` checker](#undefinedfunction-checker)
   - [`undefinedMethod` checker](#undefinedmethod-checker)
   - [`undefinedProperty` checker](#undefinedproperty-checker)
   - [`undefinedTrait` checker](#undefinedtrait-checker)
   - [`undefinedVariable` checker](#undefinedvariable-checker)
   - [`unimplemented` checker](#unimplemented-checker)
   - [`unused` checker](#unused-checker)
   - [`useEval` checker](#useeval-checker)
   - [`useExitOrDie` checker](#useexitordie-checker)
   - [`useSleep` checker](#usesleep-checker)
   - [`varShadow` checker](#varshadow-checker)
 - Disabled by default
   - [`argsReverse` checker](#argsreverse-checker)
   - [`arrayAccess` checker](#arrayaccess-checker)
   - [`classMembersOrder` checker](#classmembersorder-checker)
   - [`complexity` checker](#complexity-checker)
   - [`deprecatedUntagged` checker](#deprecateduntagged-checker)
   - [`errorSilence` checker](#errorsilence-checker)
   - [`getTypeMisUse` checker (autofixable)](#gettypemisuse-checker)
   - [`langDeprecated` checker](#langdeprecated-checker)
   - [`missingPhpdoc` checker](#missingphpdoc-checker)
   - [`packaging` checker](#packaging-checker)
   - [`parentNotFound` checker](#parentnotfound-checker)
   - [`propNullDefault` checker (autofixable)](#propnulldefault-checker)
   - [`redundantCast` checker](#redundantcast-checker)
   - [`returnAssign` checker](#returnassign-checker)
   - [`switchDefault` checker](#switchdefault-checker)
   - [`trailingComma` checker (autofixable)](#trailingcomma-checker)
   - [`typeHint` checker](#typehint-checker)
   - [`voidResultUsed` checker](#voidresultused-checker)
## Enabled

### `accessLevel` checker

#### Description

Report erroneous member access.

#### Non-compliant code:
```php
$x->privateMethod(); // privateMethod is private and can't be accessed.
```

#### Compliant code:
```php
$x->publicMethod();
```
<p><br></p>


### `alwaysNull` checker

#### Description

Report when use to always null object.

#### Non-compliant code:
```php
if ($obj == null && $obj->method()) { ... }
```

#### Compliant code:
```php
if ($obj != null && $obj->method()) { ... }
```
<p><br></p>


### `argCount` checker

#### Description

Report mismatching args count inside call expressions.

#### Non-compliant code:
```php
array_combine($keys) // The function takes at least two arguments.
```

#### Compliant code:
```php
array_combine($keys, $values)
```
<p><br></p>


### `argsOrder` checker

#### Description

Report suspicious arguments order.

#### Non-compliant code:
```php
// It is possible that the arguments are in the wrong order, since 
// searching for a substring in a character does not make sense.
strpos('/', $s);
```

#### Compliant code:
```php
strpos($s, '/');
```
<p><br></p>


### `arraySyntax` checker

> Auto fix available

#### Description

Report usages of old `array()` syntax.

#### Non-compliant code:
```php
array(1, 2)
```

#### Compliant code:
```php
[1, 2]
```
<p><br></p>


### `assignOp` checker

> Auto fix available

#### Description

Report assignments that can be simplified.

#### Non-compliant code:
```php
$x = $x + $y;
```

#### Compliant code:
```php
$x += $y;
```
<p><br></p>


### `badTraitUse` checker

#### Description

Report misuse of traits.

#### Non-compliant code:
```php
trait A {}
function f(A $a) {} // Traits cannot be used as type hints.
```

#### Compliant code:
```php
class A {}
function f(A $a) {}
```
<p><br></p>


### `bareTry` checker

#### Description

Report `try` blocks without `catch/finally`.

#### Non-compliant code:
```php
try {
  doit();
}
// Missing catch or finally blocks.
```

#### Compliant code:
```php
try {
  doit();
} catch (Exception $e) {
  // Handle $e.
}
```
<p><br></p>


### `bitwiseOps` checker

> Auto fix available

#### Description

Report suspicious usage of bitwise operations.

#### Non-compliant code:
```php
if ($isURL & $verify) { ... } // Bitwise AND on two bool looks suspicious,
```

#### Compliant code:
```php
if ($isURL && $verify) { ... }
```
<p><br></p>


### `callSimplify` checker

> Auto fix available

#### Description

Report call expressions that can be simplified.

#### Non-compliant code:
```php
in_array($k, array_keys($this->data))
```

#### Compliant code:
```php
array_key_exists($k, $this->data)
```
<p><br></p>


### `callStatic` checker

#### Description

Report static calls of instance methods and vice versa.

#### Non-compliant code:
```php
$object::instance_method() // instance_method is not a static method.
```

#### Compliant code:
```php
$object->instance_method()
```
<p><br></p>


### `caseBreak` checker

#### Description

Report `switch` cases without `break`.

#### Non-compliant code:
```php
switch ($v) {
case 1:
  echo "one"; // May want to insert a "break" here.
case 2:
  echo "this fallthrough is intentional";
  // fallthrough
case 3:
  echo "two or three";
}
```

#### Compliant code:
```php
switch ($v) {
case 1:
  echo "one";
  break;
case 2:
  echo "this fallthrough is intentional";
  // fallthrough
case 3:
  echo "two or three";
}
```
<p><br></p>


### `caseContinue` checker

#### Description

Report suspicious `continue` usages inside `switch` cases.

#### Non-compliant code:
```php
switch ($v) {
case STOP:
  continue; // Continue inside a switch is equivalent to break.
case INC:
  $x++;
  break;
}
```

#### Compliant code:
```php
switch ($v) {
case STOP:
  break;
case INC:
  $x++;
  break;
}
```
<p><br></p>


### `catchOrder` checker

#### Description

Report erroneous `catch` order in `try` statements.

#### Non-compliant code:
```php
try {
  // Some code.
} catch (Exception $e) {
  // This will catch both Exception and TimeoutException.
} catch (TimeoutException $e) {
  // This is a dead code.
}
```

#### Compliant code:
```php
try {
  // Some code.
} catch (TimeoutException $e) {
  // Ok, it can catch TimeoutException.
} catch (Exception $e) {
  // Ok, it will catch everything else.
}
```
<p><br></p>


### `concatenationPrecedence` checker

#### Description

Report when use unparenthesized expression containing both `.` and binary operator.

#### Non-compliant code:
```php
"id: " . $id - 10
```

#### Compliant code:
```php
"id: " . ($id - 10)
```
<p><br></p>


### `constCase` checker

> Auto fix available

#### Description

Report built-in constants that are not in the lower case.

#### Non-compliant code:
```php
return TRUE;
```

#### Compliant code:
```php
return true;
```
<p><br></p>


### `countUse` checker

> Auto fix available

#### Description

Report comparisons `count(...)` which are always `false` or `true`.

#### Non-compliant code:
```php
if (count($arr) >= 0) { ... }
```

#### Compliant code:
```php
if (count($arr) != 0) { ... }
```
<p><br></p>


### `deadCode` checker

#### Description

Report potentially unreachable code.

#### Non-compliant code:
```php
thisFunctionAlwaysExits();
foo(); // Dead code.
```

#### Compliant code:
```php
foo();
thisFunctionAlwaysExits();
```
<p><br></p>


### `deprecated` checker

#### Description

Report usages of deprecated symbols.

#### Non-compliant code:
```php
/**
 * @deprecated Use g() instead
 */
function f() {}

f();
```

#### Compliant code:
```php
/**
 * @deprecated Use g() instead
 */
function f() {}

g();
```
<p><br></p>


### `discardExpr` checker

#### Description

Report expressions that are evaluated but not used.

#### Non-compliant code:
```php
if ($cond) {
  [$v, $err]; // Result expression is not used anywhere.
}
```

#### Compliant code:
```php
if ($cond) {
  return [$v, $err];
}
```
<p><br></p>


### `discardVar` checker

#### Description

Report the use of variables that were supposed to be unused, like `$_`.

#### Non-compliant code:
```php
$_ = some();
echo $_;
```

#### Compliant code:
```php
$someVal = some();
echo $someVal;
```
<p><br></p>


### `dupArrayKeys` checker

#### Description

Report duplicated keys in array literals.

#### Non-compliant code:
```php
[A => 1, B => 2, A => 3] // Key A is duplicated.
```

#### Compliant code:
```php
[A => 1, B => 2, C => 3]
```
<p><br></p>


### `dupBranchBody` checker

#### Description

Report suspicious conditional branches that execute the same action.

#### Non-compliant code:
```php
// Regardless of the condition, the result will always be the same.
$pickLeft ? foo($left) : foo($left)
```

#### Compliant code:
```php
$pickLeft ? foo($left) : foo($right)
```
<p><br></p>


### `dupCatch` checker

#### Description

Report duplicated `catch` clauses.

#### Non-compliant code:
```php
try {
  // some code
} catch (Exception1 $e) {
} catch (Exception1 $e) {} // <- Possibly the typo.
```

#### Compliant code:
```php
try {
  // some code
} catch (Exception1 $e) {
} catch (Exception2 $e) {}
```
<p><br></p>


### `dupCond` checker

#### Description

Report duplicated conditions in `switch` and `if/else` statements.

#### Non-compliant code:
```php
if ($status == OK) {
  return "OK";
} elseif ($status == OK) { // Duplicated condition.
  return "NOT OK";
} else {
  return "UNKNOWN";
}
```

#### Compliant code:
```php
if ($status == OK) {
  return "OK";
} elseif ($status == NOT_OK) {
  return "NOT OK";
} else {
  return "UNKNOWN";
}
```
<p><br></p>


### `dupGlobal` checker

#### Description

Report repeated global statements over variables.

#### Non-compliant code:
```php
global $x, $y, $x; // $x was already mentioned in global.
```

#### Compliant code:
```php
global $x, $y;
```
<p><br></p>


### `dupSubExpr` checker

#### Description

Report suspicious duplicated operands in expressions.

#### Non-compliant code:
```php
return $x[$i] < $x[$i]; // The left and right expressions are the same.
```

#### Compliant code:
```php
return $x[$i] < $x[$j];
```
<p><br></p>


### `emptyStmt` checker

#### Description

Report redundant empty statements that can be safely removed.

#### Non-compliant code:
```php
echo $foo;; // Second semicolon is unnecessary here.
```

#### Compliant code:
```php
echo $foo;
```
<p><br></p>


### `emptyStringCheck` checker

#### Description

Report string emptyness checking using `strlen(...)`.

#### Non-compliant code:
```php
if (strlen($string)) { ... }
```

#### Compliant code:
```php
if ($string !== "") { ... }
```
<p><br></p>


### `forLoop` checker

#### Description

Report potentially erroneous `for` loops.

#### Non-compliant code:
```php
for ($i = 0; $i < 100; $i--) { ... }
```

#### Compliant code:
```php
for ($i = 0; $i < 100; $i++) { ... }
```
<p><br></p>


### `implicitModifiers` checker

#### Description

Report implicit modifiers.

#### Non-compliant code:
```php
class Foo {
  function f() {} // The access modifier is implicit.
}
```

#### Compliant code:
```php
class Foo {
  public function f() {}
}
```
<p><br></p>


### `indexingSyntax` checker

> Auto fix available

#### Description

Report the use of curly braces for indexing.

#### Non-compliant code:
```php
$x{0}
```

#### Compliant code:
```php
$x[0]
```
<p><br></p>


### `intNeedle` checker

#### Description

Report using an integer for `$needle` argument of `str*` functions.

#### Non-compliant code:
```php
strpos("hello", 10)
```

#### Compliant code:
```php
strpos("hello", chr(10))
```
<p><br></p>


### `intOverflow` checker

#### Description

Report potential integer overflows that may result in unexpected behavior.

#### Non-compliant code:
```php
// Better to use a constant to avoid accidental overflow and float conversion.
return -9223372036854775808;
```

#### Compliant code:
```php
return PHP_INT_MIN;
```
<p><br></p>


### `invalidDocblock` checker

#### Description

Report malformed PHPDoc comments.

#### Non-compliant code:
```php
@property $foo // Property type is missing.
```

#### Compliant code:
```php
@property Foo $foo
```
<p><br></p>


### `invalidDocblockRef` checker

#### Description

Report invalid symbol references inside PHPDoc.

#### Non-compliant code:
```php
@see MyClass
```

#### Compliant code:
```php
@see \Foo\MyClass
```
<p><br></p>


### `invalidDocblockType` checker

#### Description

Report potential issues in PHPDoc types.

#### Non-compliant code:
```php
@var []int $xs
```

#### Compliant code:
```php
@var int[] $xs
```
<p><br></p>


### `invalidExtendClass` checker

#### Description

Report inheritance from the final class.

#### Non-compliant code:
```php
final class Foo {}
class Boo extends Foo {}
```

#### Compliant code:
```php
class Foo {}
class Boo extends Foo {}
```
<p><br></p>


### `invalidNew` checker

#### Description

Report trait or interface usages in `new` expressions.

#### Non-compliant code:
```php
// It is forbidden to create instances of traits or interfaces.
return new SomeTrait();
```

#### Compliant code:
```php
return new SomeClass();
```
<p><br></p>


### `keywordCase` checker

#### Description

Report keywords that are not in the lower case.

#### Non-compliant code:
```php
RETURN $x;
```

#### Compliant code:
```php
return $x;
```
<p><br></p>


### `linterError` checker

#### Description

Report internal linter error.

<p><br></p>

### `magicMethodDecl` checker

#### Description

Report issues in magic method declarations.

#### Non-compliant code:
```php
class Foo {
  private function __call($method, $args) {} // The magic method __call() must have public visibility.
  public static function __set($name, $value) {} // The magic method __set() cannot be static.
}
```

#### Compliant code:
```php
class Foo {
  public function __call($method, $args) {}
  public function __set($name, $value) {}
}
```
<p><br></p>


### `maybeUndefined` checker

#### Description

Report usages of potentially undefined symbols.

#### Non-compliant code:
```php
if ($cond) {
  $v = 10;
}
return $v; // $v may be undefined.
```

#### Compliant code:
```php
$v = 0; // Default value.
if ($cond) {
  $v = 10;
}
return $v;
```
<p><br></p>


### `methodSignatureMismatch` checker

#### Description

Report a method signature mismatch in inheritance.

#### Non-compliant code:
```php
class Foo {
  final public function f() {}
}

class Boo extends Foo {
  public function f() {} // Foo::f is final.
}
```

#### Compliant code:
```php
class Foo {
  public function f() {}
}

class Boo extends Foo {
  public function f() {}
}
```
<p><br></p>


### `misspellComment` checker

#### Description

Report commonly misspelled words in comments.

#### Non-compliant code:
```php
/** This is our performace test. */
function performance_test() {}
```

#### Compliant code:
```php
/** This is our performance test. */
function performance_test() {}
```
<p><br></p>


### `misspellName` checker

#### Description

Report commonly misspelled words in symbol names.

#### Non-compliant code:
```php
function performace_test() ...
```

#### Compliant code:
```php
function performance_test() ...
```
<p><br></p>


### `mixedArrayKeys` checker

#### Description

Report array literals that have both implicit and explicit keys.

#### Non-compliant code:
```php
['a', 5 => 'b'] // Both explicit and implicit keys are used.
```

#### Compliant code:
```php
[0 => 'a', 5 => 'b']
```
<p><br></p>


### `nameMismatch` checker

#### Description

Report symbol case mismatches.

#### Non-compliant code:
```php
class Foo {}
// The spelling is in lower case, although the class definition begins with an uppercase letter.
$foo = new foo();
```

#### Compliant code:
```php
class Foo {}
$foo = new Foo();
```
<p><br></p>


### `nestedTernary` checker

#### Description

Report an unspecified order in a nested ternary operator.

#### Non-compliant code:
```php
$_ = 1 ? 2 : 3 ? 4 : 5; // There is no clear order of execution.
```

#### Compliant code:
```php
$_ = (1 ? 2 : 3) ? 4 : 5;
// or
$_ = 1 ? 2 : (3 ? 4 : 5);
```
<p><br></p>


### `newAbstract` checker

#### Description

Report abstract classes usages in `new` expressions.

#### Non-compliant code:
```php
// It is forbidden to create instances of abstract classes.
return new AbstractFactory();
```

#### Compliant code:
```php
return new NonAbstractFactory();
```
<p><br></p>


### `nonPublicInterfaceMember` checker

#### Description

Report illegal non-public access level in interfaces.

#### Non-compliant code:
```php
interface Iface {
  function a();
  public function b();
  private function c(); // Methods in an interface cannot be private.
  protected function d(); // Methods in an interface cannot be protected.
}
```

#### Compliant code:
```php
interface Iface {
  function a();
  public function b();
  public function c();
  public function d();
}
```
<p><br></p>


### `notExplicitNullableParam` checker

> Auto fix available

#### Description

Report not nullable param with explicit null default value.

#### Non-compliant code:
```php
function f(string $str = null);
```

#### Compliant code:
```php
function f(?string $str = null);
```
<p><br></p>


### `notNullSafetyFunctionArgumentArrayDimFetch` checker

#### Description

Report not nullsafety call array.

#### Non-compliant code:
```php
class A {
    public string $value = 'Hello';
}

function test(A $a): void {
    echo $a->value;
}

$arr = [new A(), null];
test($arr[1]);
```

#### Compliant code:
```php
reported not safety call
```
<p><br></p>


### `notNullSafetyFunctionArgumentConstFetch` checker

#### Description

Report not nullsafety call

#### Non-compliant code:
```php
function f(A $klass);
						f(null);
```

#### Compliant code:
```php
reported that null passed to non-nullable parameter.
```
<p><br></p>


### `notNullSafetyFunctionArgumentFunctionCall` checker

#### Description

Report not nullsafety function call.

#### Non-compliant code:
```php
class A {
    public static function hello(): ?string {
        return "Hello!";
    }
}

function test(A $s): void {
    echo $s;
}

function testNullable(): ?A{
	return new A();
}

test(testNullable());
```

#### Compliant code:
```php
reported not safety call
```
<p><br></p>


### `notNullSafetyFunctionArgumentList` checker

#### Description

Report not nullsafety call for null list

#### Non-compliant code:
```php
test(list($a) = [null]);
```

#### Compliant code:
```php
reported not safety call
```
<p><br></p>


### `notNullSafetyFunctionArgumentPropertyFetch` checker

#### Description

Report not nullsafety fetching property in function argument.

#### Non-compliant code:
```php

class User {
    public $name = "lol";
}

$user = new User();
$user = null;
echo $user->name;
```

#### Compliant code:
```php
reported not safety call
```
<p><br></p>


### `notNullSafetyFunctionArgumentStaticFunctionCall` checker

#### Description

Report not nullsafety call with static function call usage.

#### Non-compliant code:
```php
class A {
    public static function hello(): ?string {
        return "Hello!";
    }
}

function test(string $s): void {
    echo $s;
}

test(A::hello());
```

#### Compliant code:
```php
reported not safety call
```
<p><br></p>


### `notNullSafetyFunctionArgumentVariable` checker

#### Description

Report not nullsafety call

#### Non-compliant code:
```php
function f(A $klass);
						f(null);
```

#### Compliant code:
```php
reported not safety call with null in variable.
```
<p><br></p>


### `notNullSafetyFunctionCall` checker

#### Description

Report not nullsafety function call.

#### Non-compliant code:
```php
function getUserOrNull(): ?User { echo "test"; }

$getUserOrNull()->test();
```

#### Compliant code:
```php
reported not safety function call
```
<p><br></p>


### `notNullSafetyPropertyFetch` checker

#### Description

Report not nullsafety property fetch.

#### Non-compliant code:
```php

class User {
    public $name = "lol";
}

$user = new User();
$user = null;
echo $user->name;
```

#### Compliant code:
```php
reported not safety call
```
<p><br></p>


### `notNullSafetyStaticFunctionCall` checker

#### Description

Report not nullsafety function call.

#### Non-compliant code:
```php
class A {
    public static function hello(): ?string {
        return "Hello!";
    }
}

function test(string $s): void {
    echo $s;
}

test(A::hello());
```

#### Compliant code:
```php
reported not safety static function call
```
<p><br></p>


### `notNullSafetyVariable` checker

#### Description

Report not nullsafety call

#### Non-compliant code:
```php
$user = new User();

$user = null;

echo $user->name;
```

#### Compliant code:
```php
reported not safety call with null in variable.
```
<p><br></p>


### `offBy1` checker

> Auto fix available

#### Description

Report potential off-by-one mistakes.

#### Non-compliant code:
```php
$a[count($a)]
```

#### Compliant code:
```php
$a[count($a)-1]
```
<p><br></p>


### `oldStyleConstructor` checker

#### Description

Report old-style (PHP4) class constructors.

#### Non-compliant code:
```php
class Foo {
  // Constructor in the old style of PHP 4.
  public function Foo($v) { $this->v = $v; }
}
```

#### Compliant code:
```php
class Foo {
  public function __construct($v) { $this->v = $v; }
}
```
<p><br></p>


### `paramClobber` checker

#### Description

Report assignments that overwrite params prior to their usage.

#### Non-compliant code:
```php
function api_get_video($user_id) {
  // The arguments are assigned a new value before using the value passed to the function.
  $user_id = 0;
  return get_video($user_id);
}
```

#### Compliant code:
```php
function api_get_video($user_id) {
  $user_id = $user_id ?: 0;
  return get_video($user_id);
}
```
<p><br></p>


### `parentConstructor` checker

#### Description

Report missing `parent::__construct` calls in class constructors.

#### Non-compliant code:
```php
class Foo extends Bar {
  public function __construct($x, $y) {
    // Lost call to parent constructor.
    $this->y = $y;
  }
}
```

#### Compliant code:
```php
class Foo extends Bar {
  public function __construct($x, $y) {
    parent::__construct($x);
    $this->y = $y;
  }
}
```
<p><br></p>


### `phpAliases` checker

> Auto fix available

#### Description

Report php aliases functions.

#### Non-compliant code:
```php
join("", []);
```

#### Compliant code:
```php
implode("", []);
```
<p><br></p>


### `precedence` checker

#### Description

Report potential operation precedence issues.

#### Non-compliant code:
```php
$x & $mask == 0; // == has higher precedence than &
```

#### Compliant code:
```php
($x & $mask) == 0
```
<p><br></p>


### `printf` checker

#### Description

Report issues in printf-like function calls.

#### Non-compliant code:
```php
sprintf("id=%d") // Lost argument for '%d' specifier.
```

#### Compliant code:
```php
sprintf("id=%d", $id)
```
<p><br></p>


### `redundantGlobal` checker

#### Description

Report global statement over superglobal variables (which is redundant).

#### Non-compliant code:
```php
global $Foo, $_GET; // $_GET is superglobal.
```

#### Compliant code:
```php
global $Foo;
```
<p><br></p>


### `regexpSimplify` checker

#### Description

Report regular expressions that can be simplified.

#### Non-compliant code:
```php
preg_match('/x(?:a|b|c){0,}/', $s) // The regex can be simplified.
```

#### Compliant code:
```php
preg_match('/x[abc]*/', $s)
```
<p><br></p>


### `regexpSyntax` checker

#### Description

Report regexp syntax errors.

<p><br></p>

### `regexpVet` checker

#### Description

Report suspicious regexp patterns.

#### Non-compliant code:
```php
preg_match('a\d+a', $s); // 'a' is not a valid delimiter.
```

#### Compliant code:
```php
preg_match('/\d+/', $s);
```
<p><br></p>


### `reverseAssign` checker

#### Description

Report a reverse assign with unary plus or minus.

#### Non-compliant code:
```php
$a =+ 100;
```

#### Compliant code:
```php
$a += 100;
```
<p><br></p>


### `selfAssign` checker

#### Description

Report self-assignment of variables.

#### Non-compliant code:
```php
$x = $x;
```

#### Compliant code:
```php
$x = $y;
```
<p><br></p>


### `stdInterface` checker

#### Description

Report issues related to std PHP interfaces.

<p><br></p>

### `strangeCast` checker

#### Description

Report a strange way of type cast.

#### Non-compliant code:
```php
$x.""
```

#### Compliant code:
```php
(string)$x
```
<p><br></p>


### `strictCmp` checker

#### Description

Report not-strict-enough comparisons.

#### Non-compliant code:
```php
in_array("what", $s)
```

#### Compliant code:
```php
in_array("what", $s, true)
```
<p><br></p>


### `stringInterpolationDeprecated` checker

#### Description

Report deprecated string interpolation style

#### Non-compliant code:
```php
${variable}
```

#### Compliant code:
```php
{$variable}
```
<p><br></p>


### `stripTags` checker

#### Description

Report invalid `strip_tags` function usage.

#### Non-compliant code:
```php
$s = strip_tags($s, '<br/>') // Error, self-closing tags are ignored. 
```

#### Compliant code:
```php
$s = strip_tags($s, '<br>')
```
<p><br></p>


### `switchEmpty` checker

#### Description

Report `switch` with empty body.

#### Non-compliant code:
```php
switch ($a) {}
```

#### Compliant code:
```php
switch ($a) {
  case 1:
    // do something
    break;
}
```
<p><br></p>


### `switchSimplify` checker

#### Description

Report possibility to rewrite `switch` with the `if`.

#### Non-compliant code:
```php
switch ($a) {
  case 1:
    echo 1;
    break;
}
```

#### Compliant code:
```php
if ($a == 1) {
  echo 1;
}
```
<p><br></p>


### `syntax` checker

#### Description

Report syntax errors.

#### Non-compliant code:
```php
foo(1]
```

#### Compliant code:
```php
foo(1)
```
<p><br></p>


### `ternarySimplify` checker

> Auto fix available

#### Description

Report ternary expressions that can be simplified.

#### Non-compliant code:
```php
$x ? $x : $y
```

#### Compliant code:
```php
$x ?: $y
```
<p><br></p>


### `unaryRepeat` checker

> Auto fix available

#### Description

Report the repetition of unary (`!` or `~`) operators in a row.

#### Non-compliant code:
```php
echo !!$a;
```

#### Compliant code:
```php
echo (bool) $a;
```
<p><br></p>


### `undefinedClass` checker

#### Description

Report usages of undefined class or interface.

#### Non-compliant code:
```php
$foo = new UndefinedClass;
```

#### Compliant code:
```php
$foo = new DefinedClass;
```
<p><br></p>


### `undefinedConstant` checker

#### Description

Report usages of undefined constant.

#### Non-compliant code:
```php
echo PI;
```

#### Compliant code:
```php
echo M_PI;
```
<p><br></p>


### `undefinedFunction` checker

#### Description

Report usages of undefined function.

#### Non-compliant code:
```php
undefinedFunc();
```

#### Compliant code:
```php
definedFunc();
```
<p><br></p>


### `undefinedMethod` checker

#### Description

Report usages of undefined method.

#### Non-compliant code:
```php
class Foo {
  public function method() {};
}

(new Foo)->method2(); // method2 is undefined.
```

#### Compliant code:
```php
class Foo {
  public function method() {}
}

(new Foo)->method();
```
<p><br></p>


### `undefinedProperty` checker

#### Description

Report usages of undefined property.

#### Non-compliant code:
```php
class Foo {
  public string $prop;
}

(new Foo)->prop2; // prop2 is undefined.
```

#### Compliant code:
```php
class Foo {
  public string $prop;
}

(new Foo)->prop;
```
<p><br></p>


### `undefinedTrait` checker

#### Description

Report usages of undefined trait.

#### Non-compliant code:
```php
class Foo {
  use UndefinedTrait;
}
```

#### Compliant code:
```php
class Foo {
  use DefinedTrait;
}
```
<p><br></p>


### `undefinedVariable` checker

#### Description

Report usages of undefined variable.

#### Non-compliant code:
```php
echo $undefinedVar;
```

#### Compliant code:
```php
$definedVar = 100;
echo $definedVar;
```
<p><br></p>


### `unimplemented` checker

#### Description

Report classes that don't implement their contract.

#### Non-compliant code:
```php
class MyObj implements Serializable {
  public function serialize() { /* ... */ }
  // Lost implementation of the unserialize method.
}
```

#### Compliant code:
```php
class MyObj implements Serializable {
  public function serialize() { /* ... */ }
  public function unserialize(string $s) { /* ... */ }
}
```
<p><br></p>


### `unused` checker

#### Description

Report potentially unused variables.

#### Non-compliant code:
```php
$result = calculateResult(); // Unused $result.
return [$err];
```

#### Compliant code:
```php
$result = calculateResult();
return [$result, $err];
```
<p><br></p>


### `useEval` checker

#### Description

Report using `eval` function.

#### Non-compliant code:
```php
eval("2 + 2");
```

#### Compliant code:
```php
// no eval
```
<p><br></p>


### `useExitOrDie` checker

#### Description

Report using `exit` or `die` functions.

#### Non-compliant code:
```php
exit(1);
```

#### Compliant code:
```php
// no exit
```
<p><br></p>


### `useSleep` checker

#### Description

Report using `sleep` function.

#### Non-compliant code:
```php
sleep(10);
```

#### Compliant code:
```php
// no sleep
```
<p><br></p>


### `varShadow` checker

#### Description

Report the shadow of an existing variable.

#### Non-compliant code:
```php
function f(int $a) {
  // The $a variable hides the $a argument.
  foreach ([1, 2] as $a) {
    echo $a;
  }
}
```

#### Compliant code:
```php
function f(int $a) {
  foreach ([1, 2] as $b) {
    echo $b;
  }
}
```
<p><br></p>

## Disabled

### `argsReverse` checker

#### Description

Report using variables as arguments in reverse order.

#### Non-compliant code:
```php
function makeHello(string $name, int $age) {
  echo "Hello ${$name}-${$age}";
}

function main(): void {
  $name = "John";
  $age = 18;
  makeHello($age, $name); // The name should come first, and then the age.
}
```

#### Compliant code:
```php
function makeHello(string $name, int $age) {
  echo "Hello ${$name}-${$age}";
}

function main(): void {
  $name = "John";
  $age = 18;
  makeHello($name, $age);
}
```
<p><br></p>


### `arrayAccess` checker

#### Description

Report array access to non-array objects.

#### Non-compliant code:
```php
return $foo[0]; // $foo value may not implement ArrayAccess
```

#### Compliant code:
```php
if ($foo instanceof ArrayAccess) { 
  return $foo[0];
}
```
<p><br></p>


### `classMembersOrder` checker

#### Description

Report the wrong order of the class members.

#### Non-compliant code:
```php
class A {
  // In the class, constants and properties should go first, and then methods.
  public function func() {}
  const B = 1;
  public $c = 2;
}
```

#### Compliant code:
```php
class A {
  const B = 1;
  public $c = 2;
  public function func() {}
}
```
<p><br></p>


### `complexity` checker

#### Description

Report funcs/methods that are too complex.

#### Non-compliant code:
```php
function checkRights() {
  // Super big function.
}
```

#### Compliant code:
```php
function checkRights() {
  return true; // Or 42 if you need int-typed result.
}
```
<p><br></p>


### `deprecatedUntagged` checker

#### Description

Report usages of deprecated symbols if the `@deprecated` tag has no description (see `deprecated` check).

#### Non-compliant code:
```php
/**
 * @deprecated
 */
function f() {}

f();
```

#### Compliant code:
```php
/**
 * @deprecated
 */
function f() {}

g();
```
<p><br></p>


### `errorSilence` checker

#### Description

Report using `@`.

#### Non-compliant code:
```php
@f();
```

#### Compliant code:
```php
f();
```
<p><br></p>


### `getTypeMisUse` checker

> Auto fix available

#### Description

Report call gettype function.

#### Non-compliant code:
```php
if (gettype($a) == "string") { ... }
```

#### Compliant code:
```php
if (is_string($a)) { ... }
```
<p><br></p>


### `langDeprecated` checker

#### Description

Report the use of deprecated (per language spec) features.

#### Non-compliant code:
```php
$a = (real)100; // 'real' has been deprecated.
$_ = is_real($a);
```

#### Compliant code:
```php
$a = (float)100;
$_ = is_float($a);
```
<p><br></p>


### `missingPhpdoc` checker

#### Description

Report missing PHPDoc on public methods.

#### Non-compliant code:
```php
public function process($acts, $config) {
  // Does something very complicated.
}
```

#### Compliant code:
```php

/**
 * Process executes all $acts in a new context.
 * Processed $acts should never be processed again.
 *
 * @param Act[] $acts - acts to execute
 * @param array $config - options
 */
public function process($acts, $config) {
  // Does something very complicated.
}
```
<p><br></p>


### `packaging` checker

#### Description

Report call @internal method outside @package.

#### Non-compliant code:
```php
// file Boo.php 

namespace BooPackage; 

/** 
 * @package BooPackage 
 * @internal 
 */ 
class Boo { 
  public static function b() {} 
} 

// file Foo.php 

namespace FooPackage;

/** 
 * @package FooPackage 
 */ 
class Foo { 
  public static function f() {}

  /**
   * @internal
   */
  public static function fInternal() {}
}

// file Main.php

namespace Main;

use BooPackage\Boo;
use FooPackage\Foo;

class Main {
  public static function main(): void {
    Foo::f(); // ok, call non-internal method outside FooPackage

    Boo::b(); // error, call internal method inside other package
    Foo::fInternal(); // error, call internal method inside other package
  }
}
```

#### Compliant code:
```php
// file Boo.php 

namespace BooPackage; 

/** 
 * @package BooPackage 
 * @internal 
 */ 
class Boo { 
  public static function b() {} 
} 

// file Foo.php 

namespace BooPackage;

/** 
 * @package BooPackage 
 */ 
class Foo { 
  public static function f() {}

  /**
   * @internal
   */
  public static function fInternal() {}
}

// file Main.php

namespace BooPackage;

/**
 * @package BooPackage
 */
class Main {
  public static function main(): void {
    Foo::f(); // ok, call internal method inside same package

    Boo::b(); // ok, call internal method inside same package
    Foo::fInternal(); // ok, call internal method inside same package
  }
}
```
<p><br></p>


### `parentNotFound` checker

#### Description

Report using `parent::` in a class without a parent class.

#### Non-compliant code:
```php
class Foo {
  public function f() {
    parent::b(); // Class Foo has no parent.
  }
}
```

#### Compliant code:
```php
class Foo extends Boo {
  public function f() {
    parent::b(); // Ok.
  }
}
```
<p><br></p>


### `propNullDefault` checker

> Auto fix available

#### Description

Report a null assignment for a not nullable property.

#### Non-compliant code:
```php
class Foo {
  /**
   * @var Boo $item
   */
  public $item = null; // The type of the property is not nullable, but it is assigned null.
}
```

#### Compliant code:
```php
class Foo {
  /**
   * @var Boo $item
   */
  public $item;
}
```
<p><br></p>


### `redundantCast` checker

#### Description

Report redundant type casts.

#### Non-compliant code:
```php
return (int)10; // The expression is already of type int.
```

#### Compliant code:
```php
return 10;
```
<p><br></p>


### `returnAssign` checker

#### Description

Report the use of assignment in the `return` statement.

#### Non-compliant code:
```php
return $a = 100;
```

#### Compliant code:
```php
return $a;
```
<p><br></p>


### `switchDefault` checker

#### Description

Report the lack or wrong position of `default`.

#### Non-compliant code:
```php
switch ($a) {
  case 1:
    echo 1;
    break;
}
```

#### Compliant code:
```php
switch ($a) {
  case 1:
    echo 1;
    break;
  default:
    echo 2;
    break;
}
```
<p><br></p>


### `trailingComma` checker

> Auto fix available

#### Description

Report the absence of a comma for the last element in a multi-line array.

#### Non-compliant code:
```php
$_ = [
  10,
  20 // Lost comma at the end for a multi-line array.
]
```

#### Compliant code:
```php
$_ = [
  10,
  20,
]
```
<p><br></p>


### `typeHint` checker

#### Description

Report misuse of type hints.

#### Non-compliant code:
```php
// The array typehint is too generic, you need to specify a specialization or mixed[] in PHPDoc.
function f(array $a) {}
```

#### Compliant code:
```php
/**
 * @param mixed[] $a
 */
function f(array $a) {}
```
<p><br></p>


### `voidResultUsed` checker

#### Description

Report usages of the void-type expressions

#### Non-compliant code:
```php
$x = var_dump($v); // var_dump returns void.
```

#### Compliant code:
```php
$x = print_r($v, true);
```
<p><br></p>

