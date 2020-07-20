<?php

class TestClass {}

$bad = new NonExisting();

$v = new TestClass();
var_dump($v->foo);
