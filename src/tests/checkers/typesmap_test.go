package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/meta"
)

func TestEqualsMatching(t *testing.T) {
	testCases := [][]meta.TypesMap{
		[]meta.TypesMap{
			meta.NewEmptyTypesMap(1),
			meta.NewEmptyTypesMap(1),
		},
		[]meta.TypesMap{
			meta.NewTypesMapFromMap(map[string]struct{}{"string": struct{}{}, "int": struct{}{}}),
			meta.NewTypesMapFromMap(map[string]struct{}{"string": struct{}{}, "int": struct{}{}}),
		},
	}

	for _, testCase := range testCases {
		if !testCase[0].Equals(testCase[1]) {
			t.Errorf("%v and %v must match", testCase[0], testCase[1])
		}
		if !testCase[1].Equals(testCase[0]) {
			t.Errorf("%v and %v must match", testCase[1], testCase[0])
		}
	}
}

func TestEqualNonMatching(t *testing.T) {
	testCases := [][]meta.TypesMap{
		[]meta.TypesMap{
			meta.NewEmptyTypesMap(1),
			meta.NewTypesMapFromMap(map[string]struct{}{"int": struct{}{}}),
		},
		[]meta.TypesMap{
			meta.NewTypesMapFromMap(map[string]struct{}{"string": struct{}{}}),
			meta.NewTypesMapFromMap(map[string]struct{}{"int": struct{}{}}),
		},
	}

	for _, testCase := range testCases {
		if testCase[0].Equals(testCase[1]) {
			t.Errorf("%v and %v must not match", testCase[0], testCase[1])
		}
		if testCase[1].Equals(testCase[0]) {
			t.Errorf("%v and %v must not match", testCase[1], testCase[0])
		}
	}
}
