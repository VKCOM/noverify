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
	AfterLeaveFile()
	BeforeEnterNode(walker.Walkable)
	AfterEnterNode(walker.Walkable)
	BeforeLeaveNode(walker.Walkable)
	AfterLeaveNode(walker.Walkable)
}

// BlockCheckerDefaults is a type for embedding into checkers to
// get default (empty) BlockChecker implementations.
//
// You can "override" any required methods while ignoring the others.
//
// The benefit is higher backwards-compatibility.
// If new methods are added to BlockChecker, you wouldn't need
// to change your code right away (especially if you don't need a new hook).
type BlockCheckerDefaults struct{}

func (BlockCheckerDefaults) BeforeEnterNode(walker.Walkable) {}
func (BlockCheckerDefaults) AfterEnterNode(walker.Walkable)  {}
func (BlockCheckerDefaults) BeforeLeaveNode(walker.Walkable) {}
func (BlockCheckerDefaults) AfterLeaveNode(walker.Walkable)  {}

// RootCheckerDefaults is a type for embedding into checkers to
// get default (empty) RootChecker implementations.
//
// You can "override" any required methods while ignoring the others.
//
// The benefit is higher backwards-compatibility.
// If new methods are added to RootChecker, you wouldn't need
// to change your code right away (especially if you don't need a new hook).
type RootCheckerDefaults struct{}

func (RootCheckerDefaults) AfterLeaveFile()                 {}
func (RootCheckerDefaults) BeforeEnterNode(walker.Walkable) {}
func (RootCheckerDefaults) AfterEnterNode(walker.Walkable)  {}
func (RootCheckerDefaults) BeforeLeaveNode(walker.Walkable) {}
func (RootCheckerDefaults) AfterLeaveNode(walker.Walkable)  {}

// RootContext is the context for root checker to run on.
type RootContext struct {
	w *RootWalker
}

// Report records linter warning of specified level.
// chechName is a key that identifies the "checker" (diagnostic name) that found
// issue being reported.
func (ctx *RootContext) Report(n node.Node, level int, checkName, msg string, args ...interface{}) {
	ctx.w.Report(n, level, checkName, msg, args...)
}

// Scope returns variables declared at root level.
func (ctx *RootContext) Scope() *meta.Scope {
	return ctx.w.Scope()
}

// ClassParseState returns class parse state (namespace, class, etc).
func (ctx *RootContext) ClassParseState() *meta.ClassParseState {
	return ctx.w.ClassParseState()
}

// State returns state that can be modified and passed into block context
func (ctx *RootContext) State() map[string]interface{} {
	return ctx.w.State()
}

// Filename returns the full file name of the file being analyzed.
func (ctx *RootContext) Filename() string {
	return ctx.w.filename
}

// BlockContext is the context for block checker.
type BlockContext struct {
	w *BlockWalker
}

// Report records linter warning of specified level.
// chechName is a key that identifies the "checker" (diagnostic name) that found
// issue being reported.
func (ctx *BlockContext) Report(n node.Node, level int, checkName, msg string, args ...interface{}) {
	ctx.w.Report(n, level, checkName, msg, args...)
}

// Scope returns variables declared in this block.
func (ctx *BlockContext) Scope() *meta.Scope {
	return ctx.w.Scope()
}

// ClassParseState returns class parse state (namespace, class, etc).
func (ctx *BlockContext) ClassParseState() *meta.ClassParseState {
	return ctx.w.ClassParseState()
}

// RootState returns state from root context.
func (ctx *BlockContext) RootState() map[string]interface{} {
	return ctx.w.RootState()
}

// IsRootLevel reports whether we are analysing root-level code currently.
func (ctx *BlockContext) IsRootLevel() bool {
	return ctx.w.IsRootLevel()
}

// IsStatement reports whether or not specified node is a statement.
func (ctx *BlockContext) IsStatement(n node.Node) bool {
	return ctx.w.IsStatement(n)
}

func (ctx *BlockContext) PrematureExitFlags() int {
	return ctx.w.PrematureExitFlags()
}

// BlockCheckerCreateFunc is a factory function for BlockChecker
type BlockCheckerCreateFunc func(*BlockContext) BlockChecker

// RootCheckerCreateFunc is a factory function for RootChecker
type RootCheckerCreateFunc func(*RootContext) RootChecker

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
