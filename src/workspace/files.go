package workspace

import (
	"log"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

type ReadCallback func(ch chan FileInfo)

// ReadFilenames returns callback that reads filenames into channel
func ReadFilenames(filenames []string, filter *FilenameFilter, phpExtensions []string) ReadCallback {
	return func(ch chan FileInfo) {
		for _, filename := range filenames {
			readFilenames(ch, filename, filter, phpExtensions)
		}
	}
}

func readFilenames(ch chan<- FileInfo, filename string, filter *FilenameFilter, phpExtensions []string) {
	absFilename, err := filepath.Abs(filename)
	if err == nil {
		filename = absFilename
	}

	if filter == nil {
		// No-op filter that doesn't track gitignore files.
		filter = &FilenameFilter{}
	}

	// If we use stat here, it will return file info of an entry
	// pointed by a symlink (if filename is a link).
	// lstat is required for a symlink test below to succeed.
	// If we ever want to permit top-level (CLI args) symlinks,
	// caller should resolve them to a files that are pointed by them.
	st, err := os.Lstat(filename)
	if err != nil {
		log.Fatalf("Could not stat file %s: %s", filename, err.Error())
	}
	if st.Mode()&os.ModeSymlink != 0 {
		// filepath.Walk does not follow symlinks, but it does
		// accept it as a root argument without an error.
		// godirwalk.Walk can traverse symlinks with FollowSymbolicLinks=true,
		// but we don't use it. It will give an error if root is
		// a symlink, so we avoid calling Walk() on them.
		return
	}

	if !st.IsDir() {
		if filter.IgnoreFile(filename) {
			return
		}

		if !isPHPExtension(filename, phpExtensions) {
			return
		}

		ch <- FileInfo{
			Name: filename,
		}
		return
	}

	// Start with a sentinel "" path to make last(gitignorePaths) safe
	// without a length check.
	gitignorePaths := []string{""}

	walkOptions := &godirwalk.Options{
		Unsorted: true,

		Callback: func(path string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				if filter.IgnoreDir(path) {
					return filepath.SkipDir
				}
				// During indexing phase and with -gitignore=false
				// we don't want to do extra FS operations.
				if !filter.GitignoreIsEnabled() {
					return nil
				}

				matcher, err := ParseGitignoreFromDir(path)
				if err != nil {
					log.Printf("read .gitignore: %v", err)
				}
				if matcher != nil {
					gitignorePaths = append(gitignorePaths, path)
					filter.GitignorePush(path, matcher)
				}
				return nil
			}

			if !isPHPExtension(path, phpExtensions) {
				return nil
			}
			if filter.IgnoreFile(path) {
				return nil
			}

			ch <- FileInfo{
				Name: path,
			}
			return nil
		},
	}

	if filter.GitignoreIsEnabled() {
		walkOptions.PostChildrenCallback = func(path string, de *godirwalk.Dirent) error {
			topGitignorePath := gitignorePaths[len(gitignorePaths)-1]
			if topGitignorePath == path {
				gitignorePaths = gitignorePaths[:len(gitignorePaths)-1]
				filter.GitignorePop(path)
			}
			return nil
		}
	}

	if err := godirwalk.Walk(filename, walkOptions); err != nil {
		log.Fatalf("Could not walk filepath %s (%v)", filename, err)
	}
}

func isPHPExtension(filename string, phpExtensions []string) bool {
	fileExt := filepath.Ext(filename)
	if fileExt == "" {
		return false
	}

	fileExt = fileExt[len("."):]

	for _, ext := range phpExtensions {
		if fileExt == ext {
			return true
		}
	}

	return false
}
