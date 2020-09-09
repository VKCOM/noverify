<?php

class Foo {}

function f() {
    $_ = array();
    $_ = array(1,2,);
    $_ = array(1,2,3);
    $_ = array(new Foo, new Foo);
    $_ = array("info" => new Foo, new Foo);
    $_ = array("info" => new Foo, "for" => 2, "home" => function() {});

    // So far, we cannot replace recursively, so replacement in
    // multidimensional arrays will only be performed for internal arrays.
    $_ = array(array(1,2,3), array(1,2,3));
    $_ = array(array("info" => new Foo, "for" => 2, "home" => function() {}), array(1,2,3));
}
