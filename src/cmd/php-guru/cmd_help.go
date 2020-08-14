package main

import (
	"flag"

	"github.com/VKCOM/noverify/src/cmd/php-guru/guru"
)

func cmdHelp(ctx *guru.Context) (int, error) {
	flag.Parse()

	printSupportedCommands(commands)

	return 0, nil
}
