package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDangerousCondition1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
if(true){
}
`)
	test.Expect = []string{
		`Potential dangerous bool value: you have constant bool value in condition at _file0.php:2`,
	}
	test.RunAndMatch()
}

func TestDangerousCondition2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

$a = true;
if(true||$a){
echo "test";
}

if(1||$a||1||true||false||0){
}
`)
	test.Expect = []string{
		`Potential dangerous bool value: you have constant bool value in condition at _file0.php:4`,
		`Potential dangerous value: you have constant int value that interpreted as bool at _file0.php:8`,
		`Potential dangerous value: you have constant int value that interpreted as bool at _file0.php:8`,
		`Potential dangerous bool value: you have constant bool value in condition at _file0.php:8`,
		`Potential dangerous bool value: you have constant bool value in condition at _file0.php:8`,
		`Potential dangerous value: you have constant int value that interpreted as bool at _file0.php:8`,
	}
	test.RunAndMatch()
}

func TestDangerousCondition3(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

$a = true;
if($a && false && true && 1 && 0){
}

`)
	test.Expect = []string{
		`Potential dangerous bool value: you have constant bool value in condition at _file0.php:4`,
		`Potential dangerous bool value: you have constant bool value in condition at _file0.php:4`,
		`Potential dangerous value: you have constant int value that interpreted as bool at _file0.php:4`,
		`Potential dangerous value: you have constant int value that interpreted as bool at _file0.php:4`,
	}
	test.RunAndMatch()
}
