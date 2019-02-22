package main

import (
	"log"

	"github.com/VKCOM/noverify/src/cmd"
)

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	// You can register your own rules here, see src/linter/custom.go

	cmd.Main()
}
