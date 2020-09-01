package main

import (
	"go/types"
)

func formatType(typ types.Type) string {
	return types.TypeString(typ, func(pkg *types.Package) string {
		return pkg.Name()
	})
}
