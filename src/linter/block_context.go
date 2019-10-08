package linter

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

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
}

// copyBlockContext returns a copy of the context.
//
// The copy does not inherit some properties, like deadCodeReported.
func copyBlockContext(ctx *blockContext) *blockContext {
	return &blockContext{
		sc:            ctx.sc.Clone(),
		customTypes:   append([]solver.CustomType{}, ctx.customTypes...),
		innermostLoop: ctx.innermostLoop,
		insideLoop:    ctx.insideLoop,
	}
}
