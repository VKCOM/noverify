package stubsutil

import (
	"github.com/VKCOM/noverify/src/meta"
)

func InitPos(pos *meta.ElementPosition, filename string, line, endLine, character, length int32) {
	*pos = meta.ElementPosition{
		Line:      line,
		EndLine:   endLine,
		Character: character,
		Length:    length,
		Filename:  filename,
	}
}

func NewFuncParam(name string, typ meta.TypesMap) meta.FuncParam {
	return meta.FuncParam{Name: name, Typ: typ}
}

func NewRefFuncParam(name string, typ meta.TypesMap) meta.FuncParam {
	return meta.FuncParam{Name: name, Typ: typ, IsRef: true}
}
