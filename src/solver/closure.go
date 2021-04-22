package solver

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/types"
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

// ClosureCallerInfo containing information about the function that called the closure.
type ClosureCallerInfo struct {
	Name     string      // caller function name
	ArgTypes []types.Map // types for each arg for call caller function
}

func GetClosure(name ir.Node, sc *meta.Scope, cs *meta.ClassParseState) (meta.FuncInfo, bool) {
	nmf, ok := name.(*ir.SimpleVar)
	if !ok {
		return meta.FuncInfo{}, false
	}

	var fi meta.FuncInfo
	sv := &ir.SimpleVar{Name: nmf.Name}

	tp := ExprType(sc, cs, sv)
	found := tp.Find(func(typ string) bool {
		if !strings.HasPrefix(typ, `\Closure$`) {
			return false
		}

		funcInfo, ok := cs.Info.GetFunction(typ)
		if !ok {
			return false
		}

		fi = funcInfo
		return true
	})

	if found {
		return fi, true
	}

	return meta.FuncInfo{}, false
}
