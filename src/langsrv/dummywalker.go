package langsrv

import (
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

type dummyWalker struct{}

func (d *dummyWalker) EnterChildNode(key string, w walker.Walkable) {}
func (d *dummyWalker) LeaveChildNode(key string, w walker.Walkable) {}
func (d *dummyWalker) EnterChildList(key string, w walker.Walkable) {}
func (d *dummyWalker) LeaveChildList(key string, w walker.Walkable) {}
