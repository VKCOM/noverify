package stubsutil

import (
	"github.com/VKCOM/noverify/src/meta"
)

func NewFuncParam(name string, typ meta.TypesMap) meta.FuncParam {
	return meta.FuncParam{Name: name, Typ: typ}
}

func NewRefFuncParam(name string, typ meta.TypesMap) meta.FuncParam {
	return meta.FuncParam{Name: name, Typ: typ, IsRef: true}
}
