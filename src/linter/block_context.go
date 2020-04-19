package linter

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/astutil"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/solver"
)

type customMethod struct {
	obj  node.Node
	name string
}

// blockContext is a state that is used to hold inner blocks info.
//
// When BlockWalker enters another block, new context is created.
// When it leaves that block, previous context is restored.
// The BlockWalker itself is not copied and, instead, re-used as is.
type blockContext struct {
	exitFlags         int // if block always breaks code flow then there will be exitFlags
	containsExitFlags int // if block sometimes breaks code flow then there will be containsExitFlags

	deadCodeReported bool

	// Fields below should be copied during context cloning.

	sc            *meta.Scope
	innermostLoop loopKind
	// insideLoop is true if any number of statements above there is enclosing loop.
	// innermostLoop is not enough for this, since we can be inside switch while
	// having for loop outside of that switch.
	insideLoop  bool
	customTypes []solver.CustomType

	customMethods   []customMethod
	customFunctions []string
}

func (ctx *blockContext) addCustomFunction(functionName string) {
	ctx.customFunctions = append(ctx.customFunctions, functionName)
}

func (ctx *blockContext) customFunctionExists(nm node.Node) bool {
	for _, functionName := range ctx.customFunctions {
		if meta.NameNodeEquals(nm, functionName) {
			return true
		}
	}
	return false
}

func (ctx *blockContext) addCustomMethod(obj node.Node, methodName string) {
	ctx.customMethods = append(ctx.customMethods, customMethod{
		obj:  obj,
		name: methodName,
	})
}

func (ctx *blockContext) customMethodExists(obj node.Node, methodName string) bool {
	for _, m := range ctx.customMethods {
		if m.name == methodName && astutil.NodeEqual(m.obj, obj) {
			return true
		}
	}
	return false
}

// copyBlockContext returns a copy of the context.
//
// The copy does not inherit some properties, like deadCodeReported.
func copyBlockContext(ctx *blockContext) *blockContext {
	return &blockContext{
		sc:              ctx.sc.Clone(),
		customTypes:     append([]solver.CustomType{}, ctx.customTypes...),
		customMethods:   append([]customMethod{}, ctx.customMethods...),
		customFunctions: append([]string{}, ctx.customFunctions...),
		innermostLoop:   ctx.innermostLoop,
		insideLoop:      ctx.insideLoop,
	}
}
