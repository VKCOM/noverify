package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDangerousConditionExplicitBool(t *testing.T) {
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

func TestDangerousConditionElseIf(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
if(1){
} elseif (true){

}


`)
	test.Expect = []string{
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
	}
	test.RunAndMatch()
}

func TestDangerousConditionExplicitBoolMultiOr(t *testing.T) {
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

func TestDangerousConditionExplicitBoolMultiAnd(t *testing.T) {
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

func TestDangerousConditionExplicitBoolMultiOrAnd(t *testing.T) {
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

func TestDangerousConditionCycles(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

while(true){
}

do{

}while(1);
`)
	test.Expect = []string{
		`Potential dangerous bool value: you have constant bool value in condition`,
		`Potential dangerous value: you have constant int value that interpreted as bool`,
	}
	test.RunAndMatch()
}
