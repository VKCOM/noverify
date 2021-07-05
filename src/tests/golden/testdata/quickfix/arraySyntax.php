<?php

class Foo {}

$_ = array(1,2,3);
$_ = array( new Foo,    "Hello",
        1,2,
        3,4,
);

function f() {
    $_ = array();
    $_ = array(1,2,);
    $_ = array(1,2,3);
    $_ = array  (  1,   2, 3  )   ;
    $_ =   array  (  1,   2, 3  )   ;
    $_ = array(new Foo, new Foo);
    $_ = array("info" => new Foo, new Foo);
    $_ = array("info" => new Foo, "for" => 2, "home" => function() {});

    $_ = array(
        "info" => new Foo,
        new Foo
    );
    $_ = array(
        1,2,
        3,4,
    );
    $_ = array( new Foo,    "Hello",
        1,2,
        3,4,
    );

    b(array(1,2));

    // So far, we cannot replace recursively, so replacement in
    // multidimensional arrays will only be performed for external arrays.
    $_ = array(array(1,2,3), array(1,2,3));
    $_ = array(array("info" => new Foo, "for" => 2, "home" => function() {}), array(1,2,3));
}

function b($a) {
    switch ($a) {
    case 1:
        return array(1, 2, 3);
    case 2:
        return array("info" => new Foo, "for" => 2, "home" => function() {});
    }
}
