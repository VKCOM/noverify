package types

import (
	"testing"
)

func TestEqualsMatching(t *testing.T) {
	testCases := [][]Map{
		{
			NewEmptyMap(1),
			NewEmptyMap(1),
		},
		{
			NewMapFromMap(map[string]struct{}{"string": {}, "int": {}}),
			NewMapFromMap(map[string]struct{}{"string": {}, "int": {}}),
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
	testCases := [][]Map{
		{
			NewEmptyMap(1),
			NewMapFromMap(map[string]struct{}{"int": {}}),
		},
		{
			NewMapFromMap(map[string]struct{}{"string": {}}),
			NewMapFromMap(map[string]struct{}{"int": {}}),
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
