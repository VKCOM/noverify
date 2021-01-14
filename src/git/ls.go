package git

import (
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

	out, err := execOutput("git", args...)
	if err != nil {
		return nil, err
	}

	filenames := strings.Split(strings.TrimSpace(string(out)), "\n")
	return filenames, nil
}
