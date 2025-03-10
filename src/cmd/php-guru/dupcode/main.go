package dupcode

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"sync"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/cmd/php-guru/guru"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parseutil"
	"github.com/VKCOM/noverify/src/workspace"
)

// We don't use linter "custom" API here as it doesn't bring much to the
// table for us and we don't want to depend on that API.
//
// Maybe we'll roll the custom-v2 and use it here.

// TODO: handle methods better (don't give up on self-references).

type normalizationLevel int

const (
	// -norm=0 is "no normalization".
	normNone normalizationLevel = iota
	// -norm=1 enables basic normalization.
	normFast
	// -norm=2 also does indexing, so normalization can do more.
	normSafe
	// -norm=3 also enables risky normalizations.
	normMore
)

func Main(ctx *guru.Context) (int, error) {
	var args arguments
	flag.BoolVar(&args.checkPrivate, "private", false,
		"whether private methods should be analyzed")
	flag.BoolVar(&args.checkAutogen, "autogen", false,
		"whether auto-generated files should be analyzed")
	flag.StringVar(&args.exclude, "exclude", "",
		"regexp that excludes files from the analysis")
	flag.StringVar(&args.cacheDir, "cache-dir", cmd.DefaultCacheDir(),
		"Directory for linter cache (greatly improves indexing speed)")
	flag.UintVar(&args.minComplexity, "min-complexity", 20,
		"min function complexity level threshold")
	flag.UintVar(&args.norm, "norm", uint(normFast),
		"code normalization level: 0 for none, 1 for fast-only, 2 for safe, 3 if you feel lucky")

	flag.Parse()

	var exclude *regexp.Regexp
	if args.exclude != "" {
		var err error
		exclude, err = regexp.Compile(args.exclude)
		if err != nil {
			return 1, fmt.Errorf("parse -exclude: %v", err)
		}
	}
	var normLevel normalizationLevel
	switch args.norm {
	case 0:
		normLevel = normNone
	case 1:
		normLevel = normFast
	case 2:
		normLevel = normSafe
	case 3:
		normLevel = normMore
	default:
		return 1, fmt.Errorf("invalid -norm level %d", args.norm)
	}

	targets := flag.Args()

	filter := workspace.NewFilenameFilter(exclude)
	if normLevel > normFast {
		if err := runIndexing(args.cacheDir, targets, filter); err != nil {
			return 1, err
		}
	}
	readFileNamesFunc := workspace.ReadFilenames(targets, filter, []string{"php", "inc", "php5", "phtml"})
	filenamesCh := make(chan workspace.FileInfo, 512)
	go func() {
		readFileNamesFunc(filenamesCh)
		close(filenamesCh)
	}()

	nworkers := runtime.NumCPU()
	results := make([]funcSet, nworkers)
	var wg sync.WaitGroup
	wg.Add(nworkers)
	for i := 0; i < nworkers; i++ {
		go func(workerID int) {
			irConverter := irconv.NewConverter(nil)
			workerResult := make(funcSet)
			for f := range filenamesCh {
				data, err := os.ReadFile(f.Name)
				if err != nil {
					log.Printf("read %s file: %v", f.Name, err)
				}
				if !args.checkAutogen && workspace.FileIsAutoGenerated(data) {
					continue
				}
				root, err := parseutil.ParseFile(data)
				if err != nil {
					log.Printf("parse %s file: %v", f.Name, err)
					continue
				}
				rootIR := irConverter.ConvertRoot(root)

				indexer := &fileIndexer{
					st:           &meta.ClassParseState{},
					funcs:        workerResult,
					fileContents: data,
					filename:     f.Name,
					args:         &args,
					normLevel:    normLevel,
				}
				indexer.CollectFuncs(rootIR)
			}
			results[workerID] = workerResult
			wg.Done()
		}(i)
	}
	wg.Wait()

	allFuncs := make(funcSet)
	for _, workerResult := range results {
		allFuncs.Merge(workerResult)
	}

	printDuplicates(allFuncs)

	return 0, nil
}

type arguments struct {
	checkPrivate  bool
	checkAutogen  bool
	exclude       string
	cacheDir      string
	minComplexity uint
	norm          uint
}
