package linter

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/vscode"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/walker"
)

// BlockChecker is a custom linter that is called on block level
type BlockChecker interface {
	BeforeEnterNode(walker.Walkable)
	AfterEnterNode(walker.Walkable)
	BeforeLeaveNode(walker.Walkable)
	AfterLeaveNode(walker.Walkable)
}

// RootChecker is a custom linter that should operator only at root level.
// Block level analysis (function and method bodies and all if/else/for/etc blocks) must be performed in BlockChecker.
type RootChecker interface {
	BeforeEnterNode(walker.Walkable)
	AfterEnterNode(walker.Walkable)
	BeforeLeaveNode(walker.Walkable)
	AfterLeaveNode(walker.Walkable)
}

// RootContext is the context for root checker to run on.
type RootContext interface {
	Report(n node.Node, level int, checkName, msg string, args ...interface{})
	Scope() *meta.Scope                     // get variables declared at root level
	ClassParseState() *meta.ClassParseState // get class parse state (namespace, class, etc)
	State() map[string]interface{}          // state that can be modified and passed into block context
}

// BlockContext is the context for block checker.
type BlockContext interface {
	Report(n node.Node, level int, checkName, msg string, args ...interface{})
	Scope() *meta.Scope                     // get variables declared in this block
	ClassParseState() *meta.ClassParseState // get class parse state (namespace, class, etc)
	RootState() map[string]interface{}      // state from root context
	IsRootLevel() bool                      // are we analysing root-level code currently
	IsStatement(n node.Node) bool           // whether or not specified node is a statement
	PrematureExitFlags() int
}

// BlockCheckerCreateFunc is a factory function for BlockChecker
type BlockCheckerCreateFunc func(BlockContext) BlockChecker

// RootCheckerCreateFunc is a factory function for RootChecker
type RootCheckerCreateFunc func(RootContext) RootChecker

const (
	LevelError       = 1
	LevelWarning     = 2
	LevelInformation = 3
	LevelHint        = 4
	LevelUnused      = 5
	LevelDoNotReject = 6 // do not treat this warning as a reason to reject if we get this kind of warning
	LevelSyntax      = 7
)

var vscodeLevelMap = map[int]int{
	LevelError:       vscode.Error,
	LevelWarning:     vscode.Warning,
	LevelInformation: vscode.Information,
	LevelHint:        vscode.Hint,
	LevelUnused:      vscode.Information,
	LevelDoNotReject: vscode.Warning,
	// LevelSyntax is intentionally not included here
}

var severityNames = map[int]string{
	LevelError:       "ERROR  ",
	LevelWarning:     "WARNING",
	LevelInformation: "INFO   ",
	LevelHint:        "HINT   ",
	LevelUnused:      "UNUSED ",
	LevelDoNotReject: "MAYBE  ",
	LevelSyntax:      "SYNTAX ",
}

var (
	customBlockLinters []BlockCheckerCreateFunc
	customRootLinters  []RootCheckerCreateFunc
)

// RegisterBlockChecker registers a custom block linter that will be used on block level.
func RegisterBlockChecker(c BlockCheckerCreateFunc) {
	customBlockLinters = append(customBlockLinters, c)
}

// RegisterRootChecker registers a custom root linter that will be used on root level.
func RegisterRootChecker(c RootCheckerCreateFunc) {
	customRootLinters = append(customRootLinters, c)
}
