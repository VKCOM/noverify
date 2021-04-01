package linter

import (
	"fmt"
	"io"
	"sort"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/linter/lintapi"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/workspace"
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

// CheckerInfo provides a single checker (diagnostic) metadata.
//
// This structure may change with different revisions of noverify
// and get new fields that may be used by the linter.
type CheckerInfo struct {
	// Name is a diagnostic short name.
	// If several words are needed, prefer camelCase.
	Name string

	// Default controls whether diagnostic is
	// enabled by default or it should be included by allow-checks explicitly.
	Default bool

	// Quickfix tells whether this checker can automatically fix the reported
	// issues when linter works in -fix mode.
	Quickfix bool

	// Comment is a short summary of what this diagnostic does.
	// A single descriptive sentence is a perfect format for it.
	Comment string

	// Before is a non-compliant code example (before the fix).
	// Optional, but if present, After should also be non-empty.
	Before string

	// After is a compliant code example (after the fix).
	// Optional, but if present, Before should also be non-empty.
	After string

	// Extends tells the check is created by a dynamic rule that
	// extends the internal linter rule.
	Extends bool
}

// BlockChecker is a custom linter that is called on block level
type BlockChecker interface {
	BeforeEnterNode(ir.Node)
	AfterEnterNode(ir.Node)
	BeforeLeaveNode(ir.Node)
	AfterLeaveNode(ir.Node)
}

// RootChecker is a custom linter that should operator only at root level.
// Block level analysis (function and method bodies and all if/else/for/etc blocks) must be performed in BlockChecker.
type RootChecker interface {
	BeforeEnterFile()
	AfterLeaveFile()
	BeforeEnterNode(ir.Node)
	AfterEnterNode(ir.Node)
	BeforeLeaveNode(ir.Node)
	AfterLeaveNode(ir.Node)
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

func (BlockCheckerDefaults) BeforeEnterNode(ir.Node) {}
func (BlockCheckerDefaults) AfterEnterNode(ir.Node)  {}
func (BlockCheckerDefaults) BeforeLeaveNode(ir.Node) {}
func (BlockCheckerDefaults) AfterLeaveNode(ir.Node)  {}

// RootCheckerDefaults is a type for embedding into checkers to
// get default (empty) RootChecker implementations.
//
// You can "override" any required methods while ignoring the others.
//
// The benefit is higher backwards-compatibility.
// If new methods are added to RootChecker, you wouldn't need
// to change your code right away (especially if you don't need a new hook).
type RootCheckerDefaults struct{}

func (RootCheckerDefaults) BeforeEnterFile()        {}
func (RootCheckerDefaults) AfterLeaveFile()         {}
func (RootCheckerDefaults) BeforeEnterNode(ir.Node) {}
func (RootCheckerDefaults) AfterEnterNode(ir.Node)  {}
func (RootCheckerDefaults) BeforeLeaveNode(ir.Node) {}
func (RootCheckerDefaults) AfterLeaveNode(ir.Node)  {}

// RootContext is the context for root checker to run on.
type RootContext struct {
	w *rootWalker
}

// ParsePHPDoc returns parsed phpdoc comment parts.
func (ctx *RootContext) ParsePHPDoc(doc string) phpdoc.Comment {
	return phpdoc.Parse(ctx.w.ctx.phpdocTypeParser, doc)
}

// Report records linter warning of specified level.
// chechName is a key that identifies the "checker" (diagnostic name) that found
// issue being reported.
func (ctx *RootContext) Report(n ir.Node, level int, checkName, msg string, args ...interface{}) {
	ctx.w.Report(n, level, checkName, msg, args...)
}

func (ctx *RootContext) ReportByLine(lineNumber, level int, checkName, msg string, args ...interface{}) {
	ctx.w.ReportByLine(lineNumber, level, checkName, msg, args...)
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

// File returns analyzed file.
//
// Experimental API.
func (ctx *RootContext) File() *workspace.File {
	return ctx.w.file
}

// BlockContext is the context for block checker.
type BlockContext struct {
	w *blockWalker
}

// NodePath returns a node path up to the current traversal position.
// The path includes the node that is being traversed as well.
func (ctx *BlockContext) NodePath() irutil.NodePath {
	return ctx.w.path
}

// ExprType resolves the type of e expression node.
func (ctx *BlockContext) ExprType(e ir.Node) types.Map {
	return ctx.w.exprType(e)
}

// Report records linter warning of specified level.
// chechName is a key that identifies the "checker" (diagnostic name) that found
// issue being reported.
func (ctx *BlockContext) Report(n ir.Node, level int, checkName, msg string, args ...interface{}) {
	ctx.w.r.Report(n, level, checkName, msg, args...)
}

func (ctx *BlockContext) ReportByLine(lineNumber, level int, checkName, msg string, args ...interface{}) {
	ctx.w.r.ReportByLine(lineNumber, level, checkName, msg, args...)
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
func (ctx *BlockContext) IsStatement(n ir.Node) bool {
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

// File returns the file being analyzed.
func (ctx *BlockContext) File() *workspace.File {
	return ctx.w.r.file
}

// AddQuickfix adds a new quick fix.
func (ctx *BlockContext) AddQuickfix(fix quickfix.TextEdit) {
	ctx.w.r.ctx.fixes = append(ctx.w.r.ctx.fixes, fix)
}

// BlockCheckerCreateFunc is a factory function for BlockChecker
type BlockCheckerCreateFunc func(*BlockContext) BlockChecker

// RootCheckerCreateFunc is a factory function for RootChecker
type RootCheckerCreateFunc func(*RootContext) RootChecker

const (
	LevelError    = lintapi.LevelError
	LevelWarning  = lintapi.LevelWarning
	LevelNotice   = lintapi.LevelNotice
	LevelSecurity = lintapi.LevelSecurity // Like warning, but reported without a context line
)

type CheckersRegistry struct {
	blockCheckers []BlockCheckerCreateFunc
	rootCheckers  []RootCheckerCreateFunc
	cachers       []MetaCacher
	info          map[string]CheckerInfo
}

// AddBlockChecker registers a custom block linter that will be used on block level.
func (reg *CheckersRegistry) AddBlockChecker(c BlockCheckerCreateFunc) {
	reg.blockCheckers = append(reg.blockCheckers, c)
}

// AddRootChecker registers a custom root linter that will be used on root level.
//
// Root checker indexing phase is expected to be stateless.
// If indexing results need to be saved (and cached), use RegisterRootCheckerWithCacher.
func (reg *CheckersRegistry) AddRootChecker(c RootCheckerCreateFunc) {
	reg.AddRootCheckerWithCacher(nil, c)
}

// AddRootCheckerWithCacher registers a custom root linter that will be used on root level.
// Specified cacher is used to save (and load) indexing phase results.
func (reg *CheckersRegistry) AddRootCheckerWithCacher(cacher MetaCacher, c RootCheckerCreateFunc) {
	reg.rootCheckers = append(reg.rootCheckers, c)
	if cacher != nil {
		// Validate cacher version string.
		ver := cacher.Version()
		if len(ver) > 256 {
			panic(fmt.Sprintf("register cacher %q: can't handle strings longer that 256 bytes", ver))
		}
		for _, cacher := range reg.cachers {
			if cacher.Version() == ver {
				panic(fmt.Sprintf("register cacher %q: already registered", ver))
			}
		}
	}
	reg.cachers = append(reg.cachers, cacher)
}

func (reg *CheckersRegistry) DeclareRules(rset *rules.Set) {
	for _, ruleName := range rset.Names {
		doc := rset.DocByName[ruleName]
		comment := doc.Comment
		if comment == "" {
			comment = fmt.Sprintf("%s is a dynamic rule", ruleName)
		}
		reg.DeclareChecker(CheckerInfo{
			Name:     ruleName,
			Comment:  comment,
			Default:  true,
			Quickfix: doc.Fix,
			Before:   doc.Before,
			After:    doc.After,
			Extends:  doc.Extends,
		})
	}
}

// DeclareChecker declares a checker described by an info.
// It's a good practice to declare *all* provided checks.
//
// If checker is not declared, for example, there is no way to
// make it enabled by default.
func (reg *CheckersRegistry) DeclareChecker(info CheckerInfo) {
	if info.Name == "" {
		panic("can't declare a checker with an empty name")
	}
	if _, ok := reg.info[info.Name]; ok {
		if !info.Extends {
			panic(fmt.Sprintf("the checker %q is already declared, if you want to set the checker both in the dynamic rule and in the code, then use @extends annotation", info.Name))
		}
		return
	}
	if info.Before != "" && info.After == "" {
		panic(fmt.Sprintf("%s: Before is set, but After is empty", info.Name))
	}
	if info.After != "" && info.Before == "" {
		panic(fmt.Sprintf("%s: After is set, but Before is empty", info.Name))
	}
	reg.info[info.Name] = info
}

// ListDeclared returns a list of all checkers that were declared.
// Slice is sorted by checker names.
func (reg *CheckersRegistry) ListDeclared() []CheckerInfo {
	checks := make([]CheckerInfo, 0, len(reg.info))
	for _, c := range reg.info {
		checks = append(checks, c)
	}
	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Name < checks[j].Name
	})
	return checks
}
