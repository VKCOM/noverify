package phpdoc

import (
	"reflect"
	"testing"
)

func TestParseSimple(t *testing.T) {
	want := []CommentPart{
		{
			Line:       4,
			Name:       "param",
			Params:     []string{"$param", "int", "Here", "goes", "the", "description"},
			ParamsText: "$param int  Here goes the description",
		},
		{
			Line:       5,
			Name:       "param",
			Params:     []string{"array<int,string>", "$arr", "Array", "of", "int", "to", "string"},
			ParamsText: "array<int, string> $arr  Array of int to string",
		},
		{
			Line:       6,
			Name:       "param",
			Params:     []string{"array<int,array<string,stdclass>>", "$arr_nested", "Array", "of", "nested", "arrays"},
			ParamsText: "array<int, array<string, stdclass> > $arr_nested  Array of nested arrays",
		},
		{
			Line:       7,
			Name:       "param",
			Params:     []string{"$arr_nested", "array<int,array<string,stdclass>>", "Array", "of", "nested", "arrays"},
			ParamsText: "$arr_nested array<int, array<string, stdclass> >  Array of nested arrays",
		},
		{
			Line:       8,
			Name:       "var",
			Params:     []string{"int"},
			ParamsText: "int",
		},
		{
			Line:       9,
			Name:       "var",
			Params:     []string{"array<int>"},
			ParamsText: "array<int>",
		},
		{
			Line:       10,
			Name:       "var",
			Params:     []string{"array<int,string>"},
			ParamsText: "array<int,string>",
		},
		{
			Line:       11,
			Name:       "var",
			Params:     []string{"array<int,string>"},
			ParamsText: "array< int, string >",
		},
		{
			Line:   12,
			Name:   "var",
			Params: []string{"array<int,array<string,stdclass>>"},
			ParamsText: "array<int, array<string, stdclass>	>",
		},
		{
			Line:       13,
			Name:       "return",
			Params:     []string{"int", "some", "result"},
			ParamsText: "int   some    result",
		},
	}

	got := Parse(`/**
	 * Some description
	 *
	 * @param   $param int  Here goes the description
	 * @param  array<int, string> $arr  Array of int to string
	 * @param  array<int, array<string, stdclass> > $arr_nested  Array of nested arrays
	 * @param  $arr_nested array<int, array<string, stdclass> >  Array of nested arrays
	 * @var int
	 * @var array<int>
	 * @var array<int,string>
	 * @var array< int, string >
	 * @var array<int, array<string, stdclass>	>
	 * @return int   some    result
	*/`)

	if len(got) != len(want) {
		t.Fatalf("len(got) != len(want): %d != %d", len(got), len(want))
	}

	for i, g := range got {
		w := want[i]

		if !reflect.DeepEqual(g, w) {
			t.Errorf("%d: got %#v, want %#v", i, g, w)
		}
	}
}
