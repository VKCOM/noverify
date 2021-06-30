package autogen

import (
	"github.com/VKCOM/noverify/src/types"
)

type ShapeTypeProp struct {
	Key   string
	Types []types.Type
}

type ShapeTypeInfo struct {
	Name  string
	Props []ShapeTypeProp
}
