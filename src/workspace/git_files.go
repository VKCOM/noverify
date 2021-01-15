package workspace

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/monochromegane/go-gitignore"

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/meta"
)

// ParseGitignoreFromDir tries to parse a gitignore file at path/.gitignore.
// If no such file exists, <nil, nil> is returned.
func ParseGitignoreFromDir(path string) (gitignore.IgnoreMatcher, error) {
	f, err := os.Open(filepath.Join(path, ".gitignore"))
	switch {
	case os.IsNotExist(err):
		return nil, nil // No gitignore file, not an error
	case err != nil:
		return nil, err // Some unexpected error (e.g. access failure)
	}
	defer f.Close()
	matcher := gitignore.NewGitIgnoreFromReader(path, f)
	return matcher, nil
}

// ReadChangesFromWorkTree returns callback that reads files from workTree dir that are changed
func ReadChangesFromWorkTree(dir string, changes []git.Change) ReadCallback {
	return func(ch chan FileInfo) {
		for _, c := range changes {
			if c.Type == git.Deleted {
				continue
			}

			if !isPHPExtension(c.NewName) {
				continue
			}

			filename := filepath.Join(dir, c.NewName)

			contents, err := ioutil.ReadFile(filename)
			if err != nil {
				log.Fatalf("Could not read file %s: %s", filename, err.Error())
			}

			ch <- FileInfo{
				Name:     filename,
				Contents: contents,
			}
		}
	}
}

// ReadFilesFromGit parses file contents in the specified commit
func ReadFilesFromGit(repo, commitSHA1 string, ignoreRegex *regexp.Regexp) ReadCallback {
	catter, err := git.NewCatter(repo)
	if err != nil {
		log.Fatalf("Could not start catter: %s", err.Error())
	}

	tree, err := git.GetTreeSHA1(catter, commitSHA1)
	if err != nil {
		log.Fatalf("Could not get tree sha1: %s", err.Error())
	}

	suffixes := makePHPExtensionSuffixes()

	return func(ch chan FileInfo) {
		start := time.Now()
		idx := 0

		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				return isPHPExtensionBytes(filename, suffixes)
			},
			func(filename string, contents []byte) {
				idx++
				if time.Since(start) >= 2*time.Second {
					start = time.Now()
					action := "Indexed"
					if meta.IsIndexingComplete() {
						action = "Analyzed"
					}
					log.Printf("%s %d files from git", action, idx)
				}

				if ignoreRegex != nil && ignoreRegex.MatchString(filename) {
					return
				}

				ch <- FileInfo{
					Name:     filename,
					Contents: contents,
				}
			},
		)

		if err != nil {
			log.Fatalf("Could not walk: %s", err.Error())
		}
	}
}

// ReadOldFilesFromGit parses file contents in the specified commit, the old version
func ReadOldFilesFromGit(repo, commitSHA1 string, changes []git.Change) ReadCallback {
	changedMap := make(map[string][]git.LineRange, len(changes))
	for _, ch := range changes {
		if ch.Type == git.Added {
			continue
		}
		changedMap[ch.OldName] = append(changedMap[ch.OldName], ch.OldLineRanges...)
	}

	catter, err := git.NewCatter(repo)
	if err != nil {
		log.Fatalf("Could not start catter: %s", err.Error())
	}

	tree, err := git.GetTreeSHA1(catter, commitSHA1)
	if err != nil {
		log.Fatalf("Could not get tree sha1: %s", err.Error())
	}

	suffixes := makePHPExtensionSuffixes()

	return func(ch chan FileInfo) {
		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				if !isPHPExtensionBytes(filename, suffixes) {
					return false
				}

				_, ok := changedMap[string(filename)]
				return ok
			},
			func(filename string, contents []byte) {
				ch <- FileInfo{
					Name:       filename,
					Contents:   contents,
					LineRanges: changedMap[filename],
				}
			},
		)

		if err != nil {
			log.Fatalf("Could not walk: %s", err.Error())
		}
	}
}

// ReadFilesFromGitWithChanges parses file contents in the specified commit, but only specified ranges
func ReadFilesFromGitWithChanges(repo, commitSHA1 string, changes []git.Change) ReadCallback {
	changedMap := make(map[string][]git.LineRange, len(changes))
	for _, ch := range changes {
		if ch.Type == git.Deleted {
			// TODO: actually support deletes too
			continue
		}

		changedMap[ch.NewName] = append(changedMap[ch.NewName], ch.LineRanges...)
	}

	catter, err := git.NewCatter(repo)
	if err != nil {
		log.Fatalf("Could not start catter: %s", err.Error())
	}

	tree, err := git.GetTreeSHA1(catter, commitSHA1)
	if err != nil {
		log.Fatalf("Could not get tree sha1: %s", err.Error())
	}

	suffixes := makePHPExtensionSuffixes()

	return func(ch chan FileInfo) {
		err = catter.Walk(
			"",
			tree,
			func(filename []byte) bool {
				if !isPHPExtensionBytes(filename, suffixes) {
					return false
				}

				_, ok := changedMap[string(filename)]
				return ok
			},
			func(filename string, contents []byte) {
				ch <- FileInfo{
					Name:       filename,
					Contents:   contents,
					LineRanges: changedMap[filename],
				}
			},
		)

		if err != nil {
			log.Fatalf("Could not walk: %s", err.Error())
		}
	}
}

func makePHPExtensionSuffixes() [][]byte {
	res := make([][]byte, 0, len(PHPExtensions))
	for _, ext := range PHPExtensions {
		res = append(res, []byte("."+ext))
	}
	return res
}

func isPHPExtensionBytes(filename []byte, suffixes [][]byte) bool {
	for _, suffix := range suffixes {
		if bytes.HasSuffix(filename, suffix) {
			return true
		}
	}

	return false
}
