package main

import (
	"flag"
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type arguments struct {
	debug bool
}

func main() {
	log.SetFlags(0)

	var args arguments
	flag.BoolVar(&args.debug, "v", false,
		`enable debug info output`)
	flag.Parse()

	ctx := context{args: args}

	steps := []struct {
		name string
		fn   func(ctx *context) error
	}{
		{"validate args", doValidateArgs},
		{"init context", doInitContext},
		{"generate equal", doGenerateEqual},
		{"generate clone", doGenerateClone},
		{"generate walk", doGenerateWalk},
		{"generate get freefloating", doGenerateGetFreeFloating},
		{"generate get node kind", doGenerateGetNodeKind},
		{"generate get position", doGenerateGetPosition},
		{"generate iterate tokens", doGenerateIterateTokens},
	}

	for _, step := range steps {
		ctx.Debugf("step %s started", step.name)
		if err := step.fn(&ctx); err != nil {
			log.Fatalf("%s: error: %+v", step.name, err)
		}
	}
}

func doValidateArgs(ctx *context) error {
	// Nothing to do right now.
	return nil
}

func doInitContext(ctx *context) error {
	ctx.fset = token.NewFileSet()
	ctx.date = time.Now()

	gitOutput, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("get git revision: %v: %s", err, gitOutput)
	}
	ctx.gitCommit = strings.TrimSpace(string(gitOutput))

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %v", err)
	}
	ctx.rootDir = wd

	typechecker := &types.Config{
		Importer: importer.ForCompiler(ctx.fset, "source", nil),
	}
	irPkg, err := loadPackage(typechecker, ctx, ctx.rootDir)
	if err != nil {
		return fmt.Errorf("load ir package: %v", err)
	}
	ctx.irPkg = irPkg

	sort.Slice(ctx.irPkg.types, func(i, j int) bool {
		return ctx.irPkg.types[i].name < ctx.irPkg.types[j].name
	})

	return nil
}

func doGenerateEqual(ctx *context) error {
	g := &genEqual{ctx: ctx}
	return g.Run()
}

func doGenerateClone(ctx *context) error {
	g := &genClone{ctx: ctx}
	return g.Run()
}

func doGenerateWalk(ctx *context) error {
	g := &genWalk{ctx: ctx}
	return g.Run()
}

func doGenerateGetFreeFloating(ctx *context) error {
	g := &genGetFreeFloating{ctx: ctx}
	return g.Run()
}

func doGenerateGetNodeKind(ctx *context) error {
	g := &genGetNodeKind{ctx: ctx}
	return g.Run()
}

func doGenerateGetPosition(ctx *context) error {
	g := &genGetPosition{ctx: ctx}
	return g.Run()
}

func doGenerateIterateTokens(ctx *context) error {
	g := &genIterate{ctx: ctx}
	return g.Run()
}
