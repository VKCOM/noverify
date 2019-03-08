package linter

import (
	"bufio"
	"crypto/md5"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/VKCOM/noverify/src/meta"
)

const cacheVersion = 24

var (
	errWrongVersion = errors.New("Wrong cache version")

	initFileReadTime  int64
	initCacheReadTime int64
)

type fileMeta struct {
	Scope             *meta.Scope
	Classes           meta.ClassesMap
	Traits            meta.ClassesMap
	Functions         meta.FunctionsMap
	Constants         meta.ConstantsMap
	FunctionOverrides meta.FunctionsOverrideMap
}

// Parse file and fill in the meta info. Can use cache.
func Parse(filename string, contents []byte, encoding string) error {
	if CacheDir == "" {
		_, w, err := ParseContents(filename, contents, encoding, nil)
		if w != nil {
			updateMetaInfo(filename, &w.meta)
		}
		return err
	}

	h := md5.New()

	if contents == nil {
		start := time.Now()
		fp, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fp.Close()
		if _, err := io.Copy(h, fp); err != nil {
			return err
		}
		atomic.AddInt64(&initFileReadTime, int64(time.Since(start)))
	} else {
		h.Write(contents)
	}

	contentsHash := fmt.Sprintf("%x", h.Sum(nil))

	cacheFilenamePart := filename

	// windows user supplied full path to directory to be analyzed,
	// but windows paths does not support ":" in the middle
	if len(filename) > 2 && filename[0] >= 'A' && filename[0] <= 'Z' && filename[1] == ':' {
		cacheFilenamePart = string(filename[0:1]) + "_" + string(filename[2:])
	}

	cacheFile := filepath.Join(CacheDir, cacheFilenamePart+"."+contentsHash)

	start := time.Now()
	fp, err := os.Open(cacheFile)
	if err != nil {
		_, w, err := ParseContents(filename, contents, encoding, nil)
		if err != nil {
			return err
		}

		return createMetaCacheFile(filename, cacheFile, &w.meta)
	}
	defer fp.Close()

	if err := restoreMetaFromCache(filename, fp); err != nil {
		// do not really care about why exactly reading from cache failed
		os.Remove(cacheFile)

		_, w, err := ParseContents(filename, contents, encoding, nil)
		if err != nil {
			return err
		}

		return createMetaCacheFile(filename, cacheFile, &w.meta)
	}

	atomic.AddInt64(&initCacheReadTime, int64(time.Since(start)))
	return nil
}

func createMetaCacheFile(filename, cacheFile string, m *fileMeta) error {
	tmpPath := cacheFile + ".tmp"
	os.MkdirAll(filepath.Dir(tmpPath), 0777)

	// TODO: some kind of file-based locking
	fp, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()
	defer os.Remove(tmpPath)

	wr := bufio.NewWriter(fp)
	if err := wr.WriteByte(cacheVersion); err != nil {
		return err
	}

	enc := gob.NewEncoder(wr)

	if err := enc.Encode(m); err != nil {
		return err
	}

	if err := wr.Flush(); err != nil {
		return err
	}
	
	// Windows clearly does not want to allow to rename unclosed files
	if runtime.GOOS == "windows" {
		fp.Close()
		os.Remove(cacheFile)
	}


	if err := os.Rename(tmpPath, cacheFile); err != nil {
		return err
	}

	// if using cache, this is the only proper place to update meta info:
	// after all cache meta info was successfully written to disk
	updateMetaInfo(filename, m)
	return nil
}

func restoreMetaFromCache(filename string, rd io.Reader) error {
	var m fileMeta

	bufrd := bufio.NewReader(rd)

	ver, err := bufrd.ReadByte()
	if err != nil {
		return err
	}

	if ver != cacheVersion {
		return errWrongVersion
	}

	dec := gob.NewDecoder(bufrd)
	if err := dec.Decode(&m); err != nil {
		return err
	}

	updateMetaInfo(filename, &m)
	return nil
}

func updateMetaInfo(filename string, m *fileMeta) {
	if meta.IsIndexingComplete() {
		panic("Trying to update meta info when not indexing")
	}

	meta.Info.Lock()
	defer meta.Info.Unlock()

	meta.Info.DeleteMetaForFileNonLocked(filename)

	meta.Info.AddFilenameNonLocked(filename)
	meta.Info.AddClassesNonLocked(filename, m.Classes)
	meta.Info.AddTraitsNonLocked(filename, m.Traits)
	meta.Info.AddFunctionsNonLocked(filename, m.Functions)
	meta.Info.AddConstantsNonLocked(filename, m.Constants)
	meta.Info.AddFunctionsOverridesNonLocked(filename, m.FunctionOverrides)

	if m.Scope != nil {
		meta.Info.AddToGlobalScopeNonLocked(filename, m.Scope)
	}
}
