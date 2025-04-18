package linter

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/VKCOM/noverify/src/meta"
)

// cacheVersions is a magic number that helps to distinguish incompatible caches.
//
// Version log:
//
//	27 - added Static field to meta.FuncInfo
//	28 - array type parsed as mixed[]
//	29 - updated type inference for ClassConstFetch
//	30 - resolve ClassConstFetch to a wrapped type string
//	31 - fixed plus operator type inference for arrays
//	32 - replaced Static:bool with Flags:uint8 in meta.FuncInfo
//	33 - support parsing of array<k,v> and list<type>
//	34 - support parsing of ?ClassName as "ClassName|null"
//	35 - added Flags:uint8 to meta.ClassInfo
//	36 - added FuncAbstract bit to FuncFlags
//	     added FuncFinal bit to FuncFlags
//	     added ClassFinal bit to ClassFlags
//	     FuncInfo now stores original function name
//	     ClassInfo now stores original class name
//	37 - added ClassShape bit to ClassFlags
//	     changed meta.scopeVar bool fields representation
//	38 - replaced TypesMap.immutable:bool with flags:uint8.
//	     added mapPrecise flag to mark precise type maps.
//	39 - added new field Value in ConstantInfo
//	40 - changed string const value storage (no quotes)
//	41 - const-folding affected const definition values
//	42 - bool-typed consts are now stored in meta info
//	43 - define'd const values stored in cache
//	44 - rename ConstantInfo => ConstInfo
//	45 - added Mixins field to meta.ClassInfo
//	46 - changed the way of inferring the return type of functions and methods
//	47 - forced cache version invalidation due to the #921
//	48 - renamed meta.TypesMap to types.Map; this affects gob encoding
//	49 - for shape, names are now generated using the keys that make up this shape
//	50 - added Flags field for meta.PropertyInfo
//	51 - added anonymous classes
//	52 - renamed all PhpDoc and Phpdoc with PHPDoc
//	53 - added DeprecationInfo for functions and methods and support for some attributes
//	54 - forced cache version invalidation due to the #1165
//	55 - updated go version 1.16 -> 1.21
//	56 - added isVariadic to meta.FuncInfo
//	57 - added DeprecationInfo for property and const
const cacheVersion = 57

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

func writeMetaCache(w *bufio.Writer, root *rootWalker) error {
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

func createMetaCacheFile(filename, cacheFile string, root *rootWalker) error {
	tmpPath := cacheFile + ".tmp"
	if err := os.MkdirAll(filepath.Dir(tmpPath), 0777); err != nil {
		return err
	}

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
	updateMetaInfo(root.ctx.st.Info, filename, &root.meta)
	return nil
}

func readMetaCache(r io.Reader, cachers []MetaCacher, filename string, dst *fileMeta) error {
	bufrd := bufio.NewReader(r)
	if err := readMetaCacheHeader(cachers, bufrd); err != nil {
		return err
	}

	dec := gob.NewDecoder(bufrd)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := customCachersDecode(cachers, filename, bufrd); err != nil {
		return err
	}
	return nil
}

func restoreMetaFromCache(info *meta.Info, cachers []MetaCacher, filename string, rd io.Reader) error {
	var m fileMeta
	if err := readMetaCache(rd, cachers, filename, &m); err != nil {
		return err
	}

	updateMetaInfo(info, filename, &m)
	return nil
}

func updateMetaInfo(info *meta.Info, filename string, m *fileMeta) {
	if info.IsIndexingComplete() {
		panic("Trying to update meta info when not indexing")
	}

	info.Lock()
	defer info.Unlock()

	info.DeleteMetaForFileNonLocked(filename)

	info.AddFilenameNonLocked(filename)
	info.AddClassesNonLocked(filename, m.Classes)
	info.AddTraitsNonLocked(filename, m.Traits)
	info.AddFunctionsNonLocked(filename, m.Functions)
	info.AddConstantsNonLocked(filename, m.Constants)
	info.AddFunctionsOverridesNonLocked(filename, m.FunctionOverrides)

	if m.Scope != nil {
		info.AddToGlobalScopeNonLocked(filename, m.Scope)
	}
}

func writeMetaCacheHeader(wr *bufio.Writer, root *rootWalker) error {
	if err := wr.WriteByte(cacheVersion); err != nil {
		return err
	}

	for i := range root.custom {
		cacher := root.config.Checkers.cachers[i]
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

func customCachersEncode(wr *bufio.Writer, root *rootWalker) error {
	for i, c := range root.custom {
		cacher := root.config.Checkers.cachers[i]
		if cacher == nil {
			continue
		}
		if err := cacher.Encode(wr, c); err != nil {
			return err
		}
	}

	return nil
}

func readMetaCacheHeader(cachers []MetaCacher, rd *bufio.Reader) error {
	ver, err := rd.ReadByte()
	if err != nil {
		return err
	}

	if ver != cacheVersion {
		return errWrongVersion
	}

	var versionBuf [256]byte
	for _, cacher := range cachers {
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

func customCachersDecode(cachers []MetaCacher, filename string, rd *bufio.Reader) error {
	for _, cacher := range cachers {
		if cacher == nil {
			continue
		}
		if err := cacher.Decode(rd, filename); err != nil {
			return err
		}
	}

	return nil
}
