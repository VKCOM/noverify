package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// UntrackedFiles returns filenames of untracked files obtained with git ls-files command.
func UntrackedFiles(gitDir string) ([]string, error) {
	args := []string{
		"--git-dir=" + gitDir,
		"ls-files",
		"--others",
		"--exclude-standard",
	}

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, string(out))
	}

	filenames := strings.Split(strings.TrimSpace(string(out)), "\n")
	return filenames, nil
}
