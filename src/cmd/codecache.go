package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
)

func visitedCachePath() string {
	if linter.CacheDir == "" {
		return ""
	}

	return filepath.Join(linter.CacheDir, "visited-cache.json")
}

func readVisitedCache() map[string]linter.StatCacheEntry {
	cachePath := visitedCachePath()
	if cachePath == "" {
		return nil
	}

	fp, err := os.Open(cachePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to open cache file: %v", err)
		}
		return nil
	}
	defer fp.Close()

	var res map[string]linter.StatCacheEntry
	if err := json.NewDecoder(fp).Decode(&res); err != nil {
		log.Printf("Failed decoding visited cache: %v", err)
		return nil
	}

	return res
}

func writeVisitedCache(cache map[string]linter.StatCacheEntry) {
	cachePath := visitedCachePath()
	if cachePath == "" {
		return
	}

	os.MkdirAll(filepath.Dir(cachePath), 0777)

	fp, err := os.OpenFile(cachePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Failed to open cache file: %v", err)
		return
	}
	defer fp.Close()

	if err := json.NewEncoder(fp).Encode(cache); err != nil {
		log.Printf("Failed writing visited cache: %v", err)
		return
	}
}

func dumpCodeCache(dir string, packagePrefix string) (exitCode int, err error) {
	log.Printf("Dumping code cache into %s", dir)

	if err := os.MkdirAll(dir, 0777); err != nil {
		return 2, err
	}

	type funcInfo struct {
		name      string
		cacheFile string
	}

	allFuncs := make(map[string][]funcInfo)

	for filename, cache := range meta.Info.GetAllPerFileCaches() {
		h := md5.New()
		h.Write([]byte(filename))
		filenameHash := fmt.Sprintf("%x", h.Sum(nil))

		goFileDir := "p" + filenameHash[0:2]
		goFileName := filepath.Join(dir, goFileDir, filenameHash+".go")
		goFuncName := "Load_" + filenameHash

		if err := os.MkdirAll(filepath.Dir(goFileName), 0777); err != nil {
			return 2, err
		}
		allFuncs[goFileDir] = append(allFuncs[goFileDir], funcInfo{name: goFuncName, cacheFile: cache.CacheFilename})

		fp, err := os.OpenFile(goFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			return 2, err
		}

		_, err = fmt.Fprintf(fp, `
package %s

import "github.com/VKCOM/noverify/src/meta"

// %s returns cache for file %s
func %s() *meta.PerFileCache {
	return &meta.PerFileCache{
		Scope: %#v,
		Classes: %#v,
		Constants: %#v,
		FunctionOverrides: %#v,
		Functions: %#v,
		Traits: %#v,
	}
}

`, goFileDir, goFuncName, filename, goFuncName, cache.Scope, cache.Classes, cache.Constants, cache.FunctionOverrides, cache.Functions, cache.Traits)

		if err != nil {
			return 2, err
		}

		if err := fp.Close(); err != nil {
			return 2, err
		}
	}

	globalIndexFile, err := os.OpenFile(filepath.Join(dir, "codecache.go"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return 2, err
	}
	defer globalIndexFile.Close()

	if _, err := fmt.Fprintf(globalIndexFile, "package noverifycache\n\nimport \"github.com/VKCOM/noverify/src/meta\"\n"); err != nil {
		return 2, err
	}

	for pkg := range allFuncs {
		if _, err := fmt.Fprintf(globalIndexFile, "import %q\n", packagePrefix+"/"+pkg); err != nil {
			return 2, err
		}
	}

	if _, err := fmt.Fprintf(globalIndexFile, "\nvar CodeCacheMap = map[string]func()*meta.PerFileCache{\n"); err != nil {
		return 2, err
	}

	for pkg, funcs := range allFuncs {
		for _, f := range funcs {
			if _, err := fmt.Fprintf(globalIndexFile, "%q: %s,\n", f.cacheFile, pkg+"."+f.name); err != nil {
				return 2, err
			}
		}
	}

	if _, err := fmt.Fprintf(globalIndexFile, "}\n"); err != nil {
		return 2, err
	}

	return 0, nil
}
