package linter

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
)

type rootContext struct {
	typeNormalizer   typeNormalizer
	phpdocTypeParser *phpdoc.TypeParser

	// TODO: move class parse state here?
}

func newRootContext(st *meta.ClassParseState) rootContext {
	return rootContext{
		typeNormalizer:   typeNormalizer{st: st},
		phpdocTypeParser: phpdoc.NewTypeParser(),
	}
}

func newTypesMap(ctx *rootContext, types []meta.Type) meta.TypesMap {
	ctx.typeNormalizer.NormalizeTypes(types)
	return meta.NewTypesMapFromTypes(types)
}
