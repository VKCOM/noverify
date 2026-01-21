package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/workspace"
)

func gitParseUntracked(l *LinterRunner) []*linter.Report {
	if !l.flags.GitIncludeUntracked {
		return nil
	}

	filenames, err := git.UntrackedFiles(l.flags.GitRepo)
	if err != nil {
		log.Fatalf("get untracked files: %v", err)
	}

	return l.linter.AnalyzeFiles(workspace.ReadFilenames(filenames, nil, l.config.PhpExtensions))
}

func parseIndexOnlyFiles(l *LinterRunner) {
	if l.flags.IndexOnlyFiles == "" {
		return
	}
	filenames := strings.Split(l.flags.IndexOnlyFiles, ",")
	l.linter.AnalyzeFiles(workspace.ReadFilenames(filenames, nil, l.config.PhpExtensions))
}

// Not the best name, and not the best function signature.
// Refactor this function whenever you get the idea how to separate logic better.
func gitRepoComputeReportsFromCommits(l *LinterRunner, logArgs, diffArgs []string) (oldReports, reports []*linter.Report, changes []git.Change, changeLog []git.Commit, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	changeLog, err := git.Log(l.flags.GitRepo, logArgs)
	if err != nil {
		log.Panicf("Could not get commits in range %+v: %s", logArgs, err.Error())
	}

	if shouldRun := analyzeGitAuthorsWhiteList(l, changeLog); !shouldRun {
		return nil, nil, nil, nil, false
	}

	changes, err = git.Diff(l.flags.GitRepo, "", diffArgs)
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if l.flags.GitFullDiff {
		resetMetaInfo(l)
		if err := loadEmbeddedStubs(l.linter); err != nil {
			log.Panicf("Load embedded stubs: %v", err)
		}

		start := time.Now()
		l.linter.AnalyzeFiles(workspace.ReadFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitFrom, nil, l.config.PhpExtensions))
		parseIndexOnlyFiles(l)
		log.Printf("Indexed old commit in %s", time.Since(start))

		l.linter.MetaInfo().SetIndexingComplete(true)

		start = time.Now()
		oldReports = l.linter.AnalyzeFiles(workspace.ReadFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitFrom, l.config.ExcludeRegex, l.config.PhpExtensions))
		log.Printf("Parsed old commit for %s (%d reports)", time.Since(start), len(oldReports))

		resetMetaInfo(l)
		if err := loadEmbeddedStubs(l.linter); err != nil {
			log.Panicf("Load embedded stubs: %v", err)
		}

		start = time.Now()
		l.linter.AnalyzeFiles(workspace.ReadFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitTo, nil, l.config.PhpExtensions))
		parseIndexOnlyFiles(l)
		log.Printf("Indexed new commit in %s", time.Since(start))

		l.linter.MetaInfo().SetIndexingComplete(true)

		start = time.Now()
		reports = l.linter.AnalyzeFiles(workspace.ReadFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitTo, l.config.ExcludeRegex, l.config.PhpExtensions))
		log.Printf("Parsed new commit in %s (%d reports)", time.Since(start), len(reports))
	} else {
		start := time.Now()
		l.linter.AnalyzeFiles(workspace.ReadFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitTo, nil, l.config.PhpExtensions))
		parseIndexOnlyFiles(l)
		log.Printf("Indexing complete in %s", time.Since(start))

		l.linter.MetaInfo().SetIndexingComplete(true)

		start = time.Now()
		oldReports = l.linter.AnalyzeFiles(workspace.ReadOldFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitFrom, changes, l.config.PhpExtensions))
		log.Printf("Parsed old files versions for %s", time.Since(start))

		start = time.Now()
		l.linter.MetaInfo().SetIndexingComplete(false)
		parseIndexOnlyFiles(l)
		l.linter.AnalyzeFiles(workspace.ReadFilesFromGitWithChanges(l.flags.GitRepo, l.flags.Mutable.GitCommitTo, changes, l.config.PhpExtensions))
		l.linter.MetaInfo().SetIndexingComplete(true)
		log.Printf("Indexed files versions for %s", time.Since(start))

		start = time.Now()
		reports = l.linter.AnalyzeFiles(workspace.ReadFilesFromGitWithChanges(l.flags.GitRepo, l.flags.Mutable.GitCommitTo, changes, l.config.PhpExtensions))
		log.Printf("Parsed new file versions in %s", time.Since(start))
	}

	return oldReports, reports, changes, changeLog, true
}

func gitRepoComputeReportsFromLocalChanges(l *LinterRunner) (oldReports, reports []*linter.Report, changes []git.Change, ok bool) {
	// TODO(quasilyte): hard to replace fatalf with error return here. Use panicf for now.

	if l.flags.GitWorkTree == "" {
		return nil, nil, nil, false
	}

	// compute changes for working copy (staged + unstaged changes combined starting with the commit being pushed)
	changes, err := git.Diff(l.flags.GitRepo, l.flags.GitWorkTree, []string{l.flags.Mutable.GitCommitFrom})
	if err != nil {
		log.Panicf("Could not compute git diff: %s", err.Error())
	}

	if len(changes) == 0 {
		return nil, nil, nil, false
	}

	log.Printf("You have changes in your work tree, showing diff between %s and work tree", l.flags.Mutable.GitCommitFrom)

	start := time.Now()
	l.linter.AnalyzeFiles(workspace.ReadFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitFrom, nil, l.config.PhpExtensions))
	parseIndexOnlyFiles(l)
	log.Printf("Indexing complete in %s", time.Since(start))

	l.linter.MetaInfo().SetIndexingComplete(true)

	start = time.Now()
	oldReports = l.linter.AnalyzeFiles(workspace.ReadOldFilesFromGit(l.flags.GitRepo, l.flags.Mutable.GitCommitFrom, changes, l.config.PhpExtensions))
	log.Printf("Parsed old files versions for %s", time.Since(start))

	start = time.Now()
	l.linter.MetaInfo().SetIndexingComplete(false)
	l.linter.AnalyzeFiles(workspace.ReadChangesFromWorkTree(l.flags.GitWorkTree, changes, l.config.PhpExtensions))
	parseIndexOnlyFiles(l)
	gitParseUntracked(l)
	l.linter.MetaInfo().SetIndexingComplete(true)
	log.Printf("Indexed new files versions for %s", time.Since(start))

	start = time.Now()
	reports = l.linter.AnalyzeFiles(workspace.ReadChangesFromWorkTree(l.flags.GitWorkTree, changes, l.config.PhpExtensions))
	reports = append(reports, gitParseUntracked(l)...)
	log.Printf("Parsed new file versions in %s", time.Since(start))

	return oldReports, reports, changes, true
}

func gitMain(runner *LinterRunner, ctx *AppContext) (status int, err error) {
	var (
		oldReports, reports []*linter.Report
		diffArgs            []string
		changes             []git.Change
		changeLog           []git.Commit
		ok                  bool
	)

	logArgs, diffArgs, err := prepareGitArgs(runner)
	if err != nil {
		return 0, err
	}

	oldReports, reports, changes, ok = gitRepoComputeReportsFromLocalChanges(runner)
	if !ok {
		oldReports, reports, changes, changeLog, ok = gitRepoComputeReportsFromCommits(runner, logArgs, diffArgs)
		if !ok {
			return 0, nil
		}
	}

	start := time.Now()
	diff, err := linter.DiffReports(runner.flags.GitRepo, diffArgs, changes, changeLog, oldReports, reports, 8)
	if err != nil {
		return 0, fmt.Errorf("Could not compute reports diff: %v", err)
	}
	log.Printf("Computed reports diff for %s", time.Since(start))

	stat := processReports(runner, ctx.MainConfig, diff)
	status = processReportsStat(ctx, stat)

	return status, nil
}

func analyzeGitAuthorsWhiteList(l *LinterRunner, changeLog []git.Commit) (shouldRun bool) {
	if l.flags.GitAuthorsWhitelist != "" {
		whiteList := make(map[string]bool)
		for _, name := range strings.Split(l.flags.GitAuthorsWhitelist, ",") {
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

func prepareGitArgs(l *LinterRunner) (logArgs, diffArgs []string, err error) {
	if l.flags.GitPushArg != "" {
		args := strings.Fields(l.flags.GitPushArg)
		if len(args) != 3 {
			return nil, nil, fmt.Errorf("Unexpected format of push arguments, expected only 3 columns: %s", l.flags.GitPushArg)
		}
		// args[2] is a git ref (branch name), but its unused.
		l.flags.Mutable.GitCommitFrom, l.flags.Mutable.GitCommitTo = args[0], args[1]
	}
	if l.flags.Mutable.GitCommitFrom == git.Zero {
		l.flags.Mutable.GitCommitFrom = "master"
	}

	if !l.flags.GitSkipFetch {
		start := time.Now()
		log.Printf("Fetching origin master to ORIGIN_MASTER")
		if err := git.Fetch(l.flags.GitRepo, "master", "ORIGIN_MASTER"); err != nil {
			return nil, nil, fmt.Errorf("Could not fetch ORIGIN_MASTER: %v", err.Error())
		}
		log.Printf("Fetched for %s", time.Since(start))
	}

	if !l.flags.GitDisableCompensateMaster {
		fromAndMaster, err := git.MergeBase(l.flags.GitRepo, "ORIGIN_MASTER", l.flags.Mutable.GitCommitFrom)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", l.flags.Mutable.GitCommitFrom)
		}

		toAndMaster, err := git.MergeBase(l.flags.GitRepo, "ORIGIN_MASTER", l.flags.Mutable.GitCommitTo)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not compute merge base between ORIGIN_MASTER and %s", l.flags.Mutable.GitCommitTo)
		}

		// check if master was merged in between the commits
		if fromAndMaster != toAndMaster {
			l.flags.Mutable.GitCommitFrom = toAndMaster
		}
	}

	logArgs = []string{l.flags.Mutable.GitCommitFrom + ".." + l.flags.Mutable.GitCommitTo}
	diffArgs = []string{l.flags.Mutable.GitCommitFrom + ".." + l.flags.Mutable.GitCommitTo}

	return logArgs, diffArgs, nil
}

// This function is a kludge to make old git-related code work without many modifications.
func resetMetaInfo(l *LinterRunner) {
	l.linter = linter.NewLinter(l.config)
}
