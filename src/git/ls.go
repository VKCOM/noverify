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

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}

	return strings.Split(trimmed, "\n"), nil
}

// TopLevel returns the absolute path of the working tree root
// that is associated with the given git directory.
func TopLevel(gitDir string) (string, error) {
	args := []string{
		"--git-dir=" + gitDir,
		"rev-parse",
		"--show-toplevel",
	}

	out, err := execOutput("git", args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
