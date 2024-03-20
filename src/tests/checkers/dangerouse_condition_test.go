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

if(1){
}

`)
	test.Expect = []string{
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
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
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous value: you have constant int value that interpreted`,
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
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
	}
	test.RunAndMatch()
}

func TestDangerousCondition4(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

$a = true;
if($a || false && 1 || true){
}

`)
	test.Expect = []string{
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
	}
	test.RunAndMatch()
}
