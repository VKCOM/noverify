package linter

import (
	"bytes"

	"github.com/VKCOM/noverify/src/phpdoc"
)

// WorkerContext is a state that is shared between all worker-owned
// RootWalker's and BlockWalker's.
//
// A worker is a separate goroutine that processed the incoming files.
//
// Since workerContext is worker-bound, that state is never accessed
// from different threads, so we can re-use it without synchronization.
type WorkerContext struct {
	phpdocTypeParser *phpdoc.TypeParser

	scratchBuf bytes.Buffer
}

func NewWorkerContext() *WorkerContext {
	return &WorkerContext{
		phpdocTypeParser: phpdoc.NewTypeParser(),
		scratchBuf:       bytes.Buffer{},
	}
}
