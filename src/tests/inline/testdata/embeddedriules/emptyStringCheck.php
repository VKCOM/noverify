<?php

function emptyStringCheck() {
    $x = "hello";

    if (strlen($x)) { // want `use '$x !== ""' instead`
        echo 1;
    }

    if (mb_strlen($x)) { // want `use '$x !== ""' instead`
        echo 1;
    }

    if ($x || strlen($x)) { // want `use '$x !== ""' instead`
        echo 1;
    }

    if (!strlen($x)) { // want `use '$x === ""' instead`
        echo 1;
    }

    if (!mb_strlen($x)) { // want `use '$x === ""' instead`
        echo 1;
    }
}
