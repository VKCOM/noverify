package parser

import (
	"github.com/VKCOM/noverify/src/php/parser/errors"
	"github.com/VKCOM/noverify/src/php/parser/node"
)

// Parser interface
type Parser interface {
	Parse() int
	GetPath() string
	GetRootNode() *node.Root
	GetErrors() []*errors.Error
	WithFreeFloating()
}
