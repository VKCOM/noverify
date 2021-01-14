package linter

import (
	"github.com/VKCOM/noverify/src/ir"
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
