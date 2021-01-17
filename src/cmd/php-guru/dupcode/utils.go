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

func runIndexing(cacheDir string, targets []string, filter *workspace.FilenameFilter) error {
	config := linter.NewConfig()
	config.CacheDir = cacheDir

	// If we don't do this, the program will hang.
	go linter.MemoryLimiterThread(0)

	// Handle stubs.
	filenames := stubs.AssetNames()
	if err := cmd.LoadEmbeddedStubs(config, filenames); err != nil {
		return err
	}

	// Handle workspace files.
	l := linter.NewLinter(config)
	l.AnalyzeFiles(workspace.ReadFilenames(targets, filter, []string{"php", "inc", "php5", "phtml"}))

	meta.SetIndexingComplete(true)
	return nil
}
