package linter

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/git"
)

const (
	// IgnoreLinterMessage is a commit message that you specify if you want to cancel linter checks for this changeset
	IgnoreLinterMessage = "@linter disable"
)

func init() {
	allChecks := []CheckInfo{
		{
			Name:    "discardExpr",
			Default: true,
			Comment: `Report expressions that are evaluated but not used.`,
		},

		{
			Name:    "voidResultUsed",
			Default: true,
			Comment: `Report usages of the void-type expressions`,
		},

		{
			Name:    "keywordCase",
			Default: true,
			Comment: `Report keywords that are not in the lower case.`,
		},

		{
			Name:    "accessLevel",
			Default: true,
			Comment: `Report erroneous member access.`,
		},

		{
			Name:    "argCount",
			Default: true,
			Comment: `Report mismatching args count inside call expressions.`,
		},

		{
			Name:    "arrayAccess",
			Default: true,
			Comment: `Report array access to non-array objects.`,
		},

		{
			Name:    "bitwiseOps",
			Default: true,
			Comment: `Report suspicious usage of bitwise operations.`,
		},

		{
			Name:    "mixedArrayKeys",
			Default: true,
			Comment: `Report array literals that have both implicit and explicit keys.`,
		},

		{
			Name:    "dupArrayKeys",
			Default: true,
			Comment: `Report duplicated keys in array literals.`,
		},

		{
			Name:    "arraySyntax",
			Default: true,
			Comment: `Report usages of old array() syntax.`,
		},

		{
			Name:    "bareTry",
			Default: true,
			Comment: `Report try blocks without catch/finally.`,
		},

		{
			Name:    "caseBreak",
			Default: true,
			Comment: `Report switch cases without break.`,
		},

		{
			Name:    "complexity",
			Default: true,
			Comment: `Report funcs/methods that are too complex.`,
		},

		{
			Name:    "deadCode",
			Default: true,
			Comment: `Report potentially unreachable code.`,
		},

		{
			Name:    "phpdocLint",
			Default: true,
			Comment: `Report malformed phpdoc comments.`,
		},

		{
			Name:    "phpdocType",
			Default: true,
			Comment: `Report potential issues in phpdoc types.`,
		},

		{
			Name:    "phpdoc",
			Default: true,
			Comment: `Report missing phpdoc on public methods.`,
		},

		{
			Name:    "stdInterface",
			Default: true,
			Comment: `Report issues related to std PHP interfaces.`,
		},

		{
			Name:    "syntax",
			Default: true,
			Comment: `Report syntax errors.`,
		},

		{
			Name:    "undefined",
			Default: true,
			Comment: `Report usages of potentially undefined symbols.`,
		},

		{
			Name:    "unused",
			Default: true,
			Comment: `Report potentially unused variables.`,
		},

		{
			Name:    "redundantCast",
			Default: false,
			Comment: `Report redundant type casts.`,
		},

		{
			Name:    "caseContinue",
			Default: true,
			Comment: `Report suspicious 'continue' usages inside switch cases.`,
		},

		{
			Name:    "deprecated",
			Default: false, // Experimental
			Comment: `Report usages of deprecated symbols.`,
		},

		{
			Name:    "callStatic",
			Default: true,
			Comment: `Report static calls of instance methods and vice versa.`,
		},

		{
			Name:    "oldStyleConstructor",
			Default: true,
			Comment: `Report old-style (PHP4) class constructors.`,
		},

		{
			Name:    "misspellName",
			Default: true,
			Comment: `Report commonly misspelled words in symbol names.`,
		},

		{
			Name:    "misspellComment",
			Default: true,
			Comment: `Report commonly misspelled words in comments.`,
		},
	}

	for _, info := range allChecks {
		DeclareCheck(info)
	}
}

// Report is a linter report message.
type Report struct {
	checkName  string
	startLn    string
	startChar  int
	startLine  int
	endChar    int
	level      int
	msg        string
	filename   string
	isDisabled bool // user-defined flag that file should not be linted
}

// CheckName returns report associated check name.
func (r *Report) CheckName() string {
	return r.checkName
}

// MarshalJSON is used to write report in its JSON representation.
//
// Used for -output-json option.
func (r *Report) MarshalJSON() ([]byte, error) {
	type jsonReport struct {
		CheckName string `json:"check_name"`
		Severity  string `json:"severity"`
		Context   string `json:"context"`
		Message   string `json:"message"`
		Filename  string `json:"filename"`
		Line      int    `json:"line"`
		StartChar int    `json:"start_char"`
		EndChar   int    `json:"end_char"`
	}

	b, err := json.Marshal(jsonReport{
		CheckName: r.checkName,
		Severity:  strings.TrimSpace(severityNames[r.level]),
		Context:   r.startLn,
		Message:   r.msg,
		Filename:  r.filename,
		Line:      r.startLine,
		StartChar: r.startChar,
		EndChar:   r.endChar,
	})
	return b, err
}

func (r *Report) String() string {
	contextLn := strings.Builder{}
	for i, ch := range r.startLn {
		if i == r.startChar {
			break
		}
		if ch == '\t' {
			contextLn.WriteRune(ch)
		} else {
			contextLn.WriteByte(' ')
		}
	}

	if r.endChar > r.startChar {
		contextLn.WriteString(strings.Repeat("^", r.endChar-r.startChar))
	}

	msg := r.msg
	if r.checkName != "" {
		msg = r.checkName + ": " + msg
	}
	return fmt.Sprintf("%s %s at %s:%d\n%s\n%s", severityNames[r.level], msg, r.filename, r.startLine, r.startLn, contextLn.String())
}

// IsCritical returns whether or not we need to reject whole commit when found this kind of report.
func (r *Report) IsCritical() bool {
	return r.level != LevelDoNotReject
}

// IsDisabledByUser returns whether or not user thinks that this file should not be checked
func (r *Report) IsDisabledByUser() bool {
	return r.isDisabled
}

// GetFilename returns report filename
func (r *Report) GetFilename() string {
	return r.filename
}

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

func isUnderscore(s string) bool {
	return s == "_"
}

// unquote returns unquoted version of s, if there are any quotes.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '\'' || s[0] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func linterError(filename, format string, args ...interface{}) {
	log.Printf("error: "+filename+": "+format, args...)
}
