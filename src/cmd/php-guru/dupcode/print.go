package dupcode

import (
	"fmt"
	"sort"
)

func printDuplicates(funcs funcSet) {
	funcScore := func(fn *funcInfo) int {
		baseScore := len(fn.code)
		return baseScore * len(fn.dups)
	}

	funcList := make([]*funcInfo, 0, len(funcs))
	for _, fn := range funcs {
		if len(fn.dups) == 0 {
			continue // No duplicates
		}
		funcList = append(funcList, fn)
	}
	sort.Slice(funcList, func(i, j int) bool {
		return funcScore(funcList[i]) > funcScore(funcList[j])
	})

	dupLines := 0
	totalFuncs := 0
	for _, fn := range funcList {
		fmt.Printf("%s: %s()\n", fn.declPos, fn.name)
		for _, dup := range fn.dups {
			fmt.Printf("%s: %s()\n", dup.declPos, dup.name)
		}
		fmt.Println(string(fn.code))
		fmt.Println("----")
		dupLines += fn.linesOfCode * len(fn.dups)
		totalFuncs += len(fn.dups) + 1
	}

	fmt.Printf("%d duplicated patterns reported\n", len(funcList))
	fmt.Printf("%d funcs included\n", totalFuncs)
	fmt.Printf("%d duplicated lines of code\n", dupLines)
}
