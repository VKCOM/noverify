package solver

import (
	"reflect"

	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

// Like reflect.DeepEqual but knows how to compare node.Node by ignoring
// freefloating text and positions
func nodeAwareDeepEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}

	t1 := reflect.TypeOf(a)
	t2 := reflect.TypeOf(b)
	if t1 != t2 {
		return false
	}

	nodeType := reflect.TypeOf((*node.Node)(nil)).Elem()
	if t1.Implements(nodeType) {
		return nodeDeepEqual(a.(node.Node), b.(node.Node))
	} else if t1.Kind() == reflect.Slice && t1.Elem().Implements(nodeType) {
		return nodeSliceDeepEqual(a.([]node.Node), b.([]node.Node))
	} else {
		return reflect.DeepEqual(a, b)
	}
}

func nodeDeepEqual(a, b node.Node) bool {
	v1 := reflect.ValueOf(a)
	v2 := reflect.ValueOf(b)
	if v1.Type() != v2.Type() {
		panic("nodeDeepEqual should only be called on nodes of the same type")
	}

	// node.Node is normally implemented by a pointer to a struct
	if v1.Kind() == reflect.Ptr {
		v1 = v1.Elem()
		v2 = v2.Elem()
	}
	if v1.Kind() != reflect.Struct {
		panic("Expected node.Node to be implemented by a (pointer to) struct")
	}

	ffType := reflect.TypeOf((freefloating.Collection)(nil))
	posType := reflect.TypeOf((*position.Position)(nil))
	for i := 0; i < v1.NumField(); i++ {
		f1 := v1.Field(i)
		f2 := v2.Field(i)
		// Ignore freefloating text and positions
		if f1.Type() == ffType || f1.Type() == posType {
			continue
		}

		if !nodeAwareDeepEqual(f1.Interface(), f2.Interface()) {
			return false
		}
	}

	return true
}

func nodeSliceDeepEqual(a, b []node.Node) bool {
	l := len(a)
	if l != len(b) {
		return false
	}

	for i, n := range a {
		if !nodeAwareDeepEqual(n, b[i]) {
			return false
		}
	}
	return true
}
