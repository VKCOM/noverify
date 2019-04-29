package main

import (
	"bufio"
	"io"
	"log"
	"sync"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	// You can register your own rules here, see src/linter/custom.go

	var fw fileReaderWrapper
	cmd.MainHooks.AfterFlagParse = func() {
		fw.encoding = linter.DefaultEncoding
		linter.WrapFileReader = fw.WrapFileReader
	}
	cmd.Main()
}

func (fw *fileReaderWrapper) WrapFileReader(filename string, r io.ReadCloser) io.ReadCloser {
	if fw.encoding != "windows-1251" {
		return r
	}

	bufRd := fw.bufPool.Get().(*bufio.Reader)
	bufRd.Reset(r)
	return &readCloser{
		Reader: transform.NewReader(r, charmap.Windows1251.NewDecoder()),
		closeFunc: func() error {
			fw.bufPool.Put(bufRd)
			return r.Close()
		},
	}
}

type fileReaderWrapper struct {
	bufPool  sync.Pool
	encoding string
}

type readCloser struct {
	io.Reader
	closeFunc func() error
}

func (rc *readCloser) Close() error {
	return rc.closeFunc()
}
