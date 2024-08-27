<?php

function notNullableString(string $filed = null) {}

function notNullableCallable(callable $a = null) {}

class MyClass1 {}

class MyClass2 {
    public function myMethod(MyClass1 $a = null) {}
}

function nullableArray(array $a = null) {}

function multipleArgsExample(string $a, int $b = null, bool $c = null) {}

function nullableOrString(null|string $a = null) {}

function mixedParam(mixed $a = null) {}
