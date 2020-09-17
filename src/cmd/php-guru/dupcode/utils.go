package dupcode

import (
	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/cmd/stubs"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/workspace"
)

func findNode(root ir.Node, pred func(n ir.Node) bool) bool {
	found := false
	irutil.Inspect(root, func(n ir.Node) bool {
		if found {
			return false
		}
		if pred(n) {
			found = true
		}
		if v, ok := n.(*ir.SimpleVar); ok {
			if v.Name == "this" {
				found = true
			}
		}
		return !found
	})
	if found {
		return true
	}
	return found
}

func hasModifier(list []*ir.Identifier, key string) bool {
	for _, x := range list {
		if x.Value == key {
			return true
		}
	}
	return false
}

func runIndexing(cacheDir string, targets []string, filter *workspace.FilenameFilter) {
	linter.CacheDir = cacheDir
	linter.AnalysisFiles = targets

	// If we don't do this, the program will hang.
	go linter.MemoryLimiterThread()

	// Handle stubs.
	filenames := stubs.AssetNames()
	cmd.LoadEmbeddedStubs(filenames)

	// Handle workspace files.
	linter.ParseFilenames(workspace.ReadFilenames(targets, filter), nil, nil)

	meta.SetIndexingComplete(true)
}
