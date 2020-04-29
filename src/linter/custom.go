package linter

import (
	"fmt"
	"io"
	"sort"

	"github.com/VKCOM/noverify/src/linter/lintapi"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/vscode"
)

// MetaCacher is an interface for integrating checker-specific
// indexing results into NoVerify cache.
//
// Usually, every vendor contains a global meta object that
// can implement MetaCacher and be associated with a relevant root checker.
type MetaCacher interface {
	// Version returns a unique cache version identifier.
	// When underlying cache structure is updated, version
	// should return different value.
	//
	// Preferably something unique, prefixed with a vendor
	// name, like `mylints-1.0.0` or `extension-abc4`.
	//
	// Returned value is written before Encode() is called to
	// the same writer. It's also read from the reader before
	// Decode() is invoked.
	Version() string

	// Encode stores custom meta cache part data into provided writer.
	// RootChecker is expected to carry the necessary indexing phase results.
	Encode(io.Writer, RootChecker) error

	// Decode loads custom meta cache part data from provided reader.
	// Those results are used insted of running the associated indexer.
	Decode(r io.Reader, filename string) error
}

// CheckInfo provides a single check (diagnostic) metadata.
//
// This structure may change with different revisions of noverify
// and get new fields that may be used by the linter.
type CheckInfo struct {
	// Name is a diagnostic short name.
	// If several words are needed, prefer camelCase.
	Name string

	// Default controls whether diagnostic is
	// enabled by default or it should be included by allow-checks explicitly.
	Default bool

	// Comment is a short summary of what this diagnostic does.
	// A single descriptive sentence is a perfect format for it.
	Comment string
}

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
	return ctx.w.scope()
}

// ClassParseState returns class parse state (namespace, class, etc).
func (ctx *RootContext) ClassParseState() *meta.ClassParseState {
	return ctx.w.ctx.st
}

// State returns state that can be modified and passed into block context
func (ctx *RootContext) State() map[string]interface{} {
	return ctx.w.state()
}

// Filename returns the file name of the file being analyzed.
func (ctx *RootContext) Filename() string {
	return ctx.w.ctx.st.CurrentFile
}

// FileContents returns analyzed file source code.
// Caller should not modify the returned slice.
//
// Experimental API.
func (ctx *RootContext) FileContents() []byte {
	return ctx.w.fileContents
}

// BlockContext is the context for block checker.
type BlockContext struct {
	w *BlockWalker
}

// Report records linter warning of specified level.
// chechName is a key that identifies the "checker" (diagnostic name) that found
// issue being reported.
func (ctx *BlockContext) Report(n node.Node, level int, checkName, msg string, args ...interface{}) {
	ctx.w.r.Report(n, level, checkName, msg, args...)
}

// Scope returns variables declared in this block.
func (ctx *BlockContext) Scope() *meta.Scope {
	return ctx.w.ctx.sc
}

// ClassParseState returns class parse state (namespace, class, etc).
func (ctx *BlockContext) ClassParseState() *meta.ClassParseState {
	return ctx.w.r.ctx.st
}

// RootState returns state from root context.
func (ctx *BlockContext) RootState() map[string]interface{} {
	return ctx.w.r.state()
}

// IsRootLevel reports whether we are analysing root-level code currently.
func (ctx *BlockContext) IsRootLevel() bool {
	return ctx.w.rootLevel
}

// IsStatement reports whether or not specified node is a statement.
func (ctx *BlockContext) IsStatement(n node.Node) bool {
	_, ok := ctx.w.statements[n]
	return ok
}

func (ctx *BlockContext) PrematureExitFlags() int {
	return ctx.w.ctx.exitFlags
}

// Filename returns the file name of the file being analyzed.
func (ctx *BlockContext) Filename() string {
	return ctx.w.r.ctx.st.CurrentFile
}

// BlockCheckerCreateFunc is a factory function for BlockChecker
type BlockCheckerCreateFunc func(*BlockContext) BlockChecker

// RootCheckerCreateFunc is a factory function for RootChecker
type RootCheckerCreateFunc func(*RootContext) RootChecker

const (
	LevelError       = lintapi.LevelError
	LevelWarning     = lintapi.LevelWarning
	LevelInformation = lintapi.LevelInformation
	LevelHint        = lintapi.LevelHint
	LevelUnused      = lintapi.LevelUnused
	LevelDoNotReject = lintapi.LevelMaybe
	LevelSyntax      = lintapi.LevelSyntax
	LevelSecurity    = lintapi.LevelSecurity // Like warning, but reported without a context line
)

var vscodeLevelMap = map[int]int{
	LevelError:       vscode.Error,
	LevelWarning:     vscode.Warning,
	LevelInformation: vscode.Information,
	LevelHint:        vscode.Hint,
	LevelUnused:      vscode.Information,
	LevelDoNotReject: vscode.Warning,
	LevelSecurity:    vscode.Warning,
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
	LevelSecurity:    "WARNING",
}

var (
	customBlockLinters []BlockCheckerCreateFunc
	customRootLinters  []RootCheckerCreateFunc
	metaCachers        []MetaCacher
	checksInfoRegistry = map[string]CheckInfo{}
)

// RegisterBlockChecker registers a custom block linter that will be used on block level.
func RegisterBlockChecker(c BlockCheckerCreateFunc) {
	customBlockLinters = append(customBlockLinters, c)
}

// RegisterRootChecker registers a custom root linter that will be used on root level.
//
// Root checker indexing phase is expected to be stateless.
// If indexing results need to be saved (and cached), use RegisterRootCheckerWithCacher.
func RegisterRootChecker(c RootCheckerCreateFunc) {
	RegisterRootCheckerWithCacher(nil, c)
}

// RegisterRootCheckerWithCacher registers a custom root linter that will be used on root level.
// Specified cacher is used to save (and load) indexing phase results.
func RegisterRootCheckerWithCacher(cacher MetaCacher, c RootCheckerCreateFunc) {
	customRootLinters = append(customRootLinters, c)
	if cacher != nil {
		// Validate cacher version string.
		ver := cacher.Version()
		if len(ver) > 256 {
			panic(fmt.Sprintf("register cacher %q: can't handle strings longer that 256 bytes", ver))
		}
		for _, cacher := range metaCachers {
			if cacher.Version() == ver {
				panic(fmt.Sprintf("register cacher %q: already registered", ver))
			}
		}
	}
	metaCachers = append(metaCachers, cacher)
}

// DeclareCheck declares a check described by an info.
// It's a good practice to declare *all* provided checks.
//
// If check is not declared, for example, there is no way to
// make it enabled by default.
func DeclareCheck(info CheckInfo) {
	if info.Name == "" {
		panic("can't declare a check with an empty name")
	}
	if _, ok := checksInfoRegistry[info.Name]; ok {
		panic(fmt.Sprintf("check %q already declared", info.Name))
	}
	checksInfoRegistry[info.Name] = info
}

// GetDeclaredChecks returns a list of all checks that were declared.
// Slice is sorted by check names.
func GetDeclaredChecks() []CheckInfo {
	checks := make([]CheckInfo, 0, len(checksInfoRegistry))
	for _, c := range checksInfoRegistry {
		checks = append(checks, c)
	}
	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Name < checks[j].Name
	})
	return checks
}
