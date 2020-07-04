package linter

import (
	"fmt"
	"path/filepath"

	"github.com/VKCOM/noverify/src/baseline"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
)

type rootContext struct {
	st *meta.ClassParseState

	typeNormalizer   typeNormalizer
	phpdocTypeParser *phpdoc.TypeParser

	// shapes is a list of generated shape types for the current file.
	shapes []shapeTypeInfo

	baseline     baseline.FileProfile
	hashCounters map[uint64]int // Allocated lazily
}

func newRootContext(st *meta.ClassParseState) rootContext {
	var p baseline.FileProfile
	if BaselineProfile != nil {
		filename := filepath.Base(st.CurrentFile)
		p = BaselineProfile.Files[filename]
	}
	return rootContext{
		typeNormalizer:   typeNormalizer{st: st},
		phpdocTypeParser: phpdoc.NewTypeParser(),
		st:               st,
		baseline:         p,
	}
}

func (ctx *rootContext) generateShapeName() string {
	// We'll probably generate names for anonymous classes in the
	// same way in future. All auto-generated names should end with "$".
	// `\shape$` prefix makes it easy to check whether a type
	// is a shape without looking it up inside classes map.
	return fmt.Sprintf(`\shape$%s$%d$`, ctx.st.CurrentFile, len(ctx.shapes))
}

func newTypesMap(ctx *rootContext, types []meta.Type) meta.TypesMap {
	ctx.typeNormalizer.NormalizeTypes(types)
	return meta.NewTypesMapFromTypes(types)
}
