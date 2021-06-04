package linter

import (
	"path/filepath"

	"github.com/VKCOM/noverify/src/baseline"
	"github.com/VKCOM/noverify/src/linter/autogen"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/types"
)

type rootContext struct {
	*WorkerContext

	st *meta.ClassParseState

	typeNormalizer typeNormalizer

	// shapes is a list of generated shape types for the current file.
	shapes map[string]autogen.ShapeTypeInfo

	baseline     baseline.FileProfile
	hashCounters map[uint64]int // Allocated lazily

	fixes []quickfix.TextEdit
}

func newRootContext(config *Config, workerCtx *WorkerContext, st *meta.ClassParseState) rootContext {
	var p baseline.FileProfile
	if config.BaselineProfile != nil {
		filename := filepath.Base(st.CurrentFile)
		p = config.BaselineProfile.Files[filename]
	}
	return rootContext{
		WorkerContext: workerCtx,

		typeNormalizer: typeNormalizer{st: st, kphp: config.KPHP},
		st:             st,
		baseline:       p,
		shapes:         map[string]autogen.ShapeTypeInfo{},
	}
}

func newTypesMap(ctx *rootContext, typs []types.Type) types.Map {
	ctx.typeNormalizer.NormalizeTypes(typs)
	return types.NewMapFromTypes(typs)
}
