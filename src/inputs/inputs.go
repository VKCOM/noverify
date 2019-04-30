package inputs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// SourceInput interface describes linter source code inputs that can come from
// files and buffers.
//
// One of the responsibilities of the SourceInput is to return readers
// that can be treated like they always have UTF-8 encoded data.
// If real data has different encoding, it's up to the implementation
// to detect and fix that.
type SourceInput interface {
	NewReader(filename string) (ReadCloseSizer, error)
	NewBytesReader(filename string, data []byte) (ReadCloseSizer, error)
}

// ReadCloseSizer is the interface that groups io.ReadCloser and Size methods.
type ReadCloseSizer interface {
	io.ReadCloser
	Size() int
}

// NewReadCloseSizer turns ReadCloser into ReadCloserSizer by adding Size method
// that always returns initially bound size.
func NewReadCloseSizer(r io.ReadCloser, size int) ReadCloseSizer {
	return readCloserSizer{
		ReadCloser: r,
		size:       size,
	}
}

// NewDefaultSourceInput returns the default SourceInput implementation that
// operates with filesystem directly.
//
// It assumes that all sources are UTF-8 encoded.
func NewDefaultSourceInput() SourceInput {
	return defaultSourceInput{}
}

type defaultSourceInput struct{}

func (defaultSourceInput) NewReader(filename string) (ReadCloseSizer, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %v", filename, err)
	}

	st, err := fp.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat file %s: %v", filename, err)
	}

	size := int(st.Size())
	return NewReadCloseSizer(fp, size), nil
}

func (defaultSourceInput) NewBytesReader(filename string, data []byte) (ReadCloseSizer, error) {
	return NewReadCloseSizer(ioutil.NopCloser(bytes.NewReader(data)), len(data)), nil
}

type readCloserSizer struct {
	io.ReadCloser
	size int
}

func (r readCloserSizer) Size() int {
	return r.size
}
