package linter

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/solver"
)

func (d *RootWalker) checkClass(classNode *ir.ClassStmt) {
	d.checkClassModifiers(classNode.Modifiers)
	d.checkClassImplements(classNode.Implements)
	d.checkClassExtends(classNode.Extends)

	doc := d.parseClassPHPDoc(classNode, classNode.PhpDoc)
	d.checkClassPHPDoc(classNode, classNode.PhpDoc)
	d.reportPhpdocErrors(classNode, doc.errs)

	d.checkCommentMisspellings(classNode.ClassName, classNode.PhpDocComment)
	d.checkIdentMisspellings(classNode.ClassName)
}

func (d *RootWalker) checkClassModifiers(modifiers []*ir.Identifier) {
	for _, m := range modifiers {
		d.checkLowerCaseModifier(m)
	}
}

func (d *RootWalker) checkClassImplements(impl *ir.ClassImplementsStmt) {
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

func (d *RootWalker) checkClassExtends(exts *ir.ClassExtendsStmt) {
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

func (d *RootWalker) checkClassPHPDoc(n ir.Node, doc []phpdoc.CommentPart) {
	if len(doc) == 0 {
		return
	}

	for _, part := range doc {
		d.checkPHPDocRef(n, part)
	}
}

func (d *RootWalker) checkClassMethod(m *ir.ClassMethodStmt) {
	class := d.getOrCreateCurrentClass()

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
	d.addClassMethodThisVariableToScope(modif, sc)
	d.addClassMethodParamsToScope(method, sc)
	d.addScope(m, sc)

	handleMethodInfo := d.handleFuncStmts(method.Params, nil, convertNodeToStmts(m.Stmt), sc)

	// checks
	d.checkClassMethodOldStyleConstructor(m, name)
	d.checkParentConstructorCall(m.MethodName, handleMethodInfo.callsParentConstructor)

	d.checkClassMethodComplexity(m)
	d.checkClassMethodPhpDoc(m, name, modif, insideInterface)
	d.checkClassMethodParams(m)
	d.checkClassMethodTraversable(m, name, method)
	d.checkClassMagicMethod(m.MethodName, name, modif, len(m.Params))

	d.checkIdentMisspellings(m.MethodName)
	d.checkCommentMisspellings(m.MethodName, m.PhpDocComment)
}

func (d *RootWalker) checkClassMethodOldStyleConstructor(meth *ir.ClassMethodStmt, nm string) {
	lastDelim := strings.IndexByte(d.ctx.st.CurrentClass, '\\')
	if strings.EqualFold(d.ctx.st.CurrentClass[lastDelim+1:], nm) {
		_, isClass := d.currentClassNode.(*ir.ClassStmt)
		if isClass {
			d.Report(meth.MethodName, LevelDoNotReject, "oldStyleConstructor", "Old-style constructor usage, use __construct instead")
		}
	}
}

func (d *RootWalker) checkClassMethodPhpDoc(m *ir.ClassMethodStmt, name string, modif methodModifiers, insideInterface bool) {
	if m.PhpDocComment == "" && modif.accessLevel == meta.Public {
		// Permit having "__call" and other magic method without comments.
		if !insideInterface && !strings.HasPrefix(name, "_") {
			d.Report(m.MethodName, LevelDoNotReject, "phpdoc", "Missing PHPDoc for %q public method", name)
		}
	}

	doc := d.parsePHPDoc(m.MethodName, m.PhpDoc, m.Params)
	d.reportPhpdocErrors(m.MethodName, doc.errs)
}

func (d *RootWalker) checkClassMethodComplexity(m *ir.ClassMethodStmt) {
	pos := ir.GetPosition(m)
	if funcSize := pos.EndLine - pos.StartLine; funcSize > maxFunctionLines {
		d.Report(m.MethodName, LevelDoNotReject, "complexity", "Too big method: more than %d lines", maxFunctionLines)
	}
}

func (d *RootWalker) checkClassMethodParams(m *ir.ClassMethodStmt) {
	for _, p := range m.Params {
		d.checkVarnameMisspellings(p, p.(*ir.Parameter).Variable.Name)
	}
}

func (d *RootWalker) checkClassMethodTraversable(m *ir.ClassMethodStmt, name string, method meta.FuncInfo) {
	if !meta.IsIndexingComplete() {
		return
	}
	if name != "getIterator" {
		return
	}
	if !solver.Implements(d.ctx.st.CurrentClass, `\IteratorAggregate`) {
		return
	}

	implementsTraversable := method.Typ.Find(func(typ string) bool {
		return solver.Implements(typ, `\Traversable`)
	})

	if !implementsTraversable {
		d.Report(m.MethodName, LevelError, "stdInterface", "Objects returned by %s::getIterator() must be traversable or implement interface \\Iterator", d.ctx.st.CurrentClass)
	}
}

func (d *RootWalker) checkClassMagicMethod(meth ir.Node, name string, modif methodModifiers, countArgs int) {
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

func (d *RootWalker) checkInterface(n *ir.InterfaceStmt) {
	d.checkKeywordCase(n, "interface")
	d.checkCommentMisspellings(n.InterfaceName, n.PhpDocComment)
	if !strings.HasSuffix(n.InterfaceName.Value, "able") {
		d.checkIdentMisspellings(n.InterfaceName)
	}
}

func (d *RootWalker) checkTrait(n *ir.TraitStmt) {
	d.checkKeywordCase(n, "trait")
	d.checkCommentMisspellings(n.TraitName, n.PhpDocComment)
	d.checkIdentMisspellings(n.TraitName)
}