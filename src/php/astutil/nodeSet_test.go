package astutil

import (
	"fmt"
	"testing"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
)

func BenchmarkNodeSet(b *testing.B) {
	runBench := func(n int) {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			cases := createSwitchCases(n)
			b.ResetTimer()
			var nset NodeSet
			for i := 0; i < b.N; i++ {
				nset.Reset()
				for _, c := range cases {
					if !nset.Add(c) {
						b.Error("unexpected duplicate reported")
					}
				}
			}
		})
	}

	runBench(1)
	runBench(5)
	runBench(6)
	runBench(100)
	runBench(400)
	runBench(1000)
}

func createFQN(parts ...string) *name.FullyQualified {
	nm := &name.FullyQualified{
		Parts: make([]node.Node, len(parts)),
	}
	for i, s := range parts {
		nm.Parts[i] = &name.NamePart{Value: s}
	}
	return nm
}

func createSwitchCases(ncases int) []node.Node {
	cases := make([]node.Node, ncases)
	for i := range cases {
		cases[i] = &stmt.Case{
			Cond: &expr.ClassConstFetch{
				Class:        createFQN("Namespace", "Foo"),
				ConstantName: &node.Identifier{Value: fmt.Sprintf("Bar%d", i)},
			},
		}
	}
	return cases
}
