package phpdoc

import (
	"reflect"
	"testing"
)

func TestParseSimple(t *testing.T) {
	expected := []CommentPart{
		{
			Line:       4,
			Name:       "param",
			Params:     []string{"$param", "int", "Here", "goes", "the", "description"},
			ParamsText: "$param int  Here goes the description",
		},
		{
			Line:       5,
			Name:       "return",
			Params:     []string{"int", "some", "result"},
			ParamsText: "int   some    result",
		},
	}

	actual := Parse(`/**
	 * Some description
	 *
	 * @param   $param int  Here goes the description
	 * @return int   some    result
	*/`)

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Actual parsed structure is different from what we expected: %+v", actual)
	}
}
