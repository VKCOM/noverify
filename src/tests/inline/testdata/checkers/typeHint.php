<?php

class Boo {}

class FooBad {
    public parent $a; // want `Cannot use 'parent' typehint when current class has no parent`

    public function f1(): parent { // want `Cannot use 'parent' typehint when current class has no parent`
        return new Boo;
    }

    public function f2(parent $a) { // want `Cannot use 'parent' typehint when current class has no parent`
        return new Boo;
    }
}

class FooOk extends Boo {
    public parent $a;

    public function f(): parent {
        return new Boo;
    }

    public function f2(parent $a) {
        return new Boo;
    }
}
