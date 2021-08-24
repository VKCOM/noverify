# Checkers

## Brief statistics

| Total checks | Checks enabled by default | Disabled checks by default | Autofixable checks |
| ------------ | ------------------------- | -------------------------- | ------------------ |
| 95           | 86                        | 9                         | 12                 |

## Table of contents
 - Enabled by default
   - [`accessLevel` checker](#accesslevel-checker)
   - [`alwaysNull` checker](#alwaysnull-checker)
   - [`argCount` checker](#argcount-checker)
   - [`argsOrder` checker](#argsorder-checker)
   - [`arrayAccess` checker](#arrayaccess-checker)
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
   - [`complexity` checker](#complexity-checker)
   - [`concatenationPrecedence` checker](#concatenationprecedence-checker)
   - [`constCase` checker (autofixable)](#constcase-checker)
   - [`countUse` checker (autofixable)](#countuse-checker)
   - [`deadCode` checker](#deadcode-checker)
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
   - [`errorSilence` checker](#errorsilence-checker)
   - [`forLoop` checker](#forloop-checker)
   - [`implicitModifiers` checker](#implicitmodifiers-checker)
   - [`indexingSyntax` checker (autofixable)](#indexingsyntax-checker)
   - [`intNeedle` checker](#intneedle-checker)
   - [`intOverflow` checker](#intoverflow-checker)
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
   - [`offBy1` checker (autofixable)](#offby1-checker)
   - [`oldStyleConstructor` checker](#oldstyleconstructor-checker)
   - [`paramClobber` checker](#paramclobber-checker)
   - [`parentConstructor` checker](#parentconstructor-checker)
   - [`phpdocLint` checker](#phpdoclint-checker)
   - [`phpdocRef` checker](#phpdocref-checker)
   - [`phpdocType` checker](#phpdoctype-checker)
   - [`precedence` checker](#precedence-checker)
   - [`printf` checker](#printf-checker)
   - [`redundantGlobal` checker](#redundantglobal-checker)
   - [`regexpSimplify` checker](#regexpsimplify-checker)
   - [`regexpSyntax` checker](#regexpsyntax-checker)
   - [`regexpVet` checker](#regexpvet-checker)
   - [`returnAssign` checker](#returnassign-checker)
   - [`reverseAssign` checker](#reverseassign-checker)
   - [`selfAssign` checker](#selfassign-checker)
   - [`stdInterface` checker](#stdinterface-checker)
   - [`strangeCast` checker](#strangecast-checker)
   - [`strictCmp` checker](#strictcmp-checker)
   - [`stripTags` checker](#striptags-checker)
   - [`switchDefault` checker](#switchdefault-checker)
   - [`switchEmpty` checker](#switchempty-checker)
   - [`switchSimplify` checker](#switchsimplify-checker)
   - [`syntax` checker](#syntax-checker)
   - [`ternarySimplify` checker (autofixable)](#ternarysimplify-checker)
   - [`unaryRepeat` checker (autofixable)](#unaryrepeat-checker)
   - [`undefinedConstant` checker](#undefinedconstant-checker)
   - [`undefinedFunction` checker](#undefinedfunction-checker)
   - [`undefinedMethod` checker](#undefinedmethod-checker)
   - [`undefinedProperty` checker](#undefinedproperty-checker)
   - [`undefinedType` checker](#undefinedtype-checker)
   - [`undefinedVariable` checker](#undefinedvariable-checker)
   - [`unimplemented` checker](#unimplemented-checker)
   - [`unused` checker](#unused-checker)
   - [`varShadow` checker](#varshadow-checker)
   - [`voidResultUsed` checker](#voidresultused-checker)
 - Disabled by default
   - [`argsReverse` checker](#argsreverse-checker)
   - [`classMembersOrder` checker](#classmembersorder-checker)
   - [`deprecated` checker](#deprecated-checker)
   - [`langDeprecated` checker](#langdeprecated-checker)
   - [`missingPhpdoc` checker](#missingphpdoc-checker)
   - [`propNullDefault` checker (autofixable)](#propnulldefault-checker)
   - [`redundantCast` checker](#redundantcast-checker)
   - [`trailingComma` checker (autofixable)](#trailingcomma-checker)
   - [`typeHint` checker](#typehint-checker)
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


### `linterError` checker

#### Description

Report internal linter error.



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


### `phpdocLint` checker

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


### `phpdocRef` checker

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


### `phpdocType` checker

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


### `regexpSyntax` checker

#### Description

Report regexp syntax errors.



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


### `stdInterface` checker

#### Description

Report issues related to std PHP interfaces.



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


### `undefinedType` checker

#### Description

Report usages of undefined type.

#### Non-compliant code:
```php
class Foo extends UndefinedClass {}
```

#### Compliant code:
```php
class Foo extends DefinedClass {}
```


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


### `deprecated` checker

#### Description

Report usages of deprecated symbols.

#### Non-compliant code:
```php
ereg($pat, $s) // The ereg function has been deprecated.
```

#### Compliant code:
```php
preg_match($pat, $s)
```


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

