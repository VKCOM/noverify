package linter

import (
	"path/filepath"

	"github.com/VKCOM/noverify/src/baseline"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/quickfix"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
)

type rootContext struct {
	*WorkerContext

	st *meta.ClassParseState

	typeNormalizer types.Normalizer

	// shapes is a list of generated shape types for the current file.
	shapes types.ShapesMap

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

	classFQNProvider := func(name string) (string, bool) {
		return solver.GetClassName(st, &ir.Name{Value: name})
	}

	return rootContext{
		WorkerContext: workerCtx,

		typeNormalizer: types.NewNormalizer(classFQNProvider, config.KPHP),
		st:             st,
		baseline:       p,
		shapes:         types.ShapesMap{},
	}
}
