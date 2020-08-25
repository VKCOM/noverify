package phpdoc

import (
	"fmt"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{``, `Invalid=""`},

		// Names.
		{`a`, `Name="a"`},
		{`\`, `Name="\"`},
		{`foo`, `Name="foo"`},
		{`\A\B`, `Name="\A\B"`},
		{`$this`, `Name="$this"`},

		// Ints.
		{`0`, `Int="0"`},
		{`1249`, `Int="1249"`},

		// Special names.
		{`*`, `SpecialName="*"`},
		{`...`, `SpecialName="..."`},

		// Parens.
		{`(x)`, `Paren="(x)"{Name="x"}`},
		{`((x))`, `Paren="((x))"{Paren="(x)"{Name="x"}}`},
		{`()`, `Paren="()"{Invalid=""}`},

		// Unclosed parens.
		{`(x`, `Paren="(x"{Name="x"}`},
		{`((x`, `Paren="((x"{Paren="(x"{Name="x"}}`},

		// Arrays.
		{`int[]`, `Array="int[]"{Name="int"}`},
		{`int[][]`, `Array="int[][]"{Array="int[]"{Name="int"}}`},

		// Prefix arrays.
		{`[]int`, `PrefixArray="[]int"{Name="int"}`},
		{`[][]int`, `PrefixArray="[][]int"{PrefixArray="[]int"{Name="int"}}`},

		// Unions.
		{`x|y`, `Union="x|y"{Name="x" Name="y"}`},
		{`x|y|z`, `Union="x|y|z"{Name="x" Name="y" Name="z"}`},
		{`(x|)`, `Paren="(x|)"{Union="x|"{Name="x" Invalid=""}}`},
		{`x|`, `Union="x|"{Name="x" Invalid=""}`},
		{`[](int|float)`, `PrefixArray="[](int|float)"{Paren="(int|float)"{Union="int|float"{Name="int" Name="float"}}}`},

		// Intersections.
		{`x&y`, `Inter="x&y"{Name="x" Name="y"}`},
		{`x&y&\z`, `Inter="x&y&\z"{Name="x" Name="y" Name="\z"}`},

		// Nullable types.
		{`?x`, `Nullable="?x"{Name="x"}`},
		{`??x`, `Nullable="??x"{Nullable="?x"{Name="x"}}`},

		// Negated (Not) types.
		{`!x`, `Not="!x"{Name="x"}`},
		{`!?x`, `Not="!?x"{Nullable="?x"{Name="x"}}`},

		// Generic types.
		{`A<>`, `Generic="A<>"{Name="A"}`},
		{`A<`, `Generic="A<"{Name="A"}`},
		{`list<int>`, `Generic="list<int>"{Name="list" Name="int"}`},
		{`A<B,C>`, `Generic="A<B,C>"{Name="A" Name="B" Name="C"}`},
		{`A< T1 , T2 >`, `Generic="A< T1 , T2 >"{Name="A" Name="T1" Name="T2"}`},
		{`A<T,>`, `Generic="A<T,>"{Name="A" Name="T"}`},
		{`A<int[],B|C>`, `Generic="A<int[],B|C>"{Name="A" Array="int[]"{Name="int"} Union="B|C"{Name="B" Name="C"}}`},
		{`array<int>`, `Generic="array<int>"{Name="array" Name="int"}`},
		{`array<int,string>`, `Generic="array<int,string>"{Name="array" Name="int" Name="string"}`},
		{`array<int, array<string, stdclass> >`, `Generic="array<int, array<string, stdclass> >"{Name="array" Name="int" Generic="array<string, stdclass>"{Name="array" Name="string" Name="stdclass"}}`},
		{`?A<B>`, `Nullable="?A<B>"{Generic="A<B>"{Name="A" Name="B"}}`},

		// Alternative generic syntax 1.
		{`tuple(*)`, `GenericParen="tuple(*)"{Name="tuple" SpecialName="*"}`},
		{`tuple(int,float)`, `GenericParen="tuple(int,float)"{Name="tuple" Name="int" Name="float"}`},
		{`tuple(T)|false`, `Union="tuple(T)|false"{GenericParen="tuple(T)"{Name="tuple" Name="T"} Name="false"}`},
		{`tuple(int|float,bool)`, `GenericParen="tuple(int|float,bool)"{Name="tuple" Union="int|float"{Name="int" Name="float"} Name="bool"}`},

		// Alternative generic syntax 2.
		{`tuple{int, int}`, `GenericBrace="tuple{int, int}"{Name="tuple" Name="int" Name="int"}`},
		{`array{a: int, b: float}`, `GenericBrace="array{a: int, b: float}"{Name="array" KeyVal="a: int"{Name="a" Name="int"} KeyVal="b: float"{Name="b" Name="float"}}`},
		{`array{a : int}`, `GenericBrace="array{a : int}"{Name="array" KeyVal="a : int"{Name="a" Name="int"}}`},

		// Typed callable (see #537).
		// TODO: add variadic params support.
		{`callable() : void`, `TypedCallable="callable() : void"{Name="void"}`},
		{`callable(A):B`, `TypedCallable="callable(A):B"{Name="B" Name="A"}`},
		{`(callable (A, B) : C)`, `Paren="(callable (A, B) : C)"{TypedCallable="callable (A, B) : C"{Name="C" Name="A" Name="B"}}`},
		{`?callable() : int`, `Nullable="?callable() : int"{TypedCallable="callable() : int"{Name="int"}}`},
		{`callable(A) : callable(B) : int`, `TypedCallable="callable(A) : callable(B) : int"{TypedCallable="callable(B) : int"{Name="int" Name="B"} Name="A"}`},

		// The typed callable without return type is parsed as generic type.
		// Only () shape is recognized.
		{`callable(A, B)`, `GenericParen="callable(A, B)"{Name="callable" Name="A" Name="B"}`},
		{`callable()`, `GenericParen="callable()"{Name="callable"}`},
		{`callable<A>:B`, `KeyVal="callable<A>:B"{Generic="callable<A>"{Name="callable" Name="A"} Name="B"}`},

		// KeyVal types.
		{`name:int`, `KeyVal="name:int"{Name="name" Name="int"}`},
		{`array{foo: int}`, `GenericBrace="array{foo: int}"{Name="array" KeyVal="foo: int"{Name="foo" Name="int"}}`},
		{`array{0: int}`, `GenericBrace="array{0: int}"{Name="array" KeyVal="0: int"{Int="0" Name="int"}}`},
		{`shape{s?: string}`, `GenericBrace="shape{s?: string}"{Name="shape" KeyVal="s?: string"{Optional="s?"{Name="s"} Name="string"}}`},
		{`foo(A):B`, `KeyVal="foo(A):B"{GenericParen="foo(A)"{Name="foo" Name="A"} Name="B"}`},

		// MemberType.
		{`\Foo::CONST`, `MemberType="\Foo::CONST"{Name="\Foo" Name="CONST"}`},
		{`\Foo::CONST|\Foo::CONST2`, `Union="\Foo::CONST|\Foo::CONST2"{MemberType="\Foo::CONST"{Name="\Foo" Name="CONST"} MemberType="\Foo::CONST2"{Name="\Foo" Name="CONST2"}}`},
		{`\Foo::$a|\Foo::CONST2`, `Union="\Foo::$a|\Foo::CONST2"{MemberType="\Foo::$a"{Name="\Foo" Name="$a"} MemberType="\Foo::CONST2"{Name="\Foo" Name="CONST2"}}`},
		// MemberType and other.
		{`\Foo::CONST|string`, `Union="\Foo::CONST|string"{MemberType="\Foo::CONST"{Name="\Foo" Name="CONST"} Name="string"}`},
		{`\Foo::$a|x&(y|z)`, `Union="\Foo::$a|x&(y|z)"{MemberType="\Foo::$a"{Name="\Foo" Name="$a"} Inter="x&(y|z)"{Name="x" Paren="(y|z)"{Union="y|z"{Name="y" Name="z"}}}}`},
		{`\Foo::CONST|shape(i:int, ...)`, `Union="\Foo::CONST|shape(i:int, ...)"{MemberType="\Foo::CONST"{Name="\Foo" Name="CONST"} GenericParen="shape(i:int, ...)"{Name="shape" KeyVal="i:int"{Name="i" Name="int"} SpecialName="..."}}`},
		{`\Boo::CONST|?x[]`, `Union="\Boo::CONST|?x[]"{MemberType="\Boo::CONST"{Name="\Boo" Name="CONST"} Array="?x[]"{Nullable="?x"{Name="x"}}}`},
		{`Foo::$a|[](int|float)`, `Union="Foo::$a|[](int|float)"{MemberType="Foo::$a"{Name="Foo" Name="$a"} PrefixArray="[](int|float)"{Paren="(int|float)"{Union="int|float"{Name="int" Name="float"}}}}`},
		{`self::CONST|?callable() : int`, `Union="self::CONST|?callable() : int"{MemberType="self::CONST"{Name="self" Name="CONST"} Nullable="?callable() : int"{TypedCallable="callable() : int"{Name="int"}}}`},
		{`self::$a|tuple(T)|false`, `Union="self::$a|tuple(T)|false"{MemberType="self::$a"{Name="self" Name="$a"} GenericParen="tuple(T)"{Name="tuple" Name="T"} Name="false"}`},
		{`\Space\Foo::CONST|A<>`, `Union="\Space\Foo::CONST|A<>"{MemberType="\Space\Foo::CONST"{Name="\Space\Foo" Name="CONST"} Generic="A<>"{Name="A"}}`},
		{`Foo\Boo::CONST|!?x`, `Union="Foo\Boo::CONST|!?x"{MemberType="Foo\Boo::CONST"{Name="Foo\Boo" Name="CONST"} Not="!?x"{Nullable="?x"{Name="x"}}}`},

		// Intersection types has higher priority that union types.
		{`x&y|z`, `Union="x&y|z"{Inter="x&y"{Name="x" Name="y"} Name="z"}`},
		{`x&(y|z)`, `Inter="x&(y|z)"{Name="x" Paren="(y|z)"{Union="y|z"{Name="y" Name="z"}}}`},

		// Mixing different kinds of expressions.
		{`?x|?y`, `Union="?x|?y"{Nullable="?x"{Name="x"} Nullable="?y"{Name="y"}}`},
		{`?x[]`, `Array="?x[]"{Nullable="?x"{Name="x"}}`},
		{`??x[]`, `Array="??x[]"{Nullable="??x"{Nullable="?x"{Name="x"}}}`},
		{`??x[][]`, `Array="??x[][]"{Array="??x[]"{Nullable="??x"{Nullable="?x"{Name="x"}}}}`},
		{`?x[][]`, `Array="?x[][]"{Array="?x[]"{Nullable="?x"{Name="x"}}}`},
		{`[]x|y`, `Union="[]x|y"{PrefixArray="[]x"{Name="x"} Name="y"}`},
		{`?[]x|y`, `Union="?[]x|y"{Nullable="?[]x"{PrefixArray="[]x"{Name="x"}} Name="y"}`},
		{`!x|?x`, `Union="!x|?x"{Not="!x"{Name="x"} Nullable="?x"{Name="x"}}`},
		{`!x&?x`, `Inter="!x&?x"{Not="!x"{Name="x"} Nullable="?x"{Name="x"}}`},
		{`shape(i:int, ...)`, `GenericParen="shape(i:int, ...)"{Name="shape" KeyVal="i:int"{Name="i" Name="int"} SpecialName="..."}`},

		// Some whitespace should be tolerated.
		{`(x | y)`, `Paren="(x | y)"{Union="x | y"{Name="x" Name="y"}}`},
		{`( x| y)`, `Paren="( x| y)"{Union="x| y"{Name="x" Name="y"}}`},
		{` ( (string))`, `Paren="( (string))"{Paren="(string)"{Name="string"}}`},
		{` ((string ) ) `, `Paren="((string ) )"{Paren="(string )"{Name="string"}}`},
		{`( [] int)`, `Paren="( [] int)"{PrefixArray="[] int"{Name="int"}}`},

		// If no postfix/infix token is found, the parser stops.
		{`x?y`, `Optional="x?"{Name="x"}`},
		{`x[]y`, `Array="x[]"{Name="x"}`},
		{`() $x`, `Paren="()"{Invalid=""}`},
		{`@ @`, `Invalid="@"`},
		{`@ @ | x`, `Invalid="@"`},
		{`@ @| x`, `Invalid="@"`},
		{`x| @ @`, `Union="x| "{Name="x" Invalid=" "}`},
		{`x &$x`, `Name="x"`},
		{`x [ ][  ]`, `Name="x"`},
		{`tuple {int, int}`, `Name="tuple"`},
		{`x |y`, `Name="x"`},
		{`x| y`, `Union="x| "{Name="x" Invalid=" "}`},
		{`[] int`, `PrefixArray="[] "{Invalid=" "}`},

		// Unknown expressions.
		{`-foo`, `Unknown="-foo"{Name="foo"}`},
		{`--foo`, `Unknown="--foo"{Name="foo"}`},
		{`x|@foo`, `Union="x|@foo"{Name="x" Unknown="@foo"{Name="foo"}}`},

		// Invalid expressions.
		{`@`, `Invalid="@"`},
		{`@#%`, `Invalid="@#%"`},
		{`x|@`, `Union="x|@"{Name="x" Invalid="@"}`},
		{`x|@@`, `Union="x|@@"{Name="x" Invalid="@@"}`},
		{`A<|b`, `Generic="A<|b"{Name="A" Unknown="|b"{Name="b"}}`},
		{`A<,>`, `Generic="A<,>"{Name="A" Invalid=","}`},
		{`A<,,>`, `Generic="A<,,>"{Name="A" Invalid=",,"}`},
		{`A<,>|B`, `Union="A<,>|B"{Generic="A<,>"{Name="A" Invalid=","} Name="B"}`},
		{`..`, `Invalid=".."`},

		// Array types without closing ']' or without element type.
		{`[`, `PrefixArray="["{Invalid=""}`},
		{`T[`, `Array="T["{Name="T"}`},
		{`T[|b`, `Union="T[|b"{Array="T["{Name="T"} Name="b"}`},
		{`[T`, `PrefixArray="[T"{Name="T"}`},
		{`[T|b`, `Union="[T|b"{PrefixArray="[T"{Name="T"} Name="b"}`},
	}

	p := NewTypeParser()
	for _, test := range tests {
		typ := p.Parse(test.input)
		have := exprSyntax(typ.Expr)
		if have != test.want {
			t.Errorf("parse(`%s`):\nhave: %s\nwant: %s", test.input, have, test.want)
		}
	}
}

func exprSyntax(e TypeExpr) string {
	kind := e.Kind.String()
	switch e.Shape {
	case ShapeArrayPrefix:
		kind = "PrefixArray"
	case ShapeGenericParen:
		kind = "GenericParen"
	case ShapeGenericBrace:
		kind = "GenericBrace"
	}
	if len(e.Args) == 0 {
		return fmt.Sprintf(`%s="%s"`, kind, e.Value)
	}
	args := make([]string, len(e.Args))
	for i, a := range e.Args {
		args[i] = exprSyntax(a)
	}
	return fmt.Sprintf(`%s="%s"{%s}`, kind, e.Value, strings.Join(args, " "))
}
