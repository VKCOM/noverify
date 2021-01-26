package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"path"
	"path/filepath"
)

func loadPackage(typechecker *types.Config, ctx *context, dir string) (*packageData, error) {
	pkgName := path.Base(filepath.ToSlash(dir))
	parsedPkgs, err := parser.ParseDir(ctx.fset, dir, nil, 0)
	if err != nil {
		return nil, err
	}
	astPkg, ok := parsedPkgs[pkgName]
	if !ok {
		return nil, fmt.Errorf("package %s not found", pkgName)
	}
	astFiles := make([]*ast.File, 0, len(astPkg.Files))
	for _, f := range astPkg.Files {
		astFiles = append(astFiles, f)
	}
	var info types.Info
	typesPkg, err := typechecker.Check(pkgName, ctx.fset, astFiles, &info)
	if err != nil {
		return nil, err
	}
	root := typesPkg.Scope()
	result := &packageData{
		scope: root,
	}
	nodeObject := root.Lookup("Node")
	nodeIface, ok := nodeObject.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("can't find ir.Node type")
	}
	ctx.nodeIface = nodeIface
	for _, sym := range root.Names() {
		tn, ok := root.Lookup(sym).(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		structType, ok := named.Underlying().(*types.Struct)
		if !ok {
			continue
		}
		if !types.Implements(types.NewPointer(named), nodeIface) {
			continue
		}
		result.types = append(result.types, &typeData{
			name: tn.Name(),
			info: structType,
		})
	}

	return result, nil
}
