package autogen

import (
	"github.com/VKCOM/noverify/src/meta"
)

type ShapeTypeProp struct {
	Key   string
	Types []meta.Type
}

type ShapeTypeInfo struct {
	Name  string
	Props []ShapeTypeProp
}
