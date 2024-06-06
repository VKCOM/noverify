package linter

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/phpdoctypes"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/utils"
	"github.com/VKCOM/noverify/src/workspace"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/client9/misspell"
)

type rootChecker struct {
	walker *rootWalker

	file       *workspace.File
	normalizer types.Normalizer
	info       *meta.Info

	scope  *meta.Scope
	state  *meta.ClassParseState
	parser *phpdoc.TypeParser

	currentClassNodeStack *irutil.NodePath

	// TypoFixer is a rule set for English typos correction.
	// If nil, no misspell checking is performed.
	// See github.com/client9/misspell for details.
	typoFixer *misspell.Replacer

	quickfix *QuickFixGenerator
}

func newRootChecker(walker *rootWalker, quickfix *QuickFixGenerator) *rootChecker {
	c := &rootChecker{
		file:                  walker.file,
		walker:                walker,
		normalizer:            walker.ctx.typeNormalizer,
		info:                  walker.metaInfo(),
		state:                 walker.ctx.st,
		scope:                 walker.scope(),
		parser:                walker.ctx.phpdocTypeParser,
		currentClassNodeStack: &walker.currentClassNodeStack,
		quickfix:              quickfix,
	}
	if walker.config != nil {
		c.typoFixer = walker.config.TypoFixer
	}
	return c
}

func (r *rootChecker) CheckFunctionTypeHint(fun *ir.FunctionStmt) {
	for _, comment := range fun.Doc.Parsed {
		var typeContainer, ok = comment.(*phpdoc.TypeVarCommentPart)
		if !ok {
			continue
		}

		if typeContainer.Name() != "param" {
			continue
		}

		var typeParam = typeContainer.Type.Source

		for _, param := range fun.Params {
			var typedParam, ok = param.(*ir.Parameter)
			if ok {
				var variable = typedParam.Variable

				// maybe we don`t need it and order the same as param
				if variable.Name != typeContainer.Var[1:] {
					continue
				}

				var paramType, ok = typedParam.VariableType.(*ir.Name)
				if paramType != nil && ok {
					// TODO: quickFix -> remove @param from typeHint
					break
				}

				converted := phpdoctypes.ToRealType(r.normalizer.ClassFQNProvider(), r.normalizer.KPHP(), typeContainer.Type)
				if cap(converted.Types) > 1 {
					continue
				}

				if !types.IsTrivial(converted.Types[0].Elem) && !types.IsClass(converted.Types[0].Elem) {
					continue
				}

				// TODO: quickFix -> remove @param from typeHint
				var varDollar = typeContainer.Var
				var variableWithType = typeParam + " " + varDollar
				r.walker.Report(variable, LevelWarning, "implicitParamType", "Type for %s can be wrote explicitly from typeHint", varDollar)
				r.walker.addQuickFix("implicitParamType", r.quickfix.FunctionParamTypeReplacementFromTypeHint(variable, variableWithType))
			}
		}
	}
}

func (r *rootChecker) CheckFunction(fun *ir.FunctionStmt) bool {
	r.CheckKeywordCase(fun, "function")

	sc := meta.NewScope()
	pos := ir.GetPosition(fun)

	if funcSize := pos.EndLine - pos.StartLine; funcSize > maxFunctionLines {
		r.walker.Report(fun.FunctionName, LevelNotice, "complexity", "Too big function: more than %d lines", maxFunctionLines)
	}

	r.CheckCommentMisspellings(fun.FunctionName, fun.Doc.Raw)
	r.CheckIdentMisspellings(fun.FunctionName)

	// Check stage.
	errors := r.CheckPHPDoc(fun, fun.Doc, fun.Params)
	r.reportPHPDocErrors(errors)

	doc := phpdoctypes.Parse(fun.Doc, fun.Params, r.normalizer)
	phpDocReturnType := doc.ReturnType
	phpDocParamTypes := doc.ParamTypes

	returnTypeHint, ok := r.walker.parseTypeHintNode(fun.ReturnType)
	if ok && !doc.Inherit {
		r.CheckFuncReturnType(fun.FunctionName, fun.FunctionName.Value, returnTypeHint, phpDocReturnType)
	}
	r.CheckTypeHintNode(fun.ReturnType, "return type")

	funcParams := r.walker.parseFuncParams(fun.Params, phpDocParamTypes, sc, nil)
	r.CheckFuncParams(fun.FunctionName, fun.Params, funcParams, phpDocParamTypes)

	r.walker.handleFuncStmts(funcParams.params, nil, fun.Stmts, sc)

	r.CheckFunctionTypeHint(fun)
	return false
}

func (r *rootChecker) CheckPropertyList(pl *ir.PropertyListStmt) bool {
	accessImplicit := true

	for _, m := range pl.Modifiers {
		r.CheckModifierKeywordCase(m)

		switch strings.ToLower(m.Value) {
		case "public", "protected", "private":
			accessImplicit = false
		}
	}

	if accessImplicit {
		target := "property"
		if len(pl.Properties) > 1 {
			target = "properties"
		}
		r.walker.Report(pl, LevelNotice, "implicitModifiers", "Specify the access modifier for %s explicitly", target)
	}

	docblockType := r.walker.parsePHPDocVar(pl.Doc)

	r.CheckCommentMisspellings(pl, pl.Doc.Raw)
	r.CheckPHPDocVar(pl, pl.Doc, docblockType)

	typeHintType, ok := r.walker.parseTypeHintNode(pl.Type)
	if ok && !types.TypeHintHasMoreAccurateType(typeHintType, docblockType) {
		r.walker.Report(pl, LevelNotice, "typeHint", "Specify the type for the property in PHPDoc, 'array' type hint too generic")
	}

	r.CheckTypeHintNode(pl.Type, "property type")

	for _, p := range pl.Properties {
		prop := p.(*ir.PropertyStmt)

		// We need to clone the types, because otherwise, if several
		// properties are written in one definition, and null was
		// assigned to the first, then all properties become nullable.
		propTypes := docblockType.Clone().Append(typeHintType)

		r.CheckAssignNullToNotNullableProperty(prop, propTypes)
	}

	return true
}

func (r *rootChecker) reportPHPDocErrors(errs PHPDocErrors) {
	for _, err := range errs.types {
		r.walker.ReportPHPDoc(err.Location, LevelNotice, "invalidDocblockType", err.Message)
	}
	for _, err := range errs.lint {
		r.walker.ReportPHPDoc(err.Location, LevelWarning, "invalidDocblock", err.Message)
	}
}

func (r *rootChecker) CheckPHPDoc(n ir.Node, doc phpdoc.Comment, actualParams []ir.Node) (errors PHPDocErrors) {
	if doc.Raw == "" {
		return errors
	}

	if phpdoc.IsSuspicious([]byte(doc.Raw)) {
		errors.pushLint(
			PHPDocLine(n, 1),
			"Multiline PHPDoc comment should start with /**, not /*",
		)
	}

	actualParamNames := make(map[string]struct{}, len(actualParams))
	for _, p := range actualParams {
		p := p.(*ir.Parameter)
		actualParamNames[p.Variable.Name] = struct{}{}
	}

	var curParam int

	for _, rawPart := range doc.Parsed {
		r.checkPHPDocRef(n, rawPart)

		if rawPart.Name() == "return" {
			part := rawPart.(*phpdoc.TypeCommentPart)

			converted := phpdoctypes.ToRealType(r.normalizer.ClassFQNProvider(), r.normalizer.KPHP(), part.Type)

			if converted.Warning != "" {
				errors.pushType(
					PHPDocLineField(n, part.Line(), 1),
					converted.Warning,
				)
			}

			returnType := types.NewMapWithNormalization(r.normalizer, converted.Types)

			if returnType.Contains("void") && returnType.Len() > 1 {
				errors.pushType(
					PHPDocLineField(n, part.Line(), 1),
					"Void type can only be used as a standalone type for the return type",
				)
			}

			r.checkUndefinedClassesInPHPDoc(n, returnType, part)
			continue
		}

		// Rest is for @param handling.

		if rawPart.Name() != "param" {
			continue
		}

		part := rawPart.(*phpdoc.TypeVarCommentPart)
		switch {
		case part.Var == "":
			errors.pushLint(
				PHPDocLineField(n, part.Line(), 1),
				"Malformed @param tag (maybe var is missing?)",
			)

		case part.Type.IsEmpty():
			errors.pushLint(
				PHPDocLineField(n, part.Line(), 1),
				"Malformed @param %s tag (maybe type is missing?)", part.Var,
			)

			continue
		}

		if part.VarIsFirst {
			// Phpstorm gives the same message.
			errors.pushLint(
				PHPDocLine(n, part.Line()),
				"Non-canonical order of variable and type",
			)
		}

		variable := part.Var
		if !strings.HasPrefix(variable, "$") {
			if len(actualParams) > curParam {
				variable = actualParams[curParam].(*ir.Parameter).Variable.Name
			}
		}
		if _, ok := actualParamNames[strings.TrimPrefix(variable, "$")]; !ok {
			errors.pushLint(
				PHPDocLineField(n, part.Line(), 2),
				"@param for non-existing argument %s", variable,
			)
			continue
		}

		curParam++

		converted := phpdoctypes.ToRealType(r.normalizer.ClassFQNProvider(), r.normalizer.KPHP(), part.Type)

		if converted.Warning != "" {
			errors.pushType(
				PHPDocLineField(n, part.Line(), 1),
				converted.Warning,
			)
		}

		var param phpdoctypes.Param
		param.Typ = types.NewMapWithNormalization(r.normalizer, converted.Types)

		if param.Typ.Contains("void") {
			errors.pushType(
				PHPDocLineField(n, part.Line(), 1),
				"Void type can only be used as a standalone type for the return type",
			)
		}

		r.checkUndefinedClassesInPHPDoc(n, param.Typ, part)
	}

	return errors
}

func (r *rootChecker) checkPHPDocRef(n ir.Node, part phpdoc.CommentPart) {
	if !r.info.IsIndexingComplete() {
		return
	}

	switch part.Name() {
	case "mixin":
		r.checkPHPDocMixinRef(n, part)
	case "see":
		r.checkPHPDocSeeRef(n, part)
	}
}

func (r *rootChecker) checkPHPDocSeeRef(n ir.Node, part phpdoc.CommentPart) {
	params := part.(*phpdoc.RawCommentPart).Params
	if len(params) == 0 {
		return
	}

	// @see supports a comma-separated list of refs.
	var refs []string
	for _, p := range params {
		refs = append(refs, strings.TrimSuffix(p, ","))
		if !strings.HasSuffix(p, ",") {
			break
		}
	}

	for _, ref := range refs {
		// Sometimes people write references like `foo()` `foo...` `foo@`.
		ref = strings.TrimRight(ref, "().;@")
		if !r.isValidPHPDocRef(ref) {
			r.walker.ReportPHPDoc(
				PHPDocLineField(n, part.Line(), 1),
				LevelWarning, "invalidDocblockRef", "@see tag refers to unknown symbol %s", ref,
			)
		}
	}
}

func (r *rootChecker) checkPHPDocMixinRef(n ir.Node, part phpdoc.CommentPart) {
	rawPart, ok := part.(*phpdoc.RawCommentPart)
	if !ok {
		return
	}

	params := rawPart.Params
	if len(params) == 0 {
		return
	}

	name, ok := solver.GetClassName(r.state, &ir.Name{
		Value: params[0],
	})

	if !ok {
		return
	}

	if _, ok := r.info.GetClass(name); !ok {
		r.walker.ReportPHPDoc(
			PHPDocLineField(n, part.Line(), 1),
			LevelWarning, "invalidDocblockRef", "@mixin tag refers to unknown class %s", name,
		)
	}
}

func (r *rootChecker) checkUndefinedClassesInPHPDoc(n ir.Node, typesMap types.Map, part phpdoc.CommentPart) {
	if !r.info.IsIndexingComplete() {
		return
	}

	resolved := solver.ResolveTypes(r.info, r.state.CurrentClass, typesMap, solver.ResolverMap{})
	typesMap = types.NewMapFromMap(resolved)

	typesMap.Iterate(func(className string) {
		if types.IsShape(className) {
			shape, ok := r.info.GetClass(className)
			if ok {
				for _, info := range shape.Properties {
					info.Typ.Iterate(func(typ string) {
						if !types.IsClass(typ) {
							return
						}

						r.checkUndefinedClass(typ, part, n)
					})
				}
			}
			return
		}

		if types.IsArray(className) {
			arrayType := types.ArrayType(className)
			if types.IsClass(arrayType) {
				r.checkUndefinedClass(arrayType, part, n)
			}
			return
		}

		if !types.IsClass(className) {
			return
		}

		r.checkUndefinedClass(className, part, n)
	})
}

func (r *rootChecker) checkUndefinedClass(className string, part phpdoc.CommentPart, n ir.Node) {
	// While there is no template support, this hack saves you unnecessary bugs.
	if strings.HasSuffix(className, `\T`) {
		return
	}

	_, ok := r.info.GetClassOrTrait(className)
	if ok {
		return
	}
	partNum := 1
	if varPart, ok := part.(*phpdoc.TypeVarCommentPart); ok && varPart.VarIsFirst {
		partNum = 2
	}

	r.walker.ReportPHPDoc(PHPDocLineField(n, part.Line(), partNum),
		LevelError, "undefinedClass",
		"Class or interface named %s does not exist", className,
	)
}

func (r *rootChecker) isValidPHPDocRef(ref string) bool {
	// Skip:
	// - URLs
	// - Things that can be a filename (e.g. "foo.php")
	// - Wildcards (e.g. "self::FOO*")
	// - Issue references (e.g. "#1393" "BACK-103")
	// - RFCs
	if strings.Contains(ref, "http:") || strings.Contains(ref, "https:") {
		return true // OK: URL?
	}
	if strings.Contains(ref, "RFC") {
		return true
	}
	if strings.ContainsAny(ref, ".*-#") {
		return true
	}

	// expandName tries to convert s symbol into fully qualified form.
	expandName := func(s string) string {
		s, ok := solver.GetClassName(r.state, &ir.Name{Value: s})
		if !ok {
			return s
		}
		return s
	}

	isValidGlobalVar := func(ref string) bool {
		// Since we don't have an exhaustive list of globals,
		// we can't tell for sure whether a variable ref is correct.
		return true
	}

	isValidClassSymbol := func(ref string) bool {
		parts := strings.Split(ref, "::")
		if len(parts) != 2 {
			return false
		}
		typeName, symbolName := expandName(parts[0]), parts[1]
		if symbolName == "class" {
			_, ok := r.info.GetClass(typeName)
			return ok
		}
		if strings.HasPrefix(symbolName, "$") {
			return classHasProp(r.state, typeName, symbolName)
		}
		if _, ok := solver.FindMethod(r.info, typeName, symbolName); ok {
			return true
		}
		if _, _, ok := solver.FindConstant(r.info, typeName, symbolName); ok {
			return true
		}
		return false
	}

	isValidSymbol := func(ref string) bool {
		if !strings.HasPrefix(ref, `\`) {
			if r.currentClassNodeStack.Current() != nil {
				className := r.state.CurrentClass
				if _, ok := solver.FindMethod(r.info, className, ref); ok {
					return true // OK: class method reference
				}
				if _, _, ok := solver.FindConstant(r.info, className, ref); ok {
					return true // OK: class constant reference
				}
				if classHasProp(r.state, className, ref) {
					return true // OK: class prop reference
				}
			}

			// Functions and constants fall back in global namespace resolving.
			// See https://www.php.net/manual/en/language.namespaces.fallback.php
			globalRef := `\` + ref
			if _, ok := r.info.GetFunction(globalRef); ok {
				return true // OK: function reference
			}
			if _, ok := r.info.GetConstant(globalRef); ok {
				return true // OK: here's the const reference
			}
		}
		fqnRef := expandName(ref)
		if _, ok := r.info.GetFunction(fqnRef); ok {
			return true // OK: FQN function reference
		}
		if _, ok := r.info.GetClass(fqnRef); ok {
			return true // OK: FQN class reference
		}
		if _, ok := r.info.GetConstant(fqnRef); ok {
			return true // OK: FQN const reference
		}
		return false
	}

	switch {
	case strings.Contains(ref, "::"):
		return isValidClassSymbol(ref)
	case strings.HasPrefix(ref, "$"):
		return isValidGlobalVar(ref)
	default:
		return isValidSymbol(ref)
	}
}

func (r *rootChecker) CheckNameCase(n ir.Node, nameUsed, nameExpected string) {
	if nameUsed == "" || nameExpected == "" {
		return
	}
	if nameUsed != nameExpected {
		r.walker.Report(n, LevelWarning, "nameMismatch", "%s should be spelled %s",
			nameUsed, nameExpected)
	}
}

func (r *rootChecker) CheckKeywordCase(n ir.Node, keyword string) {
	toks := irutil.Keywords(n)
	if toks == nil {
		return
	}

	tok := toks[0]

	switch n := n.(type) {
	case *ir.YieldFromExpr:
		r.compareKeywordWithTokenCase(n, toks[0], "yield")
		r.compareKeywordWithTokenCase(n, toks[1], "from")

	case *ir.ElseIfStmt:
		if !n.Merged {
			r.compareKeywordWithTokenCase(n, toks[0], "if")
			r.compareKeywordWithTokenCase(n, toks[1], "else")
		} else {
			r.compareKeywordWithTokenCase(n, tok, "elseif")
		}

	default:
		r.compareKeywordWithTokenCase(n, tok, keyword)
	}
}

func (r *rootChecker) compareKeywordWithTokenCase(n ir.Node, tok *token.Token, keyword string) {
	wantKwd := keyword
	haveKwd := tok.Value
	if wantKwd != string(haveKwd) {
		r.walker.Report(n, LevelWarning, "keywordCase", "Use %s instead of %s",
			wantKwd, haveKwd)
	}
}

func (r *rootChecker) CheckTypeHintNode(n ir.Node, place string) {
	if !r.info.IsIndexingComplete() || n == nil {
		return
	}

	// We need to check this part without normalization, since
	// otherwise parent will be replaced with the class name.
	typeList := types.TypeHintTypes(n)
	for _, typ := range typeList {
		if typ.Elem == "parent" && r.state.CurrentClass != "" {
			if r.state.CurrentParentClass == "" {
				r.walker.Report(n, LevelError, "typeHint", "Cannot use 'parent' typehint when current class has no parent")
			}
		}
	}

	_, inTrait := r.currentClassNodeStack.Current().(*ir.TraitStmt)

	typesMap := types.NewMapWithNormalization(r.normalizer, typeList)

	typesMap.Iterate(func(typ string) {
		if types.IsClass(typ) {
			className := typ

			_, hasTrait := r.info.GetTrait(className)
			if hasTrait && !inTrait {
				r.walker.Report(n, LevelWarning, "badTraitUse", "Cannot use trait %s as a typehint for %s", strings.TrimPrefix(className, `\`), place)
			}

			class, hasClass := r.info.GetClass(className)

			if !hasClass && !hasTrait {
				r.walker.Report(n, LevelError, "undefinedClass",
					"Class or interface named %s does not exist", className,
				)
			}

			r.CheckNameCase(n, className, class.Name)
		}
	})
}

func (r *rootChecker) CheckFuncParams(funcName *ir.Identifier, params []ir.Node, funcParams parseFuncParamsResult, phpDocParamTypes phpdoctypes.ParamsMap) {
	for _, param := range params {
		r.checkFuncParam(param.(*ir.Parameter))
	}

	r.checkParamsTypeHint(funcName, funcParams, phpDocParamTypes)
}

func (r *rootChecker) checkFuncParam(p *ir.Parameter) {
	r.CheckVarNameMisspellings(p, p.Variable.Name)

	// TODO(quasilyte): DefaultValue can only contain constant expressions.
	// Could run special check over them to detect the potential fatal errors.
	irutil.Inspect(p.DefaultValue, func(w ir.Node) bool {
		if n, ok := w.(*ir.ArrayExpr); ok && !n.ShortSyntax {
			r.walker.Report(n, LevelNotice, "arraySyntax", "Use the short form '[]' instead of the old 'array()'")

			r.walker.addQuickFix("arraySyntax", r.quickfix.Array(n))
		}
		return true
	})

	r.CheckTypeHintFunctionParam(p)
}

func (r *rootChecker) CheckTypeHintFunctionParam(p *ir.Parameter) {
	if !r.info.IsIndexingComplete() {
		return
	}

	r.CheckTypeHintNode(p.VariableType, "parameter type")
}

func (r *rootChecker) checkParamsTypeHint(funcName *ir.Identifier, funcParams parseFuncParamsResult, phpDocParamTypes phpdoctypes.ParamsMap) {
	for param, typeHintType := range funcParams.paramsTypeHint {
		var phpDocType types.Map

		if phpDocParamType, ok := phpDocParamTypes[param]; ok {
			phpDocType = phpDocParamType.Typ
		}

		if !types.TypeHintHasMoreAccurateType(typeHintType, phpDocType) {
			r.walker.Report(funcName, LevelNotice, "typeHint", "Specify the type for the parameter $%s in PHPDoc, 'array' type hint too generic", param)
		}
	}
}

func (r *rootChecker) CheckFuncReturnType(fun ir.Node, funcName string, returnTypeHint, phpDocReturnType types.Map) {
	if !types.TypeHintHasMoreAccurateType(returnTypeHint, phpDocReturnType) {
		r.walker.Report(fun, LevelNotice, "typeHint", "Specify the return type for the function %s in PHPDoc, 'array' type hint too generic", funcName)
	}
}

func (r *rootChecker) CheckCommentMisspellings(n ir.Node, s string) {
	// Try to avoid checking for symbol names and references.
	r.checkMisspellings(n, s, "misspellComment", utils.IsCapitalized)
}

func (r *rootChecker) CheckVarNameMisspellings(n ir.Node, s string) {
	r.checkMisspellings(n, s, "misspellName", func(string) bool {
		return false
	})
}

func (r *rootChecker) CheckIdentMisspellings(n *ir.Identifier) {
	// Before PHP got context-sensitive lexer, it was common to use
	// method names to avoid parsing errors.
	// We can't suggest a fix that leads to a parsing error.
	// To avoid false positives, skip PHP keywords.
	r.checkMisspellings(n, n.Value, "misspellName", utils.IsPHPKeyword)
}

func (r *rootChecker) checkMisspellings(n ir.Node, s string, label string, skip func(string) bool) {
	if !r.info.IsIndexingComplete() {
		return
	}
	if r.typoFixer == nil {
		return
	}
	_, changes := r.typoFixer.Replace(s)
	for _, c := range changes {
		if skip(c.Corrected) || skip(c.Original) {
			continue
		}
		r.walker.Report(n, LevelNotice, label, `"%s" is a misspelling of "%s"`, c.Original, c.Corrected)
	}
}

func (r *rootChecker) CheckAssignNullToNotNullableProperty(prop *ir.PropertyStmt, propTypes types.Map) {
	assignNull := false

	if expr, ok := prop.Expr.(*ir.ConstFetchExpr); ok {
		assignNull = strings.EqualFold(expr.Constant.Value, "null")
	}

	if assignNull && !propTypes.Empty() {
		onlyClasses := true
		nullable := propTypes.Find(func(typ string) bool {
			if !types.IsClass(typ) && typ != "null" {
				onlyClasses = false
			}
			return typ == "null"
		})

		if !nullable && onlyClasses {
			r.walker.Report(prop, LevelNotice, "propNullDefault", "Assigning null to a not nullable property")
			r.walker.addQuickFix("propNullDefault", r.quickfix.NullForNotNullableProperty(prop))
		}
	}
}

func (r *rootChecker) CheckModifierKeywordCase(m *ir.Identifier) {
	lcase := strings.ToLower(m.Value)
	if lcase != m.Value {
		r.walker.Report(m, LevelWarning, "keywordCase", "Use %s instead of %s",
			lcase, m.Value)
	}
}

func (r *rootChecker) CheckOldStyleConstructor(meth *ir.ClassMethodStmt) {
	lastDelim := strings.LastIndexByte(r.state.CurrentClass, '\\')
	methodName := meth.MethodName.Value
	className := r.state.CurrentClass[lastDelim+1:]

	if !strings.EqualFold(className, methodName) {
		return
	}

	_, inClass := r.currentClassNodeStack.Current().(*ir.ClassStmt)
	if !inClass {
		return
	}

	r.walker.Report(meth.MethodName, LevelNotice, "oldStyleConstructor", "Old-style constructor usage, use __construct instead")
}

func (r *rootChecker) CheckPHPDocVar(n ir.Node, doc phpdoc.Comment, typ types.Map) {
	if phpdoc.IsSuspicious([]byte(doc.Raw)) {
		r.walker.ReportPHPDoc(PHPDocLine(n, 1),
			LevelWarning, "invalidDocblock",
			"Multiline PHPDoc comment should start with /**, not /*",
		)
	}

	for _, part := range doc.Parsed {
		r.checkPHPDocRef(n, part)
		part, ok := part.(*phpdoc.TypeVarCommentPart)
		if ok && part.Name() == "var" {
			converted := phpdoctypes.ToRealType(r.normalizer.ClassFQNProvider(), r.normalizer.KPHP(), part.Type)

			if converted.Warning != "" {
				field := 1
				if part.VarIsFirst {
					field = 2
				}
				r.walker.ReportPHPDoc(PHPDocLineField(n, part.Line(), field),
					LevelNotice, "invalidDocblockType",
					converted.Warning,
				)
			}

			r.checkUndefinedClassesInPHPDoc(n, typ, part)
		}
	}
}

func (r *rootChecker) CheckParentConstructorCall(n ir.Node, parentConstructorCalled bool) {
	if !r.info.IsIndexingComplete() {
		return
	}

	class, ok := r.currentClassNodeStack.Current().(*ir.ClassStmt)
	if !ok || class.Extends == nil {
		return
	}
	m, ok := solver.FindMethod(r.info, r.state.CurrentParentClass, `__construct`)
	if !ok || m.Info.AccessLevel == meta.Private || m.Info.IsAbstract() {
		return
	}

	if !parentConstructorCalled {
		r.walker.Report(n, LevelWarning, "parentConstructor", "Missing parent::__construct() call")
	}
}

func (r *rootChecker) CheckMagicMethod(meth ir.Node, name string, modif methodModifiers, countArgs int) {
	const Any = -1
	var (
		canBeStatic    bool
		canBeNonPublic bool
		mustBeStatic   bool

		numArgsExpected int
	)

	switch name {
	case "__call",
		"__set":
		canBeStatic = false
		canBeNonPublic = false
		numArgsExpected = 2

	case "__get",
		"__isset",
		"__unset":
		canBeStatic = false
		canBeNonPublic = false
		numArgsExpected = 1

	case "__toString":
		canBeStatic = false
		canBeNonPublic = false
		numArgsExpected = 0

	case "__invoke",
		"__debugInfo":
		canBeStatic = false
		canBeNonPublic = false
		numArgsExpected = Any

	case "__construct":
		canBeStatic = false
		canBeNonPublic = true
		numArgsExpected = Any

	case "__destruct", "__clone":
		canBeStatic = false
		canBeNonPublic = true
		numArgsExpected = 0

	case "__callStatic":
		canBeStatic = true
		canBeNonPublic = false
		mustBeStatic = true
		numArgsExpected = 2

	case "__sleep",
		"__wakeup",
		"__serialize",
		"__unserialize",
		"__set_state":
		canBeNonPublic = true
		canBeStatic = true
		numArgsExpected = Any

	default:
		return // Not a magic method
	}

	switch {
	case mustBeStatic && !modif.static:
		r.walker.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() must be static", name)
	case !canBeStatic && modif.static:
		r.walker.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() cannot be static", name)
	}
	if !canBeNonPublic && modif.accessLevel != meta.Public {
		r.walker.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() must have public visibility", name)
	}

	if countArgs != numArgsExpected && numArgsExpected != Any {
		r.walker.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() must take exactly %d argument", name, numArgsExpected)
	}
}

func (r *rootChecker) CheckExtends(class ir.Node, currentClass meta.ClassInfo, extends *ir.ClassExtendsStmt) {
	if extends == nil {
		return
	}

	r.CheckKeywordCase(extends, "extends")

	className, ok := solver.GetClassName(r.state, extends.ClassName)
	if ok {
		r.CheckClassInherit(class, extends.ClassName, currentClass, className)
	}
}

func (r *rootChecker) CheckImplements(class ir.Node, currentClass meta.ClassInfo, implements *ir.ClassImplementsStmt) {
	if implements == nil {
		return
	}

	r.CheckKeywordCase(implements, "implements")

	for _, tr := range implements.InterfaceNames {
		interfaceName, ok := solver.GetClassName(r.state, tr)
		if ok {
			currentClass.Interfaces[interfaceName] = struct{}{}
			r.CheckInterfaceImplemented(class, tr, currentClass, interfaceName)
		}
	}
}

func (r *rootChecker) CheckTraitImplemented(classNode, name ir.Node, currentClass meta.ClassInfo, nameUsed string) {
	if !r.info.IsIndexingComplete() {
		return
	}
	trait, ok := r.info.GetTrait(nameUsed)
	if !ok {
		r.ReportUndefinedTrait(name, nameUsed)
		return
	}
	r.checkImplemented(classNode, name, nameUsed, currentClass, trait)
}

func (r *rootChecker) CheckClassInherit(classNode, extendsClassNameNode ir.Node, currentClass meta.ClassInfo, nameUsed string) {
	if !r.info.IsIndexingComplete() {
		return
	}

	class, ok := r.info.GetClass(nameUsed)
	if !ok {
		r.ReportUndefinedClass(extendsClassNameNode, nameUsed)
		return
	}

	r.checkClassExtends(extendsClassNameNode, currentClass, class)
	r.checkImplemented(classNode, extendsClassNameNode, nameUsed, currentClass, class)
}

func (r *rootChecker) checkClassExtends(extendsClassNameNode ir.Node, currentClass, otherClass meta.ClassInfo) {
	if otherClass.IsFinal() {
		r.walker.Report(extendsClassNameNode, LevelError, "invalidExtendClass", "Class %s may not inherit from final class %s", currentClass.Name, otherClass.Name)
	}
}

func (r *rootChecker) checkClassImplemented(classNode, extendsClassNameNode ir.Node, currentClass meta.ClassInfo, nameUsed string) {
	if !r.info.IsIndexingComplete() {
		return
	}

	class, ok := r.info.GetClass(nameUsed)
	if !ok {
		r.ReportUndefinedClass(extendsClassNameNode, nameUsed)
		return
	}

	r.checkImplemented(classNode, extendsClassNameNode, nameUsed, currentClass, class)
}

func (r *rootChecker) CheckInterfaceImplemented(classNode, name ir.Node, currentClass meta.ClassInfo, nameUsed string) {
	r.checkClassImplemented(classNode, name, currentClass, nameUsed)
}

func (r *rootChecker) checkImplemented(classNode, name ir.Node, nameUsed string, currentClass, otherClass meta.ClassInfo) {
	if r.state.IsTrait || currentClass.IsAbstract() {
		return
	}

	r.CheckNameCase(name, nameUsed, otherClass.Name)
	visited := make(map[string]struct{}, 4)
	r.checkImplementedStep(classNode, name, nameUsed, otherClass, visited)
}

func (r *rootChecker) checkImplementedStep(classNode, name ir.Node, className string, otherClass meta.ClassInfo, visited map[string]struct{}) {
	// TODO: check that method signatures are compatible?
	if _, ok := visited[className]; ok {
		return
	}
	visited[className] = struct{}{}

	for _, ifaceMethod := range otherClass.Methods.H {
		m, ok := solver.FindMethod(r.info, r.state.CurrentClass, ifaceMethod.Name)
		if !ok || !m.Implemented {
			r.walker.Report(name, LevelError, "unimplemented", "Class %s must implement %s::%s method",
				r.state.CurrentClass, className, ifaceMethod.Name)
			continue
		}
		if m.Info.Name != ifaceMethod.Name {
			r.walker.Report(name, LevelNotice, "nameMismatch", "%s::%s should be spelled as %s::%s",
				r.state.CurrentClass, m.Info.Name, className, ifaceMethod.Name)

		}
		if ifaceMethod.IsFinal() && ifaceMethod.AccessLevel != meta.Private {
			methodNode := irutil.FindClassMethodNode(classNode, ifaceMethod.Name)
			if methodNode != nil {
				r.walker.Report(methodNode, LevelError, "methodSignatureMismatch",
					"Method %s::%s is declared final and cannot be overridden",
					otherClass.Name, ifaceMethod.Name)
			}
		}
	}

	for _, ifaceName := range otherClass.ParentInterfaces {
		iface, ok := r.info.GetClass(ifaceName)
		if ok {
			r.checkImplementedStep(classNode, name, ifaceName, iface, visited)
		}
	}

	if otherClass.Parent != "" {
		class, ok := r.info.GetClass(otherClass.Parent)
		if ok {
			r.checkImplementedStep(classNode, name, otherClass.Parent, class, visited)
		}
	}
}

func (r *rootChecker) ReportUndefinedClass(n ir.Node, name string) {
	r.walker.Report(n, LevelError, "undefinedClass", "Class or interface named %s does not exist", name)
}

func (r *rootChecker) ReportUndefinedTrait(n ir.Node, name string) {
	r.walker.Report(n, LevelError, "undefinedTrait", "Trait named %s does not exist", name)
}
