package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// BlameResult is the result of git blame operation
type BlameResult struct {
	Lines map[int]string // map from line number to SHA1 commit
}

// Blame returns lines that have changed in commit range specified by refSpec with respected commits and line numbers
func Blame(gitDir string, refspec []string, filename string) (BlameResult, error) {
	args := make([]string, 0, 6+len(refspec))
	args = append(args, "--git-dir="+gitDir, "--no-pager", "blame", "--abbrev=40")
	args = append(args, refspec...)
	args = append(args, "--", filename)

	cmd := exec.Command("git", args...)
	defer cmd.Wait()

	out, err := cmd.Output()
	if err != nil {
		return BlameResult{}, err
	}

	var res = BlameResult{Lines: make(map[int]string)}

	lines := strings.Split(string(out), "\n")
	var lineNum int
	for _, ln := range lines {
		lineNum++

		if ln == "" {
			continue
		}

		idx := strings.IndexByte(ln, ' ')
		if idx < 0 {
			return BlameResult{}, fmt.Errorf("Bad blame line: %s", ln)
		}

		commit := ln[0:idx]

		if len(commit) < CommitHashLen {
			return BlameResult{}, fmt.Errorf("Too short commit: %s", commit)
		}

		if commit[0] == '^' {
			continue
		}

		res.Lines[lineNum] = commit
	}

	return res, nil
}
