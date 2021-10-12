package git

import (
	"bytes"
	"errors"
	"log"
)

// MergeBase computes merge base between commits one and two
func MergeBase(gitDir string, one, two string) (res string, err error) {
	out, err := execOutput("git", "--git-dir="+gitDir, "merge-base", one, two)
	if err != nil {
		return "", err
	}

	out = bytes.TrimSpace(out)

	if len(out) != CommitHashLen {
		return "", errors.New("too short hash")
	}

	log.Printf("Merge base between %s and %s is %s", one, two, out)

	return string(out), nil
}
