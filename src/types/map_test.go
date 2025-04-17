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

func TestMapUnion(t *testing.T) {
	tests := []struct {
		a, b     Map
		expected Map
	}{
		{
			NewMapFromMap(map[string]struct{}{"int": {}, "string": {}}),
			NewMapFromMap(map[string]struct{}{"float": {}}),
			NewMapFromMap(map[string]struct{}{"int": {}, "string": {}, "float": {}}),
		},
		{
			NewMapFromMap(map[string]struct{}{"int": {}}),
			NewMapFromMap(map[string]struct{}{"int": {}}),
			NewMapFromMap(map[string]struct{}{"int": {}}),
		},
		{
			NewEmptyMap(0),
			NewMapFromMap(map[string]struct{}{"bool": {}}),
			NewMapFromMap(map[string]struct{}{"bool": {}}),
		},
		{
			NewEmptyMap(0),
			NewEmptyMap(0),
			NewEmptyMap(0),
		},
	}

	for _, tt := range tests {
		result := tt.a.Union(tt.b)
		if !result.Equals(tt.expected) {
			t.Errorf("Union failed: %v ∪ %v = %v, expected %v", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMapIntersect(t *testing.T) {
	tests := []struct {
		a, b     Map
		expected Map
	}{
		{
			NewMapFromMap(map[string]struct{}{"int": {}, "string": {}}),
			NewMapFromMap(map[string]struct{}{"int": {}, "float": {}}),
			NewMapFromMap(map[string]struct{}{"int": {}}),
		},
		{
			NewMapFromMap(map[string]struct{}{"bool": {}}),
			NewMapFromMap(map[string]struct{}{"int": {}}),
			NewEmptyMap(0),
		},
		{
			NewEmptyMap(0),
			NewMapFromMap(map[string]struct{}{"string": {}}),
			NewEmptyMap(0),
		},
		{
			NewMapFromMap(map[string]struct{}{"a": {}, "b": {}}),
			NewMapFromMap(map[string]struct{}{"a": {}, "b": {}}),
			NewMapFromMap(map[string]struct{}{"a": {}, "b": {}}),
		},
	}

	for _, tt := range tests {
		result := tt.a.Intersect(tt.b)
		if !result.Equals(tt.expected) {
			t.Errorf("Intersect failed: %v ∩ %v = %v, expected %v", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMapIsClass(t *testing.T) {
	testCases := []struct {
		m        Map
		expected bool
	}{
		// Class
		{NewMapFromMap(map[string]struct{}{`\MyClass`: {}}), true},

		// Array
		{NewMapFromMap(map[string]struct{}{`\MyClass[]`: {}}), false},

		// Shape
		{NewMapFromMap(map[string]struct{}{`\shape$foo`: {}}), false},

		// Closure
		{NewMapFromMap(map[string]struct{}{`\Closure$123`: {}}), false},

		// Scalar
		{NewMapFromMap(map[string]struct{}{"int": {}}), false},

		// Multi types
		{NewMapFromMap(map[string]struct{}{`\MyClass`: {}, `int`: {}}), false},
	}

	for _, tc := range testCases {
		if got := tc.m.IsClass(); got != tc.expected {
			t.Errorf("IsClass() = %v, want %v for %v", got, tc.expected, tc.m)
		}
	}
}

func TestMapIsBoolean(t *testing.T) {
	testCases := []struct {
		m        Map
		expected bool
	}{
		{NewMapFromMap(map[string]struct{}{"bool": {}}), true},

		{NewMapFromMap(map[string]struct{}{"0400bool": {}}), true},

		{NewMapFromMap(map[string]struct{}{"true": {}}), true},
		{NewMapFromMap(map[string]struct{}{"false": {}}), true},

		{NewMapFromMap(map[string]struct{}{"int": {}}), false},
		{NewMapFromMap(map[string]struct{}{`\MyClass`: {}}), false},

		{NewMapFromMap(map[string]struct{}{"bool": {}, "int": {}}), false},
	}

	for _, tc := range testCases {
		if got := tc.m.IsBoolean(); got != tc.expected {
			t.Errorf("IsBoolean() = %v, want %v for %v", got, tc.expected, tc.m)
		}
	}
}
