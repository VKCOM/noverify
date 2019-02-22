package git

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
)

// MergeBase computes merge base between commits one and two
func MergeBase(gitDir string, one, two string) (res string, err error) {
	cmd := exec.Command("git", "--git-dir="+gitDir, "merge-base", one, two)
	defer cmd.Wait()

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	out = bytes.TrimSpace(out)

	if len(out) != CommitHashLen {
		return "", errors.New("Too short hash")
	}

	log.Printf("merge base between %s and %s is %s", one, two, out)

	return string(out), nil
}
