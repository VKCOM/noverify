package checkers_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

func BenchmarkExprType(b *testing.B) {
	test := linttest.NewSuite(b)
	test.AddFile(`<?php
/** @var int[] $xs */
$xs = [1, 2, 3];

function f1()       { global $xs; return $xs[1]; }
function f2(int $x) { return (int)f1() + $x; }
function f3()       { return f2(10); }

class Foo {
  public $i = 10;

  public function m1() {
    return f3() + $this->i;
  }

  public static function create() {
    return new Foo();
  }
}
`)
	test.RunLinter()
	meta.SetIndexingComplete(true)

	newName := func(nm string) *ir.Name {
		stringParts := strings.Split(nm, `\`)
		nameParts := make([]ir.Node, len(stringParts))
		for i, p := range stringParts {
			nameParts[i] = &ir.NamePart{Value: p}
		}
		return &ir.Name{Parts: nameParts}
	}
	f1call := &ir.FunctionCallExpr{Function: newName("f1")}
	f3call := &ir.FunctionCallExpr{Function: newName("f3")}
	foovar := &ir.SimpleVar{Name: "foo"}
	m4call := &ir.MethodCallExpr{Variable: foovar, Method: &ir.Identifier{Value: "m1"}}
	newpropfetch := &ir.PropertyFetchExpr{
		Variable: &ir.StaticCallExpr{
			Call:  &ir.Identifier{Value: "create"},
			Class: newName("Foo"),
		},
		Property: &ir.Identifier{Value: "i"},
	}

	st := &meta.ClassParseState{}
	sc := meta.NewScope()

	sc.AddVarName("foo", meta.NewTypesMap(`\Foo|int|null`), "test", meta.VarAlwaysDefined)

	b.Run("simplevar", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = solver.ExprType(sc, st, foovar)
		}
	})
	b.Run("f1call", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = solver.ExprType(sc, st, f1call)
		}
	})
	b.Run("f3call", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = solver.ExprType(sc, st, f3call)
		}
	})
	b.Run("m4call", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = solver.ExprType(sc, st, m4call)
		}
	})
	b.Run("newpropfetch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = solver.ExprType(sc, st, newpropfetch)
		}
	})
}
