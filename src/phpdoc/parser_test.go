package phpdoc

import (
	"reflect"
	"testing"
)

func TestParseSimple(t *testing.T) {
	expected := []CommentPart{
		{Name: "param", Params: []string{"$param", "int", "Here", "goes", "the", "description"}},
		{Name: "return", Params: []string{"int", "some", "result"}},
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
