<?php

final class FooFinal {}
class FooNotFinal {}
class BooFinal extends FooFinal {} // want `Class \BooFinal may not inherit from final class \FooFinal`
class BooNotFinal extends FooNotFinal {} // ok
class ZooFinal extends BooFinal {} // ok
