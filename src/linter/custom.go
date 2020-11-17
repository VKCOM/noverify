package linter

import (
	"fmt"
	"io"
	"sort"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linter/lintapi"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/rules"
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
	w *RootWalker
}

// ParsePHPDoc returns parsed phpdoc comment parts.
func (ctx *RootContext) ParsePHPDoc(doc string) []phpdoc.CommentPart {
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

// NodePath returns a node path up to the current traversal position.
// The path includes the node that is being traversed as well.
func (ctx *BlockContext) NodePath() NodePath {
	return ctx.w.path
}

// ExprType resolves the type of e expression node.
func (ctx *BlockContext) ExprType(e ir.Node) meta.TypesMap {
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

// FileContent returns the content of the file being analyzed.
func (ctx *BlockContext) FileContent() []byte {
	return ctx.w.r.fileContents
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

func DeclareRules(rset *rules.Set) {
	for _, ruleName := range rset.Names {
		doc := rset.DocByName[ruleName]
		comment := doc.Comment
		if comment == "" {
			comment = fmt.Sprintf("%s is a dynamic rule", ruleName)
		}
		DeclareCheck(CheckInfo{
			Name:     ruleName,
			Comment:  comment,
			Default:  true,
			Quickfix: doc.Fix,
			Before:   doc.Before,
			After:    doc.After,
		})
	}
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
	if info.Before != "" && info.After == "" {
		panic(fmt.Sprintf("%s: Before is set, but After is empty", info.Name))
	}
	if info.After != "" && info.Before == "" {
		panic(fmt.Sprintf("%s: After is set, but Before is empty", info.Name))
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
