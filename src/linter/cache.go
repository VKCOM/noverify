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

// cacheVersions is a magic number that helps to distinguish incompatible caches.
//
// Version log:
//     27 - added Static field to meta.FuncInfo
//     28 - array type parsed as mixed[]
//     29 - updated type inference for ClassConstFetch
//     30 - resolve ClassConstFetch to a wrapped type string
//     31 - fixed plus operator type inference for arrays
//     32 - replaced Static:bool with Flags:uint8 in meta.FuncInfo
//     33 - support parsing of array<k,v> and list<type>
//     34 - support parsing of ?ClassName as "ClassName|null"
//     35 - added Flags:uint8 to meta.ClassInfo
//     36 - added FuncAbstract bit to FuncFlags
//          added FuncFinal bit to FuncFlags
//          added ClassFinal bit to ClassFlags
//          FuncInfo now stores original function name
//          ClassInfo now stores original class name
//     37 - added ClassShape bit to ClassFlags
//          changed meta.scopeVar bool fields representation
//     38 - replaced TypesMap.immutable:bool with flags:uint8.
//          added mapPrecise flag to mark precise type maps.
const cacheVersion = 38

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

// IndexFile parses the file and fills in the meta info. Can use cache.
func IndexFile(filename string, contents []byte) error {
	if CacheDir == "" {
		_, w, err := ParseContents(filename, contents, nil)
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

	volumeName := filepath.VolumeName(filename)

	// windows user supplied full path to directory to be analyzed,
	// but windows paths does not support ":" in the middle
	if len(volumeName) == 2 && volumeName[1] == ':' {
		cacheFilenamePart = filename[0:1] + "_" + filename[2:]
	}

	cacheFile := filepath.Join(CacheDir, cacheFilenamePart+"."+contentsHash)

	start := time.Now()
	fp, err := os.Open(cacheFile)
	if err != nil {
		_, w, err := ParseContents(filename, contents, nil)
		if err != nil {
			return err
		}

		return createMetaCacheFile(filename, cacheFile, w)
	}
	defer fp.Close()

	if err := restoreMetaFromCache(filename, fp); err != nil {
		// do not really care about why exactly reading from cache failed
		os.Remove(cacheFile)

		_, w, err := ParseContents(filename, contents, nil)
		if err != nil {
			return err
		}

		return createMetaCacheFile(filename, cacheFile, w)
	}

	atomic.AddInt64(&initCacheReadTime, int64(time.Since(start)))
	return nil
}

func writeMetaCache(w *bufio.Writer, root *RootWalker) error {
	if err := writeMetaCacheHeader(w, root); err != nil {
		return err
	}
	enc := gob.NewEncoder(w)
	if err := enc.Encode(&root.meta); err != nil {
		return err
	}
	if err := customCachersEncode(w, root); err != nil {
		return err
	}
	return nil
}

func createMetaCacheFile(filename, cacheFile string, root *RootWalker) error {
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
	if err := writeMetaCache(wr, root); err != nil {
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
	updateMetaInfo(filename, &root.meta)
	return nil
}

func readMetaCache(r io.Reader, filename string, dst *fileMeta) error {
	bufrd := bufio.NewReader(r)
	if err := readMetaCacheHeader(bufrd); err != nil {
		return err
	}

	dec := gob.NewDecoder(bufrd)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := customCachersDecode(filename, bufrd); err != nil {
		return err
	}
	return nil
}

func restoreMetaFromCache(filename string, rd io.Reader) error {
	var m fileMeta
	if err := readMetaCache(rd, filename, &m); err != nil {
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

func writeMetaCacheHeader(wr *bufio.Writer, root *RootWalker) error {
	if err := wr.WriteByte(cacheVersion); err != nil {
		return err
	}

	for i := range root.custom {
		cacher := metaCachers[i]
		if cacher == nil {
			continue
		}
		ver := cacher.Version()
		if len(ver) > 256 {
			return fmt.Errorf("cacher version %q is too long (%d bytes)", ver, len(ver))
		}
		if err := wr.WriteByte(byte(len(ver))); err != nil {
			return fmt.Errorf("write cacher version %q len: %v", ver, err)
		}
		if _, err := wr.WriteString(ver); err != nil {
			return fmt.Errorf("write cacher version %q: %v", ver, err)
		}
	}

	return nil
}

func customCachersEncode(wr *bufio.Writer, root *RootWalker) error {
	for i, c := range root.custom {
		cacher := metaCachers[i]
		if cacher == nil {
			continue
		}
		if err := cacher.Encode(wr, c); err != nil {
			return err
		}
	}

	return nil
}

func readMetaCacheHeader(rd *bufio.Reader) error {
	ver, err := rd.ReadByte()
	if err != nil {
		return err
	}

	if ver != cacheVersion {
		return errWrongVersion
	}

	var versionBuf [256]byte
	for _, cacher := range metaCachers {
		if cacher == nil {
			continue
		}

		versionLen, err := rd.ReadByte()
		if err != nil {
			return err
		}
		if _, err := rd.Read(versionBuf[:versionLen]); err != nil {
			return err
		}
		ver := string(versionBuf[:versionLen])

		if ver != cacher.Version() {
			return errWrongVersion
		}
	}

	return nil
}

func customCachersDecode(filename string, rd *bufio.Reader) error {
	for _, cacher := range metaCachers {
		if cacher == nil {
			continue
		}
		if err := cacher.Decode(rd, filename); err != nil {
			return err
		}
	}

	return nil
}
