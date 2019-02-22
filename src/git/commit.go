package git

import (
	"bytes"
	"errors"
	"strings"
)

const (
	CommitHashLen = 40
	Zero          = "0000000000000000000000000000000000000000"
)

// GetTreeSHA1 returns SHA1 of tree object for the commit
func GetTreeSHA1(catter *ObjectCatter, commitSHA1 string) (string, error) {
	obj, err := catter.Get(commitSHA1)
	if err != nil {
		return "", err
	}

	lines := bytes.SplitN(obj.Contents, []byte("\n"), 2)
	if !bytes.HasPrefix(lines[0], []byte("tree ")) {
		return "", errors.New("Unexpected commit object format")
	}

	return strings.TrimPrefix(string(lines[0]), "tree "), nil
}
