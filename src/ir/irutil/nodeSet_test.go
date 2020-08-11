package irutil

import (
	"fmt"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
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

func createFQN(parts ...string) *ir.FullyQualifiedName {
	nm := &ir.FullyQualifiedName{
		Parts: make([]ir.Node, len(parts)),
	}
	for i, s := range parts {
		nm.Parts[i] = &ir.NamePart{Value: s}
	}
	return nm
}

func createSwitchCases(ncases int) []ir.Node {
	cases := make([]ir.Node, ncases)
	for i := range cases {
		cases[i] = &ir.CaseStmt{
			Cond: &ir.ClassConstFetchExpr{
				Class:        createFQN("Namespace", "Foo"),
				ConstantName: &ir.Identifier{Value: fmt.Sprintf("Bar%d", i)},
			},
		}
	}
	return cases
}
