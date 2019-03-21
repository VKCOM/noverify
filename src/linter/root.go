package linter

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/state"
	"github.com/VKCOM/noverify/src/vscode"
	"github.com/z7zmey/php-parser/comment"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/php7"
	"github.com/z7zmey/php-parser/position"
	"github.com/z7zmey/php-parser/printer"
	"github.com/z7zmey/php-parser/walker"
)

const (
	maxFunctionLines = 150
)

// RootWalker is used to analyze root scope. Mostly defines, function and class definitions are analyzed.
//
// Current list of annotated checks:
//	- complexity
//	- modifiers
//	- phpdoc
//	- stdInterface
//	- syntax
//	- unused
type RootWalker struct {
	filename string
	comments comment.Comments

	lineRanges []git.LineRange

	custom      []RootChecker
	customBlock []BlockCheckerCreateFunc
	customState map[string]interface{}

	// internal state
	meta fileMeta

	st               *meta.ClassParseState
	currentClassNode node.Node

	disabledFlag bool // user-defined flag that file should not be linted

	reports []*Report

	// state required for both language server and reports creation
	Positions      position.Positions
	Lines          [][]byte
	LinesPositions []int

	// exposed meta-information for language server to use
	Scopes      map[node.Node]*meta.Scope
	Diagnostics []vscode.Diagnostic
}

// Report is a linter report message.
type Report struct {
	checkName  string
	startLn    string
	startChar  int
	startLine  int
	endChar    int
	level      int
	msg        string
	filename   string
	isDisabled bool // user-defined flag that file should not be linted
}

// CheckName returns report associated check name.
func (r *Report) CheckName() string {
	return r.checkName
}

func (r *Report) String() string {
	contextLn := strings.Builder{}
	for i, ch := range string(r.startLn) {
		if i == r.startChar {
			break
		}
		if ch == '\t' {
			contextLn.WriteRune(ch)
		} else {
			contextLn.WriteByte(' ')
		}
	}

	if r.endChar > r.startChar {
		contextLn.WriteString(strings.Repeat("^", r.endChar-r.startChar))
	}

	msg := r.msg
	if r.checkName != "" {
		msg = r.checkName + ": " + msg
	}
	return fmt.Sprintf("%s %s at %s:%d\n%s\n%s", severityNames[r.level], msg, r.filename, r.startLine, r.startLn, contextLn.String())
}

// IsCritical returns whether or not we need to reject whole commit when found this kind of report.
func (r *Report) IsCritical() bool {
	return r.level != LevelDoNotReject
}

// IsDisabledByUser returns whether or not user thinks that this file should not be checked
func (r *Report) IsDisabledByUser() bool {
	return r.isDisabled
}

// GetFilename returns report filename
func (r *Report) GetFilename() string {
	return r.filename
}

type phpDocParamEl struct {
	optional bool
	typ      *meta.TypesMap
}

type phpDocParamsMap map[string]phpDocParamEl

// NewWalkerForLangServer creates a copy of RootWalker to make full analysis of a file
func NewWalkerForLangServer(prev *RootWalker) *RootWalker {
	return &RootWalker{
		filename:       prev.filename,
		Positions:      prev.Positions,
		comments:       prev.comments,
		LinesPositions: prev.LinesPositions,
		Lines:          prev.Lines,
		lineRanges:     prev.lineRanges,
		st:             &meta.ClassParseState{},
	}
}

// NewWalkerForReferencesSearcher allows to access full context of a parser so that we can perform complex
// searches if needed.
func NewWalkerForReferencesSearcher(filename string, block BlockCheckerCreateFunc) *RootWalker {
	d := &RootWalker{
		filename:    filename,
		st:          &meta.ClassParseState{},
		customBlock: []BlockCheckerCreateFunc{block},
	}
	return d
}

// InitFromParser initializes common fields that are needed for RootWalker work
func (d *RootWalker) InitFromParser(contents []byte, parser *php7.Parser) {
	lines := bytes.Split(contents, []byte("\n"))
	linesPositions := make([]int, len(lines))
	pos := 0
	for idx, ln := range lines {
		linesPositions[idx] = pos
		pos += len(ln) + 1
	}

	d.Positions = parser.GetPositions()
	d.comments = parser.GetComments()
	d.LinesPositions = linesPositions
	d.Lines = lines
}

// InitCustom is needed to initialize walker state
func (d *RootWalker) InitCustom() {
	d.custom = nil
	for _, createFn := range customRootLinters {
		d.custom = append(d.custom, createFn(d))
	}

	d.customBlock = customBlockLinters
}

// UpdateMetaInfo is intended to be used in tests. Do not use it directly!
func (d *RootWalker) UpdateMetaInfo() {
	updateMetaInfo(d.filename, &d.meta)
}

// Scope returns root-level variable scope if applicable.
func (d *RootWalker) Scope() *meta.Scope {
	if d.meta.Scope == nil {
		d.meta.Scope = meta.NewScope()
	}
	return d.meta.Scope
}

// State allows for custom hooks to store state between entering root context and block context.
func (d *RootWalker) State() map[string]interface{} {
	if d.customState == nil {
		d.customState = make(map[string]interface{})
	}
	return d.customState
}

// GetReports returns collected reports for this file.
func (d *RootWalker) GetReports() []*Report {
	return d.reports
}

// ClassParseState returns class parse state (namespace, current class, etc)
func (d *RootWalker) ClassParseState() *meta.ClassParseState {
	return d.st
}

// EnterNode is invoked at every node in hierarchy
func (d *RootWalker) EnterNode(w walker.Walkable) (res bool) {
	res = true

	for _, c := range d.custom {
		c.BeforeEnterNode(w)
	}

	if n, ok := w.(node.Node); ok {
		for _, c := range d.comments[n] {
			d.handleComment(c)
		}
	}

	state.EnterNode(d.st, w)

	switch n := w.(type) {
	case *stmt.Interface:
		d.currentClassNode = n
	case *stmt.Class:
		d.currentClassNode = n
		cl := d.getClass()
		for _, tr := range n.Implements {
			interfaceName, ok := solver.GetClassName(d.st, tr)
			if ok {
				cl.Interfaces[interfaceName] = struct{}{}
			}
		}
	case *stmt.Trait:
		d.currentClassNode = n
	case *stmt.TraitUse:
		cl := d.getClass()
		for _, tr := range n.Traits {
			traitName, ok := solver.GetClassName(d.st, tr)
			if ok {
				cl.Traits[traitName] = struct{}{}
			}
		}
	case *assign.Assign:
		v, ok := n.Variable.(*expr.Variable)
		if !ok {
			break
		}

		d.Scope().AddVar(v, solver.ExprTypeLocal(d.meta.Scope, d.st, n.Expression), "global variable", true)
	case *stmt.Function:
		res = d.enterFunction(n)
	case *stmt.PropertyList:
		res = d.enterPropertyList(n)
	case *stmt.ClassConstList:
		res = d.enterClassConstList(n)
	case *stmt.ClassMethod:
		res = d.enterClassMethod(n)
	case *expr.FunctionCall:
		res = d.enterFunctionCall(n)
	case *stmt.ConstList:
		res = d.enterConstList(n)
	}

	for _, c := range d.custom {
		c.AfterEnterNode(w)
	}

	return res
}

func (d *RootWalker) parseStartPos(pos *position.Position) (startLn []byte, startChar int) {
	if pos.StartLine >= 1 && len(d.Lines) > pos.StartLine {
		startLn = d.Lines[pos.StartLine-1]
		p := d.LinesPositions[pos.StartLine-1]
		if pos.StartPos > p {
			startChar = pos.StartPos - p - 1
		}
	}

	return startLn, startChar
}

// Report registers a single report message about some found problem.
func (d *RootWalker) Report(n node.Node, level int, checkName, msg string, args ...interface{}) {
	if !meta.IsIndexingComplete() {
		return
	}

	var pos position.Position

	if n == nil {
		// Hack to parse syntax error message from php-parser.
		// When in language server mode, do not map syntax errors in order not to
		// complain about unfinished piece of code that user is currently writing.
		if strings.Contains(msg, "syntax error") && strings.Contains(msg, " at line ") && !LangServer {
			// it is in form "Syntax error: syntax error: unexpected '*' at line 4"
			if lastIdx := strings.LastIndexByte(msg, ' '); lastIdx > 0 {
				lineNumStr := msg[lastIdx+1:]
				lineNum, err := strconv.Atoi(lineNumStr)
				if err == nil {
					pos.StartLine = lineNum
					pos.EndLine = lineNum
					msg = msg[0:lastIdx]
					msg = strings.TrimSuffix(msg, " at line")
				}
			}
		}
	} else {
		pos = *d.Positions[n]
	}

	var endLn []byte
	var endChar int

	startLn, startChar := d.parseStartPos(&pos)

	if pos.EndLine >= 1 && len(d.Lines) > pos.EndLine {
		endLn = d.Lines[pos.EndLine-1]
		p := d.LinesPositions[pos.EndLine-1]
		if pos.EndPos > p {
			endChar = pos.EndPos - p
		}
	} else {
		endLn = startLn
	}

	if endChar == 0 {
		endChar = len(endLn)
	}

	if LangServer {
		severity, ok := vscodeLevelMap[level]
		if ok {
			diag := vscode.Diagnostic{
				Code:     msg,
				Message:  fmt.Sprintf(msg, args...),
				Severity: severity,
				Range: vscode.Range{
					Start: vscode.Position{Line: pos.StartLine - 1, Character: startChar},
					End:   vscode.Position{Line: pos.EndLine - 1, Character: endChar},
				},
			}

			if level == LevelUnused {
				diag.Tags = append(diag.Tags, 1 /* Unnecessary */)
			}

			d.Diagnostics = append(d.Diagnostics, diag)
		}
	} else {
		d.reports = append(d.reports, &Report{
			checkName:  checkName,
			startLn:    string(startLn),
			startChar:  startChar,
			startLine:  pos.StartLine,
			endChar:    endChar,
			level:      level,
			filename:   d.filename,
			msg:        fmt.Sprintf(msg, args...),
			isDisabled: d.disabledFlag,
		})
	}
}

func (d *RootWalker) reportUndefinedVariable(s *expr.Variable, maybeHave bool) {
	name, ok := s.VarName.(*node.Identifier)
	if !ok {
		d.Report(s, LevelInformation, "undefined", "Variable variable used")
		return
	}

	if _, ok := superGlobals[name.Value]; ok {
		return
	}

	if maybeHave {
		d.Report(s, LevelInformation, "undefined", "Variable might have not been defined: %s", name.Value)
	} else {
		d.Report(s, LevelError, "undefined", "Undefined variable: %s", name.Value)
	}
}

// FmtNode is used for debug purposes and returns string representation of a specified node.
func FmtNode(n node.Node) string {
	var b bytes.Buffer
	printer.NewPrinter(&b, " ").Print(n)
	return b.String()
}

func (d *RootWalker) handleComment(c comment.Comment) {
	str := c.String()
	if !phpdoc.IsPHPDoc(str) {
		return
	}

	for _, ln := range phpdoc.Parse(str) {
		if ln.Name != "linter" {
			continue
		}

		for _, p := range ln.Params {
			if p == "disable" {
				d.disabledFlag = true
			}
		}
	}
}

func (d *RootWalker) handleFuncStmts(params []meta.FuncParam, uses, stmts []node.Node, sc *meta.Scope) (returnTypes *meta.TypesMap, prematureExitFlags int) {
	b := &BlockWalker{sc: sc, r: d, unusedVars: make(map[string][]node.Node), nonLocalVars: make(map[string]struct{})}
	for _, createFn := range d.customBlock {
		b.custom = append(b.custom, createFn(b))
	}

	for _, useExpr := range uses {
		u := useExpr.(*expr.ClosureUse)
		v := u.Variable.(*expr.Variable)
		varName := v.VarName.(*node.Identifier).Value

		typ, ok := sc.GetVarNameType(varName)
		if !ok {
			typ = meta.NewTypesMap("TODO_use_var")
		}

		sc.AddVar(v, typ, "use", true)

		if !u.ByRef {
			b.unusedVars[varName] = append(b.unusedVars[varName], v)
		} else {
			b.nonLocalVars[varName] = struct{}{}
		}
	}

	for _, p := range params {
		if p.IsRef {
			b.nonLocalVars[p.Name] = struct{}{}
		}
	}
	for _, s := range stmts {
		b.addStatement(s)
		s.Walk(b)
	}
	b.flushUnused()

	// we can mark function as exiting abnormally if and only if
	// it only exits with die; or throw; and does not exit
	// using return; or any other control structure
	cleanFlags := b.exitFlags & (FlagDie | FlagThrow)

	if b.exitFlags == cleanFlags && (b.containsExitFlags&FlagReturn) == 0 {
		prematureExitFlags = cleanFlags
	}

	return b.returnTypes, prematureExitFlags
}

func (d *RootWalker) getElementPos(n node.Node) meta.ElementPosition {
	pos := d.Positions[n]
	_, startChar := d.parseStartPos(pos)

	return meta.ElementPosition{
		Filename:  d.filename,
		Character: int32(startChar),
		Line:      int32(pos.StartLine),
		EndLine:   int32(pos.EndLine),
		Length:    int32(pos.EndPos - pos.StartPos),
	}
}

func (d *RootWalker) addScope(n node.Node, sc *meta.Scope) {
	if d.Scopes == nil {
		d.Scopes = make(map[node.Node]*meta.Scope)
	}
	d.Scopes[n] = sc
}

type methodModifiers struct {
	abstract    bool
	static      bool
	accessLevel meta.AccessLevel
	final       bool
}

func (d *RootWalker) parseMethodModifiers(meth *stmt.ClassMethod) (res methodModifiers) {
	res.accessLevel = meta.Public

	for _, m := range meth.Modifiers {
		switch v := m.(*node.Identifier).Value; v {
		case "abstract":
			res.abstract = true
		case "static":
			res.static = true
		case "public":
			res.accessLevel = meta.Public
		case "private":
			res.accessLevel = meta.Private
		case "protected":
			res.accessLevel = meta.Protected
		case "final":
			res.final = true
		default:
			d.Report(m, LevelWarning, "modifiers: Unrecognized method modifier: %s", v)
		}
	}

	return res
}

func (d *RootWalker) getClass() meta.ClassInfo {
	var m meta.ClassesMap

	if d.st.IsTrait {
		if d.meta.Traits == nil {
			d.meta.Traits = make(meta.ClassesMap)
		}
		m = d.meta.Traits
	} else {
		if d.meta.Classes == nil {
			d.meta.Classes = make(meta.ClassesMap)
		}
		m = d.meta.Classes
	}

	cl, ok := m[d.st.CurrentClass]
	if !ok {
		cl = meta.ClassInfo{
			Pos:              d.getElementPos(d.currentClassNode),
			Parent:           d.st.CurrentParentClass,
			ParentInterfaces: d.st.CurrentParentInterfaces,
			Interfaces:       make(map[string]struct{}),
			Traits:           make(map[string]struct{}),
			Methods:          make(meta.FunctionsMap),
			Properties:       make(meta.PropertiesMap),
			Constants:        make(meta.ConstantsMap),
		}

		m[d.st.CurrentClass] = cl
	}

	return cl
}

func (d *RootWalker) enterPropertyList(pl *stmt.PropertyList) bool {
	cl := d.getClass()

	isStatic := false
	accessLevel := meta.Public

	for _, m := range pl.Modifiers {
		switch m.(*node.Identifier).Value {
		case "public":
			accessLevel = meta.Public
		case "protected":
			accessLevel = meta.Protected
		case "private":
			accessLevel = meta.Private
		case "static":
			isStatic = true
		}
	}

	for _, pNode := range pl.Properties {
		p := pNode.(*stmt.Property)

		nm := p.Variable.(*expr.Variable).VarName.(*node.Identifier).Value

		typ, errText := d.parsePHPDocVar(p.PhpDocComment)
		if errText != "" {
			d.Report(p.Variable, LevelWarning, "phpdoc", errText)
		}

		if p.Expr != nil {
			typ = typ.Append(solver.ExprTypeLocal(d.meta.Scope, d.st, p.Expr))
		}

		if isStatic {
			nm = "$" + nm
		}

		// TODO: handle duplicate property
		cl.Properties[nm] = meta.PropertyInfo{
			Pos:         d.getElementPos(p),
			Typ:         typ.Immutable(),
			AccessLevel: accessLevel,
		}
	}

	return true
}

func (d *RootWalker) enterClassConstList(s *stmt.ClassConstList) bool {
	cl := d.getClass()
	accessLevel := meta.Public

	for _, m := range s.Modifiers {
		switch m.(*node.Identifier).Value {
		case "public":
			accessLevel = meta.Public
		case "protected":
			accessLevel = meta.Protected
		case "private":
			accessLevel = meta.Private
		}
	}

	for _, cNode := range s.Consts {
		c := cNode.(*stmt.Constant)

		nm := c.ConstantName.(*node.Identifier).Value
		typ := solver.ExprTypeLocal(d.meta.Scope, d.st, c.Expr)

		// TODO: handle duplicate constant
		cl.Constants[nm] = meta.ConstantInfo{
			Pos:         d.getElementPos(c),
			Typ:         typ.Immutable(),
			AccessLevel: accessLevel,
		}
	}

	return true
}

func (d *RootWalker) enterClassMethod(meth *stmt.ClassMethod) bool {
	nm := meth.MethodName.(*node.Identifier).Value

	pos := d.Positions[meth]

	if funcSize := pos.EndLine - pos.StartLine; funcSize > maxFunctionLines {
		d.Report(meth.MethodName, LevelDoNotReject, "complexity", "Too big method: more than %d lines", maxFunctionLines)
	}

	modif := d.parseMethodModifiers(meth)

	sc := meta.NewScope()
	if !modif.static {
		sc.AddVarName("this", meta.NewTypesMap(d.st.CurrentClass).Immutable(), "instance method", true)
		sc.SetInInstanceMethod(true)
	}

	var specifiedReturnType *meta.TypesMap
	if typ, ok := d.parseTypeNode(meth.ReturnType); ok {
		specifiedReturnType = typ
	}

	phpdocReturnType, phpDocParamTypes, phpDocError := d.parsePHPDoc(meth.PhpDocComment, meth.Params)

	if phpDocError != "" {
		d.Report(meth.MethodName, LevelInformation, "phpdoc", "PHPDoc is incorrect: %s", phpDocError)
	}

	params, minParamsCnt := d.parseFuncArgs(meth.Params, phpDocParamTypes, sc)
	actualReturnTypes, exitFlags := d.handleFuncStmts(params, nil, meth.Stmts, sc)

	d.addScope(meth, sc)

	// TODO: handle duplicate method
	class := d.getClass()
	typ := meta.MergeTypeMaps(phpdocReturnType, actualReturnTypes, specifiedReturnType).Immutable()

	class.Methods[nm] = meta.FuncInfo{
		Params:       params,
		Pos:          d.getElementPos(meth),
		Typ:          typ,
		MinParamsCnt: minParamsCnt,
		AccessLevel:  modif.accessLevel,
		ExitFlags:    exitFlags,
	}

	if nm == "getIterator" && meta.IsIndexingComplete() && solver.Implements(d.st.CurrentClass, `\IteratorAggregate`) {
		implementsTraversable := false
		typ.Iterate(func(typ string) {
			if implementsTraversable {
				return
			}

			if solver.Implements(typ, `\Traversable`) {
				implementsTraversable = true
			}
		})

		if !implementsTraversable {
			d.Report(meth.MethodName, LevelError, "stdInterface", "Objects returned by %s::getIterator() must be traversable or implement interface \\Iterator", d.st.CurrentClass)
		}
	}

	return false
}

func (d *RootWalker) parsePHPDocVar(doc string) (m *meta.TypesMap, phpDocError string) {
	if doc == "" {
		return m, ""
	}

	lines := strings.Split(doc, "\n")
	for idx, ln := range lines {
		ln = strings.TrimSpace(ln)
		if len(ln) == 0 {
			phpDocError = fmt.Sprintf("empty line %d", idx)
			continue
		}

		ln = strings.TrimPrefix(ln, "/**")
		ln = strings.TrimPrefix(ln, "*")
		ln = strings.TrimSuffix(ln, "*/")

		if !strings.Contains(ln, "@var") {
			continue
		}

		fields := strings.Fields(ln)
		if len(fields) >= 2 && fields[0] == "@var" {
			m = meta.NewTypesMap(d.maybeAddNamespace(fields[1]))
		}
		continue
	}

	return m, phpDocError
}

func (d *RootWalker) maybeAddNamespace(typStr string) string {
	if typStr == "" {
		return ""
	}

	classNames := strings.Split(typStr, `|`)
	for idx, className := range classNames {
		// ignore things like \tuple(*)
		if braceIdx := strings.IndexByte(className, '('); braceIdx >= 0 {
			className = className[0:braceIdx]
		}

		// 0 for "bool", 1 for "bool[]", 2 for "bool[][]" and so on
		arrayDim := 0
		for strings.HasSuffix(className, "[]") {
			arrayDim++
			className = strings.TrimSuffix(className, "[]")
		}

		if len(className) == 0 {
			continue
		}

		switch className {
		case "bool", "boolean", "true", "false", "double", "float", "string", "int", "array", "resource", "mixed", "null", "callable", "void", "object":
			continue
		}

		if className[0] == '\\' {
			continue
		}

		if className[0] <= meta.WMax {
			log.Printf("Bad type: '%s' in file %s", className, d.filename)
			classNames[idx] = ""
			continue
		}

		fullClassName, ok := solver.GetClassName(d.st, meta.StringToName(className))
		if !ok {
			classNames[idx] = ""
			continue
		}

		if arrayDim > 0 {
			fullClassName += strings.Repeat("[]", arrayDim)
		}

		classNames[idx] = fullClassName
	}

	return strings.Join(classNames, "|")
}

func (d *RootWalker) parsePHPDoc(doc string, actualParams []node.Node) (returnType *meta.TypesMap, types phpDocParamsMap, phpDocError string) {
	returnType = &meta.TypesMap{}

	if doc == "" {
		return returnType.Immutable(), types, phpDocError
	}

	types = make(phpDocParamsMap, len(actualParams))

	var curParam int

	lines := strings.Split(doc, "\n")
	for idx, ln := range lines {
		ln = strings.TrimSpace(ln)
		if len(ln) == 0 {
			phpDocError = fmt.Sprintf("empty line %d", idx)
			continue
		}

		ln = strings.TrimPrefix(ln, "/**")
		ln = strings.TrimPrefix(ln, "*")
		ln = strings.TrimSuffix(ln, "*/")

		if strings.Contains(ln, "@return") {
			fields := strings.Fields(ln)
			if len(fields) >= 2 && fields[0] == "@return" {
				returnType = meta.NewTypesMap(d.maybeAddNamespace(fields[1]))
			}
			continue
		}

		if !strings.Contains(ln, "@param") {
			continue
		}

		fields := strings.Fields(ln)

		if len(fields) < 2 || fields[0] != "@param" {
			continue
		}

		typ := fields[1]
		optional := strings.Contains(ln, "[optional]")
		var variable string
		if len(fields) >= 3 {
			variable = fields[2]
		}

		if strings.HasPrefix(typ, "$") && !strings.HasPrefix(variable, "$") {
			variable, typ = typ, variable
		}

		if !strings.HasPrefix(variable, "$") {
			if len(actualParams) > curParam {
				variable = actualParams[curParam].(*node.Parameter).Variable.(*expr.Variable).VarName.(*node.Identifier).Value
			} else {
				phpDocError = fmt.Sprintf("too many @param tags on line %d", idx)
				continue
			}
		}

		curParam++

		variable = strings.TrimPrefix(variable, "$")
		types[variable] = phpDocParamEl{
			optional: optional,
			typ:      meta.NewTypesMap(d.maybeAddNamespace(typ)),
		}
	}

	return returnType.Immutable(), types, phpDocError
}

// parse type info, e.g. "string" in "someFunc() : string { ... }"
func (d *RootWalker) parseTypeNode(n node.Node) (typ *meta.TypesMap, ok bool) {
	if n == nil {
		return nil, false
	}

	switch t := n.(type) {
	case *name.Name:
		typ = meta.NewTypesMap(d.maybeAddNamespace(meta.NameToString(t)))
	case *name.FullyQualified:
		typ = meta.NewTypesMap(meta.FullyQualifiedToString(t))
	case *node.Identifier:
		typ = meta.NewTypesMap(t.Value)
	}

	return typ, typ != nil
}

func (d *RootWalker) parseFuncArgs(params []node.Node, parTypes phpDocParamsMap, sc *meta.Scope) (args []meta.FuncParam, minArgs int) {
	args = make([]meta.FuncParam, 0, len(params))
	for _, param := range params {
		p := param.(*node.Parameter)
		v := p.Variable.(*expr.Variable)
		parTyp := parTypes[v.VarName.(*node.Identifier).Value]

		if !parTyp.typ.IsEmpty() {
			sc.AddVar(v, parTyp.typ, "param", true)
		}

		typ := parTyp.typ

		if p.DefaultValue == nil && !parTyp.optional {
			minArgs++
		}

		if p.VariableType != nil {
			if varTyp, ok := d.parseTypeNode(p.VariableType); ok {
				typ = varTyp
			}
		} else if typ.IsEmpty() && p.DefaultValue != nil {
			typ = solver.ExprTypeLocal(sc, d.st, p.DefaultValue)
		}

		if p.Variadic {
			arrTyp := meta.NewEmptyTypesMap(typ.Len())
			typ.Iterate(func(t string) { arrTyp = arrTyp.AppendString(meta.WrapArrayOf(t)) })
			typ = arrTyp
		}

		sc.AddVar(v, typ, "param", true)

		par := meta.FuncParam{
			Typ:   typ.Immutable(),
			IsRef: p.ByRef,
		}

		if id, ok := v.VarName.(*node.Identifier); ok {
			par.Name = id.Value
		}

		args = append(args, par)
	}
	return args, minArgs
}

func (d *RootWalker) enterFunction(fun *stmt.Function) bool {
	nm := d.st.Namespace + `\` + fun.FunctionName.(*node.Identifier).Value
	pos := d.Positions[fun]

	if funcSize := pos.EndLine - pos.StartLine; funcSize > maxFunctionLines {
		d.Report(fun.FunctionName, LevelDoNotReject, "complexity", "Too big function: more than %d lines", maxFunctionLines)
	}

	var specifiedReturnType *meta.TypesMap
	if typ, ok := d.parseTypeNode(fun.ReturnType); ok {
		specifiedReturnType = typ
	}

	phpdocReturnType, phpDocParamTypes, phpDocError := d.parsePHPDoc(fun.PhpDocComment, fun.Params)

	if phpDocError != "" {
		d.Report(fun.FunctionName, LevelInformation, "phpdoc: PHPDoc is incorrect: %s", phpDocError)
	}

	if d.meta.Functions == nil {
		d.meta.Functions = make(meta.FunctionsMap)
	}

	sc := meta.NewScope()

	params, minParamsCnt := d.parseFuncArgs(fun.Params, phpDocParamTypes, sc)

	actualReturnTypes, exitFlags := d.handleFuncStmts(params, nil, fun.Stmts, sc)
	d.addScope(fun, sc)

	d.meta.Functions[nm] = meta.FuncInfo{
		Params:       params,
		Pos:          d.getElementPos(fun),
		Typ:          meta.MergeTypeMaps(phpdocReturnType, actualReturnTypes, specifiedReturnType).Immutable(),
		MinParamsCnt: minParamsCnt,
		ExitFlags:    exitFlags,
	}

	return false
}

func isQuote(r rune) bool {
	return r == '"' || r == '\''
}

func (d *RootWalker) enterFunctionCall(s *expr.FunctionCall) bool {
	nm, ok := s.Function.(*name.Name)
	if !ok {
		return true
	}

	if d.st.Namespace == `\PHPSTORM_META` && meta.NameEquals(nm, `override`) {
		return d.handleOverride(s)
	}

	if !meta.NameEquals(nm, `define`) || len(s.Arguments) < 2 {
		// TODO: actually we could warn about bogus defines
		return true
	}

	arg, ok := s.Arguments[0].(*node.Argument)
	if !ok {
		return true
	}

	str, ok := arg.Expr.(*scalar.String)
	if !ok {
		return true
	}

	valueArg, ok := s.Arguments[1].(*node.Argument)
	if !ok {
		return true
	}

	if d.meta.Constants == nil {
		d.meta.Constants = make(meta.ConstantsMap)
	}

	d.meta.Constants[`\`+strings.TrimFunc(str.Value, isQuote)] = meta.ConstantInfo{
		Pos: d.getElementPos(s),
		Typ: solver.ExprTypeLocal(d.meta.Scope, d.st, valueArg.Expr),
	}
	return true
}

// Handle e.g. "override(\array_shift(0), elementType(0));"
// which means "return type of array_shift() is the type of element of first function parameter"
func (d *RootWalker) handleOverride(s *expr.FunctionCall) bool {
	if len(s.Arguments) != 2 {
		return true
	}

	arg0, ok := s.Arguments[0].(*node.Argument)
	if !ok {
		return true
	}

	arg1, ok := s.Arguments[1].(*node.Argument)
	if !ok {
		return true
	}

	fc0, ok := arg0.Expr.(*expr.FunctionCall)
	if !ok {
		return true
	}

	fc1, ok := arg1.Expr.(*expr.FunctionCall)
	if !ok {
		return true
	}

	fnNameNode, ok := fc0.Function.(*name.FullyQualified)
	if !ok {
		return true
	}

	overrideNameNode, ok := fc1.Function.(*name.Name)
	if !ok {
		return true
	}

	if len(fc1.Arguments) != 1 {
		return true
	}

	fc1Arg0, ok := fc1.Arguments[0].(*node.Argument)
	if !ok {
		return true
	}

	argNumNode, ok := fc1Arg0.Expr.(*scalar.Lnumber)
	if !ok {
		return true
	}

	argNum, err := strconv.Atoi(argNumNode.Value)
	if err != nil {
		return true
	}

	var overrideTyp meta.OverrideType
	switch {
	case meta.NameEquals(overrideNameNode, `type`):
		overrideTyp = meta.OverrideArgType
	case meta.NameEquals(overrideNameNode, `elementType`):
		overrideTyp = meta.OverrideElementType
	default:
		return true
	}

	fnName := meta.FullyQualifiedToString(fnNameNode)

	if d.meta.FunctionOverrides == nil {
		d.meta.FunctionOverrides = make(meta.FunctionsOverrideMap)
	}

	d.meta.FunctionOverrides[fnName] = meta.FuncInfoOverride{
		OverrideType: overrideTyp,
		ArgNum:       argNum,
	}

	return true
}

func (d *RootWalker) enterConstList(lst *stmt.ConstList) bool {
	if d.meta.Constants == nil {
		d.meta.Constants = make(meta.ConstantsMap)
	}

	for _, sNode := range lst.Consts {
		s := sNode.(*stmt.Constant)

		id, ok := s.ConstantName.(*node.Identifier)
		if !ok {
			continue
		}

		nm := d.st.Namespace + `\` + id.Value

		d.meta.Constants[nm] = meta.ConstantInfo{
			Pos: d.getElementPos(s),
			Typ: solver.ExprTypeLocal(d.meta.Scope, d.st, s.Expr),
		}
	}

	return false
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *RootWalker) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *RootWalker) LeaveNode(n walker.Walkable) {
	for _, c := range d.custom {
		c.BeforeLeaveNode(n)
	}

	switch n.(type) {
	case *stmt.Class, *stmt.Interface, *stmt.Trait:
		d.getClass() // populate classes map

		d.currentClassNode = nil
	}

	state.LeaveNode(d.st, n)

	for _, c := range d.custom {
		c.AfterLeaveNode(n)
	}
}
