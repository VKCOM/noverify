package main

import (
	"log"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverifycache"
)

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
	cmd.Main(&cmd.MainConfig{
		CodeCache: noverifycache.CodeCacheMap,
	})
}
