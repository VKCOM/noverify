package linter

import (
	"sort"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/git"
)

const (
	// IgnoreLinterMessage is a commit message that you specify if you want to cancel linter checks for this changeset
	IgnoreLinterMessage = "@linter disable"
)

// DiffReports returns only reports that are new.
// Pass diffArgs=nil if we are called from diff in working copy.
func DiffReports(gitRepo string, diffArgs []string, changesList []git.Change, changeLog []git.Commit, oldList, newList []*Report, maxConcurrency int) (res []*Report, err error) {
	ignoreCommits := make(map[string]struct{})
	for _, c := range changeLog {
		if strings.Contains(c.Message, IgnoreLinterMessage) {
			ignoreCommits[c.Hash] = struct{}{}
		}
	}

	old := reportListToMap(oldList)
	new := reportListToMap(newList)
	changes := gitChangesToMap(changesList)

	var mu sync.Mutex
	var wg sync.WaitGroup

	var resErr error

	limitCh := make(chan struct{}, maxConcurrency)

	for filename, list := range new {
		wg.Add(1)
		go func(filename string, list []*Report) {
			limitCh <- struct{}{}
			defer func() { <-limitCh }()
			defer wg.Done()

			var oldName string

			c, ok := changes[filename]
			if ok {
				oldName = c.OldName
			} else {
				oldName = filename // full diff mode
			}

			reports, err := diffReportsList(gitRepo, ignoreCommits, diffArgs, filename, c, old[oldName], list)
			if err != nil {
				mu.Lock()
				resErr = err
				mu.Unlock()
				return
			}

			mu.Lock()
			res = append(res, reports...)
			mu.Unlock()
		}(filename, list)
	}

	wg.Wait()

	if resErr != nil {
		return nil, err
	}

	return res, nil
}

type lineRangeChange struct {
	old, new git.LineRange
}

// compute blame only if refspec is not nil
func blameIfNeeded(gitDir string, refspec []string, filename string) (git.BlameResult, error) {
	if refspec == nil {
		return git.BlameResult{}, nil
	}

	return git.Blame(gitDir, refspec, filename)
}

func fmtReports(list []*Report) []byte {
	var reports []string

	for _, r := range list {
		reports = append(reports, r.String())
	}

	return []byte(strings.Join(reports, "\n") + "\n")
}

func diffReportsList(gitRepo string, ignoreCommits map[string]struct{}, diffArgs []string, filename string, c git.Change, oldList, newList []*Report) (res []*Report, err error) {
	var blame git.BlameResult

	if c.Valid {
		blame, err = blameIfNeeded(gitRepo, diffArgs, filename)
		if err != nil {
			return nil, err
		}
	}

	changesMap := make(map[int]lineRangeChange, len(c.OldLineRanges))

	for idx, r := range c.OldLineRanges {
		for i := r.From; i <= r.To; i++ {
			changesMap[i] = lineRangeChange{old: r, new: c.LineRanges[idx]}
		}
	}

	old, oldMaxLine := reportListToPerLineMap(oldList)
	new, newMaxLine := reportListToPerLineMap(newList)

	var maxLine = oldMaxLine
	if newMaxLine > maxLine {
		maxLine = newMaxLine
	}

	var oldLine, newLine int

	for i := 0; i < maxLine; i++ {
		oldLine++
		newLine++

		ch, ok := changesMap[oldLine]
		// just deletion
		if ok && ch.new.HaveRange && ch.new.Range == 0 {
			oldLine = ch.old.To
			newLine-- // cancel the increment of newLine, because code was deleted, no new lines added
			continue
		}

		res = maybeAppendReports(res, new, old, newLine, oldLine, blame, ignoreCommits)

		if ok {
			oldLine = 0 // all changes and additions must be checked
			for j := newLine + 1; j <= ch.new.To; j++ {
				newLine = j
				res = maybeAppendReports(res, new, old, newLine, oldLine, blame, ignoreCommits)
			}
			oldLine = ch.old.To
		}
	}

	return res, nil
}

func maybeAppendReports(res []*Report, new, old map[int][]*Report, newLine, oldLine int, blame git.BlameResult, ignoreCommits map[string]struct{}) []*Report {
	newReports, ok := new[newLine]

	if !ok {
		return res
	}

	if _, ok := old[oldLine]; ok {
		return res
	}

	changedCommit := blame.Lines[newLine]

	if _, ok := ignoreCommits[changedCommit]; ok {
		return res
	}

	return append(res, newReports...)
}

func reportListToPerLineMap(list []*Report) (res map[int][]*Report, maxLine int) {
	res = make(map[int][]*Report)

	for _, l := range list {
		res[l.startLine] = append(res[l.startLine], l)
		if l.startLine > maxLine {
			maxLine = l.startLine
		}
	}

	return res, maxLine
}

func gitChangesToMap(changes []git.Change) map[string]git.Change {
	res := make(map[string]git.Change)
	for _, c := range changes {
		res[c.NewName] = c
	}
	return res
}

func reportListToMap(list []*Report) map[string][]*Report {
	res := make(map[string][]*Report)

	for _, r := range list {
		res[r.filename] = append(res[r.filename], r)
	}

	for _, l := range res {
		sort.Slice(l, func(i, j int) bool {
			return l[i].startLine < l[j].startLine
		})
	}

	return res
}
