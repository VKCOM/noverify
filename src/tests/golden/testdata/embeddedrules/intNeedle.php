<?php

$str = "hello";

$_ = strpos($str, 10);
$_ = strpos($str, (string)10); // ok
$_ = strpos($str, chr(10)); // ok

$_ = strrpos($str, 10);
$_ = strrpos($str, (string)10); // ok
$_ = strrpos($str, chr(10)); // ok

$_ = stripos($str, 10);
$_ = stripos($str, (string)10); // ok
$_ = stripos($str, chr(10)); // ok

$_ = strripos($str, 10);
$_ = strripos($str, (string)10); // ok
$_ = strripos($str, chr(10)); // ok

$_ = strstr($str, 10);
$_ = strstr($str, (string)10); // ok
$_ = strstr($str, chr(10)); // ok

$_ = strchr($str, 10);
$_ = strchr($str, (string)10); // ok
$_ = strchr($str, chr(10)); // ok

$_ = strrchr($str, 10);
$_ = strrchr($str, (string)10); // ok
$_ = strrchr($str, chr(10)); // ok

$_ = stristr($str, 10);
$_ = stristr($str, (string)10); // ok
$_ = stristr($str, chr(10)); // ok
