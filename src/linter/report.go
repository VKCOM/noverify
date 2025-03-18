package linter

import (
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/git"
)

const (
	// IgnoreLinterMessage is a commit message that you specify if you want to cancel linter checks for this changeset
	IgnoreLinterMessage = "@linter disable"
)

func addBuiltinCheckers(reg *CheckersRegistry) {
	allChecks := []CheckerInfo{
		{
			Name:     "stripTags",
			Default:  true,
			Quickfix: false,
			Comment:  "Report invalid `strip_tags` function usage.",
			Before:   `$s = strip_tags($s, '<br/>') // Error, self-closing tags are ignored. `,
			After:    `$s = strip_tags($s, '<br>')`,
		},

		{
			Name:     "notNullSafety",
			Default:  true,
			Quickfix: false,
			Comment:  "Report not nullsafety call",
			Before: `function f(A $klass);
						f(null);`,
			After: `reported not safety call`,
		},

		{
			Name:     "notExplicitNullableParam",
			Default:  true,
			Quickfix: true,
			Comment:  "Report not nullable param with explicit null default value.",
			Before:   `function f(string $str = null);`,
			After:    `function f(?string $str = null);`,
		},

		{
			Name:     "emptyStmt",
			Default:  true,
			Quickfix: false,
			Comment:  `Report redundant empty statements that can be safely removed.`,
			Before:   `echo $foo;; // Second semicolon is unnecessary here.`,
			After:    `echo $foo;`,
		},

		{
			Name:     "intOverflow",
			Default:  true,
			Quickfix: false,
			Comment:  `Report potential integer overflows that may result in unexpected behavior.`,
			Before: `// Better to use a constant to avoid accidental overflow and float conversion.
return -9223372036854775808;`,
			After: `return PHP_INT_MIN;`,
		},

		{
			Name:     "phpAliases",
			Default:  true,
			Quickfix: true,
			Comment:  `Report php aliases functions.`,
			Before:   `join("", []);`,
			After:    `implode("", []);`,
		},

		{
			Name:     "discardExpr",
			Default:  true,
			Quickfix: false,
			Comment:  `Report expressions that are evaluated but not used.`,
			Before: `if ($cond) {
  [$v, $err]; // Result expression is not used anywhere.
}`,
			After: `if ($cond) {
  return [$v, $err];
}`,
		},

		{
			Name:     "voidResultUsed",
			Default:  false,
			Quickfix: false,
			Comment:  `Report usages of the void-type expressions`,
			Before:   `$x = var_dump($v); // var_dump returns void.`,
			After:    `$x = print_r($v, true);`,
		},

		{
			Name:     "keywordCase",
			Default:  true,
			Quickfix: false,
			Comment:  `Report keywords that are not in the lower case.`,
			Before:   `RETURN $x;`,
			After:    `return $x;`,
		},

		{
			Name:     "constCase",
			Default:  true,
			Quickfix: true,
			Comment:  `Report built-in constants that are not in the lower case.`,
			Before:   `return TRUE;`,
			After:    `return true;`,
		},

		{
			Name:     "accessLevel",
			Default:  true,
			Quickfix: false,
			Comment:  `Report erroneous member access.`,
			Before:   `$x->privateMethod(); // privateMethod is private and can't be accessed.`,
			After:    `$x->publicMethod();`,
		},

		{
			Name:     "argCount",
			Default:  true,
			Quickfix: false,
			Comment:  `Report mismatching args count inside call expressions.`,
			Before:   `array_combine($keys) // The function takes at least two arguments.`,
			After:    `array_combine($keys, $values)`,
		},

		{
			Name:     "redundantGlobal",
			Default:  true,
			Quickfix: false,
			Comment:  `Report global statement over superglobal variables (which is redundant).`,
			Before:   `global $Foo, $_GET; // $_GET is superglobal.`,
			After:    `global $Foo;`,
		},

		{
			Name:     "arrayAccess",
			Default:  false,
			Quickfix: false,
			Comment:  `Report array access to non-array objects.`,
			Before:   `return $foo[0]; // $foo value may not implement ArrayAccess`,
			After: `if ($foo instanceof ArrayAccess) { 
  return $foo[0];
}`,
		},

		{
			Name:     "mixedArrayKeys",
			Default:  true,
			Quickfix: false,
			Comment:  `Report array literals that have both implicit and explicit keys.`,
			Before:   `['a', 5 => 'b'] // Both explicit and implicit keys are used.`,
			After:    `[0 => 'a', 5 => 'b']`,
		},

		{
			Name:     "dupGlobal",
			Default:  true,
			Quickfix: false,
			Comment:  `Report repeated global statements over variables.`,
			Before:   `global $x, $y, $x; // $x was already mentioned in global.`,
			After:    `global $x, $y;`,
		},

		{
			Name:     "dupArrayKeys",
			Default:  true,
			Quickfix: false,
			Comment:  `Report duplicated keys in array literals.`,
			Before:   `[A => 1, B => 2, A => 3] // Key A is duplicated.`,
			After:    `[A => 1, B => 2, C => 3]`,
		},

		{
			Name:     "dupCond",
			Default:  true,
			Quickfix: false,
			Comment:  "Report duplicated conditions in `switch` and `if/else` statements.",
			Before: `if ($status == OK) {
  return "OK";
} elseif ($status == OK) { // Duplicated condition.
  return "NOT OK";
} else {
  return "UNKNOWN";
}`,
			After: `if ($status == OK) {
  return "OK";
} elseif ($status == NOT_OK) {
  return "NOT OK";
} else {
  return "UNKNOWN";
}`,
		},

		{
			Name:     "dupBranchBody",
			Default:  true,
			Quickfix: false,
			Comment:  `Report suspicious conditional branches that execute the same action.`,
			Before: `// Regardless of the condition, the result will always be the same.
$pickLeft ? foo($left) : foo($left)`,
			After: `$pickLeft ? foo($left) : foo($right)`,
		},

		{
			Name:     `dupSubExpr`,
			Default:  true,
			Quickfix: false,
			Comment:  `Report suspicious duplicated operands in expressions.`,
			Before:   `return $x[$i] < $x[$i]; // The left and right expressions are the same.`,
			After:    `return $x[$i] < $x[$j];`,
		},

		{
			Name:     "arraySyntax",
			Default:  true,
			Quickfix: true,
			Comment:  "Report usages of old `array()` syntax.",
			Before:   `array(1, 2)`,
			After:    `[1, 2]`,
		},

		{
			Name:     "bareTry",
			Default:  true,
			Quickfix: false,
			Comment:  "Report `try` blocks without `catch/finally`.",
			Before: `try {
  doit();
}
// Missing catch or finally blocks.`,
			After: `try {
  doit();
} catch (Exception $e) {
  // Handle $e.
}`,
		},

		{
			Name:     "caseBreak",
			Default:  true,
			Quickfix: false,
			Comment:  "Report `switch` cases without `break`.",
			Before: `switch ($v) {
case 1:
  echo "one"; // May want to insert a "break" here.
case 2:
  echo "this fallthrough is intentional";
  // fallthrough
case 3:
  echo "two or three";
}`,
			After: `switch ($v) {
case 1:
  echo "one";
  break;
case 2:
  echo "this fallthrough is intentional";
  // fallthrough
case 3:
  echo "two or three";
}`,
		},

		{
			Name:     "complexity",
			Default:  false,
			Quickfix: false,
			Comment:  `Report funcs/methods that are too complex.`,
			Before: `function checkRights() {
  // Super big function.
}`,
			After: `function checkRights() {
  return true; // Or 42 if you need int-typed result.
}`,
		},

		{
			Name:     "deadCode",
			Default:  true,
			Quickfix: false,
			Comment:  `Report potentially unreachable code.`,
			Before: `thisFunctionAlwaysExits();
foo(); // Dead code.`,
			After: `foo();
thisFunctionAlwaysExits();`,
		},

		{
			Name:     "invalidDocblock",
			Default:  true,
			Quickfix: false,
			Comment:  `Report malformed PHPDoc comments.`,
			Before:   `@property $foo // Property type is missing.`,
			After:    `@property Foo $foo`,
		},

		{
			Name:     "invalidDocblockType",
			Default:  true,
			Quickfix: false,
			Comment:  `Report potential issues in PHPDoc types.`,
			Before:   `@var []int $xs`,
			After:    `@var int[] $xs`,
		},

		{
			Name:     "invalidDocblockRef",
			Default:  true,
			Quickfix: false,
			Comment:  `Report invalid symbol references inside PHPDoc.`,
			Before:   `@see MyClass`,
			After:    `@see \Foo\MyClass`,
		},

		{
			Name:     "missingPhpdoc",
			Default:  false,
			Quickfix: false,
			Comment:  `Report missing PHPDoc on public methods.`,
			Before: `public function process($acts, $config) {
  // Does something very complicated.
}`,
			After: `
/**
 * Process executes all $acts in a new context.
 * Processed $acts should never be processed again.
 *
 * @param Act[] $acts - acts to execute
 * @param array $config - options
 */
public function process($acts, $config) {
  // Does something very complicated.
}`,
		},

		{
			Name:     "stdInterface",
			Default:  true,
			Quickfix: false,
			Comment:  `Report issues related to std PHP interfaces.`,
		},

		{
			Name:     "unimplemented",
			Default:  true,
			Quickfix: false,
			Comment:  `Report classes that don't implement their contract.`,
			Before: `class MyObj implements Serializable {
  public function serialize() { /* ... */ }
  // Lost implementation of the unserialize method.
}`,
			After: `class MyObj implements Serializable {
  public function serialize() { /* ... */ }
  public function unserialize(string $s) { /* ... */ }
}`,
		},

		{
			Name:     "syntax",
			Default:  true,
			Quickfix: false,
			Comment:  `Report syntax errors.`,
			Before:   `foo(1]`,
			After:    `foo(1)`,
		},

		{
			Name:     "undefinedClass",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of undefined class or interface.`,
			Before:   `$foo = new UndefinedClass;`,
			After:    `$foo = new DefinedClass;`,
		},

		{
			Name:     "undefinedTrait",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of undefined trait.`,
			Before: `class Foo {
  use UndefinedTrait;
}`,
			After: `class Foo {
  use DefinedTrait;
}`,
		},

		{
			Name:     "undefinedProperty",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of undefined property.`,
			Before: `class Foo {
  public string $prop;
}

(new Foo)->prop2; // prop2 is undefined.`,
			After: `class Foo {
  public string $prop;
}

(new Foo)->prop;`,
		},

		{
			Name:     "undefinedMethod",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of undefined method.`,
			Before: `class Foo {
  public function method() {};
}

(new Foo)->method2(); // method2 is undefined.`,
			After: `class Foo {
  public function method() {}
}

(new Foo)->method();`,
		},

		{
			Name:     "undefinedConstant",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of undefined constant.`,
			Before:   `echo PI;`,
			After:    `echo M_PI;`,
		},

		{
			Name:     "undefinedFunction",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of undefined function.`,
			Before:   `undefinedFunc();`,
			After:    `definedFunc();`,
		},

		{
			Name:     "undefinedVariable",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of undefined variable.`,
			Before:   `echo $undefinedVar;`,
			After: `$definedVar = 100;
echo $definedVar;`,
		},

		{
			Name:     "maybeUndefined",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of potentially undefined symbols.`,
			Before: `if ($cond) {
  $v = 10;
}
return $v; // $v may be undefined.`,
			After: `$v = 0; // Default value.
if ($cond) {
  $v = 10;
}
return $v;`,
		},

		{
			Name:     "unused",
			Default:  true,
			Quickfix: false,
			Comment:  `Report potentially unused variables.`,
			Before: `$result = calculateResult(); // Unused $result.
return [$err];`,
			After: `$result = calculateResult();
return [$result, $err];`,
		},

		{
			Name:     "redundantCast",
			Default:  false,
			Quickfix: false,
			Comment:  `Report redundant type casts.`,
			Before:   `return (int)10; // The expression is already of type int.`,
			After:    `return 10;`,
		},

		{
			Name:     "newAbstract",
			Default:  true,
			Quickfix: false,
			Comment:  "Report abstract classes usages in `new` expressions.",
			Before: `// It is forbidden to create instances of abstract classes.
return new AbstractFactory();`,
			After: `return new NonAbstractFactory();`,
		},

		{
			Name:     "invalidNew",
			Default:  true,
			Quickfix: false,
			Comment:  "Report trait or interface usages in `new` expressions.",
			Before: `// It is forbidden to create instances of traits or interfaces.
return new SomeTrait();`,
			After: `return new SomeClass();`,
		},

		{
			Name:     "regexpSimplify",
			Default:  true,
			Quickfix: false,
			Comment:  `Report regular expressions that can be simplified.`,
			Before:   `preg_match('/x(?:a|b|c){0,}/', $s) // The regex can be simplified.`,
			After:    `preg_match('/x[abc]*/', $s)`,
		},

		{
			Name:     "regexpVet",
			Default:  true,
			Quickfix: false,
			Comment:  `Report suspicious regexp patterns.`,
			Before:   `preg_match('a\d+a', $s); // 'a' is not a valid delimiter.`,
			After:    `preg_match('/\d+/', $s);`,
		},

		{
			Name:     "regexpSyntax",
			Default:  true,
			Quickfix: false,
			Comment:  `Report regexp syntax errors.`,
		},

		{
			Name:     "caseContinue",
			Default:  true,
			Quickfix: false,
			Comment:  "Report suspicious `continue` usages inside `switch` cases.",
			Before: `switch ($v) {
case STOP:
  continue; // Continue inside a switch is equivalent to break.
case INC:
  $x++;
  break;
}`,
			After: `switch ($v) {
case STOP:
  break;
case INC:
  $x++;
  break;
}`,
		},

		{
			Name:     "deprecated",
			Default:  true,
			Quickfix: false,
			Comment:  `Report usages of deprecated symbols.`,
			Before: `/**
 * @deprecated Use g() instead
 */
function f() {}

f();`,
			After: `/**
 * @deprecated Use g() instead
 */
function f() {}

g();`,
		},

		{
			Name:     "callStatic",
			Default:  true,
			Quickfix: false,
			Comment:  `Report static calls of instance methods and vice versa.`,
			Before:   `$object::instance_method() // instance_method is not a static method.`,
			After:    `$object->instance_method()`,
		},

		{
			Name:     "parentConstructor",
			Default:  true,
			Quickfix: false,
			Comment:  "Report missing `parent::__construct` calls in class constructors.",
			Before: `class Foo extends Bar {
  public function __construct($x, $y) {
    // Lost call to parent constructor.
    $this->y = $y;
  }
}`,
			After: `class Foo extends Bar {
  public function __construct($x, $y) {
    parent::__construct($x);
    $this->y = $y;
  }
}`,
		},

		{
			Name:     "oldStyleConstructor",
			Default:  true,
			Quickfix: false,
			Comment:  `Report old-style (PHP4) class constructors.`,
			Before: `class Foo {
  // Constructor in the old style of PHP 4.
  public function Foo($v) { $this->v = $v; }
}`,
			After: `class Foo {
  public function __construct($v) { $this->v = $v; }
}`,
		},

		{
			Name:     "stringInterpolationDeprecated",
			Default:  true,
			Quickfix: false,
			Comment:  `Report deprecated string interpolation style`,
			Before:   `${variable}`,
			After:    `{$variable}`,
		},

		{
			Name:     "misspellName",
			Default:  true,
			Quickfix: false,
			Comment:  `Report commonly misspelled words in symbol names.`,
			Before:   `function performace_test() ...`, //nolint:misspell // misspelled on purpose
			After:    `function performance_test() ...`,
		},

		//nolint:misspell // misspelled on purpose
		{
			Name:     "misspellComment",
			Default:  true,
			Quickfix: false,
			Comment:  `Report commonly misspelled words in comments.`,
			Before: `/** This is our performace test. */
function performance_test() {}`,
			After: `/** This is our performance test. */
function performance_test() {}`,
		},

		{
			Name:     "nonPublicInterfaceMember",
			Default:  true,
			Quickfix: false,
			Comment:  `Report illegal non-public access level in interfaces.`,
			Before: `interface Iface {
  function a();
  public function b();
  private function c(); // Methods in an interface cannot be private.
  protected function d(); // Methods in an interface cannot be protected.
}`,
			After: `interface Iface {
  function a();
  public function b();
  public function c();
  public function d();
}`,
		},

		{
			Name:     "linterError",
			Default:  true,
			Quickfix: false,
			Comment:  `Report internal linter error.`,
		},

		{
			Name:     "magicMethodDecl",
			Default:  true,
			Quickfix: false,
			Comment:  `Report issues in magic method declarations.`,
			Before: `class Foo {
  private function __call($method, $args) {} // The magic method __call() must have public visibility.
  public static function __set($name, $value) {} // The magic method __set() cannot be static.
}`,
			After: `class Foo {
  public function __call($method, $args) {}
  public function __set($name, $value) {}
}`,
		},

		{
			Name:     "nameMismatch",
			Default:  true,
			Quickfix: false,
			Comment:  `Report symbol case mismatches.`,
			Before: `class Foo {}
// The spelling is in lower case, although the class definition begins with an uppercase letter.
$foo = new foo();`,
			After: `class Foo {}
$foo = new Foo();`,
		},

		{
			Name:     "paramClobber",
			Default:  true,
			Quickfix: false,
			Comment:  `Report assignments that overwrite params prior to their usage.`,
			Before: `function api_get_video($user_id) {
  // The arguments are assigned a new value before using the value passed to the function.
  $user_id = 0;
  return get_video($user_id);
}`,
			After: `function api_get_video($user_id) {
  $user_id = $user_id ?: 0;
  return get_video($user_id);
}`,
		},

		{
			Name:     "printf",
			Default:  true,
			Quickfix: false,
			Comment:  `Report issues in printf-like function calls.`,
			Before:   `sprintf("id=%d") // Lost argument for '%d' specifier.`,
			After:    `sprintf("id=%d", $id)`,
		},

		{
			Name:     "discardVar",
			Default:  true,
			Quickfix: false,
			Comment:  "Report the use of variables that were supposed to be unused, like `$_`.",
			Before: `$_ = some();
echo $_;`,
			After: `$someVal = some();
echo $someVal;`,
		},

		{
			Name:     "dupCatch",
			Default:  true,
			Quickfix: false,
			Comment:  "Report duplicated `catch` clauses.",
			Before: `try {
  // some code
} catch (Exception1 $e) {
} catch (Exception1 $e) {} // <- Possibly the typo.`,
			After: `try {
  // some code
} catch (Exception1 $e) {
} catch (Exception2 $e) {}`,
		},

		{
			Name:     "catchOrder",
			Default:  true,
			Quickfix: false,
			Comment:  "Report erroneous `catch` order in `try` statements.",
			Before: `try {
  // Some code.
} catch (Exception $e) {
  // This will catch both Exception and TimeoutException.
} catch (TimeoutException $e) {
  // This is a dead code.
}`,
			After: `try {
  // Some code.
} catch (TimeoutException $e) {
  // Ok, it can catch TimeoutException.
} catch (Exception $e) {
  // Ok, it will catch everything else.
}`,
		},

		{
			Name:     "trailingComma",
			Default:  false,
			Quickfix: true,
			Comment:  `Report the absence of a comma for the last element in a multi-line array.`,
			Before: `$_ = [
  10,
  20 // Lost comma at the end for a multi-line array.
]`,
			After: `$_ = [
  10,
  20,
]`,
		},

		{
			Name:     "nestedTernary",
			Default:  true,
			Quickfix: false,
			Comment:  `Report an unspecified order in a nested ternary operator.`,
			Before:   `$_ = 1 ? 2 : 3 ? 4 : 5; // There is no clear order of execution.`,
			After: `$_ = (1 ? 2 : 3) ? 4 : 5;
// or
$_ = 1 ? 2 : (3 ? 4 : 5);`,
		},

		{
			Name:     "langDeprecated",
			Default:  false,
			Quickfix: false,
			Comment:  `Report the use of deprecated (per language spec) features.`,
			Before: `$a = (real)100; // 'real' has been deprecated.
$_ = is_real($a);`,
			After: `$a = (float)100;
$_ = is_float($a);`,
		},

		{
			Name:     "badTraitUse",
			Default:  true,
			Quickfix: false,
			Comment:  `Report misuse of traits.`,
			Before: `trait A {}
function f(A $a) {} // Traits cannot be used as type hints.`,
			After: `class A {}
function f(A $a) {}`,
		},

		{
			Name:     "typeHint",
			Default:  false,
			Quickfix: false,
			Comment:  `Report misuse of type hints.`,
			Before: `// The array typehint is too generic, you need to specify a specialization or mixed[] in PHPDoc.
function f(array $a) {}`,
			After: `/**
 * @param mixed[] $a
 */
function f(array $a) {}`,
		},

		{
			Name:     "argsOrder",
			Default:  true,
			Quickfix: false,
			Comment:  `Report suspicious arguments order.`,
			Before: `// It is possible that the arguments are in the wrong order, since 
// searching for a substring in a character does not make sense.
strpos('/', $s);`,
			After: `strpos($s, '/');`,
		},

		{
			Name:     "classMembersOrder",
			Default:  false,
			Quickfix: false,
			Comment:  `Report the wrong order of the class members.`,
			Before: `class A {
  // In the class, constants and properties should go first, and then methods.
  public function func() {}
  const B = 1;
  public $c = 2;
}`,
			After: `class A {
  const B = 1;
  public $c = 2;
  public function func() {}
}`,
		},

		{
			Name:     "varShadow",
			Default:  true,
			Quickfix: false,
			Comment:  `Report the shadow of an existing variable.`,
			Before: `function f(int $a) {
  // The $a variable hides the $a argument.
  foreach ([1, 2] as $a) {
    echo $a;
  }
}`,
			After: `function f(int $a) {
  foreach ([1, 2] as $b) {
    echo $b;
  }
}`,
		},

		{
			Name:     "propNullDefault",
			Default:  false,
			Quickfix: true,
			Comment:  `Report a null assignment for a not nullable property.`,
			Before: `class Foo {
  /**
   * @var Boo $item
   */
  public $item = null; // The type of the property is not nullable, but it is assigned null.
}`,
			After: `class Foo {
  /**
   * @var Boo $item
   */
  public $item;
}`,
		},

		{
			Name:     "switchDefault",
			Default:  false,
			Quickfix: false,
			Comment:  "Report the lack or wrong position of `default`.",
			Before: `switch ($a) {
  case 1:
    echo 1;
    break;
}`,
			After: `switch ($a) {
  case 1:
    echo 1;
    break;
  default:
    echo 2;
    break;
}`,
		},

		{
			Name:     "switchSimplify",
			Default:  true,
			Quickfix: false,
			Comment:  "Report possibility to rewrite `switch` with the `if`.",
			Before: `switch ($a) {
  case 1:
    echo 1;
    break;
}`,
			After: `if ($a == 1) {
  echo 1;
}`,
		},

		{
			Name:     "switchEmpty",
			Default:  true,
			Quickfix: false,
			Comment:  "Report `switch` with empty body.",
			Before:   `switch ($a) {}`,
			After: `switch ($a) {
  case 1:
    // do something
    break;
}`,
		},

		{
			Name:     "implicitModifiers",
			Default:  true,
			Quickfix: false,
			Comment:  `Report implicit modifiers.`,
			Before: `class Foo {
  function f() {} // The access modifier is implicit.
}`,
			After: `class Foo {
  public function f() {}
}`,
		},

		{
			Name:     "invalidExtendClass",
			Default:  true,
			Quickfix: false,
			Comment:  `Report inheritance from the final class.`,
			Before: `final class Foo {}
class Boo extends Foo {}`,
			After: `class Foo {}
class Boo extends Foo {}`,
		},

		{
			Name:     "methodSignatureMismatch",
			Default:  true,
			Quickfix: false,
			Comment:  `Report a method signature mismatch in inheritance.`,
			Before: `class Foo {
  final public function f() {}
}

class Boo extends Foo {
  public function f() {} // Foo::f is final.
}`,
			After: `class Foo {
  public function f() {}
}

class Boo extends Foo {
  public function f() {}
}`,
		},

		{
			// Checker can give many false positives, however it is
			// useful for periodic checking when you can choose what
			// appears to be a real error.
			Name:     "argsReverse",
			Default:  false,
			Quickfix: false,
			Comment:  `Report using variables as arguments in reverse order.`,
			Before: `function makeHello(string $name, int $age) {
  echo "Hello ${$name}-${$age}";
}

function main(): void {
  $name = "John";
  $age = 18;
  makeHello($age, $name); // The name should come first, and then the age.
}`,
			After: `function makeHello(string $name, int $age) {
  echo "Hello ${$name}-${$age}";
}

function main(): void {
  $name = "John";
  $age = 18;
  makeHello($name, $age);
}`,
		},

		{
			Name:     "strangeCast",
			Default:  true,
			Quickfix: false,
			Comment:  `Report a strange way of type cast.`,
			Before:   `$x.""`,
			After:    `(string)$x`,
		},

		{
			Name:     "reverseAssign",
			Default:  true,
			Quickfix: false,
			Comment:  `Report a reverse assign with unary plus or minus.`,
			Before:   `$a =+ 100;`,
			After:    `$a += 100;`,
		},

		{
			Name:     "parentNotFound",
			Default:  false,
			Quickfix: false,
			Comment:  "Report using `parent::` in a class without a parent class.",
			Before: `class Foo {
  public function f() {
    parent::b(); // Class Foo has no parent.
  }
}`,
			After: `class Foo extends Boo {
  public function f() {
    parent::b(); // Ok.
  }
}`,
		},

		{
			Name:     "packaging",
			Default:  false,
			Quickfix: false,
			Comment:  "Report call @internal method outside @package.",
			Before: `// file Boo.php 

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
}`,
			After: `// file Boo.php 

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
}`,
		},

		{
			Name:     "getTypeMisUse",
			Default:  false,
			Quickfix: true,
			Comment:  `Report call gettype function.`,
			Before:   `if (gettype($a) == "string") { ... }`,
			After:    `if (is_string($a)) { ... }`,
		},
	}

	for _, info := range allChecks {
		reg.DeclareChecker(info)
	}
}

// Report is a linter report message.
type Report struct {
	CheckName string `json:"check_name"`
	Level     int    `json:"level"`
	Context   string `json:"context"`
	Message   string `json:"message"`
	Filename  string `json:"filename"`
	Line      int    `json:"line"`
	StartChar int    `json:"start_char"`
	EndChar   int    `json:"end_char"`
	Hash      uint64 `json:"hash"`
}

var severityNames = map[int]string{
	LevelError:    "ERROR",
	LevelWarning:  "WARNING",
	LevelNotice:   "MAYBE",
	LevelSecurity: "WARNING",
}

func (r *Report) Severity() string {
	return severityNames[r.Level]
}

// IsCritical returns whether or not we need to reject whole commit when found this kind of report.
func (r *Report) IsCritical() bool {
	return r.Level != LevelNotice
}

// DiffReports returns only reports that are new.
// Pass diffArgs=nil if we are called from diff in working copy.
func DiffReports(gitRepo string, diffArgs []string, changesList []git.Change, changeLog []git.Commit, oldList, newList []*Report, maxConcurrency int) (res []*Report, err error) {
	ignoreCommits := make(map[string]struct{})
	for _, c := range changeLog {
		if strings.Contains(c.Message, IgnoreLinterMessage) {
			ignoreCommits[c.Hash] = struct{}{}
		}
	}

	old := reportListToMap(oldList)
	new := reportListToMap(newList)
	changes := gitChangesToMap(changesList)

	var mu sync.Mutex
	var wg sync.WaitGroup

	var resErr error

	limitCh := make(chan struct{}, maxConcurrency)

	for filename, list := range new {
		wg.Add(1)
		go func(filename string, list []*Report) {
			limitCh <- struct{}{}
			defer func() { <-limitCh }()
			defer wg.Done()

			var oldName string

			c, ok := changes[filename]
			if ok {
				oldName = c.OldName
			} else {
				oldName = filename // full diff mode
			}

			reports, err := diffReportsList(gitRepo, ignoreCommits, diffArgs, filename, c, old[oldName], list)
			if err != nil {
				mu.Lock()
				resErr = err
				mu.Unlock()
				return
			}

			mu.Lock()
			res = append(res, reports...)
			mu.Unlock()
		}(filename, list)
	}

	wg.Wait()

	if resErr != nil {
		return nil, err
	}

	return res, nil
}

type lineRangeChange struct {
	old, new git.LineRange
}

// compute blame only if refspec is not nil
func blameIfNeeded(gitDir string, refspec []string, filename string) (git.BlameResult, error) {
	if refspec == nil {
		return git.BlameResult{}, nil
	}

	return git.Blame(gitDir, refspec, filename)
}

func diffReportsList(gitRepo string, ignoreCommits map[string]struct{}, diffArgs []string, filename string, c git.Change, oldList, newList []*Report) (res []*Report, err error) {
	var blame git.BlameResult

	if c.Valid {
		blame, err = blameIfNeeded(gitRepo, diffArgs, filename)
		if err != nil {
			return nil, err
		}
	}

	changesMap := make(map[int]lineRangeChange, len(c.OldLineRanges))

	for idx, r := range c.OldLineRanges {
		for i := r.From; i <= r.To; i++ {
			changesMap[i] = lineRangeChange{old: r, new: c.LineRanges[idx]}
		}
	}

	old, oldMaxLine := reportListToPerLineMap(oldList)
	new, newMaxLine := reportListToPerLineMap(newList)

	var maxLine = oldMaxLine
	if newMaxLine > maxLine {
		maxLine = newMaxLine
	}

	var oldLine, newLine int

	for i := 0; i < maxLine; i++ {
		oldLine++
		newLine++

		ch, ok := changesMap[oldLine]
		// just deletion
		if ok && ch.new.HaveRange && ch.new.Range == 0 {
			oldLine = ch.old.To
			newLine-- // cancel the increment of newLine, because code was deleted, no new lines added
			continue
		}

		res = maybeAppendReports(res, new, old, newLine, oldLine, blame, ignoreCommits)

		if ok {
			oldLine = 0 // all changes and additions must be checked
			for j := newLine + 1; j <= ch.new.To; j++ {
				newLine = j
				res = maybeAppendReports(res, new, old, newLine, oldLine, blame, ignoreCommits)
			}
			oldLine = ch.old.To
		}
	}

	return res, nil
}

func maybeAppendReports(res []*Report, new, old map[int][]*Report, newLine, oldLine int, blame git.BlameResult, ignoreCommits map[string]struct{}) []*Report {
	newReports, ok := new[newLine]

	if !ok {
		return res
	}

	if _, ok := old[oldLine]; ok {
		return res
	}

	changedCommit := blame.Lines[newLine]

	if _, ok := ignoreCommits[changedCommit]; ok {
		return res
	}

	return append(res, newReports...)
}

func reportListToPerLineMap(list []*Report) (res map[int][]*Report, maxLine int) {
	res = make(map[int][]*Report)

	for _, l := range list {
		res[l.Line] = append(res[l.Line], l)
		if l.Line > maxLine {
			maxLine = l.Line
		}
	}

	return res, maxLine
}

func gitChangesToMap(changes []git.Change) map[string]git.Change {
	res := make(map[string]git.Change)
	for _, c := range changes {
		res[c.NewName] = c
	}
	return res
}

func reportListToMap(list []*Report) map[string][]*Report {
	res := make(map[string][]*Report)

	for _, r := range list {
		res[r.Filename] = append(res[r.Filename], r)
	}

	for i := range res {
		l := res[i]
		sort.Slice(l, func(i, j int) bool {
			return l[i].Line < l[j].Line
		})
	}

	return res
}

func isUnderscore(s string) bool {
	return s == "_"
}

func linterError(filename, format string, args ...interface{}) {
	log.Printf("error: "+filename+": "+format, args...)
}
