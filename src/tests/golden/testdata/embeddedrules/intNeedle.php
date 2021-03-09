<?php

function getInt(): int {
    return 0;
}

/**
 * @param  bool  $a
 *
 * @return int|string
 */
function getIntOrString(bool $a) {
    if ($a) {
        return "Hello";
    }
    return 0;
}

$str = "hello";

$_ = strpos($str, 10);
$_ = strpos($str, getInt());
$_ = strpos($str, getIntOrString(true)); // ok
$_ = strpos($str, (string)10); // ok
$_ = strpos($str, chr(10)); // ok

$_ = strrpos($str, 10);
$_ = strrpos($str, getInt());
$_ = strrpos($str, getIntOrString(true)); // ok
$_ = strrpos($str, (string)10); // ok
$_ = strrpos($str, chr(10)); // ok

$_ = stripos($str, 10);
$_ = stripos($str, getInt());
$_ = stripos($str, getIntOrString(true)); // ok
$_ = stripos($str, (string)10); // ok
$_ = stripos($str, chr(10)); // ok

$_ = strripos($str, 10);
$_ = strripos($str, getInt());
$_ = strripos($str, getIntOrString(true)); // ok
$_ = strripos($str, (string)10); // ok
$_ = strripos($str, chr(10)); // ok

$_ = strstr($str, 10);
$_ = strstr($str, getInt());
$_ = strstr($str, getIntOrString(true)); // ok
$_ = strstr($str, (string)10); // ok
$_ = strstr($str, chr(10)); // ok

$_ = strchr($str, 10);
$_ = strchr($str, getInt());
$_ = strchr($str, getIntOrString(true)); // ok
$_ = strchr($str, (string)10); // ok
$_ = strchr($str, chr(10)); // ok

$_ = strrchr($str, 10);
$_ = strrchr($str, getInt());
$_ = strrchr($str, getIntOrString(true)); // ok
$_ = strrchr($str, (string)10); // ok
$_ = strrchr($str, chr(10)); // ok

$_ = stristr($str, 10);
$_ = stristr($str, getInt());
$_ = stristr($str, getIntOrString(true)); // ok
$_ = stristr($str, (string)10); // ok
$_ = stristr($str, chr(10)); // ok
