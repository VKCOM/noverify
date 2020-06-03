package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
)

func gitParseUntracked() []*linter.Report {
	if !gitIncludeUntracked {
		return nil
	}

	filenames, err := git.UntrackedFiles(gitRepo)
	if err != nil {
		log.Fatalf("get untracked files: %v", err)
	}

	return linter.ParseFilenames(linter.ReadFilenames(filenames, nil))
}

func parseIndexOnlyFiles() {
	if indexOnlyFiles == "" {
		return
	}
	filenames := strings.Split(indexOnlyFiles, ",")
	linter.ParseFilenames(linter.ReadFilenames(filenames, nil))
}

// Not the best name, and not the best function signature.
// Refactor this function whenever you get the idea how to separate logic better.
func gitRepoComputeReportsFromCommits(logArgs, diffArgs []string) (oldReports, reports []*linter.Report, changes []git.Change, changeLog []git.Commit, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	start := time.Now()
	changeLog, err := git.Log(gitRepo, logArgs)
	if err != nil {
		log.Panicf("Could not get commits in range %+v: %s", logArgs, err.Error())
	}

	if shouldRun := analyzeGitAuthorsWhiteList(changeLog); !shouldRun {
		return nil, nil, nil, nil, false
	}

	changes, err = git.Diff(gitRepo, "", diffArgs)
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if gitFullDiff {
		meta.ResetInfo()
		if err := loadEmbeddedStubs(); err != nil {
			log.Panicf("Load embedded stubs: %v", err)
		}

		start = time.Now()
		linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitFrom, nil))
		parseIndexOnlyFiles()
		log.Printf("Indexed old commit in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		oldReports = linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitFrom, linter.ExcludeRegex))
		log.Printf("Parsed old commit for %s (%d reports)", time.Since(start), len(oldReports))

		meta.ResetInfo()
		if err := loadEmbeddedStubs(); err != nil {
			log.Panicf("Load embedded stubs: %v", err)
		}

		start = time.Now()
		linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitTo, nil))
		log.Printf("Indexed new commit in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		reports = linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitTo, linter.ExcludeRegex))
		log.Printf("Parsed new commit in %s (%d reports)", time.Since(start), len(reports))
	} else {
		start = time.Now()
		linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitTo, nil))
		parseIndexOnlyFiles()
		log.Printf("Indexing complete in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		oldReports = linter.ParseFilenames(linter.ReadOldFilesFromGit(gitRepo, gitCommitFrom, changes))
		log.Printf("Parsed old files versions for %s", time.Since(start))

		start = time.Now()
		meta.SetIndexingComplete(false)
		linter.ParseFilenames(linter.ReadFilesFromGitWithChanges(gitRepo, gitCommitTo, changes))
		meta.SetIndexingComplete(true)
		log.Printf("Indexed files versions for %s", time.Since(start))

		start = time.Now()
		reports = linter.ParseFilenames(linter.ReadFilesFromGitWithChanges(gitRepo, gitCommitTo, changes))
		log.Printf("Parsed new file versions in %s", time.Since(start))
	}

	return oldReports, reports, changes, changeLog, true
}

func gitRepoComputeReportsFromLocalChanges() (oldReports, reports []*linter.Report, changes []git.Change, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	if gitWorkTree == "" {
		return nil, nil, nil, false
	}

	// compute changes for working copy (staged + unstaged changes combined starting with the commit being pushed)
	changes, err := git.Diff(gitRepo, gitWorkTree, []string{gitCommitFrom})
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if len(changes) == 0 {
		return nil, nil, nil, false
	}

	log.Printf("You have changes in your work tree, showing diff between %s and work tree", gitCommitFrom)

	start := time.Now()
	linter.ParseFilenames(linter.ReadFilesFromGit(gitRepo, gitCommitFrom, nil))
	parseIndexOnlyFiles()
	log.Printf("Indexing complete in %s", time.Since(start))

	meta.SetIndexingComplete(true)

	start = time.Now()
	oldReports = linter.ParseFilenames(linter.ReadOldFilesFromGit(gitRepo, gitCommitFrom, changes))
	log.Printf("Parsed old files versions for %s", time.Since(start))

	start = time.Now()
	meta.SetIndexingComplete(false)
	linter.ParseFilenames(linter.ReadChangesFromWorkTree(gitWorkTree, changes))
	gitParseUntracked()
	meta.SetIndexingComplete(true)
	log.Printf("Indexed new files versions for %s", time.Since(start))

	start = time.Now()
	reports = linter.ParseFilenames(linter.ReadChangesFromWorkTree(gitWorkTree, changes))
	reports = append(reports, gitParseUntracked()...)
	log.Printf("Parsed new file versions in %s", time.Since(start))

	return oldReports, reports, changes, true
}

func gitMain() (int, error) {
	var (
		oldReports, reports []*linter.Report
		diffArgs            []string
		changes             []git.Change
		changeLog           []git.Commit
		ok                  bool
	)

	// prepareGitArgs also populates global variables like fromCommit
	logArgs, diffArgs, err := prepareGitArgs()
	if err != nil {
		return 0, err
	}

	oldReports, reports, changes, ok = gitRepoComputeReportsFromLocalChanges()
	if !ok {
		oldReports, reports, changes, changeLog, ok = gitRepoComputeReportsFromCommits(logArgs, diffArgs)
		if !ok {
			return 0, nil
		}
	}

	start := time.Now()
	diff, err := linter.DiffReports(gitRepo, diffArgs, changes, changeLog, oldReports, reports, 8)
	if err != nil {
		return 0, fmt.Errorf("Could not compute reports diff: %v", err)
	}
	log.Printf("Computed reports diff for %s", time.Since(start))

	criticalReports := analyzeReports(diff)

	if criticalReports > 0 {
		log.Printf("Found %d critical issues, please fix them.", criticalReports)
		return 2, nil
	}
	log.Printf("No critical issues found. Your code is perfect.")
	return 0, nil
}

func analyzeGitAuthorsWhiteList(changeLog []git.Commit) (shouldRun bool) {
	if gitAuthorsWhitelist != "" {
		whiteList := make(map[string]bool)
		for _, name := range strings.Split(gitAuthorsWhitelist, ",") {
			whiteList[name] = true
		}

		for _, commit := range changeLog {
			if !whiteList[commit.Author] {
				log.Printf("Found commit from '%s', PHP linter not running", commit.Author)
				return false
			}
		}
	}

	return true
}

func prepareGitArgs() (logArgs, diffArgs []string, err error) {
	if gitPushArg != "" {
		args := strings.Fields(gitPushArg)
		if len(args) != 3 {
			return nil, nil, fmt.Errorf("Unexpected format of push arguments, expected only 3 columns: %s", gitPushArg)
		}
		gitCommitFrom, gitCommitTo, gitRef = args[0], args[1], args[2]
	}

	if gitCommitFrom == git.Zero {
		gitCommitFrom = "master"
	}

	if !gitSkipFetch {
		start := time.Now()
		log.Printf("Fetching origin master to ORIGIN_MASTER")
		if err := git.Fetch(gitRepo, "master", "ORIGIN_MASTER"); err != nil {
			return nil, nil, fmt.Errorf("Could not fetch ORIGIN_MASTER: %v", err.Error())
		}
		log.Printf("Fetched for %s", time.Since(start))
	}

	if !gitDisableCompensateMaster {

		fromAndMaster, err := git.MergeBase(gitRepo, "ORIGIN_MASTER", gitCommitFrom)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", gitCommitFrom)
		}

		toAndMaster, err := git.MergeBase(gitRepo, "ORIGIN_MASTER", gitCommitTo)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", gitCommitTo)
		}

		// check if master was merged in between the commits
		if fromAndMaster != toAndMaster {
			gitCommitFrom = toAndMaster
		}
	}

	logArgs = []string{gitCommitFrom + ".." + gitCommitTo}
	diffArgs = []string{gitCommitFrom + ".." + gitCommitTo}

	return logArgs, diffArgs, nil
}
