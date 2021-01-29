package solver

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
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

// ClosureCallerInfo containing information about the function that called the closure.
type ClosureCallerInfo struct {
	Name     string          // caller function name
	ArgTypes []meta.TypesMap // types for each arg for call caller function
}

func GetClosureName(fun *ir.ClosureExpr, curFunction, curFile string) string {
	pos := ir.GetPosition(fun)
	if curFunction != "" {
		curFunction = "," + curFunction
	}
	name := `\Closure(` + curFile + curFunction + "):" + fmt.Sprint(pos.StartLine)
	return name
}

func GetClosure(name ir.Node, sc *meta.Scope, cs *meta.ClassParseState) (meta.FuncInfo, bool) {
	if !cs.Info.IsIndexingComplete() {
		return meta.FuncInfo{}, false
	}

	nmf, ok := name.(*ir.SimpleVar)
	if !ok {
		return meta.FuncInfo{}, false
	}

	var fi meta.FuncInfo
	var found bool
	var sv = &ir.SimpleVar{Name: nmf.Name}

	tp := ExprTypeLocal(sc, cs, sv)
	tp.Iterate(func(typ string) {
		if strings.HasPrefix(typ, `\Closure`) {
			funcInfo, ok := cs.Info.GetFunction(typ)
			if !ok {
				return
			}
			fi = funcInfo
			found = true
		}
	})

	if found {
		return fi, true
	}

	return meta.FuncInfo{}, false
}
