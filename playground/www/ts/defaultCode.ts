const defaultCode = `<?php

class FooWithFinalMethod {
    final function f() {}
}

class BooWithSameMethod extends FooWithFinalMethod {
    function f() {}
}

abstract class AbstractClass {
    abstract public function abstractMethod() {}
}

class SomeClass extends AbstractClass {}

/**
 * @method int a
 */
class Boo {}

/**
 * @method void check()
 */
final class Foo {
    var $prop, $prop2;
  
    /**
     * @var Boo
     */
    var $p = null;

    /**
     * Instance method
     */
    function instanceMethod(int $x) {}
    
    final public static function staticMethod(int $x) { 
        echo $this->p;
    }

    public function __call($name) {}
}

/**
 * @param  array{int,Foo} $x1
 * @return array{int,Foo}
 */
function getArray(array $x) { return [0, new Foo]; }

/**
 * @param callable(int) $a
 * @param callable(int): Foo $b
 */
function mainCheck(callable $a, callable $b) {
    echo getArray();
    (new Foo)->instanceMethod();

    echo getArray(10)[1]->p->f;

    echo (new Foo)->check();

    /**
     * @return callable(int, string): Foo
     */
    $b = function (int $a) { };
    $c = $b(10);
    $c();
}

function makeHello(string $name, int $age) {
    echo "Hello \${$name}-\${$age1}";
}

function main(): void {
    $name = "John";
    $age = 18;
    echo makeHello($age, $name);
}
`
