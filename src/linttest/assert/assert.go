package assert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// DeepEqual uses google/go-cmp (https://godoc.org/github.com/google/go-cmp/cmp)
// to assert two values are equal and fails the test if they are not equal.
func DeepEqual(t *testing.T, x, y interface{}, opts ...cmp.Option) {
	if cmp.Equal(x, y, opts...) {
		return
	}
	t.Fatal(cmp.Diff(x, y, opts...))
}
