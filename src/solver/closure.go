package solver

import (
	"github.com/VKCOM/noverify/src/meta"
)

var supportedFunctions = map[string]struct{}{
	`\array_map`:            {},
	`\array_walk`:           {},
	`\array_walk_recursive`: {},
	`\array_filter`:         {},
	`\array_reduce`:         {},
	`\usort`:                {},
	`\uasort`:               {},
}

func IsSupportedFunction(name string) bool {
	_, ok := supportedFunctions[name]
	return ok
}

// Struct containing information about the function that called the closure.
type ClosureCallerInfo struct {
	Name     string          // caller function name
	ArgTypes []meta.TypesMap // types for each arg for call caller function
}
