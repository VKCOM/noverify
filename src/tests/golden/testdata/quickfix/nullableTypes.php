<?php

function f(string $filed = null) {
    return 1;
}

function notNullableCallable(callable $a = null) {
    return 0;
}

class MyClass1 {
}

class MyClass2 {
    public function myMethod(MyClass1 $a = null) {
        return 0;
    }
}

function nullableArray(array $a = null) {
	return 0;
}