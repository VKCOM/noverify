package astutil

import (
	"github.com/VKCOM/noverify/src/php/parser/node"
)

// nodeSetListMax specifies how much entries we store inside a
// slice of NodeSet.
//
// Using high value is prone to O(N^2) complexity penalty.
// Lower values mean that we need to allocate a map more often.
//
// Using NodeSet for branch if/else conditions and switch cases
// shows that we can use n=4 or n=5 as a sweet spot.
//
// PHP corpus results (if conditions + switch cases):
//	256131 only slice is used (slice=1 map=0)
//	56628  only slice is used (slice=2 map=0)
//	5844   only slice is used (slice=3 map=0)
//	1637   only slice is used (slice=4 map=0)
//	788    only slice is used (slice=5 map=0)
//	395    map is used (slice=5 map=1)
//	310    map is used (slice=5 map=2)
//	194    map is used (slice=5 map=3)
//	146    map is used (slice=5 map=5)
//	132    map is used (slice=5 map=4)
//	89     map is used (slice=5 map=6)
//
// Note that some cases include 300+ elements which would require
// us to do 50k+ comparisons during the duplicated switch case analysis.
// This is why map fallback is needed just to handle pathological cases
// without unexpected performance degradations.
//
// When not using a list with linear search, we use printed AST
// as a unique key for a map. This is the same thing staticcheck linter does.
//
// A comparison of slice-only (old) and hybrid solutions (new):
//	name            old time/op  new time/op  delta
//	NodeSet/1-8     9.42ns ± 0%  9.47ns ± 1%     ~     (p=0.222 n=4+5)
//	NodeSet/5-8      923ns ± 3%   908ns ± 1%     ~     (p=0.159 n=5+4)
//	NodeSet/6-8     1.38µs ± 2%  2.38µs ± 1%  +72.05%  (p=0.008 n=5+5)
//	NodeSet/100-8    413µs ± 4%   132µs ± 2%  -68.06%  (p=0.008 n=5+5)
//	NodeSet/400-8   6.35ms ± 1%  0.55ms ± 2%  -91.41%  (p=0.008 n=5+5)
//	NodeSet/1000-8  40.4ms ± 0%   1.4ms ± 3%  -96.46%  (p=0.008 n=5+5)
//
// Map-only implementation works 5 times slower on the most commons cases (n=1 and n=2).
const nodeSetListMax = 5

// NodeSet is a set of unique AST nodes.
//
// It's possible to avoid allocations in most cases if node set is
// reused with Reset(). If this is not possible, use NewNodeSet()
// to create a set that can do fewer allocations than a zero value set.
//
// NodeSet is not thread-safe, but since Root/Block walkers always operate
// in isolation, we can share on per RootWalker level.
type NodeSet struct {
	list []node.Node
	m    map[string]struct{}
}

// NewNodeSet returns a node set with preallocated storage.
//
// Intended to be used in places where reusing root node set is
// tricky or impossible (function is reentrant).
func NewNodeSet() NodeSet {
	return NodeSet{list: make([]node.Node, 0, nodeSetListMax)}
}

// Reset clears the set state and prepares it for another round.
func (set *NodeSet) Reset() {
	set.list = set.list[:0]
	set.m = nil
}

// Len returns the number of unique node elements present in the set.
func (set *NodeSet) Len() int {
	return len(set.list) + len(set.m)
}

// Add attempts to insert x into the node set.
//
// Returns true if element was inserted.
// If x is already in the set, false is returned
// and no insertion is performed.
func (set *NodeSet) Add(x node.Node) bool {
	for _, y := range set.list {
		if NodeEqual(x, y) {
			return false
		}
	}
	if len(set.list) < nodeSetListMax {
		set.list = append(set.list, x)
		return true
	}

	key := FmtNode(x)
	if set.m == nil {
		set.m = map[string]struct{}{}
	}
	if _, ok := set.m[key]; ok {
		return false
	}
	set.m[key] = struct{}{}
	return true
}
