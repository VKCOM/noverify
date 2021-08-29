package linter

import (
	"bytes"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/VKCOM/noverify/src/phpdoctypes"
	"github.com/VKCOM/noverify/src/utils"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/VKCOM/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/baseline"
	"github.com/VKCOM/noverify/src/constfold"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/phpgrep"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/state"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/workspace"
)

const (
	maxFunctionLines = 150
)

// rootWalker is used to analyze root scope. Mostly defines, function and class definitions are analyzed.
type rootWalker struct {
	// An associated file that is traversed by the current Root Walker.
	file *workspace.File

	custom      []RootChecker
	customBlock []BlockCheckerCreateFunc
	customState map[string]interface{}

	rootRset  *rules.ScopedSet
	localRset *rules.ScopedSet
	anyRset   *rules.ScopedSet

	ctx rootContext

	// nodeSet is a reusable node set for both root and block walkers.
	// TODO: move to WorkerContext as we store reusable objects there.
	nodeSet irutil.NodeSet

	reSimplifier *regexpSimplifier
	reVet        *regexpVet

	// internal state
	meta fileMeta

	// We need a stack here, since anonymous classes can be
	// inside common classes and anonymous.
	currentClassNodeStack irutil.NodePath

	allowDisabledRegexp *regexp.Regexp // user-defined flag that files suitable for this regular expression should not be linted
	linterDisabled      bool           // flag indicating whether linter is disabled. Flag is set to true only if the file
	// name matches the pattern and @linter disable was encountered

	// strictTypes is true if file contains `declare(strict_types=1)`.
	strictTypes bool
	strictMixed bool

	reports []*Report

	config *Config

	checkersFilter *CheckersFilter

	checker *rootChecker
}

// InitCustom is needed to initialize walker state
func (d *rootWalker) InitCustom() {
	d.custom = nil
	for _, createFn := range d.config.Checkers.rootCheckers {
		d.custom = append(d.custom, createFn(&RootContext{w: d}))
	}
	d.customBlock = append(d.customBlock, d.config.Checkers.blockCheckers...)
}

// Getters part.

// scope returns root-level variable scope if applicable.
func (d *rootWalker) scope() *meta.Scope {
	if d.meta.Scope == nil {
		d.meta.Scope = meta.NewScope()
	}
	return d.meta.Scope
}

// metaInfo returns meta info.
func (d *rootWalker) metaInfo() *meta.Info {
	return d.ctx.st.Info
}

// state allows for custom hooks to store state between entering root context and block context.
func (d *rootWalker) state() map[string]interface{} {
	if d.customState == nil {
		d.customState = make(map[string]interface{})
	}
	return d.customState
}

// File returns file for current root walker.
func (d *rootWalker) File() *workspace.File {
	return d.file
}

// Visitor part.

// EnterNode is invoked at every node in hierarchy.
func (d *rootWalker) EnterNode(n ir.Node) (res bool) {
	res = true

	for _, c := range d.custom {
		c.BeforeEnterNode(n)
	}

	n.IterateTokens(d.handleCommentToken)

	state.EnterNode(d.ctx.st, n)

	switch n := n.(type) {
	case *ir.DeclareStmt:
		for _, c := range n.Consts {
			c, ok := c.(*ir.ConstantStmt)
			if !ok {
				continue
			}
			if c.ConstantName.Value == "strict_types" {
				v, ok := c.Expr.(*ir.Lnumber)
				if ok && v.Value == "1" {
					d.strictTypes = true
				}
			}
		}

	case *ir.AnonClassExpr:
		d.currentClassNodeStack.Push(n)

		cl := d.getClass()
		className := &ir.Identifier{Value: cl.Name}

		d.checker.CheckImplements(n, cl, n.Implements)
		d.checker.CheckExtends(n, cl, n.Extends)

		d.checker.CheckCommentMisspellings(className, n.Doc.Raw)
		d.checker.CheckIdentMisspellings(className)

		doc := d.parseClassPHPDoc(className, n.Doc)
		d.reportPHPDocErrors(doc.errs)
		d.handleClassDoc(doc, &cl)

		d.meta.Classes.Set(d.ctx.st.CurrentClass, cl)

	case *ir.InterfaceStmt:
		d.currentClassNodeStack.Push(n)
		d.checker.CheckKeywordCase(n, "interface")
		d.checker.CheckCommentMisspellings(n.InterfaceName, n.Doc.Raw)
		if !strings.HasSuffix(n.InterfaceName.Value, "able") {
			d.checker.CheckIdentMisspellings(n.InterfaceName)
		}
	case *ir.ClassStmt:
		d.currentClassNodeStack.Push(n)

		cl := d.getClass()
		classFlags := d.getClassModifiers(n)

		if classFlags != 0 {
			// Since cl is not a pointer, and it's illegal to update
			// individual fields through map, we update cl and
			// then assign it back to the map.
			cl.Flags = classFlags
			d.meta.Classes.Set(d.ctx.st.CurrentClass, cl)
		}

		for _, m := range n.Modifiers {
			d.checker.CheckModifierKeywordCase(m)
		}

		d.checker.CheckImplements(n, cl, n.Implements)
		d.checker.CheckExtends(n, cl, n.Extends)

		d.checker.CheckCommentMisspellings(n.ClassName, n.Doc.Raw)
		d.checker.CheckIdentMisspellings(n.ClassName)

		doc := d.parseClassPHPDoc(n, n.Doc)
		d.reportPHPDocErrors(doc.errs)
		d.handleClassDoc(doc, &cl)

		d.meta.Classes.Set(d.ctx.st.CurrentClass, cl)

	case *ir.TraitStmt:
		d.currentClassNodeStack.Push(n)
		d.checker.CheckKeywordCase(n, "trait")
		d.checker.CheckCommentMisspellings(n.TraitName, n.Doc.Raw)
		d.checker.CheckIdentMisspellings(n.TraitName)
	case *ir.TraitUseStmt:
		d.checker.CheckKeywordCase(n, "use")
		cl := d.getClass()
		for _, tr := range n.Traits {
			traitName, ok := solver.GetClassName(d.ctx.st, tr)
			if ok {
				cl.Traits[traitName] = struct{}{}
				d.checker.CheckTraitImplemented(d.currentClassNodeStack.Current(), tr, d.getClass(), traitName)
			}
		}
	case *ir.Assign:
		v, ok := n.Variable.(*ir.SimpleVar)
		if !ok {
			break
		}

		d.scope().AddVar(v, solver.ExprTypeLocal(d.scope(), d.ctx.st, n.Expr), "global variable", meta.VarAlwaysDefined)
	case *ir.FunctionStmt:
		if d.metaInfo().IsIndexingComplete() {
			res = d.checker.CheckFunction(n)
		} else {
			res = d.enterFunction(n)
		}
	case *ir.PropertyListStmt:
		res = d.enterPropertyList(n)
	case *ir.ClassConstListStmt:
		res = d.enterClassConstList(n)
	case *ir.ClassMethodStmt:
		res = d.enterClassMethod(n)
	case *ir.FunctionCallExpr:
		res = d.enterFunctionCall(n)
	case *ir.ConstListStmt:
		res = d.enterConstList(n)

	case *ir.NamespaceStmt:
		d.checker.CheckKeywordCase(n, "namespace")
	}

	for _, c := range d.custom {
		c.AfterEnterNode(n)
	}

	if d.metaInfo().IsIndexingComplete() && d.rootRset != nil {
		kind := ir.GetNodeKind(n)
		d.runRules(n, d.scope(), d.rootRset.RulesByKind[kind])
	}

	if !res {
		// If we're not returning true from this method,
		// LeaveNode will not be called for this node.
		// But we still need to "leave" them if they
		// were entered in the ClassParseState.
		state.LeaveNode(d.ctx.st, n)
	}
	return res
}

// LeaveNode is invoked after node process.
func (d *rootWalker) LeaveNode(n ir.Node) {
	for _, c := range d.custom {
		c.BeforeLeaveNode(n)
	}

	switch n.(type) {
	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt, *ir.AnonClassExpr:
		d.getClass() // populate classes map

		d.currentClassNodeStack.Pop()
	}

	state.LeaveNode(d.ctx.st, n)

	for _, c := range d.custom {
		c.AfterLeaveNode(n)
	}
}

// beforeEnterFile is invoked before file process.
func (d *rootWalker) beforeEnterFile() {
	for _, c := range d.custom {
		c.BeforeEnterFile()
	}
}

// afterLeaveFile is invoked after file process.
func (d *rootWalker) afterLeaveFile() {
	for _, c := range d.custom {
		c.AfterLeaveFile()
	}

	if !d.metaInfo().IsIndexingComplete() {
		for _, shape := range d.ctx.shapes {
			props := make(meta.PropertiesMap)
			for _, p := range shape.Props {
				props[p.Key] = meta.PropertyInfo{
					Typ:         types.NewMapWithNormalization(d.ctx.typeNormalizer, p.Types).Immutable(),
					AccessLevel: meta.Public,
				}
			}
			cl := meta.ClassInfo{
				Name:       shape.Name,
				Properties: props,
				Flags:      meta.ClassShape,
			}
			if d.meta.Classes.H == nil {
				d.meta.Classes = meta.NewClassesMap()
			}
			d.meta.Classes.Set(shape.Name, cl)
		}
	}
}

// Handle comments part.

func (d *rootWalker) handleCommentToken(t *token.Token) bool {
	if !phpdoc.IsPHPDocToken(t) {
		return true
	}

	for _, ln := range phpdoc.Parse(d.ctx.phpdocTypeParser, string(t.Value)).Parsed {
		if ln.Name() != "linter" {
			continue
		}

		for _, p := range ln.(*phpdoc.RawCommentPart).Params {
			if p != "disable" {
				continue
			}
			if d.linterDisabled {
				needleLine := ln.Line() + t.Position.StartLine - 1 - 1
				d.ReportPHPDoc(
					PHPDocAbsoluteLineField(needleLine, 1),
					LevelWarning, "linterError", "Linter is already disabled for this file",
				)
				continue
			}
			canDisable := false
			if d.allowDisabledRegexp != nil {
				canDisable = d.allowDisabledRegexp.MatchString(d.ctx.st.CurrentFile)
			}
			d.linterDisabled = canDisable
			if !canDisable {
				needleLine := ln.Line() + t.Position.StartLine - 1 - 1
				d.ReportPHPDoc(
					PHPDocAbsoluteLineField(needleLine, 1),
					LevelWarning, "linterError", "You are not allowed to disable linter",
				)
			}
		}
	}

	return true
}

// Handle functions part.

type handleFuncResult struct {
	returnTypes            types.Map
	prematureExitFlags     int
	callsParentConstructor bool
}

func (d *rootWalker) handleArrowFuncExpr(params []meta.FuncParam, expr ir.Node, sc *meta.Scope, parentBlockWalker *blockWalker) handleFuncResult {
	b := newBlockWalker(d, sc)
	for _, createFn := range d.customBlock {
		b.custom = append(b.custom, createFn(&BlockContext{w: b}))
	}

	b.inArrowFunction = true
	parentBlockWalker.parentBlockWalkers = append(parentBlockWalker.parentBlockWalkers, parentBlockWalker)
	b.parentBlockWalkers = parentBlockWalker.parentBlockWalkers

	for _, p := range params {
		if p.IsRef {
			b.nonLocalVars[p.Name] = varRef
		}
	}

	b.addStatement(expr)
	expr.Walk(b)

	b.flushUnused()

	return handleFuncResult{
		returnTypes: b.returnTypes,
	}
}

func (d *rootWalker) handleFuncStmts(params []meta.FuncParam, uses, stmts []ir.Node, sc *meta.Scope) handleFuncResult {
	b := newBlockWalker(d, sc)
	for _, createFn := range d.customBlock {
		b.custom = append(b.custom, createFn(&BlockContext{w: b}))
	}

	for _, useExpr := range uses {
		var byRef bool
		var v *ir.SimpleVar
		switch u := useExpr.(type) {
		case *ir.ReferenceExpr:
			v = u.Variable.(*ir.SimpleVar)
			byRef = true
		case *ir.SimpleVar:
			v = u
		default:
			return handleFuncResult{}
		}

		typ, ok := sc.GetVarNameType(v.Name)
		if !ok {
			typ = types.NewMap("TODO_use_var")
		}

		sc.AddVar(v, typ, "use", meta.VarAlwaysDefined)

		if !byRef {
			b.unusedVars[v.Name] = append(b.unusedVars[v.Name], v)
		} else {
			b.nonLocalVars[v.Name] = varRef
		}
	}

	// It's OK to read from and delete from a nil map.
	// If a func/method has 0 params, don't allocate a map for it.
	if len(params) != 0 {
		b.unusedParams = make(map[string]struct{}, len(params))
	}
	for _, p := range params {
		if p.IsRef {
			b.nonLocalVars[p.Name] = varRef
		}
		if !p.IsRef && !d.config.IsDiscardVar(p.Name) {
			b.unusedParams[p.Name] = struct{}{}
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
	cleanFlags := b.ctx.exitFlags & (FlagDie | FlagThrow)

	var prematureExitFlags int
	if b.ctx.exitFlags == cleanFlags && (b.ctx.containsExitFlags&FlagReturn) == 0 {
		prematureExitFlags = cleanFlags
	}

	switch {
	case b.bareReturn && b.returnsValue:
		b.returnTypes = types.MergeMaps(b.returnTypes, types.NullType)
	case b.returnTypes.Empty() && b.returnsValue:
		b.returnTypes = types.MixedType
	}

	return handleFuncResult{
		returnTypes:            b.returnTypes,
		prematureExitFlags:     prematureExitFlags,
		callsParentConstructor: b.callsParentConstructor,
	}
}

type parseFuncParamsResult struct {
	params         []meta.FuncParam
	paramsTypeHint map[string]types.Map
	minParamsCount int
}

func (d *rootWalker) parseFuncParams(params []ir.Node, phpDocParamsTypes phpdoctypes.ParamsMap, sc *meta.Scope, closureSolver *solver.ClosureCallerInfo) (res parseFuncParamsResult) {
	if len(params) == 0 {
		return res
	}

	minArgs := 0
	args := make([]meta.FuncParam, 0, len(params))
	typeHints := make(map[string]types.Map, len(params))

	if closureSolver != nil && solver.IsSupportedFunction(closureSolver.Name) {
		return d.parseFuncArgsForCallback(params, sc, closureSolver)
	}

	for _, param := range params {
		p := param.(*ir.Parameter)
		v := p.Variable
		phpDocType := phpDocParamsTypes[v.Name]

		if !phpDocType.Typ.Empty() {
			sc.AddVarName(v.Name, phpDocType.Typ, "param", meta.VarAlwaysDefined)
		}

		paramTyp := phpDocType.Typ

		if p.DefaultValue == nil && !phpDocType.Optional && !p.Variadic {
			minArgs++
		}

		if p.VariableType != nil {
			typeHintType, ok := d.parseTypeHintNode(p.VariableType)
			if ok {
				paramTyp = typeHintType
			}

			typeHints[v.Name] = typeHintType
		} else if paramTyp.Empty() && p.DefaultValue != nil {
			paramTyp = solver.ExprTypeLocal(sc, d.ctx.st, p.DefaultValue)
			// For the type resolver default value can look like a
			// precise source of information (e.g. "false" is a precise bool),
			// but it's not assigned unconditionally.
			// If explicit argument is provided, that parameter can have
			// almost any type possible.
			paramTyp.MarkAsImprecise()
		}

		if p.Variadic {
			paramTyp = paramTyp.Map(types.WrapArrayOf)
		}

		sc.AddVarName(v.Name, paramTyp, "param", meta.VarAlwaysDefined)

		par := meta.FuncParam{
			Typ:   paramTyp.Immutable(),
			IsRef: p.ByRef,
		}

		par.Name = v.Name
		args = append(args, par)
	}

	return parseFuncParamsResult{
		params:         args,
		paramsTypeHint: typeHints,
		minParamsCount: minArgs,
	}
}

// callbackParamByIndex returns the description of the parameter for the function by its index.
func (d *rootWalker) callbackParamByIndex(param ir.Node, argType types.Map) meta.FuncParam {
	p := param.(*ir.Parameter)
	v := p.Variable

	var typ types.Map
	tp, ok := d.parseTypeHintNode(p.VariableType)
	if ok {
		typ = tp
	} else {
		typ = argType.Map(types.WrapElemOf)
	}

	arg := meta.FuncParam{
		IsRef: p.ByRef,
		Name:  v.Name,
		Typ:   typ,
	}
	return arg
}

func (d *rootWalker) parseFuncArgsForCallback(params []ir.Node, sc *meta.Scope, closureSolver *solver.ClosureCallerInfo) (res parseFuncParamsResult) {
	countParams := len(params)
	minArgs := countParams
	if countParams == 0 {
		return res
	}
	args := make([]meta.FuncParam, countParams)

	switch closureSolver.Name {
	case `\usort`, `\uasort`, `\array_reduce`:
		args[0] = d.callbackParamByIndex(params[0], closureSolver.ArgTypes[0])
		if countParams > 1 {
			args[1] = d.callbackParamByIndex(params[1], closureSolver.ArgTypes[0])
		}
	case `\array_walk`, `\array_walk_recursive`, `\array_filter`:
		args[0] = d.callbackParamByIndex(params[0], closureSolver.ArgTypes[0])
	case `\array_map`:
		args[0] = d.callbackParamByIndex(params[0], closureSolver.ArgTypes[1])
	}

	for i, param := range params {
		p := param.(*ir.Parameter)
		v := p.Variable
		var typ types.Map
		if i < len(args) {
			typ = args[i].Typ
		} else {
			typ = types.MixedType
		}

		sc.AddVarName(v.Name, typ, "param", meta.VarAlwaysDefined)
	}

	return parseFuncParamsResult{
		params:         args,
		minParamsCount: minArgs,
	}
}

func (d *rootWalker) enterFunction(fun *ir.FunctionStmt) bool {
	nm := d.ctx.st.Namespace + `\` + fun.FunctionName.Value

	if d.meta.Functions.H == nil {
		d.meta.Functions = meta.NewFunctionsMap()
	}

	// Indexing stage.
	doc := phpdoctypes.Parse(fun.Doc, fun.Params, d.ctx.typeNormalizer)
	moveShapesToContext(&d.ctx, doc.Shapes)
	d.handleClosuresFromDoc(doc.Closures)

	phpDocReturnType := doc.ReturnType
	phpDocParamTypes := doc.ParamTypes

	sc := meta.NewScope()

	returnTypeHint, _ := d.parseTypeHintNode(fun.ReturnType)
	funcParams := d.parseFuncParams(fun.Params, phpDocParamTypes, sc, nil)

	funcInfo := d.handleFuncStmts(funcParams.params, nil, fun.Stmts, sc)
	actualReturnTypes := funcInfo.returnTypes
	exitFlags := funcInfo.prematureExitFlags

	returnTypes := functionReturnType(phpDocReturnType, returnTypeHint, actualReturnTypes)

	var funcFlags meta.FuncFlags
	if solver.SideEffectFreeFunc(d.scope(), d.ctx.st, nil, fun.Stmts) {
		funcFlags |= meta.FuncPure
	}
	d.meta.Functions.Set(nm, meta.FuncInfo{
		Params:       funcParams.params,
		Name:         nm,
		Pos:          d.getElementPos(fun),
		Typ:          returnTypes.Immutable(),
		MinParamsCnt: funcParams.minParamsCount,
		Flags:        funcFlags,
		ExitFlags:    exitFlags,
		Doc:          doc.AdditionalInfo,
	})

	return false
}

// Handle functions call part.

func (d *rootWalker) enterFunctionCall(s *ir.FunctionCallExpr) bool {
	nm, ok := s.Function.(*ir.Name)
	if !ok {
		return true
	}

	name := strings.TrimPrefix(nm.Value, `\`)

	if d.ctx.st.Namespace == `\PHPSTORM_META` && name == `override` {
		return d.handleOverride(s)
	}

	if name == "define" {
		d.handleDefineCall(s)
	}

	return true
}

func (d *rootWalker) handleDefineCall(s *ir.FunctionCallExpr) {
	if len(s.Args) < 2 {
		return
	}

	arg := s.Arg(0)

	str, ok := arg.Expr.(*ir.String)
	if !ok {
		return
	}

	valueArg := s.Arg(1)

	if d.meta.Constants == nil {
		d.meta.Constants = make(meta.ConstantsMap)
	}

	value := constfold.Eval(d.ctx.st, valueArg)

	d.meta.Constants[`\`+strings.TrimFunc(str.Value, utils.IsQuote)] = meta.ConstInfo{
		Pos:   d.getElementPos(s),
		Typ:   solver.ExprTypeLocal(d.scope(), d.ctx.st, valueArg.Expr),
		Value: value,
	}
}

// handleOverride handle e.g. "override(\array_shift(0), elementType(0));"
// which means "return type of array_shift() is the type of element of first function parameter"
func (d *rootWalker) handleOverride(s *ir.FunctionCallExpr) bool {
	if len(s.Args) != 2 {
		return true
	}

	arg0 := s.Arg(0)
	arg1 := s.Arg(1)

	fc0, ok := arg0.Expr.(*ir.FunctionCallExpr)
	if !ok {
		return true
	}

	fc1, ok := arg1.Expr.(*ir.FunctionCallExpr)
	if !ok {
		return true
	}

	fnNameNode, ok := fc0.Function.(*ir.Name)
	if !ok || !fnNameNode.IsFullyQualified() {
		return true
	}

	overrideNameNode, ok := fc1.Function.(*ir.Name)
	if !ok {
		return true
	}

	if len(fc1.Args) != 1 {
		return true
	}

	fc1Arg0 := fc1.Arg(0)

	argNumNode, ok := fc1Arg0.Expr.(*ir.Lnumber)
	if !ok {
		return true
	}

	argNum, err := strconv.Atoi(argNumNode.Value)
	if err != nil {
		return true
	}

	var overrideTyp meta.OverrideType
	switch {
	case overrideNameNode.Value == `type`:
		overrideTyp = meta.OverrideArgType
	case overrideNameNode.Value == `elementType`:
		overrideTyp = meta.OverrideElementType
	default:
		return true
	}

	fnName := fnNameNode.Value

	if d.meta.FunctionOverrides == nil {
		d.meta.FunctionOverrides = make(meta.FunctionsOverrideMap)
	}

	d.meta.FunctionOverrides[fnName] = meta.FuncInfoOverride{
		OverrideType: overrideTyp,
		ArgNum:       argNum,
	}

	return true
}

// Handle class part.

func (d *rootWalker) getClassModifiers(n *ir.ClassStmt) meta.ClassFlags {
	var classFlags meta.ClassFlags
	for _, m := range n.Modifiers {
		switch {
		case strings.EqualFold("abstract", m.Value):
			classFlags |= meta.ClassAbstract
		case strings.EqualFold("final", m.Value):
			classFlags |= meta.ClassFinal
		}
	}
	return classFlags
}

func (d *rootWalker) parseClassPHPDoc(class ir.Node, doc phpdoc.Comment) classPHPDocParseResult {
	var result classPHPDocParseResult

	if doc.Raw == "" {
		return result
	}

	// TODO: allocate maps lazily.
	// Class may not have any @property or @method annotations.
	// In that case we can handle avoid map allocations.
	result.properties = make(meta.PropertiesMap)
	result.methods = meta.NewFunctionsMap()

	for _, part := range doc.Parsed {
		d.checker.checkPHPDocRef(class, part)
		switch part.Name() {
		case "property", "property-read", "property-write":
			parseClassPHPDocProperty(class, &d.ctx, &result, part.(*phpdoc.TypeVarCommentPart))
		case "method":
			parseClassPHPDocMethod(class, &d.ctx, &result, part.(*phpdoc.RawCommentPart))
		case "mixin":
			parseClassPHPDocMixin(class, d.ctx.st, &result, part.(*phpdoc.RawCommentPart))
		}
	}

	return result
}

func (d *rootWalker) handleClassDoc(doc classPHPDocParseResult, cl *meta.ClassInfo) {
	// If we ever need to distinguish @property-annotated and real properties,
	// more work will be required here.
	for name, p := range doc.properties {
		p.Pos = cl.Pos
		cl.Properties[name] = p
	}

	for name, m := range doc.methods.H {
		m.Pos = cl.Pos
		cl.Methods.H[name] = m
	}

	cl.Mixins = doc.mixins
}

func (d *rootWalker) parsePHPDocVar(doc phpdoc.Comment) (typesMap types.Map) {
	for _, part := range doc.Parsed {
		part, ok := part.(*phpdoc.TypeVarCommentPart)
		if ok && part.Name() == "var" {
			converted := phpdoctypes.ToRealType(d.ctx.typeNormalizer.ClassFQNProvider(), part.Type)
			moveShapesToContext(&d.ctx, converted.Shapes)
			d.handleClosuresFromDoc(converted.Closures)

			typesMap = types.NewMapWithNormalization(d.ctx.typeNormalizer, converted.Types)
		}
	}

	return typesMap
}

func (d *rootWalker) enterPropertyList(pl *ir.PropertyListStmt) bool {
	cl := d.getClass()

	isStatic := false
	accessLevel := meta.Public
	accessImplicit := true

	for _, m := range pl.Modifiers {
		d.checker.CheckModifierKeywordCase(m)
		switch strings.ToLower(m.Value) {
		case "public":
			accessLevel = meta.Public
			accessImplicit = false
		case "protected":
			accessLevel = meta.Protected
			accessImplicit = false
		case "private":
			accessLevel = meta.Private
			accessImplicit = false
		case "static":
			isStatic = true
		}
	}

	if accessImplicit {
		target := "property"
		if len(pl.Properties) > 1 {
			target = "properties"
		}
		d.Report(pl, LevelNotice, "implicitModifiers", "Specify the access modifier for %s explicitly", target)
	}

	d.checker.CheckCommentMisspellings(pl, pl.Doc.Raw)
	phpDocType := d.parsePHPDocVar(pl.Doc)
	d.checker.CheckPHPDocVar(pl, pl.Doc, phpDocType)

	typeHintType, ok := d.parseTypeHintNode(pl.Type)
	if ok && !types.TypeHintHasMoreAccurateType(typeHintType, phpDocType) {
		d.Report(pl, LevelNotice, "typeHint", "Specify the type for the property in PHPDoc, 'array' type hint too generic")
	}
	d.checker.CheckTypeHintNode(pl.Type, "property type")

	for _, pNode := range pl.Properties {
		prop := pNode.(*ir.PropertyStmt)

		nm := prop.Variable.Name

		// We need to clone the types, because otherwise, if several
		// properties are written in one definition, and null was
		// assigned to the first, then all properties become nullable.
		propTypes := phpDocType.Clone().Append(typeHintType)

		d.checker.CheckAssignNullToNotNullableProperty(prop, propTypes)

		if prop.Expr != nil {
			propTypes = propTypes.Append(solver.ExprTypeLocal(d.scope(), d.ctx.st, prop.Expr))
		}

		if isStatic {
			nm = "$" + nm
		}

		// TODO: handle duplicate property
		cl.Properties[nm] = meta.PropertyInfo{
			Pos:         d.getElementPos(prop),
			Typ:         propTypes.Immutable(),
			AccessLevel: accessLevel,
		}
	}

	return true
}

func (d *rootWalker) enterClassConstList(list *ir.ClassConstListStmt) bool {
	cl := d.getClass()
	accessLevel := meta.Public

	for _, m := range list.Modifiers {
		d.checker.CheckModifierKeywordCase(m)
		switch strings.ToLower(m.Value) {
		case "public":
			accessLevel = meta.Public
		case "protected":
			accessLevel = meta.Protected
		case "private":
			accessLevel = meta.Private
		}
	}

	for _, cNode := range list.Consts {
		c := cNode.(*ir.ConstantStmt)

		nm := c.ConstantName.Value
		d.checker.CheckCommentMisspellings(c, list.Doc.Raw)
		typ := solver.ExprTypeLocal(d.scope(), d.ctx.st, c.Expr)

		value := constfold.Eval(d.ctx.st, c.Expr)

		// TODO: handle duplicate constant
		cl.Constants[nm] = meta.ConstInfo{
			Pos:         d.getElementPos(c),
			Typ:         typ.Immutable(),
			AccessLevel: accessLevel,
			Value:       value,
		}
	}

	return true
}

func (d *rootWalker) enterClassMethod(meth *ir.ClassMethodStmt) bool {
	nm := meth.MethodName.Value
	_, insideInterface := d.currentClassNodeStack.Current().(*ir.InterfaceStmt)

	d.checker.CheckOldStyleConstructor(meth)

	pos := ir.GetPosition(meth)

	if funcSize := pos.EndLine - pos.StartLine; funcSize > maxFunctionLines {
		d.Report(meth.MethodName, LevelNotice, "complexity", "Too big method: more than %d lines", maxFunctionLines)
	}

	class := d.getClass()

	modif := d.parseMethodModifiers(meth)

	if modif.accessImplicit {
		methodFQN := class.Name + "::" + nm
		d.Report(meth.MethodName, LevelNotice, "implicitModifiers", "Specify the access modifier for %s method explicitly", methodFQN)
	}

	d.checker.CheckMagicMethod(meth.MethodName, nm, modif, len(meth.Params))

	sc := meta.NewScope()
	if !modif.static {
		sc.AddVarName("this", types.NewMap(d.ctx.st.CurrentClass).Immutable(), "instance method", meta.VarAlwaysDefined)
		sc.SetInInstanceMethod(true)
	}

	if meth.Doc.Raw == "" && modif.accessLevel == meta.Public {
		// Permit having "__call" and other magic method without comments.
		if !insideInterface && !strings.HasPrefix(nm, "_") {
			methodFQN := class.Name + "::" + nm
			d.Report(meth.MethodName, LevelNotice, "missingPhpdoc", "Missing PHPDoc for %s public method", methodFQN)
		}
	}
	d.checker.CheckCommentMisspellings(meth.MethodName, meth.Doc.Raw)
	d.checker.CheckIdentMisspellings(meth.MethodName)

	// Indexing stage.
	doc := phpdoctypes.Parse(meth.Doc, meth.Params, d.ctx.typeNormalizer)
	moveShapesToContext(&d.ctx, doc.Shapes)
	d.handleClosuresFromDoc(doc.Closures)

	// Check stage.
	errors := d.checker.CheckPHPDoc(meth, meth.Doc, meth.Params)
	d.reportPHPDocErrors(errors)

	phpDocReturnType := doc.ReturnType
	phpDocParamTypes := doc.ParamTypes

	returnTypeHint, ok := d.parseTypeHintNode(meth.ReturnType)
	if ok && !doc.Inherit {
		d.checker.CheckFuncReturnType(meth.MethodName, meth.MethodName.Value, returnTypeHint, phpDocReturnType)
	}
	d.checker.CheckTypeHintNode(meth.ReturnType, "return type")

	funcParams := d.parseFuncParams(meth.Params, phpDocParamTypes, sc, nil)

	d.checker.CheckFuncParams(meth.MethodName, meth.Params, funcParams, phpDocParamTypes)

	if len(class.Interfaces) != 0 {
		// If we implement interfaces, methods that take a part in this
		// can borrow types information from them.
		// Programmers sometimes leave implementing methods without a
		// comment or use @inheritdoc there.
		//
		// If method params are properly documented, it's possible to
		// derive that information, but we need to know in which
		// interface we can find that method.
		//
		// Since we don't have all interfaces during the indexing phase
		// and shouldn't update meta after it, we defer type resolving by
		// using BaseMethodParam here. We would have to lookup
		// matching interface during the type resolving.

		// Find params without type and annotate them with special
		// type that will force solver to walk interface types that
		// current class implements to have a chance of finding relevant type info.
		for i, p := range funcParams.params {
			if !p.Typ.Empty() {
				continue // Already has a type
			}

			if i > math.MaxUint8 {
				break // Current implementation limit reached
			}

			res := make(map[string]struct{})
			res[types.WrapBaseMethodParam(i, d.ctx.st.CurrentClass, nm)] = struct{}{}
			funcParams.params[i].Typ = types.NewMapFromMap(res)
			sc.AddVarName(p.Name, funcParams.params[i].Typ, "param", meta.VarAlwaysDefined)
		}
	}

	var stmts []ir.Node
	if stmtList, ok := meth.Stmt.(*ir.StmtList); ok {
		stmts = stmtList.Stmts
	}
	funcInfo := d.handleFuncStmts(funcParams.params, nil, stmts, sc)
	actualReturnTypes := funcInfo.returnTypes
	exitFlags := funcInfo.prematureExitFlags
	if nm == `__construct` {
		d.checker.CheckParentConstructorCall(meth.MethodName, funcInfo.callsParentConstructor)
	}

	returnTypes := functionReturnType(phpDocReturnType, returnTypeHint, actualReturnTypes)

	// TODO: handle duplicate method
	var funcFlags meta.FuncFlags
	if modif.static {
		funcFlags |= meta.FuncStatic
	}
	if modif.abstract {
		funcFlags |= meta.FuncAbstract
	}
	if modif.final {
		funcFlags |= meta.FuncFinal
	}
	if !insideInterface && !modif.abstract && solver.SideEffectFreeFunc(d.scope(), d.ctx.st, nil, stmts) {
		funcFlags |= meta.FuncPure
	}
	class.Methods.Set(nm, meta.FuncInfo{
		Params:       funcParams.params,
		Name:         nm,
		Pos:          d.getElementPos(meth),
		Typ:          returnTypes.Immutable(),
		MinParamsCnt: funcParams.minParamsCount,
		AccessLevel:  modif.accessLevel,
		Flags:        funcFlags,
		ExitFlags:    exitFlags,
		Doc:          doc.AdditionalInfo,
	})

	if nm == "getIterator" && d.metaInfo().IsIndexingComplete() && solver.Implements(d.metaInfo(), d.ctx.st.CurrentClass, `\IteratorAggregate`) {
		implementsTraversable := returnTypes.Find(func(typ string) bool {
			return solver.Implements(d.metaInfo(), typ, `\Traversable`)
		})

		if !implementsTraversable {
			d.Report(meth.MethodName, LevelError, "stdInterface", "Objects returned by %s::getIterator() must be traversable or implement interface \\Iterator", d.ctx.st.CurrentClass)
		}
	}

	return false
}

type methodModifiers struct {
	abstract       bool
	static         bool
	accessLevel    meta.AccessLevel
	final          bool
	accessImplicit bool
}

func (d *rootWalker) parseMethodModifiers(meth *ir.ClassMethodStmt) (res methodModifiers) {
	res.accessLevel = meta.Public
	res.accessImplicit = true

	for _, m := range meth.Modifiers {
		d.checker.CheckModifierKeywordCase(m)
		switch strings.ToLower(m.Value) {
		case "abstract":
			res.abstract = true
		case "static":
			res.static = true
		case "public":
			res.accessLevel = meta.Public
			res.accessImplicit = false
		case "private":
			res.accessLevel = meta.Private
			res.accessImplicit = false
		case "protected":
			res.accessLevel = meta.Protected
			res.accessImplicit = false
		case "final":
			res.final = true
		default:
			linterError(d.ctx.st.CurrentFile, "Unrecognized method modifier: %s", m.Value)
		}
	}

	return res
}

// Handle const list part.

func (d *rootWalker) enterConstList(lst *ir.ConstListStmt) bool {
	if d.meta.Constants == nil {
		d.meta.Constants = make(meta.ConstantsMap)
	}

	for _, sNode := range lst.Consts {
		s := sNode.(*ir.ConstantStmt)

		value := constfold.Eval(d.ctx.st, s.Expr)

		id := s.ConstantName
		nm := d.ctx.st.Namespace + `\` + id.Value

		d.meta.Constants[nm] = meta.ConstInfo{
			Pos:   d.getElementPos(s),
			Typ:   solver.ExprTypeLocal(d.scope(), d.ctx.st, s.Expr),
			Value: value,
		}
	}

	return false
}

// Utils part.

func (d *rootWalker) handleClosuresFromDoc(closures types.ClosureMap) {
	if d.meta.Functions.H == nil {
		d.meta.Functions = meta.NewFunctionsMap()
	}

	for name, closureInfo := range closures {
		var params []meta.FuncParam
		for i, paramType := range closureInfo.ParamTypes {
			params = append(params, meta.FuncParam{
				Name: fmt.Sprintf("closure param #%d", i),
				Typ:  types.NewMapWithNormalization(d.ctx.typeNormalizer, paramType),
			})
		}

		d.meta.Functions.Set(name, meta.FuncInfo{
			Params:       params,
			Name:         name,
			Typ:          types.NewMapWithNormalization(d.ctx.typeNormalizer, closureInfo.ReturnType),
			MinParamsCnt: len(closureInfo.ParamTypes),
		})
	}
}

// parseTypeHintNode parse type info, e.g. "string" in "someFunc() : string { ... }".
func (d *rootWalker) parseTypeHintNode(n ir.Node) (typ types.Map, ok bool) {
	if n == nil {
		return types.Map{}, false
	}

	typesMap := types.NormalizedTypeHintTypes(d.ctx.typeNormalizer, n)

	return typesMap, !typesMap.Empty()
}

func (d *rootWalker) reportPHPDocErrors(errs PHPDocErrors) {
	for _, err := range errs.types {
		d.ReportPHPDoc(err.Location, LevelNotice, "invalidDocblockType", err.Message)
	}
	for _, err := range errs.lint {
		d.ReportPHPDoc(err.Location, LevelWarning, "invalidDocblock", err.Message)
	}
}

func (d *rootWalker) addQuickFix(checkName string, fix quickfix.TextEdit) {
	if !d.config.ApplyQuickFixes {
		return
	}

	if !d.checkersFilter.IsEnabledReport(checkName, d.ctx.st.CurrentFile) {
		return
	}

	d.ctx.fixes = append(d.ctx.fixes, fix)
}

func (d *rootWalker) currentFunction() (meta.FuncInfo, bool) {
	name := d.ctx.st.CurrentFunction
	if name == "" {
		return meta.FuncInfo{}, false
	}

	if d.ctx.st.CurrentClass != "" {
		className, ok := solver.GetClassName(d.ctx.st, &ir.Name{Value: d.ctx.st.CurrentClass})
		if !ok {
			return meta.FuncInfo{}, false
		}

		method, ok := solver.FindMethod(d.ctx.st.Info, className, name)
		if !ok {
			return meta.FuncInfo{}, false
		}

		return method.Info, true
	}

	funcName, ok := solver.GetFuncName(d.ctx.st, &ir.Name{Value: name})
	if !ok {
		return meta.FuncInfo{}, false
	}

	fun, ok := d.ctx.st.Info.GetFunction(funcName)
	if !ok {
		return meta.FuncInfo{}, false
	}

	return fun, true
}

func (d *rootWalker) getClass() meta.ClassInfo {
	var m meta.ClassesMap

	if d.ctx.st.IsTrait {
		if d.meta.Traits.H == nil {
			d.meta.Traits = meta.NewClassesMap()
		}
		m = d.meta.Traits
	} else {
		if d.meta.Classes.H == nil {
			d.meta.Classes = meta.NewClassesMap()
		}
		m = d.meta.Classes
	}

	cl, ok := m.Get(d.ctx.st.CurrentClass)
	if !ok {
		var flags meta.ClassFlags
		if d.ctx.st.IsInterface {
			flags = meta.ClassInterface
		}

		cl = meta.ClassInfo{
			Pos:              d.getElementPos(d.currentClassNodeStack.Current()),
			Name:             d.ctx.st.CurrentClass,
			Flags:            flags,
			Parent:           d.ctx.st.CurrentParentClass,
			ParentInterfaces: d.ctx.st.CurrentParentInterfaces,
			Interfaces:       make(map[string]struct{}),
			Traits:           make(map[string]struct{}),
			Methods:          meta.NewFunctionsMap(),
			Properties:       make(meta.PropertiesMap),
			Constants:        make(meta.ConstantsMap),
		}

		m.Set(d.ctx.st.CurrentClass, cl)
	}

	return cl
}

func (d *rootWalker) parseStartPos(pos *position.Position) (startLn []byte, startChar int) {
	if pos.StartLine >= 1 && d.file.NumLines() > pos.StartLine {
		startLn = d.file.Line(pos.StartLine - 1)
		p := d.file.LinePosition(pos.StartLine - 1)
		if pos.StartPos > p {
			startChar = pos.StartPos - p
		}
	}

	return startLn, startChar
}

func (d *rootWalker) getElementPos(n ir.Node) meta.ElementPosition {
	pos := ir.GetPosition(n)
	_, startChar := d.parseStartPos(pos)

	return meta.ElementPosition{
		Filename:  d.ctx.st.CurrentFile,
		Character: int32(startChar),
		Line:      int32(pos.StartLine),
		EndLine:   int32(pos.EndLine),
		Length:    int32(pos.EndPos - pos.StartPos),
	}
}

// nodeText is used to get the string that represents the
// passed node more efficiently than irutil.FmtNode.
func (d *rootWalker) nodeText(n ir.Node) string {
	pos := ir.GetPosition(n)
	from, to := pos.StartPos, pos.EndPos
	src := d.file.Contents()
	// Taking a node from the source code preserves the original formatting
	// and is more efficient than printing it.
	if (from >= 0 && from < len(src)) && (to >= 0 && to < len(src)) {
		return string(src[from:to])
	}
	// If we can't take node out of the source text, print it.
	return irutil.FmtNode(n)
}

// Report part.

// ReportPHPDoc registers a single report message about some found problem in PHPDoc.
func (d *rootWalker) ReportPHPDoc(phpDocLocation PHPDocLocation, level int, checkName, msg string, args ...interface{}) {
	if phpDocLocation.RelativeLine {
		doc, ok := irutil.FindPHPDoc(phpDocLocation.Node, true)
		if !ok {
			// If PHPDoc for some reason was not found, give a warning to the node.
			d.Report(phpDocLocation.Node, level, checkName, msg, args...)
			return
		}

		countPHPDocLines := strings.Count(doc, "\n") + 1

		nodePos := ir.GetPosition(phpDocLocation.Node)
		if nodePos == nil {
			// If position for some reason was not found, give a warning to the node.
			d.Report(phpDocLocation.Node, level, checkName, msg, args...)
			return
		}

		// 1| <?php
		// 2|
		// 3| /**
		// 4|  * Comment
		// 5|  * @param int $a    <- phpDocLocation.Line == 3
		// 6|  */                 <- countPHPDocLines == 4
		// 7| function f($a) {}   <- nodePos.StartLine == 7
		//
		// countPHPDocLines - phpDocLocation.Line = 1
		// nodePos.StartLine - 1 = 6
		// 6 - 1 = 5 (number of the required line relative to one)
		// 5 - 1 = 4 (number of the required line relative to zero)
		phpDocLocation.Line = nodePos.StartLine - (countPHPDocLines - phpDocLocation.Line) - 1 - 1
	}

	if phpDocLocation.Line < 0 || phpDocLocation.Line >= d.file.NumLines() {
		d.Report(phpDocLocation.Node, level, checkName, msg, args...)
		return
	}

	contextLine := d.file.Line(phpDocLocation.Line)

	lineWithoutBeginning := contextLine
	// For the case when we give a warning about the wrong start
	// of PHPDoc (/* instead /**), it is not necessary to delete characters.
	if !bytes.Contains(contextLine, []byte("/*")) || bytes.Contains(contextLine, []byte("/**")) {
		lineWithoutBeginning = bytes.TrimLeft(contextLine, "/ *")
	}

	shiftFromStart := len(contextLine) - len(lineWithoutBeginning)

	parts := bytes.Fields(lineWithoutBeginning)
	if phpDocLocation.Field >= len(parts) {
		phpDocLocation.Field = 0
		phpDocLocation.WholeLine = true
	}

	var startChar int
	var endChar int

	if phpDocLocation.WholeLine {
		startChar = shiftFromStart
		endChar = len(contextLine)
	} else {
		part := parts[phpDocLocation.Field]
		shiftStart := bytes.Index(lineWithoutBeginning, part)
		shiftEnd := shiftStart + len(part)

		startChar = shiftFromStart + shiftStart
		endChar = shiftFromStart + shiftEnd
	}

	if endChar == len(contextLine) && bytes.HasSuffix(contextLine, []byte("\r")) {
		endChar--
	}

	loc := ir.Location{
		StartLine: phpDocLocation.Line,
		EndLine:   phpDocLocation.Line,
		StartChar: startChar,
		EndChar:   endChar,
	}

	d.ReportLocation(loc, level, checkName, msg, args...)
}

func (d *rootWalker) Report(n ir.Node, level int, checkName, msg string, args ...interface{}) {
	var pos position.Position

	if n == nil {
		// Hack to parse syntax error message from php-parser.
		if strings.Contains(msg, "syntax error") && strings.Contains(msg, " at line ") {
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
		nodePos := ir.GetPosition(n)
		if nodePos == nil {
			return
		}
		pos = *nodePos
	}

	var loc ir.Location

	loc.StartLine = pos.StartLine - 1
	loc.EndLine = pos.EndLine - 1
	loc.StartChar = pos.StartPos
	loc.EndChar = pos.EndPos

	if pos.StartLine >= 1 && d.file.NumLines() >= pos.StartLine {
		p := d.file.LinePosition(pos.StartLine - 1)
		if pos.StartPos >= p {
			loc.StartChar = pos.StartPos - p
		}
	}

	if pos.EndLine >= 1 && d.file.NumLines() >= pos.EndLine {
		p := d.file.LinePosition(pos.EndLine - 1)
		if pos.EndPos >= p {
			loc.EndChar = pos.EndPos - p
		}
	}

	d.ReportLocation(loc, level, checkName, msg, args...)
}

func (d *rootWalker) ReportLocation(loc ir.Location, level int, checkName, msg string, args ...interface{}) {
	if !d.metaInfo().IsIndexingComplete() {
		return
	}
	if d.file.AutoGenerated() && !d.config.CheckAutoGenerated {
		return
	}
	// We don't report anything if linter was disabled by a
	// successful @linter disable, unless it's the linterError.
	if d.linterDisabled && checkName != "linterError" {
		return
	}

	if !d.checkersFilter.IsEnabledReport(checkName, d.ctx.st.CurrentFile) {
		return
	}

	if loc.StartLine < 0 || loc.StartLine >= d.file.NumLines() {
		return
	}

	contextLine := d.file.Line(loc.StartLine)

	var hash uint64
	// If baseline is not enabled, don't waste time on hash computations.
	if d.config.ComputeBaselineHashes {
		hash = d.reportHash(&loc, contextLine, checkName, msg)
		if count := d.ctx.baseline.Count(hash); count >= 1 {
			if d.ctx.hashCounters == nil {
				d.ctx.hashCounters = make(map[uint64]int)
			}
			d.ctx.hashCounters[hash]++
			if d.ctx.hashCounters[hash] <= count {
				return
			}
		}
	}

	d.reports = append(d.reports, &Report{
		CheckName: checkName,
		Context:   string(contextLine),
		StartChar: loc.StartChar,
		EndChar:   loc.EndChar,
		Line:      loc.StartLine + 1,
		Level:     level,
		Filename:  strings.ReplaceAll(d.ctx.st.CurrentFile, "\\", "/"), // To make output stable between platforms, see #572
		Message:   fmt.Sprintf(msg, args...),
		Hash:      hash,
	})
}

// reportHash computes the ReportLocation signature hash for the baseline.
func (d *rootWalker) reportHash(loc *ir.Location, contextLine []byte, checkName, msg string) uint64 {
	// Since we store class::method scope, renaming a class would cause baseline
	// invalidation for the entire class. But this is not an issue, since in
	// a modern PHP class name always should map to a filename.
	// If we renamed a class, we probably renamed the file as well, so
	// the baseline would be invalidated anyway.
	//
	// We still get all the benefits by using method prefix: it's common
	// for different classes to have methods with similar name. We don't
	// want such methods to be considered as a single "scope".
	scope := "file"
	switch {
	case d.ctx.st.CurrentClass != "" && d.ctx.st.CurrentFunction != "":
		scope = d.ctx.st.CurrentClass + "::" + d.ctx.st.CurrentFunction
	case d.ctx.st.CurrentFunction != "":
		scope = d.ctx.st.CurrentFunction
	}

	var prevLine []byte
	var nextLine []byte
	// Adjacent lines are only included when using non-conservative baseline.
	if !d.config.ConservativeBaseline {
		// Lines are 1-based, indexes are 0-based.
		// If this function is called, we expect that lines[index] exists.
		index := loc.StartLine - 1
		if index >= 1 {
			prevLine = d.file.Line(index - 1)
		}
		if d.file.NumLines() > index+1 {
			nextLine = d.file.Line(index + 1)
		}
	}

	// Observation: using a base file name instead of its full name does not
	// introduce any "bad collisions", because we have a "scope" part
	// that cuts the risk by a big margin.
	//
	// One benefit is that it makes the baseline contents more position-independent.
	// We don't need to know a source root folder to truncate it.
	//
	// Moves like Foo/Bar/User.php -> Common/User.php do not invalidate the
	// suppress base. This is not an obvious win, but it may be a good
	// compromise to solve a use case where a class file is being moved.
	// For classes, filename is unlikely to be changed during this file move operation.
	//
	// It can't handle file duplication when Foo/Bar/User.php is *copied* to a new location.
	// It may be useful to give warnings to both *old* and *new* copies of the duplicated
	// code since some bugs should be fixed in both places.
	// We'll see how it goes.
	filename := filepath.Base(d.ctx.st.CurrentFile)

	d.ctx.scratchBuf.Reset()
	return baseline.ReportHash(&d.ctx.scratchBuf, baseline.HashFields{
		Filename:  filename,
		PrevLine:  bytes.TrimSuffix(prevLine, []byte("\r")),
		StartLine: bytes.TrimSuffix(contextLine, []byte("\r")),
		NextLine:  bytes.TrimSuffix(nextLine, []byte("\r")),
		CheckName: checkName,
		Message:   msg,
		Scope:     scope,
	})
}

func (d *rootWalker) reportUndefinedVariable(v ir.Node, maybeHave bool, path irutil.NodePath) {
	sv, ok := v.(*ir.SimpleVar)
	if !ok {
		d.Report(v, LevelWarning, "undefinedVariable", "Unknown variable variable %s used",
			utils.NameNodeToString(v))
		return
	}

	if _, ok := superGlobals[sv.Name]; ok {
		return
	}

	// For the following cases, we do not give warnings,
	// as they check for the presence of this variable.
	// $b = $a ?? 100;
	// $b = isset($a) ? $a : 100;
	needWarn := !utils.InCoalesceOrIsset(path)
	if !needWarn {
		return
	}

	if maybeHave {
		d.Report(sv, LevelWarning, "maybeUndefined", "Possibly undefined variable $%s", sv.Name)
	} else {
		d.Report(sv, LevelError, "undefinedVariable", "Cannot find referenced variable $%s", sv.Name)
	}
}

// Rules part.

func (d *rootWalker) runRules(n ir.Node, sc *meta.Scope, rlist []rules.Rule) {
	for i := range rlist {
		rule := &rlist[i]
		if d.runRule(n, sc, rule) {
			// Stop at the first matched rule per IR node.
			// Sometimes it's useful to report more, but we rely on the rules definition
			// order so we can report more specific issues instead of the
			// more generic ones whether possible.
			// This also makes rules execution faster.
			break
		}
	}
}

func (d *rootWalker) renderRuleMessage(msg string, n ir.Node, m phpgrep.MatchData, truncate bool) string {
	// "$$" stands for the entire matched node, like $0 in regexp.
	if strings.Contains(msg, "$$") {
		msg = strings.ReplaceAll(msg, "$$", d.nodeText(n))
	}

	if len(m.Capture) == 0 {
		return msg // No variables to interpolate, we're done
	}
	for _, c := range m.Capture {
		key := "$" + c.Name
		if !strings.Contains(msg, key) {
			continue
		}
		nodeString := d.nodeText(c.Node)
		// Don't interpolate strings that are too long
		// or contain a newline.
		var replacement string
		if truncate && (len(nodeString) > 60 || strings.Contains(nodeString, "\n")) {
			replacement = key
		} else {
			replacement = nodeString
		}
		msg = strings.ReplaceAll(msg, key, replacement)
	}
	return msg
}

func (d *rootWalker) runRule(n ir.Node, sc *meta.Scope, rule *rules.Rule) bool {
	m, ok := rule.Matcher.Match(n)
	if !ok {
		return false
	}

	matched := false
	if len(rule.Filters) == 0 {
		matched = true
	} else {
		for _, filterSet := range rule.Filters {
			if d.checkFilterSet(&m, sc, filterSet) {
				matched = true
				break
			}
		}
	}

	// If location is explicitly set, use named match set.
	// Otherwise peek the root target node.
	var location ir.Node
	switch {
	case matched && rule.Location != "":
		named, _ := m.CapturedByName(rule.Location)
		location = named
	case matched:
		location = n
	}

	if location == nil {
		return false
	}

	message := d.renderRuleMessage(rule.Message, n, m, true)
	d.Report(location, rule.Level, rule.Name, "%s", message)

	if d.config.ApplyQuickFixes && rule.Fix != "" {
		// As rule sets contain only enabled rules,
		// we should be OK without any filtering here.
		pos := ir.GetPosition(n)
		d.addQuickFix(rule.Name, quickfix.TextEdit{
			StartPos:    pos.StartPos,
			EndPos:      pos.EndPos,
			Replacement: d.renderRuleMessage(rule.Fix, n, m, false),
		})
	}

	return true
}

func (d *rootWalker) checkTypeFilter(wantType *phpdoc.Type, sc *meta.Scope, nn ir.Node) bool {
	if wantType == nil {
		return true
	}

	// TODO: compare without converting a TypesMap into TypeExpr?
	// Or maybe store TypeExpr inside a TypesMap instead of strings?
	// Can we use `meta.Type` for this?
	typ := solver.ExprType(sc, d.ctx.st, nn)
	haveType := typesMapToTypeExpr(d.ctx.phpdocTypeParser, typ)
	return rules.TypeIsCompatible(wantType.Expr, haveType.Expr)
}

func (d *rootWalker) checkFilterSet(m *phpgrep.MatchData, sc *meta.Scope, filterSet map[string]rules.Filter) bool {
	// TODO: pass custom types here, so both @type and @pure predicates can use it.

	for name, filter := range filterSet {
		nn, ok := m.CapturedByName(name)
		if !ok {
			continue
		}

		if !d.checkTypeFilter(filter.Type, sc, nn) {
			return false
		}
		if filter.Pure && !solver.SideEffectFree(d.scope(), d.ctx.st, nil, nn) {
			return false
		}
	}

	return true
}
