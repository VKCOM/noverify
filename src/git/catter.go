package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

// Object is a type of object and it's contents
type Object struct {
	Type     string
	Contents []byte
}

// ObjectCatter is used to get objects from git, fast
type ObjectCatter struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	err    error
}

// NewCatter spawns git so that we can do "cat-file" for objects pretty fast.
func NewCatter(gitDir string) (*ObjectCatter, error) {
	cmd := exec.Command("git", "--git-dir="+gitDir, "cat-file", "--batch")

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	res := &ObjectCatter{
		cmd:    cmd,
		stdin:  stdinPipe,
		stdout: bufio.NewReader(stdoutPipe),
	}

	go func() {
		err := cmd.Wait()
		res.mu.Lock()
		res.err = err
		res.mu.Unlock()
	}()

	return res, nil
}

func (o *ObjectCatter) Error() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.err
}

// Get returns object with its type and contents.
//
// <sha1> SP <type> SP <size> LF
// <contents> LF
func (o *ObjectCatter) Get(sha1 string) (*Object, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	_, err := o.stdin.Write([]byte(sha1 + "\n"))
	if err != nil {
		return nil, err
	}

	ln, isPrefix, err := o.stdout.ReadLine()
	if isPrefix {
		return nil, errors.New("Unexpected too long line from git")
	}

	if err != nil {
		return nil, err
	}

	parts := bytes.Fields(ln)
	if len(parts) < 3 {
		return nil, errors.New("Got bad line, perhaps object is missing")
	}

	size, err := strconv.Atoi(string(parts[2]))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, size+1)
	_, err = io.ReadFull(o.stdout, buf)
	if err != nil {
		return nil, err
	}

	return &Object{
		Type:     string(parts[1]),
		Contents: buf[0 : len(buf)-1],
	}, nil
}

// Walk traverses tree object treeSHA1 and calls cb() upon encountering any blob that matches filenameFilter()
func (o *ObjectCatter) Walk(dir string, treeSHA1 string, filenameFilter func(filename []byte) bool, cb func(filename string, contents []byte)) error {
	obj, err := o.Get(treeSHA1)
	if err != nil {
		return err
	}

	const (
		fileTyp = '1'
		dirTyp  = '4'
		fileLen = 5
		dirLen  = 4

		shaLen = CommitHashLen / 2 // raw length is 2 times less
	)

	var typeBuf []byte
	var filenameBuf []byte
	pos := 0

	for pos < len(obj.Contents) {
		typ := obj.Contents[pos]
		pos++
		switch typ {
		case fileTyp:
			typeBuf = obj.Contents[pos : pos+fileLen]
			pos += fileLen
		case dirTyp:
			typeBuf = obj.Contents[pos : pos+dirLen]
			pos += dirLen
		default:
			return fmt.Errorf("Unknown typ: %c", typ)
		}
		pos++ // space

		nameLen := bytes.IndexByte(obj.Contents[pos:], 0)
		filename := obj.Contents[pos : pos+nameLen]
		pos += nameLen + 1 // nul byte

		sha := obj.Contents[pos : pos+shaLen]
		pos += shaLen

		if typ == fileTyp {
			filenameBuf = filenameBuf[0:0]
			if dir == "" {
				filenameBuf = append(filenameBuf, filename...)
			} else {
				filenameBuf = append(filenameBuf, dir...)
				filenameBuf = append(filenameBuf, filename...)
			}

			if !filenameFilter(filenameBuf) {
				continue
			}

			fileObj, err := o.Get(fmt.Sprintf("%x", sha))
			if err != nil {
				return fmt.Errorf("Error getting object for file %s: %s", filename, err.Error())
			}

			cb(string(filenameBuf), fileObj.Contents)
		} else if typ == dirTyp {
			var filePath string
			if dir == "" {
				filePath = string(filename) + string(os.PathSeparator)
			} else {
				filePath = dir + string(filename) + string(os.PathSeparator)
			}
			o.Walk(filePath, fmt.Sprintf("%x", sha), filenameFilter, cb)
		}
	}

	_ = typeBuf

	return nil
}
