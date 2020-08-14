package dupcode

import (
	"crypto/md5"
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irfmt"
)

type codeHash string

func calculateCodeHash(list []ir.Node) codeHash {
	asBlock := &ir.StmtList{Stmts: list}
	codeText := irfmt.Node(asBlock)
	h := md5.Sum([]byte(codeText))
	return codeHash(fmt.Sprintf("%x", h))
}

type funcSet map[codeHash]*funcInfo

func (set funcSet) AddFunc(key codeHash, info *funcInfo) {
	origFunc, ok := set[key]
	if ok {
		origFunc.dups = append(origFunc.dups, info)
		return
	}
	set[key] = info
}

func (set funcSet) Merge(otherSet funcSet) {
	for hash, x := range otherSet {
		y, ok := set[hash]
		if !ok {
			set[hash] = x
			continue
		}
		y.dups = append(y.dups, x)
		y.dups = append(y.dups, x.dups...)
	}
}
