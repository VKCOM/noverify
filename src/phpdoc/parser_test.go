package phpdoc

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseSimple(t *testing.T) {
	p := NewTypeParser()
	parseType := func(s string) Type {
		return p.Parse(s).Clone()
	}
	want := []CommentPart{
		&TypeVarCommentPart{
			line:       4,
			name:       "param",
			VarIsFirst: true,
			Type:       parseType(`int  Here goes the description`),
			Var:        "$param",
			Rest:       "Here goes the description",
		},
		&TypeVarCommentPart{
			line: 5,
			name: "param",
			Var:  "$arr",
			Type: parseType(`array<int, string> $arr  Array of int to string`),
			Rest: "Array of int to string",
		},
		&TypeVarCommentPart{
			line: 6,
			name: "param",
			Var:  "$arr_nested",
			Type: parseType(`array<int, array<string, stdclass> > $arr_nested  Array of nested arrays`),
			Rest: `Array of nested arrays`,
		},
		&TypeVarCommentPart{
			line:       7,
			name:       "param",
			VarIsFirst: true,
			Type:       parseType(`array<int, array<string, stdclass> >  Array of nested arrays`),
			Var:        "$arr_nested",
			Rest:       "Array of nested arrays",
		},
		&TypeVarCommentPart{
			line: 8,
			name: "var",
			Var:  "",
			Type: parseType(`int`),
		},
		&TypeVarCommentPart{
			line: 9,
			name: "var",
			Var:  "$foo1",
			Type: parseType(`array<int> $foo1  var comment`),
			Rest: "var comment",
		},
		&TypeVarCommentPart{
			line:       10,
			name:       "var",
			VarIsFirst: true,
			Var:        "$foo2",
			Type:       parseType(`array<int,string>`),
		},
		&TypeVarCommentPart{
			line: 11,
			name: "var",
			Type: parseType(`array< int, string >`),
		},
		&TypeVarCommentPart{
			line: 12,
			name: "var",
			Type: parseType("array<int, array<string, stdclass>	>"),
		},
		&TypeCommentPart{
			line: 13,
			name: "return",
			Type: parseType(`int   some    result`),
			Rest: `some    result`,
		},
		&RawCommentPart{
			line:       14,
			name:       "unknown",
			Params:     []string{"a", "b", "c"},
			ParamsText: `a b c`,
		},
		&TypeVarCommentPart{
			line:       15,
			name:       "property-read",
			VarIsFirst: false,
			Type:       parseType(`int $readonly`),
			Var:        "$readonly",
		},
	}

	got := Parse(p, `/**
	 * Some description
	 *
	 * @param   $param int  Here goes the description
	 * @param  array<int, string> $arr  Array of int to string
	 * @param  array<int, array<string, stdclass> > $arr_nested  Array of nested arrays
	 * @param  $arr_nested array<int, array<string, stdclass> >  Array of nested arrays
	 * @var int
	 * @var  array<int> $foo1  var comment
	 * @var $foo2  array<int,string>
	 * @var array< int, string >
	 * @var array<int, array<string, stdclass>	>
	 * @return int   some    result
	 * @unknown a b c
	 * @property-read int $readonly
	*/`)

	if len(got.Parsed) != len(want) {
		t.Fatalf("len(got) != len(want): %d != %d", len(got.Parsed), len(want))
	}

	for i, g := range got.Parsed {
		w := want[i]

		if diff := cmp.Diff(g, w, cmp.Exporter(func(reflect.Type) bool { return true })); diff != "" {
			t.Errorf("%d: (-have +want):\n%s", i, diff)
		}
	}
}
