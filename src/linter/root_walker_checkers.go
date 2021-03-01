package linter

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
)

func (d *rootWalker) checkClass(classNode *ir.ClassStmt) {
	d.checkClassModifiers(classNode.Modifiers)
	d.checkClassImplements(classNode.Implements)
	d.checkClassExtends(classNode.Extends)

	doc := d.parseClassPHPDoc(classNode, classNode.PhpDoc)
	d.checkClassPHPDoc(classNode, classNode.PhpDoc)
	d.reportPhpdocErrors(classNode, doc.errs)

	d.checkCommentMisspellings(classNode.ClassName, classNode.PhpDocComment)
	d.checkIdentMisspellings(classNode.ClassName)
}

func (d *rootWalker) checkClassModifiers(modifiers []*ir.Identifier) {
	for _, m := range modifiers {
		d.checkLowerCaseModifier(m)
	}
}

func (d *rootWalker) checkClassImplements(impl *ir.ClassImplementsStmt) {
	if impl == nil {
		return
	}

	d.checkKeywordCase(impl, "implements")

	for _, tr := range impl.InterfaceNames {
		interfaceName, ok := solver.GetClassName(d.ctx.st, tr)
		if !ok {
			continue
		}

		d.checkIfaceImplemented(tr, interfaceName)
	}
}

func (d *rootWalker) checkClassExtends(exts *ir.ClassExtendsStmt) {
	if exts == nil {
		return
	}

	d.checkKeywordCase(exts, "extends")

	className, ok := solver.GetClassName(d.ctx.st, exts.ClassName)
	if !ok {
		return
	}

	d.checkClassImplemented(exts.ClassName, className)
}

func (d *rootWalker) checkClassPHPDoc(n ir.Node, doc []phpdoc.CommentPart) {
	if len(doc) == 0 {
		return
	}

	for _, part := range doc {
		d.checkPHPDocRef(n, part)
	}
}

func (d *rootWalker) checkClassMethod(m *ir.ClassMethodStmt) {
	class, ok := d.getCurrentClass()
	if !ok {
		return
	}

	// data
	name := m.MethodName.Value
	method, has := class.Methods.Get(name)
	if !has {
		return
	}
	_, insideInterface := d.currentClassNode.(*ir.InterfaceStmt)
	sc := meta.NewScope()
	modif := d.parseMethodModifiers(m)

	// state
	d.addThisVariableToClassMethodScope(modif, sc)
	d.addParamsToClassMethodScope(method, sc)

	handleMethodInfo := d.handleFuncStmts(method.Params, nil, convertNodeToStmts(m.Stmt), sc)

	// checks
	d.checkClassMethodOldStyleConstructor(m, name)
	d.checkParentConstructorCall(m.MethodName, handleMethodInfo.callsParentConstructor)

	d.checkClassMethodParams(m)
	d.checkClassMethodModifiers(m)
	d.checkClassMethodComplexity(m)
	d.checkClassMethodTraversable(m, name, method)
	d.checkClassMethodPhpDoc(m, name, modif, insideInterface)
	d.checkClassMagicMethod(m.MethodName, name, modif, len(m.Params))

	d.checkIdentMisspellings(m.MethodName)
	d.checkCommentMisspellings(m.MethodName, m.PhpDocComment)
}

func (d *rootWalker) checkClassMethodModifiers(meth *ir.ClassMethodStmt) {
	for _, m := range meth.Modifiers {
		modifier := d.checkLowerCaseModifier(m)
		switch modifier {
		case "abstract", "static", "public", "private", "protected", "final":
			// ok
		default:
			linterError(d.ctx.st.CurrentFile, "Unrecognized method modifier: %s", m.Value)
		}
	}
}

func (d *rootWalker) checkClassMethodOldStyleConstructor(meth *ir.ClassMethodStmt, nm string) {
	lastDelim := strings.IndexByte(d.ctx.st.CurrentClass, '\\')
	if strings.EqualFold(d.ctx.st.CurrentClass[lastDelim+1:], nm) {
		_, isClass := d.currentClassNode.(*ir.ClassStmt)
		if isClass {
			d.Report(meth.MethodName, LevelNotice, "oldStyleConstructor", "Old-style constructor usage, use __construct instead")
		}
	}
}

func (d *rootWalker) checkClassMethodPhpDoc(m *ir.ClassMethodStmt, name string, modif methodModifiers, insideInterface bool) {
	if m.PhpDocComment == "" && modif.accessLevel == meta.Public {
		// Permit having "__call" and other magic method without comments.
		if !insideInterface && !strings.HasPrefix(name, "_") {
			d.Report(m.MethodName, LevelNotice, "phpdoc", "Missing PHPDoc for %q public method", name)
		}
	}

	doc := d.parsePHPDoc(m.MethodName, m.PhpDoc, m.Params)
	d.reportPhpdocErrors(m.MethodName, doc.errs)
}

func (d *rootWalker) checkClassMethodComplexity(m *ir.ClassMethodStmt) {
	pos := ir.GetPosition(m)
	if funcSize := pos.EndLine - pos.StartLine; funcSize > maxFunctionLines {
		d.Report(m.MethodName, LevelNotice, "complexity", "Too big method: more than %d lines", maxFunctionLines)
	}
}

func (d *rootWalker) checkClassMethodParams(m *ir.ClassMethodStmt) {
	for _, p := range m.Params {
		d.checkVarnameMisspellings(p, p.(*ir.Parameter).Variable.Name)
	}
}

func (d *rootWalker) checkClassMethodTraversable(m *ir.ClassMethodStmt, name string, method meta.FuncInfo) {
	if name != "getIterator" {
		return
	}
	if !solver.Implements(d.metaInfo(), d.ctx.st.CurrentClass, `\IteratorAggregate`) {
		return
	}

	implementsTraversable := method.Typ.Find(func(typ string) bool {
		return solver.Implements(d.metaInfo(), typ, `\Traversable`)
	})

	if !implementsTraversable {
		d.Report(m.MethodName, LevelError, "stdInterface", "Objects returned by %s::getIterator() must be traversable or implement interface \\Iterator", d.ctx.st.CurrentClass)
	}
}

func (d *rootWalker) checkClassMagicMethod(meth ir.Node, name string, modif methodModifiers, countArgs int) {
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
		d.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() must be static", name)
	case !canBeStatic && modif.static:
		d.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() cannot be static", name)
	}
	if !canBeNonPublic && modif.accessLevel != meta.Public {
		d.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() must have public visibility", name)
	}

	if countArgs != numArgsExpected && numArgsExpected != Any {
		d.Report(meth, LevelError, "magicMethodDecl", "The magic method %s() must take exactly %d argument", name, numArgsExpected)
	}
}

func (d *rootWalker) checkInterface(n *ir.InterfaceStmt) {
	d.checkKeywordCase(n, "interface")
	d.checkCommentMisspellings(n.InterfaceName, n.PhpDocComment)
	if !strings.HasSuffix(n.InterfaceName.Value, "able") {
		d.checkIdentMisspellings(n.InterfaceName)
	}
}

func (d *rootWalker) checkTrait(n *ir.TraitStmt) {
	d.checkKeywordCase(n, "trait")
	d.checkCommentMisspellings(n.TraitName, n.PhpDocComment)
	d.checkIdentMisspellings(n.TraitName)
}

func (d *rootWalker) checkTraitUse(n *ir.TraitUseStmt) {
	d.checkKeywordCase(n, "use")

	for _, tr := range n.Traits {
		traitName, ok := solver.GetClassName(d.ctx.st, tr)
		if !ok {
			continue
		}

		d.checkTraitImplemented(tr, traitName)
	}
}

func (d *rootWalker) checkPropertyList(pl *ir.PropertyListStmt) {
	d.checkPropertyModifiers(pl)
	d.checkCommentMisspellings(pl, pl.PhpDocComment)

	for _, pNode := range pl.Properties {
		prop := pNode.(*ir.PropertyStmt)
		d.checkPHPDocVar(prop, pl.PhpDoc)
	}
}

func (d *rootWalker) checkPropertyModifiers(pl *ir.PropertyListStmt) {
	for _, m := range pl.Modifiers {
		d.checkLowerCaseModifier(m)
	}
}

func (d *rootWalker) checkClassConstList(s *ir.ClassConstListStmt) {
	d.checkConstantAccessLevel(s)
	d.checkCommentMisspellings(s, s.PhpDocComment)
}

func (d *rootWalker) checkConstantAccessLevel(s *ir.ClassConstListStmt) {
	for _, m := range s.Modifiers {
		d.checkLowerCaseModifier(m)
	}
}
