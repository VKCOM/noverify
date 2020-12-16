package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/workspace"
)

func gitParseUntracked(l *linterRunner) []*linter.Report {
	if !l.args.gitIncludeUntracked {
		return nil
	}

	filenames, err := git.UntrackedFiles(l.args.gitRepo)
	if err != nil {
		log.Fatalf("get untracked files: %v", err)
	}

	return linter.ParseFilenames(workspace.ReadFilenames(filenames, nil), l.allowDisableRegex)
}

func parseIndexOnlyFiles(l *linterRunner) {
	if l.args.indexOnlyFiles == "" {
		return
	}
	filenames := strings.Split(l.args.indexOnlyFiles, ",")
	linter.ParseFilenames(workspace.ReadFilenames(filenames, nil), l.allowDisableRegex)
}

// Not the best name, and not the best function signature.
// Refactor this function whenever you get the idea how to separate logic better.
func gitRepoComputeReportsFromCommits(l *linterRunner, logArgs, diffArgs []string) (oldReports, reports []*linter.Report, changes []git.Change, changeLog []git.Commit, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	changeLog, err := git.Log(l.args.gitRepo, logArgs)
	if err != nil {
		log.Panicf("Could not get commits in range %+v: %s", logArgs, err.Error())
	}

	if shouldRun := analyzeGitAuthorsWhiteList(l, changeLog); !shouldRun {
		return nil, nil, nil, nil, false
	}

	changes, err = git.Diff(l.args.gitRepo, "", diffArgs)
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if l.args.gitFullDiff {
		meta.ResetInfo()
		if err := loadEmbeddedStubs(); err != nil {
			log.Panicf("Load embedded stubs: %v", err)
		}

		start := time.Now()
		linter.ParseFilenames(workspace.ReadFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitFrom, nil), l.allowDisableRegex)
		parseIndexOnlyFiles(l)
		log.Printf("Indexed old commit in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		oldReports = linter.ParseFilenames(workspace.ReadFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitFrom, linter.ExcludeRegex), l.allowDisableRegex)
		log.Printf("Parsed old commit for %s (%d reports)", time.Since(start), len(oldReports))

		meta.ResetInfo()
		if err := loadEmbeddedStubs(); err != nil {
			log.Panicf("Load embedded stubs: %v", err)
		}

		start = time.Now()
		parseIndexOnlyFiles(l)
		linter.ParseFilenames(workspace.ReadFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitTo, nil), l.allowDisableRegex)
		log.Printf("Indexed new commit in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		reports = linter.ParseFilenames(workspace.ReadFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitTo, linter.ExcludeRegex), l.allowDisableRegex)
		log.Printf("Parsed new commit in %s (%d reports)", time.Since(start), len(reports))
	} else {
		start := time.Now()
		linter.ParseFilenames(workspace.ReadFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitTo, nil), l.allowDisableRegex)
		parseIndexOnlyFiles(l)
		log.Printf("Indexing complete in %s", time.Since(start))

		meta.SetIndexingComplete(true)

		start = time.Now()
		oldReports = linter.ParseFilenames(workspace.ReadOldFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitFrom, changes), l.allowDisableRegex)
		log.Printf("Parsed old files versions for %s", time.Since(start))

		start = time.Now()
		meta.SetIndexingComplete(false)
		parseIndexOnlyFiles(l)
		linter.ParseFilenames(workspace.ReadFilesFromGitWithChanges(l.args.gitRepo, l.args.mutable.gitCommitTo, changes), l.allowDisableRegex)
		meta.SetIndexingComplete(true)
		log.Printf("Indexed files versions for %s", time.Since(start))

		start = time.Now()
		reports = linter.ParseFilenames(workspace.ReadFilesFromGitWithChanges(l.args.gitRepo, l.args.mutable.gitCommitTo, changes), l.allowDisableRegex)
		log.Printf("Parsed new file versions in %s", time.Since(start))
	}

	return oldReports, reports, changes, changeLog, true
}

func gitRepoComputeReportsFromLocalChanges(l *linterRunner) (oldReports, reports []*linter.Report, changes []git.Change, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	if l.args.gitWorkTree == "" {
		return nil, nil, nil, false
	}

	// compute changes for working copy (staged + unstaged changes combined starting with the commit being pushed)
	changes, err := git.Diff(l.args.gitRepo, l.args.gitWorkTree, []string{l.args.mutable.gitCommitFrom})
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if len(changes) == 0 {
		return nil, nil, nil, false
	}

	log.Printf("You have changes in your work tree, showing diff between %s and work tree", l.args.mutable.gitCommitFrom)

	start := time.Now()
	linter.ParseFilenames(workspace.ReadFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitFrom, nil), l.allowDisableRegex)
	parseIndexOnlyFiles(l)
	log.Printf("Indexing complete in %s", time.Since(start))

	meta.SetIndexingComplete(true)

	start = time.Now()
	oldReports = linter.ParseFilenames(workspace.ReadOldFilesFromGit(l.args.gitRepo, l.args.mutable.gitCommitFrom, changes), l.allowDisableRegex)
	log.Printf("Parsed old files versions for %s", time.Since(start))

	start = time.Now()
	meta.SetIndexingComplete(false)
	linter.ParseFilenames(workspace.ReadChangesFromWorkTree(l.args.gitWorkTree, changes), l.allowDisableRegex)
	parseIndexOnlyFiles(l)
	gitParseUntracked(l)
	meta.SetIndexingComplete(true)
	log.Printf("Indexed new files versions for %s", time.Since(start))

	start = time.Now()
	reports = linter.ParseFilenames(workspace.ReadChangesFromWorkTree(l.args.gitWorkTree, changes), l.allowDisableRegex)
	reports = append(reports, gitParseUntracked(l)...)
	log.Printf("Parsed new file versions in %s", time.Since(start))

	return oldReports, reports, changes, true
}

func gitMain(l *linterRunner, cfg *MainConfig) (int, error) {
	var (
		oldReports, reports []*linter.Report
		diffArgs            []string
		changes             []git.Change
		changeLog           []git.Commit
		ok                  bool
	)

	// prepareGitArgs also populates global variables like fromCommit
	logArgs, diffArgs, err := prepareGitArgs(l)
	if err != nil {
		return 0, err
	}

	oldReports, reports, changes, ok = gitRepoComputeReportsFromLocalChanges(l)
	if !ok {
		oldReports, reports, changes, changeLog, ok = gitRepoComputeReportsFromCommits(l, logArgs, diffArgs)
		if !ok {
			return 0, nil
		}
	}

	start := time.Now()
	diff, err := linter.DiffReports(l.args.gitRepo, diffArgs, changes, changeLog, oldReports, reports, 8)
	if err != nil {
		return 0, fmt.Errorf("Could not compute reports diff: %v", err)
	}
	log.Printf("Computed reports diff for %s", time.Since(start))

	criticalReports, containsAutofixableReports := analyzeReports(l, cfg, diff)

	if containsAutofixableReports {
		log.Println("Some issues are autofixable (try using the `-fix` flag)")
	}

	if criticalReports > 0 {
		log.Printf("Found %d critical issues, please fix them.", criticalReports)
		return 2, nil
	}
	log.Printf("No critical issues found. Your code is perfect.")
	return 0, nil
}

func analyzeGitAuthorsWhiteList(l *linterRunner, changeLog []git.Commit) (shouldRun bool) {
	if l.args.gitAuthorsWhitelist != "" {
		whiteList := make(map[string]bool)
		for _, name := range strings.Split(l.args.gitAuthorsWhitelist, ",") {
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

func prepareGitArgs(l *linterRunner) (logArgs, diffArgs []string, err error) {
	if l.args.gitPushArg != "" {
		args := strings.Fields(l.args.gitPushArg)
		if len(args) != 3 {
			return nil, nil, fmt.Errorf("Unexpected format of push arguments, expected only 3 columns: %s", l.args.gitPushArg)
		}
		// args[2] is a git ref (branch name), but its unused.
		l.args.mutable.gitCommitFrom, l.args.mutable.gitCommitTo = args[0], args[1]
	}
	if l.args.mutable.gitCommitFrom == git.Zero {
		l.args.mutable.gitCommitFrom = "master"
	}

	if !l.args.gitSkipFetch {
		start := time.Now()
		log.Printf("Fetching origin master to ORIGIN_MASTER")
		if err := git.Fetch(l.args.gitRepo, "master", "ORIGIN_MASTER"); err != nil {
			return nil, nil, fmt.Errorf("Could not fetch ORIGIN_MASTER: %v", err.Error())
		}
		log.Printf("Fetched for %s", time.Since(start))
	}

	if !l.args.gitDisableCompensateMaster {
		fromAndMaster, err := git.MergeBase(l.args.gitRepo, "ORIGIN_MASTER", l.args.mutable.gitCommitFrom)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", l.args.mutable.gitCommitFrom)
		}

		toAndMaster, err := git.MergeBase(l.args.gitRepo, "ORIGIN_MASTER", l.args.mutable.gitCommitTo)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", l.args.mutable.gitCommitTo)
		}

		// check if master was merged in between the commits
		if fromAndMaster != toAndMaster {
			l.args.mutable.gitCommitFrom = toAndMaster
		}
	}

	logArgs = []string{l.args.mutable.gitCommitFrom + ".." + l.args.mutable.gitCommitTo}
	diffArgs = []string{l.args.mutable.gitCommitFrom + ".." + l.args.mutable.gitCommitTo}

	return logArgs, diffArgs, nil
}
